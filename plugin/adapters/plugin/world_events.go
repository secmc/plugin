package plugin

import (
	"fmt"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	pb "github.com/secmc/plugin/proto/generated"
)

func (m *Manager) EmitWorldLiquidFlow(ctx *world.Context, from, into cube.Pos, liquid world.Liquid, replaced world.Block) {
	m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_LIQUID_FLOW,
		Payload: &pb.EventEnvelope_WorldLiquidFlow{
			WorldLiquidFlow: &pb.WorldLiquidFlowEvent{
				World:    protoWorldRef(worldFromContext(ctx)),
				From:     protoBlockPos(from),
				To:       protoBlockPos(into),
				Liquid:   protoLiquidState(liquid),
				Replaced: protoBlockState(replaced),
			},
		},
	})
}

func (m *Manager) EmitWorldLiquidDecay(ctx *world.Context, pos cube.Pos, before, after world.Liquid) {
	m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_LIQUID_DECAY,
		Payload: &pb.EventEnvelope_WorldLiquidDecay{
			WorldLiquidDecay: &pb.WorldLiquidDecayEvent{
				World:    protoWorldRef(worldFromContext(ctx)),
				Position: protoBlockPos(pos),
				Before:   protoLiquidState(before),
				After:    protoLiquidState(after),
			},
		},
	})
}

func (m *Manager) EmitWorldLiquidHarden(ctx *world.Context, pos cube.Pos, liquidHardened, otherLiquid, newBlock world.Block) {
	m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_LIQUID_HARDEN,
		Payload: &pb.EventEnvelope_WorldLiquidHarden{
			WorldLiquidHarden: &pb.WorldLiquidHardenEvent{
				World:          protoWorldRef(worldFromContext(ctx)),
				Position:       protoBlockPos(pos),
				LiquidHardened: protoLiquidOrBlockState(liquidHardened),
				OtherLiquid:    protoLiquidOrBlockState(otherLiquid),
				NewBlock:       protoBlockState(newBlock),
			},
		},
	})
}

func (m *Manager) EmitWorldSound(ctx *world.Context, s world.Sound, pos mgl64.Vec3) {
	m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_SOUND,
		Payload: &pb.EventEnvelope_WorldSound{
			WorldSound: &pb.WorldSoundEvent{
				World:    protoWorldRef(worldFromContext(ctx)),
				Sound:    fmt.Sprintf("%T", s),
				Position: protoVec3(pos),
			},
		},
	})
}

func (m *Manager) EmitWorldFireSpread(ctx *world.Context, from, to cube.Pos) {
	m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_FIRE_SPREAD,
		Payload: &pb.EventEnvelope_WorldFireSpread{
			WorldFireSpread: &pb.WorldFireSpreadEvent{
				World: protoWorldRef(worldFromContext(ctx)),
				From:  protoBlockPos(from),
				To:    protoBlockPos(to),
			},
		},
	})
}

func (m *Manager) EmitWorldBlockBurn(ctx *world.Context, pos cube.Pos) {
	m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_BLOCK_BURN,
		Payload: &pb.EventEnvelope_WorldBlockBurn{
			WorldBlockBurn: &pb.WorldBlockBurnEvent{
				World:    protoWorldRef(worldFromContext(ctx)),
				Position: protoBlockPos(pos),
			},
		},
	})
}

func (m *Manager) EmitWorldCropTrample(ctx *world.Context, pos cube.Pos) {
	m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_CROP_TRAMPLE,
		Payload: &pb.EventEnvelope_WorldCropTrample{
			WorldCropTrample: &pb.WorldCropTrampleEvent{
				World:    protoWorldRef(worldFromContext(ctx)),
				Position: protoBlockPos(pos),
			},
		},
	})
}

func (m *Manager) EmitWorldLeavesDecay(ctx *world.Context, pos cube.Pos) {
	m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_LEAVES_DECAY,
		Payload: &pb.EventEnvelope_WorldLeavesDecay{
			WorldLeavesDecay: &pb.WorldLeavesDecayEvent{
				World:    protoWorldRef(worldFromContext(ctx)),
				Position: protoBlockPos(pos),
			},
		},
	})
}

