package handlers

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/secmc/plugin/plugin/ports"
)

type PlayerHandler struct {
	player.NopHandler
	emitter ports.EventEmitter
}

func NewPlayerHandler(emitter ports.EventEmitter) player.Handler {
	return &PlayerHandler{emitter: emitter}
}

func (h *PlayerHandler) HandleChat(ctx *player.Context, message *string) {
	if h.emitter == nil {
		return
	}
	h.emitter.EmitChat(ctx, ctx.Val(), message)
}

func (h *PlayerHandler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {
	if h.emitter == nil {
		return
	}
	h.emitter.EmitCommand(ctx, ctx.Val(), command.Name(), args)
}

func (h *PlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	if h.emitter == nil {
		return
	}
	p := ctx.Val()
	worldDim := fmt.Sprint(p.Tx().World().Dimension())
	h.emitter.EmitBlockBreak(ctx, p, pos, drops, xp, worldDim)
}

func (h *PlayerHandler) HandleQuit(p *player.Player) {
	if h.emitter == nil {
		return
	}
	h.emitter.EmitPlayerQuit(p)
}
