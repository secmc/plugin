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
	mgr    *Manager
	Player *player.Player
}

func (h *PlayerHandler) HandleChat(ctx *player.Context, message *string) {
	if h.mgr == nil || h.Player == nil {
		return
	}
	h.mgr.emitChat(ctx, h.Player, message)
}

func (h *PlayerHandler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {
	if h.mgr == nil || h.Player == nil {
		return
	}
	h.mgr.emitCommandWithArgs(ctx, h.Player, command.Name(), args)
}

func (h *PlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	if h.mgr == nil || h.Player == nil {
		return
	}
	worldDim := fmt.Sprint(ctx.Val().Tx().World().Dimension())
	h.mgr.emitBlockBreak(ctx, h.Player, pos, drops, xp, worldDim)
}

func (h *PlayerHandler) HandleQuit(p *player.Player) {
	if h.mgr == nil {
		return
	}
	h.mgr.emitPlayerQuit(p)
	h.mgr.detachPlayer(p)
}
