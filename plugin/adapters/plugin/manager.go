package plugin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/secmc/plugin/plugin/adapters/grpc"
	"github.com/secmc/plugin/plugin/config"
	"github.com/secmc/plugin/plugin/ports"
	pb "github.com/secmc/plugin/proto/generated"
)

type Manager struct {
	srv *server.Server
	log *slog.Logger

	ctx    context.Context
	cancel context.CancelFunc

	grpcServer *grpc.GrpcServer

	mu       sync.RWMutex
	plugins  map[string]*pluginProcess
	players  map[uuid.UUID]*player.Player
	commands map[string]commandBinding

	worldMu sync.RWMutex
	worlds  map[string]*world.World

	eventCounter atomic.Uint64

	playerHandlerFactory ports.PlayerHandlerFactory
	worldHandlerFactory  ports.WorldHandlerFactory
}

type commandBinding struct {
	pluginID   string
	command    string
	descriptor *pb.CommandSpec
}

const eventResponseTimeout = 250 * time.Millisecond

func NewManager(srv *server.Server, log *slog.Logger, playerHandlerFactory ports.PlayerHandlerFactory, worldHandlerFactory ports.WorldHandlerFactory) *Manager {
	if log == nil {
		log = slog.Default()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		srv:                  srv,
		log:                  log.With("component", "plugin-manager"),
		ctx:                  ctx,
		cancel:               cancel,
		plugins:              make(map[string]*pluginProcess),
		players:              make(map[uuid.UUID]*player.Player),
		commands:             make(map[string]commandBinding),
		worlds:               make(map[string]*world.World),
		playerHandlerFactory: playerHandlerFactory,
		worldHandlerFactory:  worldHandlerFactory,
	}
}

func (m *Manager) Start(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			m.log.Info("no plugin configuration found", "path", configPath)
			return nil
		}
		return err
	}

	// Start gRPC server to accept plugin connections
	address := fmt.Sprintf("127.0.0.1:%d", cfg.ServerPort)
	grpcServer, err := grpc.NewServer(address, m.handlePluginConnection)
	if err != nil {
		return fmt.Errorf("start plugin server: %w", err)
	}
	m.grpcServer = grpcServer
	m.log.Info("plugin server listening", "address", grpcServer.Address())

	// Start accepting connections in background
	go func() {
		if err := grpcServer.Serve(); err != nil {
			m.log.Error("plugin server error", "error", err)
		}
	}()

	// Launch plugin processes
	for _, pc := range cfg.Plugins {
		if pc.ID == "" {
			pc.ID = pc.Name
		}
		if pc.ID == "" {
			pc.ID = fmt.Sprintf("plugin-%s", strings.ToLower(uuid.NewString()[:8]))
		}
		proc := newPluginProcess(m, pc)
		m.mu.Lock()
		m.plugins[pc.ID] = proc
		m.mu.Unlock()
		go proc.start(m.ctx, grpcServer.Address())
	}
	return nil
}

// handlePluginConnection is called when a plugin connects to the gRPC server
func (m *Manager) handlePluginConnection(stream *grpc.GrpcStream) error {
	m.log.Info("new plugin connection")

	// Read the first message to identify the plugin
	data, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receive first message: %w", err)
	}

	msg := &pb.PluginToHost{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("decode first message: %w", err)
	}

	pluginID := msg.PluginId
	if pluginID == "" {
		return errors.New("first message missing plugin_id")
	}

	// Find the plugin process
	m.mu.RLock()
	proc, ok := m.plugins[pluginID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown plugin ID: %s", pluginID)
	}

	// Handle the first message (likely PluginHello)
	m.handlePluginMessage(proc, msg)

	// Attach the stream to the process
	if err := proc.attachStream(stream); err != nil {
		return fmt.Errorf("attach stream: %w", err)
	}

	// Block until the stream is closed (the process will handle recv loop)
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	case <-proc.done:
		return nil
	}
}

