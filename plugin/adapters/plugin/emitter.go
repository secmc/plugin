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

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/secmc/plugin/plugin/config"
	"github.com/secmc/plugin/plugin/ports"
	pb "github.com/secmc/plugin/proto/generated"
)

type Emitter struct {
	srv *server.Server
	log *slog.Logger

	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.RWMutex
	plugins  map[string]*pluginProcess
	players  map[uuid.UUID]*player.Player
	commands map[string]commandBinding

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

func NewEmitter(srv *server.Server, log *slog.Logger, playerHandlerFactory ports.PlayerHandlerFactory, worldHandlerFactory ports.WorldHandlerFactory) *Emitter {
	if log == nil {
		log = slog.Default()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Emitter{
		srv:                  srv,
		log:                  log.With("component", "plugin-manager"),
		ctx:                  ctx,
		cancel:               cancel,
		plugins:              make(map[string]*pluginProcess),
		players:              make(map[uuid.UUID]*player.Player),
		commands:             make(map[string]commandBinding),
		playerHandlerFactory: playerHandlerFactory,
		worldHandlerFactory:  worldHandlerFactory,
	}
}

func (m *Emitter) Start(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			m.log.Info("no plugin configuration found", "path", configPath)
			return nil
		}
		return err
	}
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
		go proc.start(m.ctx)
	}
	return nil
}

func (m *Emitter) Close() {
	m.cancel()
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

func (m *Emitter) AttachWorld(w *world.World) {
	if w == nil {
		return
	}
	if m.worldHandlerFactory != nil {
		handler := m.worldHandlerFactory(m)
		w.Handle(handler)
	}
}

func (m *Emitter) AttachPlayer(p *player.Player) {
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

func (m *Emitter) detachPlayer(p *player.Player) {
	m.mu.Lock()
	delete(m.players, p.UUID())
	m.mu.Unlock()
}

func (m *Emitter) EmitPlayerJoin(p *player.Player) {
	evt := &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    "PLAYER_JOIN",
		Payload: &pb.EventEnvelope_PlayerJoin{
			PlayerJoin: &pb.PlayerJoinEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
			},
		},
	}
	m.broadcastEvent(evt)
}

func (m *Emitter) EmitPlayerQuit(p *player.Player) {
	evt := &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    "PLAYER_QUIT",
		Payload: &pb.EventEnvelope_PlayerQuit{
			PlayerQuit: &pb.PlayerQuitEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
			},
		},
	}
	m.broadcastEvent(evt)
}

func (m *Emitter) EmitChat(ctx *player.Context, p *player.Player, msg *string) {
	if msg == nil {
		return
	}
	evt := &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    "CHAT",
		Payload: &pb.EventEnvelope_Chat{
			Chat: &pb.ChatEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				Message:    *msg,
			},
		},
	}
	results := m.dispatchEvent(evt, true)
	var cancelled bool
	for _, res := range results {
		if res == nil {
			continue
		}
		if res.Cancel != nil && *res.Cancel {
			cancelled = true
		}
		if chatMut := res.GetChat(); chatMut != nil {
			*msg = chatMut.Message
		}
	}
	if cancelled && ctx != nil {
		ctx.Cancel()
	}
}

func (m *Emitter) EmitCommand(ctx *player.Context, p *player.Player, cmdName string, args []string) {
	raw := "/" + cmdName
	if len(args) > 0 {
		raw += " " + strings.Join(args, " ")
	}
	evt := &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    "COMMAND",
		Payload: &pb.EventEnvelope_Command{
			Command: &pb.CommandEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				Raw:        raw,
				Command:    cmdName,
				Args:       args,
			},
		},
	}
	results := m.dispatchEvent(evt, true)
	for _, res := range results {
		if res != nil && res.Cancel != nil && *res.Cancel && ctx != nil {
			ctx.Cancel()
			break
		}
	}
}

func (m *Emitter) EmitBlockBreak(ctx *player.Context, p *player.Player, pos cube.Pos, drops *[]item.Stack, xp *int, worldDim string) {
	evt := &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    "BLOCK_BREAK",
		Payload: &pb.EventEnvelope_BlockBreak{
			BlockBreak: &pb.BlockBreakEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      worldDim,
				X:          int32(pos.X()),
				Y:          int32(pos.Y()),
				Z:          int32(pos.Z()),
			},
		},
	}
	results := m.dispatchEvent(evt, true)
	var cancelled bool
	for _, res := range results {
		if res == nil {
			continue
		}
		if res.Cancel != nil && *res.Cancel {
			cancelled = true
		}
		if bbMut := res.GetBlockBreak(); bbMut != nil {
			if drops != nil {
				*drops = convertProtoDrops(bbMut.Drops)
			}
			if bbMut.Xp != nil && xp != nil {
				*xp = int(*bbMut.Xp)
			}
		}
	}
	if cancelled && ctx != nil {
		ctx.Cancel()
	}
}

