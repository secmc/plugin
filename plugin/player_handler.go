package plugin

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
)

type PlayerHandler struct {
	player.NopHandler
	mgr *Manager
}

func (h *PlayerHandler) HandleChat(ctx *player.Context, message *string) {
	if h.mgr == nil {
		return
	}
	h.mgr.emitChat(ctx, ctx.Val(), message)
}

func (h *PlayerHandler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {
	if h.mgr == nil {
		return
	}
	h.mgr.emitCommandWithArgs(ctx, ctx.Val(), command.Name(), args)
}

func (h *PlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	if h.mgr == nil {
		return
	}
	p := ctx.Val()
	worldDim := fmt.Sprint(p.Tx().World().Dimension())
	h.mgr.emitBlockBreak(ctx, p, pos, drops, xp, worldDim)
}

func (h *PlayerHandler) HandleQuit(p *player.Player) {
	if h.mgr == nil {
		return
	}
	h.mgr.emitPlayerQuit(p)
	h.mgr.detachPlayer(p)
}
