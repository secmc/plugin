package plugin

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/df-mc/dragonfly/plugin/proto"
)

const apiVersion = "v1"

type pluginProcess struct {
	id  string
	cfg PluginConfig
	mgr *Manager
	log *slog.Logger

	cmd    *exec.Cmd
	stream *grpcStream

	sendCh chan *proto.HostToPlugin

	subscriptions sync.Map
	ready         atomic.Bool

	helloMu sync.RWMutex
	hello   *proto.PluginHello

	closed atomic.Bool
}

func newPluginProcess(m *Manager, cfg PluginConfig) *pluginProcess {
	logger := m.log.With("plugin", cfg.ID)
	if cfg.Name != "" {
		logger = logger.With("name", cfg.Name)
	}
	return &pluginProcess{
		id:     cfg.ID,
		cfg:    cfg,
		mgr:    m,
		log:    logger,
		sendCh: make(chan *proto.HostToPlugin, 64),
	}
}

func (p *pluginProcess) start(ctx context.Context) {
	address, err := p.prepareAddress()
	if err != nil {
		p.log.Error("prepare address", "error", err)
		return
	}

	if p.cfg.Command != "" {
		if err := p.launchProcess(ctx, address); err != nil {
			p.log.Error("launch plugin", "error", err)
			return
		}
	}

	stream, err := p.connectLoop(ctx, address)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			p.log.Error("connect", "error", err)
		}
		p.stopProcess()
		return
	}
	p.stream = stream

	if err := p.sendHello(); err != nil {
		p.log.Error("send hello", "error", err)
		p.Stop()
		return
	}

	go p.sendLoop()
	go p.recvLoop()
}

func (p *pluginProcess) prepareAddress() (string, error) {
	addr := p.cfg.Address
	if addr == "" {
		addr = "127.0.0.1:50051"
	}
	if strings.HasSuffix(addr, ":0") {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return "", err
		}
		actual := l.Addr().String()
		_ = l.Close()
		return actual, nil
	}
	return addr, nil
}

func (p *pluginProcess) launchProcess(ctx context.Context, address string) error {
	cmd := exec.CommandContext(ctx, p.cfg.Command, p.cfg.Args...)
	if p.cfg.WorkDir != "" {
		cmd.Dir = p.cfg.WorkDir
	}
	env := os.Environ()
	env = append(env, fmt.Sprintf("DF_PLUGIN_ID=%s", p.id))
	env = append(env, fmt.Sprintf("DF_PLUGIN_GRPC_ADDRESS=%s", address))
	for k, v := range p.cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go p.consumeOutput(stdout)
	go p.consumeOutput(stderr)
	if err := cmd.Start(); err != nil {
		return err
	}
	p.cmd = cmd
	go func() {
		if err := cmd.Wait(); err != nil && !p.closed.Load() {
			p.log.Warn("process exited", "error", err)
		}
	}()
	return nil
}

func (p *pluginProcess) consumeOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		p.log.Info(scanner.Text())
	}
}

func (p *pluginProcess) connectLoop(ctx context.Context, address string) (*grpcStream, error) {
	for {
		stream, err := dialEventStream(ctx, address, 5*time.Second)
		if err == nil {
			return stream, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		select {
		case <-time.After(300 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (p *pluginProcess) sendHello() error {
	msg := &proto.HostToPlugin{
		PluginID: p.id,
		Hello:    &proto.HostHello{APIVersion: apiVersion},
	}
	payload, err := msg.Marshal()
	if err != nil {
		return err
	}
	return p.stream.Send(payload)
}

func (p *pluginProcess) sendLoop() {
	for msg := range p.sendCh {
		if p.stream == nil {
			return
		}
		data, err := msg.Marshal()
		if err != nil {
			p.log.Error("marshal message", "error", err)
			continue
		}
		if err := p.stream.Send(data); err != nil {
			p.log.Error("send message", "error", err)
			p.Stop()
			return
		}
	}
}

func (p *pluginProcess) recvLoop() {
	for {
		data, err := p.stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				p.log.Error("receive message", "error", err)
			}
			p.Stop()
			return
		}
		msg, err := proto.UnmarshalPluginToHost(data)
		if err != nil {
			p.log.Error("decode message", "error", err)
			continue
		}
		p.mgr.handlePluginMessage(p, msg)
	}
}

func (p *pluginProcess) HasSubscription(event string) bool {
	if !p.ready.Load() {
		return false
	}
	if _, ok := p.subscriptions.Load("*"); ok {
		return true
	}
	_, ok := p.subscriptions.Load(strings.ToUpper(event))
	return ok
}

func (p *pluginProcess) updateSubscriptions(events []string) {
	p.subscriptions.Range(func(key, value any) bool {
		p.subscriptions.Delete(key)
		return true
	})
	for _, evt := range events {
		evt = strings.ToUpper(strings.TrimSpace(evt))
		if evt == "" {
			continue
		}
		p.subscriptions.Store(evt, struct{}{})
	}
	p.ready.Store(true)
}

func (p *pluginProcess) queue(msg *proto.HostToPlugin) {
	select {
	case p.sendCh <- msg:
	default:
		p.log.Warn("dropping message", "reason", "queue full")
	}
}

func (p *pluginProcess) Stop() {
	if p.closed.CompareAndSwap(false, true) {
		if p.stream != nil {
			_ = p.stream.Close()
		}
		close(p.sendCh)
		p.stopProcess()
	}
}

func (p *pluginProcess) stopProcess() {
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
}

func (p *pluginProcess) setHello(h *proto.PluginHello) {
	p.helloMu.Lock()
	defer p.helloMu.Unlock()
	p.hello = h
}

func (p *pluginProcess) helloInfo() *proto.PluginHello {
	p.helloMu.RLock()
	defer p.helloMu.RUnlock()
	return p.hello
}

func generateEventID() string {
	return uuid.New().String()
}
