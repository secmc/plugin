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

	"google.golang.org/protobuf/proto"

	"github.com/secmc/plugin/plugin/adapters/grpc"
	"github.com/secmc/plugin/plugin/config"
	pb "github.com/secmc/plugin/proto/generated"
)

const (
	apiVersion = "v1"

	sendChannelBuffer    = 64
	connectTimeout       = 5 * time.Second
	connectRetryInterval = 300 * time.Millisecond
	shutdownTimeout      = 5 * time.Second
	defaultPluginAddress = "127.0.0.1:50051"
)

type pluginProcess struct {
	id      string
	cfg     config.PluginConfig
	emitter *Emitter
	log     *slog.Logger

	cmd    *exec.Cmd
	stream *grpc.GrpcStream

	sendCh chan *pb.HostToPlugin
	done   chan struct{}
	wg     sync.WaitGroup

	subscriptions sync.Map
	ready         atomic.Bool

	helloMu sync.RWMutex
	hello   *pb.PluginHello

	closed atomic.Bool

	pendingMu sync.Mutex
	pending   map[string]chan *pb.EventResult
}

func newPluginProcess(e *Emitter, cfg config.PluginConfig) *pluginProcess {
	logger := e.log.With("plugin", cfg.ID)
	if cfg.Name != "" {
		logger = logger.With("name", cfg.Name)
	}
	return &pluginProcess{
		id:      cfg.ID,
		cfg:     cfg,
		emitter: e,
		log:     logger,
		sendCh:  make(chan *pb.HostToPlugin, sendChannelBuffer),
		done:    make(chan struct{}),
		pending: make(map[string]chan *pb.EventResult),
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

	p.wg.Add(2)
	go p.sendLoop()
	go p.recvLoop()
}

func (p *pluginProcess) prepareAddress() (string, error) {
	addr := p.cfg.Address
	if addr == "" {
		addr = defaultPluginAddress
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
	if ctx.Err() != nil {
		return ctx.Err()
	}

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
		_ = stdout.Close()
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	p.cmd = cmd

	p.wg.Add(2)
	go p.consumeOutput(stdout)
	go p.consumeOutput(stderr)

	go func() {
		if err := cmd.Wait(); err != nil && !p.closed.Load() {
			p.log.Warn("process exited", "error", err)
		}
	}()
	return nil
}

func (p *pluginProcess) consumeOutput(r io.Reader) {
	defer p.wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-p.done:
			return
		default:
			p.log.Info(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil && !p.closed.Load() {
		p.log.Error("output scanner error", "error", err)
	}
}

func (p *pluginProcess) connectLoop(ctx context.Context, address string) (*grpc.GrpcStream, error) {
	for {
		stream, err := grpc.DialEventStream(ctx, address, connectTimeout)
		if err == nil {
			return stream, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		select {
		case <-time.After(connectRetryInterval):
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
	defer p.wg.Done()
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
	defer p.wg.Done()
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
		p.emitter.handlePluginMessage(p, msg)
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

		// Wait for goroutines to finish with timeout
		done := make(chan struct{})
		go func() {
			p.wg.Wait()
			close(done)
		}()
		select {
		case <-done:
			// all go routines finished cleanly
		case <-time.After(shutdownTimeout):
			p.log.Warn("timeout waiting for goroutines to finish")
		}
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
