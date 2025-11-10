package ports

import (
	"context"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	pb "github.com/secmc/plugin/proto/generated"
)

type PluginManager interface {
	Start(configPath string) error
	Close()
}

type PluginProcess interface {
	Start(ctx context.Context) error
	Stop()
	HasSubscription(eventType string) bool
	Queue(msg *pb.HostToPlugin)
}

type Stream interface {
	Send(data []byte) error
	Recv() ([]byte, error)
	CloseSend() error
	Close() error
}

type EventEmitter interface {
	EmitPlayerJoin(p *player.Player)
	EmitPlayerQuit(p *player.Player)
	EmitChat(ctx *player.Context, p *player.Player, msg *string)
	EmitCommand(ctx *player.Context, p *player.Player, cmdName string, args []string)
	EmitBlockBreak(ctx *player.Context, p *player.Player, pos cube.Pos, drops *[]item.Stack, xp *int, worldDim string)
	BroadcastEvent(evt *pb.EventEnvelope)
	GenerateEventID() string
}

type PlayerHandlerFactory func(emitter EventEmitter) player.Handler

type WorldHandlerFactory func(emitter EventEmitter) world.Handler

type PluginService interface {
	PluginManager
	EventEmitter
	AttachWorld(w *world.World)
	AttachPlayer(p *player.Player)
}
