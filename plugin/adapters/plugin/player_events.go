package plugin

import (
	"net"
	"strings"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	pb "github.com/secmc/plugin/proto/generated/go"
)

func (m *Manager) EmitPlayerJoin(p *player.Player) {
	if p == nil {
		return
	}
	m.broadcastEvent(&pb.EventEnvelope{
		Type: pb.EventType_PLAYER_JOIN,
		Payload: &pb.EventEnvelope_PlayerJoin{
			PlayerJoin: &pb.PlayerJoinEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldRef(p),
			},
		},
	})
}

func (m *Manager) EmitPlayerQuit(p *player.Player) {
	if p == nil {
		return
	}
	m.broadcastEvent(&pb.EventEnvelope{
		Type: pb.EventType_PLAYER_QUIT,
		Payload: &pb.EventEnvelope_PlayerQuit{
			PlayerQuit: &pb.PlayerQuitEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
			},
		},
	})
	m.detachPlayer(p)
}

func (m *Manager) EmitChat(ctx *player.Context, p *player.Player, msg *string) {
	if p == nil || msg == nil {
		return
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_CHAT,
		Payload: &pb.EventEnvelope_Chat{
			Chat: &pb.ChatEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				Message:    *msg,
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.ChatMutation { return r.GetChat() },
		func(mut *pb.ChatMutation) { mutateField(msg, mut.Message) },
	)
}

func (m *Manager) EmitCommand(ctx *player.Context, p *player.Player, cmdName string, args []string) {
	if p == nil {
		return
	}
	// Normalize arguments: trim spaces and drop empties to avoid usage errors on trailing/multiple spaces.
	norm := normalizeArgs(args)
	raw := "/" + cmdName
	if len(norm) > 0 {
		raw += " " + strings.Join(norm, " ")
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_COMMAND,
		Payload: &pb.EventEnvelope_Command{
			Command: &pb.CommandEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				Raw:        raw,
				Command:    cmdName,
				Args:       norm,
			},
		},
	})
}

// normalizeArgs trims each argument and removes empty entries.
func normalizeArgs(args []string) []string {
	return strings.Fields(strings.Join(args, " "))
}

func (m *Manager) EmitBlockBreak(ctx *player.Context, p *player.Player, pos cube.Pos, drops *[]item.Stack, xp *int, worldDim string) {
	if p == nil {
		return
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_BLOCK_BREAK,
		Payload: &pb.EventEnvelope_BlockBreak{
			BlockBreak: &pb.BlockBreakEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      worldDim,
				Position:   protoBlockPos(pos),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.BlockBreakMutation { return r.GetBlockBreak() },
		func(mut *pb.BlockBreakMutation) {
			if drops != nil && mut.Drops != nil {
				*drops = convertProtoDrops(mut.Drops.Items)
			}
			mutateInt32(xp, mut.Xp)
		},
	)
}

func (m *Manager) EmitPlayerMove(ctx *player.Context, p *player.Player, newPos mgl64.Vec3, newRot cube.Rotation) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_MOVE,
		Payload: &pb.EventEnvelope_PlayerMove{
			PlayerMove: &pb.PlayerMoveEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoVec3(newPos),
				Rotation:   protoRotation(newRot),
			},
		},
	})
}

func (m *Manager) EmitPlayerJump(p *player.Player) {
	if p == nil {
		return
	}
	m.broadcastEvent(&pb.EventEnvelope{
		Type: pb.EventType_PLAYER_JUMP,
		Payload: &pb.EventEnvelope_PlayerJump{
			PlayerJump: &pb.PlayerJumpEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoVec3(p.Position()),
			},
		},
	})
}

func (m *Manager) EmitPlayerTeleport(ctx *player.Context, p *player.Player, pos mgl64.Vec3) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_TELEPORT,
		Payload: &pb.EventEnvelope_PlayerTeleport{
			PlayerTeleport: &pb.PlayerTeleportEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoVec3(pos),
			},
		},
	})
}