func (m *Manager) EmitWorldEntitySpawn(tx *world.Tx, e world.Entity) {
	m.broadcastEvent(&pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_ENTITY_SPAWN,
		Payload: &pb.EventEnvelope_WorldEntitySpawn{
			WorldEntitySpawn: &pb.WorldEntitySpawnEvent{
				World:  protoWorldRef(worldFromTx(tx)),
				Entity: protoEntityRef(e),
			},
		},
	})
}

func (m *Manager) EmitWorldEntityDespawn(tx *world.Tx, e world.Entity) {
	m.broadcastEvent(&pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_ENTITY_DESPAWN,
		Payload: &pb.EventEnvelope_WorldEntityDespawn{
			WorldEntityDespawn: &pb.WorldEntityDespawnEvent{
				World:  protoWorldRef(worldFromTx(tx)),
				Entity: protoEntityRef(e),
			},
		},
	})
}

func (m *Manager) EmitWorldExplosion(ctx *world.Context, position mgl64.Vec3, entities *[]world.Entity, blocks *[]cube.Pos, itemDropChance *float64, spawnFire *bool) {
	var entityRefs []*pb.EntityRef
	if entities != nil {
		entityRefs = protoEntityRefs(*entities)
	}
	var blockPositions []*pb.BlockPos
	if blocks != nil {
		blockPositions = protoBlockPositions(*blocks)
	}
	dropChance := 0.0
	if itemDropChance != nil {
		dropChance = *itemDropChance
	}
	spawnFireVal := false
	if spawnFire != nil {
		spawnFireVal = *spawnFire
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_EXPLOSION,
		Payload: &pb.EventEnvelope_WorldExplosion{
			WorldExplosion: &pb.WorldExplosionEvent{
				World:            protoWorldRef(worldFromContext(ctx)),
				Position:         protoVec3(position),
				AffectedEntities: entityRefs,
				AffectedBlocks:   blockPositions,
				ItemDropChance:   dropChance,
				SpawnFire:        spawnFireVal,
			},
		},
	})
	for _, res := range results {
		if res == nil {
			continue
		}
		mut := res.GetWorldExplosion()
		if mut == nil {
			continue
		}
		if entities != nil && mut.EntityUuids != nil {
			*entities = filterEntitiesByUUIDs(*entities, mut.EntityUuids)
		}
		if blocks != nil && mut.Blocks != nil {
			converted := convertProtoBlockPositionsToCube(mut.Blocks)
			if converted == nil {
				*blocks = nil
			} else {
				*blocks = converted
			}
		}
		if itemDropChance != nil && mut.ItemDropChance != nil {
			*itemDropChance = *mut.ItemDropChance
		}
		if spawnFire != nil && mut.SpawnFire != nil {
			*spawnFire = *mut.SpawnFire
		}
	}
}

func (m *Manager) EmitWorldClose(tx *world.Tx) {
	w := worldFromTx(tx)
	m.broadcastEvent(&pb.EventEnvelope{
		EventId: m.generateEventID(),
		Type:    pb.EventType_WORLD_CLOSE,
		Payload: &pb.EventEnvelope_WorldClose{
			WorldClose: &pb.WorldCloseEvent{
				World: protoWorldRef(w),
			},
		},
	})
	m.unregisterWorld(w)
}

func worldFromTx(tx *world.Tx) *world.World {
	if tx == nil {
		return nil
	}
	return tx.World()
}

func filterEntitiesByUUIDs(entities []world.Entity, uuids []string) []world.Entity {
	if len(uuids) == 0 {
		return entities
	}
	allowed := make(map[string]struct{}, len(uuids))
	for _, id := range uuids {
		if id == "" {
			continue
		}
		allowed[strings.ToLower(id)] = struct{}{}
	}
	if len(allowed) == 0 {
		return entities
	}
	filtered := make([]world.Entity, 0, len(entities))
	for _, e := range entities {
		handle := e.H()
		if handle == nil {
			continue
		}
		if _, ok := allowed[strings.ToLower(handle.UUID().String())]; ok {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
