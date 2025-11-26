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
	pb "github.com/secmc/plugin/proto/generated/go"
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
	// worldsByDim maps lowercased dimension name ("overworld","nether","end") to the world instance.
	worldsByDim map[string]*world.World
	// worldsByID maps a runtime-stable world ID (assigned by host) to the world.
	worldsByID map[string]*world.World

	eventCounter atomic.Uint64

	playerHandlerFactory ports.PlayerHandlerFactory
	worldHandlerFactory  ports.WorldHandlerFactory

	bootID string
}

// SetServer assigns the Dragonfly server instance after the manager has started.
// This enables starting the plugin transport before the server is created so that
// plugins can register custom items in their Hello message ahead of server startup.
func (m *Manager) SetServer(s *server.Server) {
	m.srv = s
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
		worldsByDim:          make(map[string]*world.World),
		worldsByID:           make(map[string]*world.World),
		playerHandlerFactory: playerHandlerFactory,
		worldHandlerFactory:  worldHandlerFactory,
		bootID:               uuid.NewString(),
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
	return m.StartWithConfig(cfg)
}

// StartWithConfig starts the plugin adapter using a pre-loaded plugin config.
func (m *Manager) StartWithConfig(cfg config.Config) error {
	// Start gRPC server to accept plugin connections
	address := cfg.ServerPort
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

// broadcastEvent sends an event which does not expect a response.
// This is for events which cannot be canceled or mutated.
func (m *Manager) broadcastEvent(envelope *pb.EventEnvelope) {
	envelope.ExpectsResponse = false
	_ = m.dispatchEvent(envelope, false)
}

func (m *Manager) dispatchEvent(envelope *pb.EventEnvelope, expectResult bool) []*pb.EventResult {
	if envelope == nil {
		return nil
	}
	if envelope.EventId == "" {
		envelope.EventId = m.generateEventID()
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
		proc.log.Debug("sending event", "event_id", envelope.EventId, "type", envelope.Type.String())
		proc.queue(msg)

		if !expectResult {
			continue
		}

		waitStart := time.Now()
		res, err := proc.waitEventResult(waitCh, eventResponseTimeout)
		pluginResponseTime := time.Since(waitStart)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				proc.log.Warn("plugin did not respond to event",
					"event_id", envelope.EventId,
					"type", envelope.Type.String(),
					"wait_ms", pluginResponseTime.Milliseconds())
			}
			proc.discardEventResult(envelope.EventId)
			continue
		}
		if res != nil {
			results = append(results, res)

			// Log timing for command events
			if envelope.Type == pb.EventType_COMMAND {
				proc.log.Debug("plugin command response received",
					"event_id", envelope.EventId,
					"plugin_response_ms", pluginResponseTime.Milliseconds(),
					"plugin_response_us", pluginResponseTime.Microseconds())
			} else {
				// General timing for non-command events
				proc.log.Debug("plugin event response received",
					"event_id", envelope.EventId,
					"type", envelope.Type.String(),
					"plugin_response_ms", pluginResponseTime.Milliseconds(),
					"plugin_response_us", pluginResponseTime.Microseconds())
			}
		}
	}
	return results
}

// dispatchEventParallel broadcasts an event to all subscribed plugins concurrently and collects results.
func (m *Manager) dispatchEventParallel(envelope *pb.EventEnvelope, expectResult bool) []*pb.EventResult {
	if envelope == nil {
		return nil
	}
	if envelope.EventId == "" {
		envelope.EventId = m.generateEventID()
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

	results := make([]*pb.EventResult, len(procs))
	var wg sync.WaitGroup
	for idx, proc := range procs {
		wg.Go(func() {
			var waitCh chan *pb.EventResult
			if expectResult {
				waitCh = proc.expectEventResult(envelope.EventId)
			}
			proc.log.Debug("sending event", "event_id", envelope.EventId, "type", envelope.Type.String())
			proc.queue(&pb.HostToPlugin{
				PluginId: proc.id,
				Payload: &pb.HostToPlugin_Event{
					Event: envelope,
				},
			})
			if !expectResult {
				return
			}
			waitStart := time.Now()
			res, err := proc.waitEventResult(waitCh, eventResponseTimeout)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					proc.log.Warn("plugin did not respond to event", "event_id", envelope.EventId, "type", envelope.Type.String())
				}
				proc.discardEventResult(envelope.EventId)
				return
			}
			pluginResponseTime := time.Since(waitStart)
			if envelope.Type != pb.EventType_COMMAND {
				proc.log.Debug("plugin event response received",
					"event_id", envelope.EventId,
					"type", envelope.Type.String(),
					"plugin_response_ms", pluginResponseTime.Milliseconds(),
					"plugin_response_us", pluginResponseTime.Microseconds())
			}
			results[idx] = res
		})
	}
	wg.Wait()
	return results
}

