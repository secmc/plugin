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
	"google.golang.org/protobuf/proto"

	pb "github.com/df-mc/dragonfly/plugin/proto/generated"
)

const apiVersion = "v1"

type pluginProcess struct {
	id  string
	cfg PluginConfig
	mgr *Manager
	log *slog.Logger

	cmd    *exec.Cmd
	stream *grpcStream

	sendCh chan *pb.HostToPlugin
	done   chan struct{}

	subscriptions sync.Map
	ready         atomic.Bool

	helloMu sync.RWMutex
	hello   *pb.PluginHello

	closed atomic.Bool

	pendingMu sync.Mutex
	pending   map[string]chan *pb.EventResult
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
		sendCh: make(chan *pb.HostToPlugin, 64),
		done:   make(chan struct{}),
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
	msg := &pb.HostToPlugin{
		PluginId: p.id,
		Payload: &pb.HostToPlugin_Hello{
			Hello: &pb.HostHello{ApiVersion: apiVersion},
		},
	}
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.stream.Send(payload)
}

func (p *pluginProcess) sendLoop() {
	for {
		select {
		case <-p.done:
			return
		case msg := <-p.sendCh:
			if msg == nil {
				continue
			}
			if p.stream == nil {
				return
			}
			data, err := proto.Marshal(msg)
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
		msg := &pb.PluginToHost{}
		if err := proto.Unmarshal(data, msg); err != nil {
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

func (p *pluginProcess) queue(msg *pb.HostToPlugin) {
	if p.closed.Load() {
		return
	}
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
		close(p.done)
		p.pendingMu.Lock()
		for id, ch := range p.pending {
			delete(p.pending, id)
			close(ch)
		}
		p.pendingMu.Unlock()
		p.stopProcess()
	}
}

func (p *pluginProcess) stopProcess() {
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
}

func (p *pluginProcess) setHello(h *pb.PluginHello) {
	p.helloMu.Lock()
	defer p.helloMu.Unlock()
	p.hello = h
}

func (p *pluginProcess) helloInfo() *pb.PluginHello {
	p.helloMu.RLock()
	defer p.helloMu.RUnlock()
	return p.hello
}

func (p *pluginProcess) expectEventResult(eventID string) chan *pb.EventResult {
	ch := make(chan *pb.EventResult, 1)
	p.pendingMu.Lock()
	if p.pending == nil {
		p.pending = make(map[string]chan *pb.EventResult)
	}
	p.pending[eventID] = ch
	p.pendingMu.Unlock()
	return ch
}

func (p *pluginProcess) waitEventResult(ch chan *pb.EventResult, timeout time.Duration) (*pb.EventResult, error) {
	select {
	case res, ok := <-ch:
		if !ok {
			return nil, context.Canceled
		}
		return res, nil
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

func (p *pluginProcess) discardEventResult(eventID string) {
	p.pendingMu.Lock()
	if ch, ok := p.pending[eventID]; ok {
		delete(p.pending, eventID)
		close(ch)
	}
	p.pendingMu.Unlock()
}

func (p *pluginProcess) deliverEventResult(res *pb.EventResult) {
	if res == nil {
		return
	}
	p.pendingMu.Lock()
	ch, ok := p.pending[res.EventId]
	if ok {
		delete(p.pending, res.EventId)
	}
	p.pendingMu.Unlock()
	if !ok {
		p.log.Warn("unexpected event result", "event_id", res.EventId)
		return
	}
	select {
	case ch <- res:
	default:
	}
	close(ch)
}

func generateEventID() string {
	return uuid.New().String()
}