func (m *Manager) Close() {
	m.cancel()

	// Stop gRPC server
	if m.grpcServer != nil {
		m.grpcServer.Stop()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for id, proc := range m.plugins {
		shutdown := &pb.HostToPlugin{
			PluginId: id,
			Payload: &pb.HostToPlugin_Shutdown{
				Shutdown: &pb.HostShutdown{Reason: "server shutting down"},
			},
		}
		proc.queue(shutdown)
		proc.Stop()
	}
	m.plugins = make(map[string]*pluginProcess)
}

func (m *Manager) AttachWorld(w *world.World) {
	if w == nil {
		return
	}
	if m.worldHandlerFactory != nil {
		handler := m.worldHandlerFactory(m)
		w.Handle(handler)
	}
	m.registerWorld(w)
}

func (m *Manager) AttachPlayer(p *player.Player) {
	if p == nil {
		return
	}
	if m.playerHandlerFactory != nil {
		handler := m.playerHandlerFactory(m)
		p.Handle(handler)
	}
	m.mu.Lock()
	m.players[p.UUID()] = p
	m.mu.Unlock()
	m.EmitPlayerJoin(p)
}

func (m *Manager) detachPlayer(p *player.Player) {
	m.mu.Lock()
	delete(m.players, p.UUID())
	m.mu.Unlock()
}

func (m *Manager) broadcastEvent(envelope *pb.EventEnvelope) {
	_ = m.dispatchEvent(envelope, false)
}

func (m *Manager) dispatchEvent(envelope *pb.EventEnvelope, expectResult bool) []*pb.EventResult {
	if envelope == nil {
		return nil
	}
	eventType := envelope.Type
	m.mu.RLock()
	procs := make([]*pluginProcess, 0, len(m.plugins))
	for _, proc := range m.plugins {
		if !proc.HasSubscription(eventType) {
			continue
		}
		procs = append(procs, proc)
	}
	m.mu.RUnlock()

	if len(procs) == 0 {
		return nil
	}

	results := make([]*pb.EventResult, 0, len(procs))
	for _, proc := range procs {
		var waitCh chan *pb.EventResult
		if expectResult {
			waitCh = proc.expectEventResult(envelope.EventId)
		}
		msg := &pb.HostToPlugin{
			PluginId: proc.id,
			Payload: &pb.HostToPlugin_Event{
				Event: envelope,
			},
		}
		proc.queue(msg)
		if !expectResult {
			continue
		}
		res, err := proc.waitEventResult(waitCh, eventResponseTimeout)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				proc.log.Warn("plugin did not respond to event", "event_id", envelope.EventId, "type", envelope.Type.String())
			}
			proc.discardEventResult(envelope.EventId)
			continue
		}
		if res != nil {
			results = append(results, res)
			if envelope.Type == pb.EventType_CHAT {
				if chatEvt := envelope.GetChat(); chatEvt != nil {
					if chatMut := res.GetChat(); chatMut != nil {
						chatEvt.Message = chatMut.Message
					}
				}
			}
		}
	}
	return results
}

func (m *Manager) emitCancellable(ctx cancelContext, envelope *pb.EventEnvelope) []*pb.EventResult {
	results := m.dispatchEvent(envelope, true)
	cancelled := false
	for _, res := range results {
		if res != nil && res.Cancel != nil && *res.Cancel {
			cancelled = true
		}
	}
	if cancelled && ctx != nil {
		ctx.Cancel()
	}
	return results
}

func (m *Manager) BroadcastEvent(evt *pb.EventEnvelope) {
	m.broadcastEvent(evt)
}

func (m *Manager) GenerateEventID() string {
	return m.generateEventID()
}

func convertProtoDrops(drops []*pb.ItemStack) []item.Stack {
	if len(drops) == 0 {
		return nil
	}
	converted := make([]item.Stack, 0, len(drops))
	for _, drop := range drops {
		if drop == nil || drop.Name == "" {
			continue
		}
		material, ok := world.ItemByName(drop.Name, int16(drop.Meta))
		if !ok {
			continue
		}
		count := int(drop.Count)
		if count <= 0 {
			continue
		}
		converted = append(converted, item.NewStack(material, count))
	}
	return converted
}

func (m *Manager) registerWorld(w *world.World) {
	if w == nil {
		return
	}
	name := strings.ToLower(w.Name())
	m.worldMu.Lock()
	m.worlds[name] = w
	m.worldMu.Unlock()
}

func (m *Manager) unregisterWorld(w *world.World) {
	if w == nil {
		return
	}
	name := strings.ToLower(w.Name())
	m.worldMu.Lock()
	if existing, ok := m.worlds[name]; ok && existing == w {
		delete(m.worlds, name)
	}
	m.worldMu.Unlock()
}

func (m *Manager) worldFromRef(ref *pb.WorldRef) *world.World {
	if ref == nil {
		return nil
	}
	name := strings.ToLower(ref.Name)
	m.worldMu.RLock()
	if name != "" {
		if w := m.worlds[name]; w != nil {
			m.worldMu.RUnlock()
			return w
		}
		for _, candidate := range m.worlds {
			if strings.EqualFold(candidate.Name(), ref.Name) {
				m.worldMu.RUnlock()
				return candidate
			}
		}
	}
	if ref.Dimension != "" {
		dim := strings.ToLower(ref.Dimension)
		for _, candidate := range m.worlds {
			if worldDimension(candidate) == dim {
				m.worldMu.RUnlock()
				return candidate
			}
		}
	}
	m.worldMu.RUnlock()
	return nil
}

func (m *Manager) handlePluginMessage(p *pluginProcess, msg *pb.PluginToHost) {
	if result := msg.GetEventResult(); result != nil {
		p.deliverEventResult(result)
	}
	if hello := msg.GetHello(); hello != nil {
		p.setHello(hello)
		m.registerCommands(p, hello.Commands)
	}
	if subscribe := msg.GetSubscribe(); subscribe != nil {
		p.updateSubscriptions(subscribe.Events)
	}
	if actions := msg.GetActions(); actions != nil {
		m.applyActions(p, actions)
	}
	if logMsg := msg.GetLog(); logMsg != nil {
		level := strings.ToLower(logMsg.Level)
		switch level {
		case "warn", "warning":
			p.log.Warn(logMsg.Message)
		case "error":
			p.log.Error(logMsg.Message)
		default:
			p.log.Info(logMsg.Message)
		}
	}
}

func (m *Manager) generateEventID() string {
	id := m.eventCounter.Add(1)
	return strconv.FormatUint(id, 10)
}