func (m *Manager) EmitPlayerChangeWorld(p *player.Player, before, after *world.World) {
	if p == nil {
		return
	}
	m.broadcastEvent(&pb.EventEnvelope{
		Type: pb.EventType_PLAYER_CHANGE_WORLD,
		Payload: &pb.EventEnvelope_PlayerChangeWorld{
			PlayerChangeWorld: &pb.PlayerChangeWorldEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				Before:     protoWorldRef(before),
				After:      protoWorldRef(after),
			},
		},
	})
}

func (m *Manager) EmitPlayerToggleSprint(ctx *player.Context, p *player.Player, after bool) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_TOGGLE_SPRINT,
		Payload: &pb.EventEnvelope_PlayerToggleSprint{
			PlayerToggleSprint: &pb.PlayerToggleSprintEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				After:      after,
			},
		},
	})
}

func (m *Manager) EmitPlayerToggleSneak(ctx *player.Context, p *player.Player, after bool) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_TOGGLE_SNEAK,
		Payload: &pb.EventEnvelope_PlayerToggleSneak{
			PlayerToggleSneak: &pb.PlayerToggleSneakEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				After:      after,
			},
		},
	})
}

func (m *Manager) EmitPlayerFoodLoss(ctx *player.Context, p *player.Player, from int, to *int) {
	if p == nil {
		return
	}
	toVal := 0
	if to != nil {
		toVal = *to
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_FOOD_LOSS,
		Payload: &pb.EventEnvelope_PlayerFoodLoss{
			PlayerFoodLoss: &pb.PlayerFoodLossEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				From:       int32(from),
				To:         int32(toVal),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerFoodLossMutation { return r.GetPlayerFoodLoss() },
		func(mut *pb.PlayerFoodLossMutation) { mutateInt32(to, mut.To) },
	)
}

func (m *Manager) EmitPlayerHeal(ctx *player.Context, p *player.Player, health *float64, src world.HealingSource) {
	if p == nil {
		return
	}
	amount := 0.0
	if health != nil {
		amount = *health
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_HEAL,
		Payload: &pb.EventEnvelope_PlayerHeal{
			PlayerHeal: &pb.PlayerHealEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				Amount:     amount,
				Source:     protoHealingSource(src),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerHealMutation { return r.GetPlayerHeal() },
		func(mut *pb.PlayerHealMutation) { mutateField(health, mut.Amount) },
	)
}

func (m *Manager) EmitPlayerHurt(ctx *player.Context, p *player.Player, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource) {
	if p == nil {
		return
	}
	dmg := 0.0
	if damage != nil {
		dmg = *damage
	}
	var immunityMS int64
	if attackImmunity != nil {
		immunityMS = attackImmunity.Milliseconds()
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_HURT,
		Payload: &pb.EventEnvelope_PlayerHurt{
			PlayerHurt: &pb.PlayerHurtEvent{
				PlayerUuid:       p.UUID().String(),
				Name:             p.Name(),
				Damage:           dmg,
				Immune:           immune,
				AttackImmunityMs: immunityMS,
				Source:           protoDamageSource(src),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerHurtMutation { return r.GetPlayerHurt() },
		func(mut *pb.PlayerHurtMutation) {
			mutateField(damage, mut.Damage)
			mutateInt64Ms(attackImmunity, mut.AttackImmunityMs)
		},
	)
}

func (m *Manager) EmitPlayerDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
	if p == nil {
		return
	}
	keep := false
	if keepInv != nil {
		keep = *keepInv
	}
	results := m.emitCancellable(nil, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_DEATH,
		Payload: &pb.EventEnvelope_PlayerDeath{
			PlayerDeath: &pb.PlayerDeathEvent{
				PlayerUuid:    p.UUID().String(),
				Name:          p.Name(),
				Source:        protoDamageSource(src),
				KeepInventory: keep,
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerDeathMutation { return r.GetPlayerDeath() },
		func(mut *pb.PlayerDeathMutation) { mutateField(keepInv, mut.KeepInventory) },
	)
}

func (m *Manager) EmitPlayerRespawn(p *player.Player, pos *mgl64.Vec3, w **world.World) {
	if p == nil {
		return
	}
	var vec *pb.Vec3
	if pos != nil {
		vec = protoVec3(*pos)
	}
	var worldRef *pb.WorldRef
	if w != nil && *w != nil {
		worldRef = protoWorldRef(*w)
	}
	envelope := &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_RESPAWN,
		Payload: &pb.EventEnvelope_PlayerRespawn{
			PlayerRespawn: &pb.PlayerRespawnEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				Position:   vec,
				World:      worldRef,
			},
		},
	}
	results := m.dispatchEvent(envelope, true)
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerRespawnMutation { return r.GetPlayerRespawn() },
		func(mut *pb.PlayerRespawnMutation) {
			if pos != nil {
				if newPos, ok := vec3FromProto(mut.Position); ok {
					*pos = newPos
				}
			}
			if w != nil {
				if mut.World == nil {
					*w = nil
				} else if newWorld := m.worldFromRef(mut.World); newWorld != nil {
					*w = newWorld
				}
			}
		},
	)
}

func (m *Manager) EmitPlayerSkinChange(ctx *player.Context, p *player.Player, sk *skin.Skin) {
	if p == nil {
		return
	}
	fullID, playFabID, persona := protoSkinSummary(sk)
	skinEvent := &pb.PlayerSkinChangeEvent{
		PlayerUuid: p.UUID().String(),
		Name:       p.Name(),
		Persona:    persona,
	}
	if fullID != "" {
		skinEvent.FullId = &fullID
	}
	if playFabID != "" {
		skinEvent.PlayFabId = &playFabID
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_SKIN_CHANGE,
		Payload: &pb.EventEnvelope_PlayerSkinChange{
			PlayerSkinChange: skinEvent,
		},
	})
}

func (m *Manager) EmitPlayerFireExtinguish(ctx *player.Context, p *player.Player, pos cube.Pos) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_FIRE_EXTINGUISH,
		Payload: &pb.EventEnvelope_PlayerFireExtinguish{
			PlayerFireExtinguish: &pb.PlayerFireExtinguishEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoBlockPos(pos),
			},
		},
	})
}

