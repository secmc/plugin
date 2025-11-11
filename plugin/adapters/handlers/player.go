package handlers

import (
	"fmt"
	"net"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/secmc/plugin/plugin/ports"
)

type PlayerHandler struct {
	player.NopHandler
	manager ports.EventManager
}

func NewPlayerHandler(manager ports.EventManager) player.Handler {
	return &PlayerHandler{manager: manager}
}

func (h *PlayerHandler) HandleChat(ctx *player.Context, message *string) {
	h.manager.EmitChat(ctx, ctx.Val(), message)
}

func (h *PlayerHandler) HandleMove(ctx *player.Context, newPos mgl64.Vec3, newRot cube.Rotation) {
	h.manager.EmitPlayerMove(ctx, ctx.Val(), newPos, newRot)
}

func (h *PlayerHandler) HandleJump(p *player.Player) {
	h.manager.EmitPlayerJump(p)
}

func (h *PlayerHandler) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {
	h.manager.EmitPlayerTeleport(ctx, ctx.Val(), pos)
}

func (h *PlayerHandler) HandleChangeWorld(p *player.Player, before, after *world.World) {
	h.manager.EmitPlayerChangeWorld(p, before, after)
}

func (h *PlayerHandler) HandleToggleSprint(ctx *player.Context, after bool) {
	h.manager.EmitPlayerToggleSprint(ctx, ctx.Val(), after)
}

func (h *PlayerHandler) HandleToggleSneak(ctx *player.Context, after bool) {
	h.manager.EmitPlayerToggleSneak(ctx, ctx.Val(), after)
}

func (h *PlayerHandler) HandleFoodLoss(ctx *player.Context, from int, to *int) {
	h.manager.EmitPlayerFoodLoss(ctx, ctx.Val(), from, to)
}

func (h *PlayerHandler) HandleHeal(ctx *player.Context, health *float64, src world.HealingSource) {
	h.manager.EmitPlayerHeal(ctx, ctx.Val(), health, src)
}

func (h *PlayerHandler) HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource) {
	h.manager.EmitPlayerHurt(ctx, ctx.Val(), damage, immune, attackImmunity, src)
}

func (h *PlayerHandler) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
	h.manager.EmitPlayerDeath(p, src, keepInv)
}

func (h *PlayerHandler) HandleRespawn(p *player.Player, pos *mgl64.Vec3, w **world.World) {
	h.manager.EmitPlayerRespawn(p, pos, w)
}

func (h *PlayerHandler) HandleSkinChange(ctx *player.Context, skin *skin.Skin) {
	h.manager.EmitPlayerSkinChange(ctx, ctx.Val(), skin)
}

func (h *PlayerHandler) HandleFireExtinguish(ctx *player.Context, pos cube.Pos) {
	h.manager.EmitPlayerFireExtinguish(ctx, ctx.Val(), pos)
}

func (h *PlayerHandler) HandleStartBreak(ctx *player.Context, pos cube.Pos) {
	h.manager.EmitPlayerStartBreak(ctx, ctx.Val(), pos)
}

func (h *PlayerHandler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {
	h.manager.EmitCommand(ctx, ctx.Val(), command.Name(), args)
}

func (h *PlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	p := ctx.Val()
	worldDim := fmt.Sprint(p.Tx().World().Dimension())
	h.manager.EmitBlockBreak(ctx, p, pos, drops, xp, worldDim)
}

func (h *PlayerHandler) HandleQuit(p *player.Player) {
	h.manager.EmitPlayerQuit(p)
}

func (h *PlayerHandler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	h.manager.EmitPlayerBlockPlace(ctx, ctx.Val(), pos, b)
}

func (h *PlayerHandler) HandleBlockPick(ctx *player.Context, pos cube.Pos, b world.Block) {
	h.manager.EmitPlayerBlockPick(ctx, ctx.Val(), pos, b)
}

func (h *PlayerHandler) HandleItemUse(ctx *player.Context) {
	h.manager.EmitPlayerItemUse(ctx, ctx.Val())
}

func (h *PlayerHandler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {
	p := ctx.Val()
	if p == nil {
		return
	}
	var block world.Block
	if tx := p.Tx(); tx != nil {
		block = tx.Block(pos)
	}
	h.manager.EmitPlayerItemUseOnBlock(ctx, p, pos, face, clickPos, block)
}

func (h *PlayerHandler) HandleItemUseOnEntity(ctx *player.Context, e world.Entity) {
	h.manager.EmitPlayerItemUseOnEntity(ctx, ctx.Val(), e)
}

func (h *PlayerHandler) HandleItemRelease(ctx *player.Context, it item.Stack, dur time.Duration) {
	h.manager.EmitPlayerItemRelease(ctx, ctx.Val(), it, dur)
}

func (h *PlayerHandler) HandleItemConsume(ctx *player.Context, it item.Stack) {
	h.manager.EmitPlayerItemConsume(ctx, ctx.Val(), it)
}

func (h *PlayerHandler) HandleAttackEntity(ctx *player.Context, e world.Entity, force, height *float64, critical *bool) {
	h.manager.EmitPlayerAttackEntity(ctx, ctx.Val(), e, force, height, critical)
}

func (h *PlayerHandler) HandleExperienceGain(ctx *player.Context, amount *int) {
	h.manager.EmitPlayerExperienceGain(ctx, ctx.Val(), amount)
}

func (h *PlayerHandler) HandlePunchAir(ctx *player.Context) {
	h.manager.EmitPlayerPunchAir(ctx, ctx.Val())
}

func (h *PlayerHandler) HandleSignEdit(ctx *player.Context, pos cube.Pos, frontSide bool, oldText, newText string) {
	h.manager.EmitPlayerSignEdit(ctx, ctx.Val(), pos, frontSide, oldText, newText)
}

func (h *PlayerHandler) HandleLecternPageTurn(ctx *player.Context, pos cube.Pos, oldPage int, newPage *int) {
	h.manager.EmitPlayerLecternPageTurn(ctx, ctx.Val(), pos, oldPage, newPage)
}

func (h *PlayerHandler) HandleItemDamage(ctx *player.Context, it item.Stack, damage int) {
	h.manager.EmitPlayerItemDamage(ctx, ctx.Val(), it, damage)
}

func (h *PlayerHandler) HandleItemPickup(ctx *player.Context, it *item.Stack) {
	h.manager.EmitPlayerItemPickup(ctx, ctx.Val(), it)
}

func (h *PlayerHandler) HandleHeldSlotChange(ctx *player.Context, from, to int) {
	h.manager.EmitPlayerHeldSlotChange(ctx, ctx.Val(), from, to)
}

func (h *PlayerHandler) HandleItemDrop(ctx *player.Context, it item.Stack) {
	h.manager.EmitPlayerItemDrop(ctx, ctx.Val(), it)
}

func (h *PlayerHandler) HandleTransfer(ctx *player.Context, addr *net.UDPAddr) {
	h.manager.EmitPlayerTransfer(ctx, ctx.Val(), addr)
}

func (h *PlayerHandler) HandleDiagnostics(p *player.Player, d session.Diagnostics) {
	h.manager.EmitPlayerDiagnostics(p, d)
}