func (m *Manager) emitCancellable(ctx cancelContext, envelope *pb.EventEnvelope) []*pb.EventResult {
	envelope.ExpectsResponse = true
	// Fire all at once and wait for all responses.
	results := m.dispatchEventParallel(envelope, true)
	cancelled := false
	for _, res := range results {
		if res != nil && res.Cancel != nil && *res.Cancel {
			cancelled = true
		}
	}
	if cancelled && ctx != nil {
		ctx.Cancel()
	}
	// If any plugin cancelled, do not apply any mutations.
	if cancelled {
		m.log.Debug("event cancelled by plugin", "event_id", envelope.EventId, "type", envelope.Type.String())
		return nil
	}
	m.log.Debug("event completed", "event_id", envelope.EventId, "type", envelope.Type.String(), "responses", len(results))
	return results
}

type PtrTo[T any] interface{ ~*T }

// applyMutations iterates results and applies mutations using the provided getter and applier functions.
// The getter returns the mutation pointer; nil mutations are skipped.
func applyMutations[T any, P PtrTo[T]](
	results []*pb.EventResult,
	getter func(*pb.EventResult) P,
	applier func(P),
) {
	for _, res := range results {
		if res == nil {
			continue
		}
		mut := getter(res)
		if mut == nil {
			continue
		}
		applier(mut)
	}
}

// mutateField applies a single field mutation if both pointers are non-nil.
func mutateField[T any](dest *T, src *T) {
	if dest != nil && src != nil {
		*dest = *src
	}
}

// mutateInt32 applies int32 to int conversion mutation.
func mutateInt32(dest *int, src *int32) {
	if dest != nil && src != nil {
		*dest = int(*src)
	}
}

// mutateInt64Ms applies int64 milliseconds to time.Duration conversion mutation.
func mutateInt64Ms(dest *time.Duration, src *int64) {
	if dest != nil && src != nil {
		*dest = time.Duration(*src) * time.Millisecond
	}
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
	dim := strings.ToLower(fmt.Sprint(w.Dimension()))
	id := fmt.Sprintf("%p", w)
	m.worldMu.Lock()
	m.worlds[name] = w
	if m.worldsByDim == nil {
		m.worldsByDim = make(map[string]*world.World)
	}
	m.worldsByDim[dim] = w
	if m.worldsByID == nil {
		m.worldsByID = make(map[string]*world.World)
	}
	m.worldsByID[id] = w
	m.worldMu.Unlock()
}

func (m *Manager) unregisterWorld(w *world.World) {
	if w == nil {
		return
	}
	name := strings.ToLower(w.Name())
	dim := strings.ToLower(fmt.Sprint(w.Dimension()))
	id := fmt.Sprintf("%p", w)
	m.worldMu.Lock()
	if existing, ok := m.worlds[name]; ok && existing == w {
		delete(m.worlds, name)
	}
	if m.worldsByDim != nil {
		if existing, ok := m.worldsByDim[dim]; ok && existing == w {
			delete(m.worldsByDim, dim)
		}
	}
	if m.worldsByID != nil {
		if existing, ok := m.worldsByID[id]; ok && existing == w {
			delete(m.worldsByID, id)
		}
	}
	m.worldMu.Unlock()
}

func (m *Manager) worldFromRef(ref *pb.WorldRef) *world.World {
	if ref == nil {
		return nil
	}

	m.worldMu.RLock()
	defer m.worldMu.RUnlock()

	// Prefer lookup by host-assigned ID when provided.
	if ref.Id != "" {
		if m.worldsByID != nil {
			if w := m.worldsByID[ref.Id]; w != nil {
				return w
			}
		}
	}

	// Prefer dimension lookup to disambiguate worlds that may share the same name (e.g., "World").
	if ref.Dimension != "" {
		dim := strings.ToLower(ref.Dimension)
		if m.worldsByDim != nil {
			if w := m.worldsByDim[dim]; w != nil {
				return w
			}
		}
	}

	// Fallback to name lookup.
	if ref.Name != "" {
		name := strings.ToLower(ref.Name)
		if w := m.worlds[name]; w != nil {
			return w
		}
	}

	return nil
}