func (m *Manager) EmitPlayerStartBreak(ctx *player.Context, p *player.Player, pos cube.Pos) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_START_BREAK,
		Payload: &pb.EventEnvelope_PlayerStartBreak{
			PlayerStartBreak: &pb.PlayerStartBreakEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoBlockPos(pos),
			},
		},
	})
}

func (m *Manager) EmitPlayerBlockPlace(ctx *player.Context, p *player.Player, pos cube.Pos, b world.Block) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_BLOCK_PLACE,
		Payload: &pb.EventEnvelope_PlayerBlockPlace{
			PlayerBlockPlace: &pb.PlayerBlockPlaceEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoBlockPos(pos),
				Block:      protoBlockState(b),
			},
		},
	})
}

func (m *Manager) EmitPlayerBlockPick(ctx *player.Context, p *player.Player, pos cube.Pos, b world.Block) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_BLOCK_PICK,
		Payload: &pb.EventEnvelope_PlayerBlockPick{
			PlayerBlockPick: &pb.PlayerBlockPickEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoBlockPos(pos),
				Block:      protoBlockState(b),
			},
		},
	})
}

func (m *Manager) EmitPlayerItemUse(ctx *player.Context, p *player.Player) {
	if p == nil {
		return
	}
	main, _ := p.HeldItems()
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ITEM_USE,
		Payload: &pb.EventEnvelope_PlayerItemUse{
			PlayerItemUse: &pb.PlayerItemUseEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Item:       protoItemStack(main),
			},
		},
	})
}

func (m *Manager) EmitPlayerItemUseOnBlock(ctx *player.Context, p *player.Player, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, b world.Block) {
	if p == nil {
		return
	}
	main, _ := p.HeldItems()
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ITEM_USE_ON_BLOCK,
		Payload: &pb.EventEnvelope_PlayerItemUseOnBlock{
			PlayerItemUseOnBlock: &pb.PlayerItemUseOnBlockEvent{
				PlayerUuid:    p.UUID().String(),
				Name:          p.Name(),
				World:         playerWorldDimension(p),
				Position:      protoBlockPos(pos),
				Face:          face.String(),
				ClickPosition: protoVec3(clickPos),
				Block:         protoBlockState(b),
				Item:          protoItemStack(main),
			},
		},
	})
}

