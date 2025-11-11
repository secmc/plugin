package handlers

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/secmc/plugin/plugin/ports"
)

var _ world.Handler = (*WorldHandler)(nil)

type WorldHandler struct {
	world.NopHandler
	manager ports.EventManager
}

func NewWorldHandler(manager ports.EventManager) world.Handler {
	return &WorldHandler{manager: manager}
}

func (h *WorldHandler) HandleClose(tx *world.Tx) {
	h.manager.EmitWorldClose(tx)
}

func (h *WorldHandler) HandleLiquidFlow(ctx *world.Context, from, into cube.Pos, liquid world.Liquid, replaced world.Block) {
	h.manager.EmitWorldLiquidFlow(ctx, from, into, liquid, replaced)
}

func (h *WorldHandler) HandleLiquidDecay(ctx *world.Context, pos cube.Pos, before, after world.Liquid) {
	h.manager.EmitWorldLiquidDecay(ctx, pos, before, after)
}

func (h *WorldHandler) HandleLiquidHarden(ctx *world.Context, pos cube.Pos, liquidHardened, otherLiquid, newBlock world.Block) {
	h.manager.EmitWorldLiquidHarden(ctx, pos, liquidHardened, otherLiquid, newBlock)
}

func (h *WorldHandler) HandleSound(ctx *world.Context, s world.Sound, pos mgl64.Vec3) {
	h.manager.EmitWorldSound(ctx, s, pos)
}

func (h *WorldHandler) HandleFireSpread(ctx *world.Context, from, to cube.Pos) {
	h.manager.EmitWorldFireSpread(ctx, from, to)
}

func (h *WorldHandler) HandleBlockBurn(ctx *world.Context, pos cube.Pos) {
	h.manager.EmitWorldBlockBurn(ctx, pos)
}

func (h *WorldHandler) HandleCropTrample(ctx *world.Context, pos cube.Pos) {
	h.manager.EmitWorldCropTrample(ctx, pos)
}

func (h *WorldHandler) HandleLeavesDecay(ctx *world.Context, pos cube.Pos) {
	h.manager.EmitWorldLeavesDecay(ctx, pos)
}

func (h *WorldHandler) HandleEntitySpawn(tx *world.Tx, e world.Entity) {
	h.manager.EmitWorldEntitySpawn(tx, e)
}

func (h *WorldHandler) HandleEntityDespawn(tx *world.Tx, e world.Entity) {
	h.manager.EmitWorldEntityDespawn(tx, e)
}

func (h *WorldHandler) HandleExplosion(ctx *world.Context, position mgl64.Vec3, entities *[]world.Entity, blocks *[]cube.Pos, itemDropChance *float64, spawnFire *bool) {
	h.manager.EmitWorldExplosion(ctx, position, entities, blocks, itemDropChance, spawnFire)
}