func (m *Manager) handlePluginMessage(p *pluginProcess, msg *pb.PluginToHost) {
	switch payload := msg.GetPayload().(type) {
	case *pb.PluginToHost_EventResult:
		p.deliverEventResult(payload.EventResult)
	case *pb.PluginToHost_Hello:
		hello := payload.Hello
		cmdNames := mapSlice(hello.Commands, func(cmd *pb.CommandSpec) string {
			if len(cmd.Aliases) > 0 {
				return fmt.Sprintf("%s (aliases: %v)", cmd.Name, cmd.Aliases)
			}
			return cmd.Name
		})
		m.log.Info(fmt.Sprintf("✓ %s v%s connected [%s]", hello.Name, hello.Version, hello.ApiVersion), "commands", cmdNames)
		p.setHello(hello)
		m.registerCommands(p, hello.Commands)
		m.registerCustomItems(p, hello.CustomItems)
		m.registerCustomBlocks(p, hello.CustomBlocks)
	case *pb.PluginToHost_Subscribe:
		subscribe := payload.Subscribe
		eventNames := mapSlice(subscribe.Events, func(evt pb.EventType) string {
			return evt.String()
		})
		pluginName := p.id
		if hello := p.helloInfo(); hello != nil && hello.Name != "" {
			pluginName = hello.Name
		}
		m.log.Info(fmt.Sprintf("  %s subscribed to %d events", pluginName, len(eventNames)), "events", eventNames)
		p.updateSubscriptions(subscribe.Events)
	case *pb.PluginToHost_Actions:
		m.applyActions(p, payload.Actions)
	case *pb.PluginToHost_Log:
		logMsg := payload.Log
		level := strings.ToLower(logMsg.Level)
		switch level {
		case "warn", "warning":
			p.log.Warn(logMsg.Message)
		case "error":
			p.log.Error(logMsg.Message)
		default:
			p.log.Info(logMsg.Message)
		}
	case *pb.PluginToHost_ServerInfo:
		var pluginNames []string

		for _, pl := range m.plugins {
			pluginNames = append(pluginNames, pl.cfg.Name)
		}
		p.sendServerInfo(pluginNames)
	default:
		p.log.Info(fmt.Sprintf("unhandled event: %#v", payload))
	}
}

// mapSlice transforms a slice using the provided function
func mapSlice[T any, R any](slice []T, fn func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func (m *Manager) generateEventID() string {
	id := m.eventCounter.Add(1)
	return strconv.FormatUint(id, 10)
}

// WaitForAnyHello waits until at least one connected plugin has sent a Hello,
// or until the timeout expires. Returns true if a Hello was observed.
// WaitForAnyPlugin waits until at least one connected plugin has sent a Hello,
// or until the timeout expires. Returns true if a Hello was observed.
func (m *Manager) WaitForAnyPlugin(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		m.mu.RLock()
		for _, proc := range m.plugins {
			if proc.helloInfo() != nil {
				m.mu.RUnlock()
				return true
			}
		}
		m.mu.RUnlock()
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// WaitForPlugins waits until all required plugin IDs have sent Hello, or until timeout.
// If requiredIDs is empty, it falls back to waiting for any plugin Hello.
func (m *Manager) WaitForPlugins(requiredIDs []string, timeout time.Duration) bool {
	if len(requiredIDs) == 0 {
		return m.WaitForAnyPlugin(timeout)
	}
	need := make(map[string]struct{}, len(requiredIDs))
	for _, id := range requiredIDs {
		if id == "" {
			continue
		}
		need[strings.ToLower(id)] = struct{}{}
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		m.mu.RLock()
		for id, proc := range m.plugins {
			lid := strings.ToLower(id)
			if _, ok := need[lid]; !ok {
				continue
			}
			if proc.helloInfo() != nil {
				delete(need, lid)
			}
		}
		m.mu.RUnlock()
		if len(need) == 0 {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}