func (m *Manager) EmitPlayerItemUseOnEntity(ctx *player.Context, p *player.Player, target world.Entity) {
	if p == nil {
		return
	}
	main, _ := p.HeldItems()
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ITEM_USE_ON_ENTITY,
		Payload: &pb.EventEnvelope_PlayerItemUseOnEntity{
			PlayerItemUseOnEntity: &pb.PlayerItemUseOnEntityEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Entity:     protoEntityRef(target),
				Item:       protoItemStack(main),
			},
		},
	})
}

func (m *Manager) EmitPlayerItemRelease(ctx *player.Context, p *player.Player, it item.Stack, dur time.Duration) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ITEM_RELEASE,
		Payload: &pb.EventEnvelope_PlayerItemRelease{
			PlayerItemRelease: &pb.PlayerItemReleaseEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Item:       protoItemStack(it),
				DurationMs: dur.Milliseconds(),
			},
		},
	})
}

func (m *Manager) EmitPlayerItemConsume(ctx *player.Context, p *player.Player, it item.Stack) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ITEM_CONSUME,
		Payload: &pb.EventEnvelope_PlayerItemConsume{
			PlayerItemConsume: &pb.PlayerItemConsumeEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Item:       protoItemStack(it),
			},
		},
	})
}

func (m *Manager) EmitPlayerAttackEntity(ctx *player.Context, p *player.Player, target world.Entity, force, height *float64, critical *bool) {
	if p == nil {
		return
	}
	main, _ := p.HeldItems()
	var forceVal, heightVal float64
	if force != nil {
		forceVal = *force
	}
	if height != nil {
		heightVal = *height
	}
	crit := false
	if critical != nil {
		crit = *critical
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ATTACK_ENTITY,
		Payload: &pb.EventEnvelope_PlayerAttackEntity{
			PlayerAttackEntity: &pb.PlayerAttackEntityEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Entity:     protoEntityRef(target),
				Force:      forceVal,
				Height:     heightVal,
				Critical:   crit,
				Item:       protoItemStack(main),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerAttackEntityMutation { return r.GetPlayerAttackEntity() },
		func(mut *pb.PlayerAttackEntityMutation) {
			mutateField(force, mut.Force)
			mutateField(height, mut.Height)
			mutateField(critical, mut.Critical)
		},
	)
}

func (m *Manager) EmitPlayerExperienceGain(ctx *player.Context, p *player.Player, amount *int) {
	if p == nil {
		return
	}
	amt := 0
	if amount != nil {
		amt = *amount
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_EXPERIENCE_GAIN,
		Payload: &pb.EventEnvelope_PlayerExperienceGain{
			PlayerExperienceGain: &pb.PlayerExperienceGainEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Amount:     int32(amt),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerExperienceGainMutation { return r.GetPlayerExperienceGain() },
		func(mut *pb.PlayerExperienceGainMutation) { mutateInt32(amount, mut.Amount) },
	)
}

func (m *Manager) EmitPlayerPunchAir(ctx *player.Context, p *player.Player) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_PUNCH_AIR,
		Payload: &pb.EventEnvelope_PlayerPunchAir{
			PlayerPunchAir: &pb.PlayerPunchAirEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
			},
		},
	})
}

func (m *Manager) EmitPlayerSignEdit(ctx *player.Context, p *player.Player, pos cube.Pos, frontSide bool, oldText, newText string) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_SIGN_EDIT,
		Payload: &pb.EventEnvelope_PlayerSignEdit{
			PlayerSignEdit: &pb.PlayerSignEditEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoBlockPos(pos),
				FrontSide:  frontSide,
				OldText:    oldText,
				NewText:    newText,
			},
		},
	})
}

