package plugin

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/proto"

	"github.com/secmc/plugin/plugin/adapters/grpc"
	"github.com/secmc/plugin/plugin/config"
	pb "github.com/secmc/plugin/proto/generated/go"
)

const (
	apiVersion = "v1"

	sendChannelBuffer = 64
	shutdownTimeout   = 5 * time.Second
)

type pluginProcess struct {
	id      string
	cfg     config.PluginConfig
	manager *Manager
	log     *slog.Logger

	cmd      *exec.Cmd
	stream   *grpc.GrpcStream
	streamMu sync.RWMutex

	sendCh chan *pb.HostToPlugin
	done   chan struct{}
	wg     sync.WaitGroup

	subscriptions sync.Map
	connected     atomic.Bool
	ready         atomic.Bool

	helloMu sync.RWMutex
	hello   *pb.PluginHello

	closed atomic.Bool

	pendingMu sync.Mutex
	pending   map[string]chan *pb.EventResult
}

func newPluginProcess(m *Manager, cfg config.PluginConfig) *pluginProcess {
	logger := m.log.With("plugin", cfg.ID)
	if cfg.Name != "" {
		logger = logger.With("name", cfg.Name)
	}
	return &pluginProcess{
		id:      cfg.ID,
		cfg:     cfg,
		manager: m,
		log:     logger,
		sendCh:  make(chan *pb.HostToPlugin, sendChannelBuffer),
		done:    make(chan struct{}),
		pending: make(map[string]chan *pb.EventResult),
	}
}

func (p *pluginProcess) start(ctx context.Context, serverAddress string) {
	if p.cfg.Command != "" {
		if err := p.launchProcess(ctx, serverAddress); err != nil {
			p.log.Error("launch plugin", "error", err)
			return
		}
	}
}

// attachStream attaches an incoming stream to this plugin process
func (p *pluginProcess) attachStream(stream *grpc.GrpcStream) error {
	p.streamMu.Lock()
	// Allow replacing a stale/closed stream to support plugin hot-reload reconnections.
	if p.stream != nil {
		_ = p.stream.Close()
		p.stream = nil
	}
	p.stream = stream
	p.streamMu.Unlock()

	p.connected.Store(true)

	if err := p.sendHello(); err != nil {
		p.log.Error("send hello", "error", err)
		p.Stop()
		return err
	}

	p.wg.Add(2)
	go p.sendLoop()
	go p.recvLoop()
	return nil
}

// clearStream drops the current gRPC stream and marks the plugin as disconnected,
// without killing the underlying process. This enables external hot-reload
// wrappers to restart the plugin and reconnect cleanly.
func (p *pluginProcess) clearStream() {
	p.streamMu.Lock()
	if p.stream != nil {
		_ = p.stream.Close()
		p.stream = nil
	}
	p.streamMu.Unlock()
	p.connected.Store(false)
}

func (p *pluginProcess) launchProcess(ctx context.Context, serverAddress string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	cmd := exec.CommandContext(ctx, p.cfg.Command, p.cfg.Args...)
	if p.cfg.WorkDir.Path != "" {
		cmd.Dir = p.cfg.WorkDir.Path
	}
	env := os.Environ()
	env = append(env, fmt.Sprintf("DF_PLUGIN_ID=%s", p.id))
	// Normalize Unix socket address for plugin clients: bare paths -> unix:/path
	passAddress := serverAddress
	if strings.HasPrefix(serverAddress, "/") {
		passAddress = "unix:" + serverAddress
	}
	env = append(env, fmt.Sprintf("DF_PLUGIN_SERVER_ADDRESS=%s", passAddress))
	env = append(env, fmt.Sprintf("DF_HOST_BOOT_ID=%s", p.manager.bootID))
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

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
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
			p.manager.log.Info(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil && !p.closed.Load() {
		p.log.Error("output scanner error", "error", err)
	}
}

func (p *pluginProcess) sendServerInfo(plugins []string) error {
	msg := &pb.HostToPlugin{
		PluginId: p.id,
		Payload: &pb.HostToPlugin_ServerInfo{
			ServerInfo: &pb.ServerInformationResponse{
				Plugins: plugins,
			},
		},
	}
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.stream.Send(payload)
}

func (p *pluginProcess) sendHello() error {
	msg := &pb.HostToPlugin{
		PluginId: p.id,
		Payload: &pb.HostToPlugin_Hello{
			Hello: &pb.HostHello{
				ApiVersion: apiVersion,
				BootId:     p.manager.bootID,
			},
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
			data, err := proto.Marshal(msg)
			if err != nil {
				p.log.Error("marshal message", "error", err)
				continue
			}
			if err := p.stream.Send(data); err != nil {
				// Treat expected shutdown conditions as non-errors.
				if st, ok := status.FromError(err); ok && (st.Code() == codes.Canceled || st.Code() == codes.Unavailable) {
					p.log.Info("connection closed", "reason", st.Code().String())
				} else if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
					p.log.Info("connection closed", "reason", "canceled")
				} else {
					p.log.Error("send message", "error", err)
				}
				// Do not kill the process on transient stream errors; allow reconnection.
				p.clearStream()
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
			if st, ok := status.FromError(err); ok {
				switch st.Code() {
				case codes.Canceled, codes.Unavailable:
					p.log.Info("connection closed", "reason", st.Code().String())
				default:
					p.log.Error("receive message", "error", err)
				}
			} else if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				p.log.Info("connection closed", "reason", "canceled")
			} else {
				p.log.Error("receive message", "error", err)
			}
			// Do not kill the process on transient stream errors; allow reconnection.
			p.clearStream()
			return
		}
		msg := &pb.PluginToHost{}
		if err := proto.Unmarshal(data, msg); err != nil {
			p.log.Error("decode message", "error", err)
			continue
		}
		p.manager.handlePluginMessage(p, msg)
	}
}

func (p *pluginProcess) HasSubscription(event pb.EventType) bool {
	if !p.ready.Load() {
		return false
	}
	if _, ok := p.subscriptions.Load(pb.EventType_EVENT_TYPE_ALL); ok {
		return true
	}
	if event == pb.EventType_EVENT_TYPE_UNSPECIFIED {
		return false
	}
	_, ok := p.subscriptions.Load(event)
	return ok
}

func (p *pluginProcess) updateSubscriptions(events []pb.EventType) {
	p.subscriptions.Range(func(key, value any) bool {
		p.subscriptions.Delete(key)
		return true
	})
	for _, evt := range events {
		if evt == pb.EventType_EVENT_TYPE_UNSPECIFIED {
			continue
		}
		p.subscriptions.Store(evt, struct{}{})
	}
	p.ready.Store(true)
}

func (p *pluginProcess) queue(msg *pb.HostToPlugin) {
	if p.closed.Load() || !p.connected.Load() {
		return
	}
	select {
	case p.sendCh <- msg:
	default:
		p.log.Warn("dropping message", "reason", "queue full")
	}
}

func (p *pluginProcess) isConnected() bool {
	return p.connected.Load()
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
	p.log.Debug("waiting for event result", "event_id", eventID)
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
	p.log.Debug("discarded event result waiter", "event_id", eventID)
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
	p.log.Debug("delivered event result", "event_id", res.EventId)
}