func (m *Emitter) broadcastEvent(envelope *pb.EventEnvelope) {
	_ = m.dispatchEvent(envelope, false)
}

func (m *Emitter) dispatchEvent(envelope *pb.EventEnvelope, expectResult bool) []*pb.EventResult {
	if envelope == nil {
		return nil
	}
	eventType := strings.ToUpper(envelope.Type)
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
				proc.log.Warn("plugin did not respond to event", "event_id", envelope.EventId, "type", envelope.Type)
			}
			proc.discardEventResult(envelope.EventId)
			continue
		}
		if res != nil {
			results = append(results, res)
			if envelope.Type == "CHAT" {
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

func (m *Emitter) BroadcastEvent(evt *pb.EventEnvelope) {
	m.broadcastEvent(evt)
}

func (m *Emitter) GenerateEventID() string {
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

func (m *Emitter) handlePluginMessage(p *pluginProcess, msg *pb.PluginToHost) {
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

func (m *Emitter) registerCommands(p *pluginProcess, specs []*pb.CommandSpec) {
	for _, spec := range specs {
		if spec == nil || spec.Name == "" {
			continue
		}
		name := strings.TrimPrefix(spec.Name, "/")

		aliases := make([]string, 0, len(spec.Aliases))
		for _, alias := range spec.Aliases {
			alias = strings.TrimPrefix(alias, "/")
			if alias == "" || alias == name {
				continue
			}
			aliases = append(aliases, alias)
		}

		binding := commandBinding{pluginID: p.id, command: name, descriptor: spec}
		m.mu.Lock()
		m.commands[name] = binding
		for _, alias := range aliases {
			m.commands[alias] = binding
		}
		m.mu.Unlock()

		cmd.Register(cmd.New(name, spec.Description, aliases, pluginCommand{mgr: m, pluginID: p.id, name: name}))
	}
}

type pluginCommand struct {
	mgr      *Emitter
	pluginID string
	name     string
}

func (c pluginCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	_, ok := src.(*player.Player)
	if !ok {
		output.Errorf("command only available to players")
		return
	}
	// No-op: PlayerHandler.HandleCommandExecution emits command events
}

func (m *Emitter) applyActions(p *pluginProcess, batch *pb.ActionBatch) {
	if batch == nil {
		return
	}
	for _, action := range batch.Actions {
		if action == nil {
			continue
		}
		switch kind := action.Kind.(type) {
		case *pb.Action_SendChat:
			m.handleSendChat(kind.SendChat)
		case *pb.Action_Teleport:
			m.handleTeleport(kind.Teleport)
		case *pb.Action_Kick:
			m.handleKick(kind.Kick)
		case *pb.Action_SetGameMode:
			m.handleSetGameMode(kind.SetGameMode)
		}
	}
}

func (m *Emitter) handleSendChat(act *pb.SendChatAction) {
	if act.TargetUuid == "" {
		for p := range m.srv.Players(nil) {
			p.Message(act.Message)
		}
		chat.Global.WriteString(act.Message)
		return
	}
	id, err := uuid.Parse(act.TargetUuid)
	if err != nil {
		return
	}
	if handle, ok := m.srv.Player(id); ok {
		handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
			if pl, ok := e.(*player.Player); ok {
				pl.Message(act.Message)
			}
		})
	}
}

func (m *Emitter) handleTeleport(act *pb.TeleportAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	if handle, ok := m.srv.Player(id); ok {
		handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
			pl, ok := e.(*player.Player)
			if !ok {
				return
			}
			pl.Teleport(mgl64.Vec3{act.X, act.Y, act.Z})
			rot := pl.Rotation()
			deltaYaw := float64(act.Yaw) - rot.Yaw()
			deltaPitch := float64(act.Pitch) - rot.Pitch()
			pl.Move(mgl64.Vec3{}, deltaYaw, deltaPitch)
		})
	}
}

func (m *Emitter) handleKick(act *pb.KickAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	if handle, ok := m.srv.Player(id); ok {
		handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
			if pl, ok := e.(*player.Player); ok {
				pl.Disconnect(act.Reason)
			}
		})
	}
}

func (m *Emitter) handleSetGameMode(act *pb.SetGameModeAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	gameMode, ok := world.GameModeByID(int(act.GameMode))
	if !ok {
		return
	}
	if handle, ok := m.srv.Player(id); ok {
		handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
			if pl, ok := e.(*player.Player); ok {
				pl.SetGameMode(gameMode)
			}
		})
	}
}

func (m *Emitter) generateEventID() string {
	id := m.eventCounter.Add(1)
	return strconv.FormatUint(id, 10)
}