func (m *Manager) EmitPlayerLecternPageTurn(ctx *player.Context, p *player.Player, pos cube.Pos, oldPage int, newPage *int) {
	if p == nil {
		return
	}
	newPageVal := oldPage
	if newPage != nil {
		newPageVal = *newPage
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_LECTERN_PAGE_TURN,
		Payload: &pb.EventEnvelope_PlayerLecternPageTurn{
			PlayerLecternPageTurn: &pb.PlayerLecternPageTurnEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Position:   protoBlockPos(pos),
				OldPage:    int32(oldPage),
				NewPage:    int32(newPageVal),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerLecternPageTurnMutation { return r.GetPlayerLecternPageTurn() },
		func(mut *pb.PlayerLecternPageTurnMutation) { mutateInt32(newPage, mut.NewPage) },
	)
}

func (m *Manager) EmitPlayerItemDamage(ctx *player.Context, p *player.Player, it item.Stack, damage int) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ITEM_DAMAGE,
		Payload: &pb.EventEnvelope_PlayerItemDamage{
			PlayerItemDamage: &pb.PlayerItemDamageEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Item:       protoItemStack(it),
				Damage:     int32(damage),
			},
		},
	})
}

func (m *Manager) EmitPlayerItemPickup(ctx *player.Context, p *player.Player, it *item.Stack) {
	if p == nil {
		return
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ITEM_PICKUP,
		Payload: &pb.EventEnvelope_PlayerItemPickup{
			PlayerItemPickup: &pb.PlayerItemPickupEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Item:       protoItemStackPtr(it),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerItemPickupMutation { return r.GetPlayerItemPickup() },
		func(mut *pb.PlayerItemPickupMutation) {
			if it == nil {
				return
			}
			if mut.Item == nil {
				*it = item.Stack{}
				return
			}
			if stack, ok := convertProtoItemStackValue(mut.Item); ok {
				*it = stack
			}
		},
	)
}

func (m *Manager) EmitPlayerHeldSlotChange(ctx *player.Context, p *player.Player, from, to int) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_HELD_SLOT_CHANGE,
		Payload: &pb.EventEnvelope_PlayerHeldSlotChange{
			PlayerHeldSlotChange: &pb.PlayerHeldSlotChangeEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				FromSlot:   int32(from),
				ToSlot:     int32(to),
			},
		},
	})
}

func (m *Manager) EmitPlayerItemDrop(ctx *player.Context, p *player.Player, it item.Stack) {
	if p == nil {
		return
	}
	m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_ITEM_DROP,
		Payload: &pb.EventEnvelope_PlayerItemDrop{
			PlayerItemDrop: &pb.PlayerItemDropEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				World:      playerWorldDimension(p),
				Item:       protoItemStack(it),
			},
		},
	})
}

func (m *Manager) EmitPlayerTransfer(ctx *player.Context, p *player.Player, addr *net.UDPAddr) {
	if p == nil {
		return
	}
	results := m.emitCancellable(ctx, &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_TRANSFER,
		Payload: &pb.EventEnvelope_PlayerTransfer{
			PlayerTransfer: &pb.PlayerTransferEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
				Address:    protoAddress(addr),
			},
		},
	})
	applyMutations(results,
		func(r *pb.EventResult) *pb.PlayerTransferMutation { return r.GetPlayerTransfer() },
		func(mut *pb.PlayerTransferMutation) {
			if addr == nil || mut.Address == nil {
				return
			}
			if newAddr := parseProtoAddress(mut.Address); newAddr != nil {
				*addr = *newAddr
			} else {
				*addr = net.UDPAddr{}
			}
		},
	)
}

func (m *Manager) EmitPlayerDiagnostics(p *player.Player, d session.Diagnostics) {
	if p == nil {
		return
	}
	evt := &pb.EventEnvelope{
		Type: pb.EventType_PLAYER_DIAGNOSTICS,
		Payload: &pb.EventEnvelope_PlayerDiagnostics{
			PlayerDiagnostics: &pb.PlayerDiagnosticsEvent{
				PlayerUuid: p.UUID().String(),
				Name:       p.Name(),
			},
		},
	}
	applyDiagnosticsFields(evt.GetPlayerDiagnostics(), d)
	m.broadcastEvent(evt)
}
