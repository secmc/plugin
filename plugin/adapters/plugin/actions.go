package plugin

import (
	"fmt"
	"strings"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/bossbar"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/dialogue"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/player/hud"
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	pb "github.com/secmc/plugin/proto/generated/go"
)

func (m *Manager) applyActions(p *pluginProcess, batch *pb.ActionBatch) {
	if batch == nil {
		return
	}

	// Group world set block actions by world.
	worldSetBlockActions := make(map[*world.World][]*pb.Action)
	otherActions := make([]*pb.Action, 0, len(batch.Actions))

	for _, action := range batch.Actions {
		if action == nil {
			continue
		}

		switch action.Kind.(type) {
		case *pb.Action_WorldSetBlock:
			kind := action.GetWorldSetBlock()
			w := m.worldFromRef(kind.GetWorld())
			if w == nil {
				m.sendActionError(p, action.GetCorrelationId(), "world not found")
				continue
			}
			worldSetBlockActions[w] = append(worldSetBlockActions[w], action)
		default:
			otherActions = append(otherActions, action)
		}
	}

	// Process batched world set block actions.
	for w, actions := range worldSetBlockActions {
		m.handleWorldSetBlockBatch(p, w, actions)
	}

	// Process other actions individually.
	for _, action := range otherActions {
		m.handleSingleAction(p, action)
	}
}

func (m *Manager) handleSingleAction(p *pluginProcess, action *pb.Action) {
	correlationID := action.GetCorrelationId()
	switch kind := action.Kind.(type) {
	case *pb.Action_SendChat:
		m.handleSendChat(kind.SendChat)
	case *pb.Action_Teleport:
		m.handleTeleport(kind.Teleport)
	case *pb.Action_Kick:
		m.handleKick(kind.Kick)
	case *pb.Action_SetGameMode:
		m.handleSetGameMode(kind.SetGameMode)
	case *pb.Action_GiveItem:
		m.handleGiveItem(kind.GiveItem)
	case *pb.Action_ClearInventory:
		m.handleClearInventory(kind.ClearInventory)
	case *pb.Action_SetHeldItem:
		m.handleSetHeldItem(kind.SetHeldItem)
	case *pb.Action_PlayerSetArmour:
		m.handlePlayerSetArmour(kind.PlayerSetArmour)
	case *pb.Action_SetHealth:
		m.handleSetHealth(kind.SetHealth)
	case *pb.Action_SetFood:
		m.handleSetFood(kind.SetFood)
	case *pb.Action_SetExperience:
		m.handleSetExperience(kind.SetExperience)
	case *pb.Action_SetVelocity:
		m.handleSetVelocity(kind.SetVelocity)
	case *pb.Action_AddEffect:
		m.handleAddEffect(kind.AddEffect)
	case *pb.Action_RemoveEffect:
		m.handleRemoveEffect(kind.RemoveEffect)
	case *pb.Action_SendTitle:
		m.handleSendTitle(kind.SendTitle)
	case *pb.Action_SendPopup:
		m.handleSendPopup(kind.SendPopup)
	case *pb.Action_SendTip:
		m.handleSendTip(kind.SendTip)
	case *pb.Action_PlaySound:
		m.handlePlaySound(kind.PlaySound)
	case *pb.Action_ExecuteCommand:
		m.handleExecuteCommand(kind.ExecuteCommand)
	case *pb.Action_WorldSetDefaultGameMode:
		m.handleWorldSetDefaultGameMode(p, correlationID, kind.WorldSetDefaultGameMode)
	case *pb.Action_WorldSetDifficulty:
		m.handleWorldSetDifficulty(p, correlationID, kind.WorldSetDifficulty)
	case *pb.Action_WorldSetTickRange:
		m.handleWorldSetTickRange(p, correlationID, kind.WorldSetTickRange)
	case *pb.Action_WorldPlaySound:
		m.handleWorldPlaySound(p, correlationID, kind.WorldPlaySound)
	case *pb.Action_WorldAddParticle:
		m.handleWorldAddParticle(p, correlationID, kind.WorldAddParticle)
	case *pb.Action_WorldSetTime:
		m.handleWorldSetTime(p, correlationID, kind.WorldSetTime)
	case *pb.Action_WorldStopTime:
		m.handleWorldStopTime(p, correlationID, kind.WorldStopTime)
	case *pb.Action_WorldStartTime:
		m.handleWorldStartTime(p, correlationID, kind.WorldStartTime)
	case *pb.Action_WorldSetSpawn:
		m.handleWorldSetSpawn(p, correlationID, kind.WorldSetSpawn)
	case *pb.Action_WorldQueryEntities:
		m.handleWorldQueryEntities(p, correlationID, kind.WorldQueryEntities)
	case *pb.Action_WorldQueryPlayers:
		m.handleWorldQueryPlayers(p, correlationID, kind.WorldQueryPlayers)
	case *pb.Action_WorldQueryEntitiesWithin:
		m.handleWorldQueryEntitiesWithin(p, correlationID, kind.WorldQueryEntitiesWithin)
	case *pb.Action_WorldQueryDefaultGameMode:
		m.handleWorldQueryDefaultGameMode(p, correlationID, kind.WorldQueryDefaultGameMode)
	case *pb.Action_WorldQueryPlayerSpawn:
		m.handleWorldQueryPlayerSpawn(p, correlationID, kind.WorldQueryPlayerSpawn)
	case *pb.Action_WorldQueryBlock:
		m.handleWorldQueryBlock(p, correlationID, kind.WorldQueryBlock)
	case *pb.Action_WorldQueryBiome:
		m.handleWorldQueryBiome(p, correlationID, kind.WorldQueryBiome)
	case *pb.Action_WorldQueryLight:
		m.handleWorldQueryLight(p, correlationID, kind.WorldQueryLight)
	case *pb.Action_WorldQuerySkyLight:
		m.handleWorldQuerySkyLight(p, correlationID, kind.WorldQuerySkyLight)
	case *pb.Action_WorldQueryTemperature:
		m.handleWorldQueryTemperature(p, correlationID, kind.WorldQueryTemperature)
	case *pb.Action_WorldQueryHighestBlock:
		m.handleWorldQueryHighestBlock(p, correlationID, kind.WorldQueryHighestBlock)
	case *pb.Action_WorldQueryRainingAt:
		m.handleWorldQueryRainingAt(p, correlationID, kind.WorldQueryRainingAt)
	case *pb.Action_WorldQuerySnowingAt:
		m.handleWorldQuerySnowingAt(p, correlationID, kind.WorldQuerySnowingAt)
	case *pb.Action_WorldQueryThunderingAt:
		m.handleWorldQueryThunderingAt(p, correlationID, kind.WorldQueryThunderingAt)
	case *pb.Action_WorldQueryLiquid:
		m.handleWorldQueryLiquid(p, correlationID, kind.WorldQueryLiquid)
	case *pb.Action_WorldSetBiome:
		m.handleWorldSetBiome(p, correlationID, kind.WorldSetBiome)
	case *pb.Action_WorldSetLiquid:
		m.handleWorldSetLiquid(p, correlationID, kind.WorldSetLiquid)
	case *pb.Action_WorldScheduleBlockUpdate:
		m.handleWorldScheduleBlockUpdate(p, correlationID, kind.WorldScheduleBlockUpdate)
	case *pb.Action_WorldBuildStructure:
		m.handleWorldBuildStructure(p, correlationID, kind.WorldBuildStructure)
	case *pb.Action_PlayerStartSprinting:
		m.handlePlayerStartSprinting(kind.PlayerStartSprinting)
	case *pb.Action_PlayerStopSprinting:
		m.handlePlayerStopSprinting(kind.PlayerStopSprinting)
	case *pb.Action_PlayerStartSneaking:
		m.handlePlayerStartSneaking(kind.PlayerStartSneaking)
	case *pb.Action_PlayerStopSneaking:
		m.handlePlayerStopSneaking(kind.PlayerStopSneaking)
	case *pb.Action_PlayerStartSwimming:
		m.handlePlayerStartSwimming(kind.PlayerStartSwimming)
	case *pb.Action_PlayerStopSwimming:
		m.handlePlayerStopSwimming(kind.PlayerStopSwimming)
	case *pb.Action_PlayerStartCrawling:
		m.handlePlayerStartCrawling(kind.PlayerStartCrawling)
	case *pb.Action_PlayerStopCrawling:
		m.handlePlayerStopCrawling(kind.PlayerStopCrawling)
	case *pb.Action_PlayerStartGliding:
		m.handlePlayerStartGliding(kind.PlayerStartGliding)
	case *pb.Action_PlayerStopGliding:
		m.handlePlayerStopGliding(kind.PlayerStopGliding)
	case *pb.Action_PlayerStartFlying:
		m.handlePlayerStartFlying(kind.PlayerStartFlying)
	case *pb.Action_PlayerStopFlying:
		m.handlePlayerStopFlying(kind.PlayerStopFlying)
	case *pb.Action_PlayerSetImmobile:
		m.handlePlayerSetImmobile(kind.PlayerSetImmobile)
	case *pb.Action_PlayerSetMobile:
		m.handlePlayerSetMobile(kind.PlayerSetMobile)
	case *pb.Action_PlayerSetSpeed:
		m.handlePlayerSetSpeed(kind.PlayerSetSpeed)
	case *pb.Action_PlayerSetFlightSpeed:
		m.handlePlayerSetFlightSpeed(kind.PlayerSetFlightSpeed)
	case *pb.Action_PlayerSetVerticalFlightSpeed:
		m.handlePlayerSetVerticalFlightSpeed(kind.PlayerSetVerticalFlightSpeed)
	case *pb.Action_PlayerSetAbsorption:
		m.handlePlayerSetAbsorption(kind.PlayerSetAbsorption)
	case *pb.Action_PlayerSetOnFire:
		m.handlePlayerSetOnFire(kind.PlayerSetOnFire)
	case *pb.Action_PlayerExtinguish:
		m.handlePlayerExtinguish(kind.PlayerExtinguish)
	case *pb.Action_PlayerSetInvisible:
		m.handlePlayerSetInvisible(kind.PlayerSetInvisible)
	case *pb.Action_PlayerSetVisible:
		m.handlePlayerSetVisible(kind.PlayerSetVisible)
	case *pb.Action_PlayerSetScale:
		m.handlePlayerSetScale(kind.PlayerSetScale)
	case *pb.Action_PlayerSetHeldSlot:
		m.handlePlayerSetHeldSlot(kind.PlayerSetHeldSlot)
	case *pb.Action_PlayerSendToast:
		m.handlePlayerSendToast(kind.PlayerSendToast)
	case *pb.Action_PlayerSendJukeboxPopup:
		m.handlePlayerSendJukeboxPopup(kind.PlayerSendJukeboxPopup)
	case *pb.Action_PlayerShowCoordinates:
		m.handlePlayerShowCoordinates(kind.PlayerShowCoordinates)
	case *pb.Action_PlayerHideCoordinates:
		m.handlePlayerHideCoordinates(kind.PlayerHideCoordinates)
	case *pb.Action_PlayerEnableInstantRespawn:
		m.handlePlayerEnableInstantRespawn(kind.PlayerEnableInstantRespawn)
	case *pb.Action_PlayerDisableInstantRespawn:
		m.handlePlayerDisableInstantRespawn(kind.PlayerDisableInstantRespawn)
	case *pb.Action_PlayerSetNameTag:
		m.handlePlayerSetNameTag(kind.PlayerSetNameTag)
	case *pb.Action_PlayerSetScoreTag:
		m.handlePlayerSetScoreTag(kind.PlayerSetScoreTag)
	case *pb.Action_PlayerShowParticle:
		m.handlePlayerShowParticle(kind.PlayerShowParticle)
	case *pb.Action_PlayerSendScoreboard:
		m.handlePlayerSendScoreboard(kind.PlayerSendScoreboard)
	case *pb.Action_PlayerRemoveScoreboard:
		m.handlePlayerRemoveScoreboard(kind.PlayerRemoveScoreboard)
	case *pb.Action_PlayerSendMenuForm:
		m.handlePlayerSendMenuForm(kind.PlayerSendMenuForm)
	case *pb.Action_PlayerSendModalForm:
		m.handlePlayerSendModalForm(kind.PlayerSendModalForm)
	case *pb.Action_PlayerSendDialogue:
		m.handlePlayerSendDialogue(p, correlationID, kind.PlayerSendDialogue)
	case *pb.Action_PlayerRespawn:
		m.handlePlayerRespawn(kind.PlayerRespawn)
	case *pb.Action_PlayerTransfer:
		m.handlePlayerTransferAction(kind.PlayerTransfer)
	case *pb.Action_PlayerKnockBack:
		m.handlePlayerKnockBack(kind.PlayerKnockBack)
	case *pb.Action_PlayerSwingArm:
		m.handlePlayerSwingArm(kind.PlayerSwingArm)
	case *pb.Action_PlayerPunchAir:
		m.handlePlayerPunchAirAction(kind.PlayerPunchAir)
	case *pb.Action_PlayerSendBossBar:
		m.handlePlayerSendBossBar(kind.PlayerSendBossBar)
	case *pb.Action_PlayerRemoveBossBar:
		m.handlePlayerRemoveBossBar(kind.PlayerRemoveBossBar)
	case *pb.Action_PlayerShowHudElement:
		m.handlePlayerShowHudElement(kind.PlayerShowHudElement)
	case *pb.Action_PlayerHideHudElement:
		m.handlePlayerHideHudElement(kind.PlayerHideHudElement)
	case *pb.Action_PlayerCloseDialogue:
		m.handlePlayerCloseDialogue(kind.PlayerCloseDialogue)
	case *pb.Action_PlayerCloseForm:
		m.handlePlayerCloseForm(kind.PlayerCloseForm)
	case *pb.Action_PlayerOpenSign:
		m.handlePlayerOpenSign(kind.PlayerOpenSign)
	case *pb.Action_PlayerEditSign:
		m.handlePlayerEditSign(kind.PlayerEditSign)
	case *pb.Action_PlayerTurnLecternPage:
		m.handlePlayerTurnLecternPage(kind.PlayerTurnLecternPage)
	case *pb.Action_PlayerHidePlayer:
		m.handlePlayerHidePlayer(kind.PlayerHidePlayer)
	case *pb.Action_PlayerShowPlayer:
		m.handlePlayerShowPlayer(kind.PlayerShowPlayer)
	case *pb.Action_PlayerRemoveAllDebugShapes:
		m.handlePlayerRemoveAllDebugShapes(kind.PlayerRemoveAllDebugShapes)
	case *pb.Action_PlayerOpenBlockContainer:
		m.handlePlayerOpenBlockContainer(kind.PlayerOpenBlockContainer)
	case *pb.Action_PlayerDropItem:
		m.handlePlayerDropItem(kind.PlayerDropItem)
	case *pb.Action_PlayerSetItemCooldown:
		m.handlePlayerSetItemCooldown(kind.PlayerSetItemCooldown)
	}
}

func (m *Manager) handleWorldSetBlockBatch(p *pluginProcess, w *world.World, actions []*pb.Action) {
	correlationIDs := make([]string, 0, len(actions))
	<-w.Exec(func(tx *world.Tx) {
		for _, action := range actions {
			act := action.GetWorldSetBlock()
			correlationID := action.GetCorrelationId()
			if act.Position == nil {
				if correlationID != "" {
					m.sendActionError(p, correlationID, "missing position")
				}
				continue
			}
			pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
			var blk world.Block
			var ok bool
			if act.Block != nil {
				blk, ok = blockFromProto(act.Block)
				if !ok {
					if correlationID != "" {
						m.sendActionError(p, correlationID, "unknown block")
					}
					continue
				}
			}
			tx.SetBlock(pos, blk, nil)
			if correlationID != "" {
				correlationIDs = append(correlationIDs, correlationID)
			}
		}
	})

	for _, correlationID := range correlationIDs {
		m.sendActionOK(p, correlationID)
	}
}

func (m *Manager) handleSendChat(act *pb.SendChatAction) {
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

	m.execMethod(id, func(pl *player.Player) {
		pl.Message(act.Message)
	})
}

func (m *Manager) handleTeleport(act *pb.TeleportAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}

	m.execMethod(id, func(pl *player.Player) {
		pos, ok := vec3FromProto(act.Position)
		if ok {
			pl.Teleport(pos)
		}
		rot, ok := vec3FromProto(act.Rotation)
		if ok {
			playerRot := pl.Rotation()
			deltaYaw := rot[1] - playerRot.Yaw()
			deltaPitch := rot[0] - playerRot.Pitch()
			if deltaYaw != 0 || deltaPitch != 0 {
				pl.Move(mgl64.Vec3{}, deltaYaw, deltaPitch)
			}
		}
	})
}

func (m *Manager) handleKick(act *pb.KickAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		pl.Disconnect(act.Reason)
	})
}

func (m *Manager) handleSetGameMode(act *pb.SetGameModeAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	gameMode, ok := world.GameModeByID(int(act.GameMode))
	if !ok {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		pl.SetGameMode(gameMode)
	})
}

func (m *Manager) handleGiveItem(act *pb.GiveItemAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		if stack, ok := convertProtoItemStackValue(act.Item); ok {
			_, _ = pl.Inventory().AddItem(stack)
		}
	})
}

func (m *Manager) handleClearInventory(act *pb.ClearInventoryAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		_ = pl.Inventory().Clear()
		pl.SetHeldItems(item.Stack{}, item.Stack{})
	})
}

func (m *Manager) handleSetHeldItem(act *pb.SetHeldItemAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		main, off := pl.HeldItems()

		if act.Main != nil {
			if s, ok := convertProtoItemStackValue(act.Main); ok {
				main = s
			}
		}
		if act.Offhand != nil {
			if s, ok := convertProtoItemStackValue(act.Offhand); ok {
				off = s
			}
		}
		pl.SetHeldItems(main, off)
	})
}

func (m *Manager) handleSetHealth(act *pb.SetHealthAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		if act.MaxHealth != nil {
			pl.SetMaxHealth(*act.MaxHealth)
		}
		current := pl.Health()
		target := act.Health
		if target > current {
			pl.Heal(target-current, entity.FoodHealingSource{})
		} else if target < current {
			pl.Hurt(current-target, entity.VoidDamageSource{})
		}
	})
}

func (m *Manager) handleSetFood(act *pb.SetFoodAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		pl.SetFood(int(act.Food))
	})
}

func (m *Manager) handleSetExperience(act *pb.SetExperienceAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		if act.Level != nil {
			pl.SetExperienceLevel(int(*act.Level))
		}
		if act.Progress != nil {
			pl.SetExperienceProgress(float64(*act.Progress))
		}
		if act.Amount != nil {
			amt := int(*act.Amount)
			if amt >= 0 {
				_ = pl.AddExperience(amt)
			} else {
				pl.RemoveExperience(-amt)
			}
		}
	})
}

func (m *Manager) handleSetVelocity(act *pb.SetVelocityAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		if v, ok := vec3FromProto(act.Velocity); ok {
			pl.SetVelocity(v)
		}
	})
}

func (m *Manager) handleAddEffect(act *pb.AddEffectAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		t, ok := effect.ByID(int(act.EffectType))
		if !ok {
			return
		}
		var e effect.Effect
		if lt, ok := t.(effect.LastingType); ok {
			d := time.Duration(act.DurationMs) * time.Millisecond
			if d <= 0 {
				e = effect.NewInfinite(lt, int(act.Level))
			} else {
				e = effect.New(lt, int(act.Level), d)
			}
		} else {
			e = effect.NewInstantWithPotency(t, int(act.Level), 1.0)
		}
		if !act.ShowParticles {
			e = e.WithoutParticles()
		}
		pl.AddEffect(e)
	})
}

func (m *Manager) handleRemoveEffect(act *pb.RemoveEffectAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		t, ok := effect.ByID(int(act.EffectType))
		if !ok {
			return
		}
		pl.RemoveEffect(t)
	})
}

func (m *Manager) handleSendTitle(act *pb.SendTitleAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		t := playerTitleFromAction(act)
		pl.SendTitle(t)
	})
}

func (m *Manager) handleSendPopup(act *pb.SendPopupAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		pl.SendPopup(act.Message)
	})
}

func (m *Manager) handleSendTip(act *pb.SendTipAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		pl.SendTip(act.Message)
	})
}

func (m *Manager) handlePlaySound(act *pb.PlaySoundAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		s := soundFromProto(act.Sound)
		pl.PlaySound(s)
	})
}

func (m *Manager) handleExecuteCommand(act *pb.ExecuteCommandAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		cmd := act.Command
		if cmd != "" && !strings.HasPrefix(cmd, "/") {
			cmd = "/" + cmd
		}
		pl.ExecuteCommand(cmd)
	})
}

func (m *Manager) handleWorldSetDefaultGameMode(p *pluginProcess, correlationID string, act *pb.WorldSetDefaultGameModeAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	mode, ok := world.GameModeByID(int(act.GameMode))
	if !ok {
		m.sendActionError(p, correlationID, "unknown game mode")
		return
	}
	w.SetDefaultGameMode(mode)
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldBuildStructure(p *pluginProcess, correlationID string, act *pb.WorldBuildStructureAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Origin == nil {
		m.sendActionError(p, correlationID, "missing origin")
		return
	}
	if act.Structure == nil {
		m.sendActionError(p, correlationID, "missing structure")
		return
	}
	ps, err := buildProtoStructure(act.Structure)
	if err != nil {
		m.sendActionError(p, correlationID, err.Error())
		return
	}
	origin := cube.Pos{int(act.Origin.X), int(act.Origin.Y), int(act.Origin.Z)}
	<-w.Exec(func(tx *world.Tx) {
		tx.BuildStructure(origin, ps)
	})
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldSetDifficulty(p *pluginProcess, correlationID string, act *pb.WorldSetDifficultyAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	diff, ok := world.DifficultyByID(int(act.Difficulty))
	if !ok {
		m.sendActionError(p, correlationID, "unknown difficulty")
		return
	}
	w.SetDifficulty(diff)
	m.sendActionOK(p, correlationID)
}

// Player movement toggles
func (m *Manager) handlePlayerStartSprinting(act *pb.PlayerStartSprintingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StartSprinting() })
}
func (m *Manager) handlePlayerStopSprinting(act *pb.PlayerStopSprintingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StopSprinting() })
}
func (m *Manager) handlePlayerStartSneaking(act *pb.PlayerStartSneakingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StartSneaking() })
}
func (m *Manager) handlePlayerStopSneaking(act *pb.PlayerStopSneakingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StopSneaking() })
}
func (m *Manager) handlePlayerStartSwimming(act *pb.PlayerStartSwimmingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StartSwimming() })
}
func (m *Manager) handlePlayerStopSwimming(act *pb.PlayerStopSwimmingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StopSwimming() })
}
func (m *Manager) handlePlayerStartCrawling(act *pb.PlayerStartCrawlingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StartCrawling() })
}
func (m *Manager) handlePlayerStopCrawling(act *pb.PlayerStopCrawlingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StopCrawling() })
}
func (m *Manager) handlePlayerStartGliding(act *pb.PlayerStartGlidingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StartGliding() })
}
func (m *Manager) handlePlayerStopGliding(act *pb.PlayerStopGlidingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StopGliding() })
}
func (m *Manager) handlePlayerStartFlying(act *pb.PlayerStartFlyingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StartFlying() })
}
func (m *Manager) handlePlayerStopFlying(act *pb.PlayerStopFlyingAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.StopFlying() })
}

// Player mobility lock
func (m *Manager) handlePlayerSetImmobile(act *pb.PlayerSetImmobileAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetImmobile() })
}
func (m *Manager) handlePlayerSetMobile(act *pb.PlayerSetMobileAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetMobile() })
}

// Player movement attributes
func (m *Manager) handlePlayerSetSpeed(act *pb.PlayerSetSpeedAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetSpeed(act.Speed) })
}
func (m *Manager) handlePlayerSetFlightSpeed(act *pb.PlayerSetFlightSpeedAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetFlightSpeed(act.FlightSpeed) })
}
func (m *Manager) handlePlayerSetVerticalFlightSpeed(act *pb.PlayerSetVerticalFlightSpeedAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetVerticalFlightSpeed(act.VerticalFlightSpeed) })
}

// Player health/status
func (m *Manager) handlePlayerSetAbsorption(act *pb.PlayerSetAbsorptionAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetAbsorption(act.Absorption) })
}
func (m *Manager) handlePlayerSetOnFire(act *pb.PlayerSetOnFireAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	d := time.Duration(act.DurationMs) * time.Millisecond
	m.execMethod(id, func(pl *player.Player) { pl.SetOnFire(d) })
}
func (m *Manager) handlePlayerExtinguish(act *pb.PlayerExtinguishAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.Extinguish() })
}
func (m *Manager) handlePlayerSetInvisible(act *pb.PlayerSetInvisibleAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetInvisible() })
}
func (m *Manager) handlePlayerSetVisible(act *pb.PlayerSetVisibleAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetVisible() })
}

// Player misc attributes
func (m *Manager) handlePlayerSetScale(act *pb.PlayerSetScaleAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SetScale(act.Scale) })
}
func (m *Manager) handlePlayerSetHeldSlot(act *pb.PlayerSetHeldSlotAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	slot := int(act.Slot)
	m.execMethod(id, func(pl *player.Player) { _ = pl.SetHeldSlot(slot) })
}

// Player UI
func (m *Manager) handlePlayerSendToast(act *pb.PlayerSendToastAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	titleText := act.Title
	message := act.Message
	m.execMethod(id, func(pl *player.Player) { pl.SendToast(titleText, message) })
}
func (m *Manager) handlePlayerSendJukeboxPopup(act *pb.PlayerSendJukeboxPopupAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	msg := act.Message
	m.execMethod(id, func(pl *player.Player) { pl.SendJukeboxPopup(msg) })
}
func (m *Manager) handlePlayerShowCoordinates(act *pb.PlayerShowCoordinatesAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.ShowCoordinates() })
}
func (m *Manager) handlePlayerHideCoordinates(act *pb.PlayerHideCoordinatesAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.HideCoordinates() })
}
func (m *Manager) handlePlayerEnableInstantRespawn(act *pb.PlayerEnableInstantRespawnAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.EnableInstantRespawn() })
}
func (m *Manager) handlePlayerDisableInstantRespawn(act *pb.PlayerDisableInstantRespawnAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.DisableInstantRespawn() })
}
func (m *Manager) handlePlayerSetNameTag(act *pb.PlayerSetNameTagAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	name := act.NameTag
	m.execMethod(id, func(pl *player.Player) { pl.SetNameTag(name) })
}
func (m *Manager) handlePlayerSetScoreTag(act *pb.PlayerSetScoreTagAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	text := act.ScoreTag
	m.execMethod(id, func(pl *player.Player) { pl.SetScoreTag(text) })
}

// Player visuals
func (m *Manager) handlePlayerShowParticle(act *pb.PlayerShowParticleAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	pos, ok := vec3FromProto(act.Position)
	if !ok {
		return
	}
	part, ok := particleFromPlayerAction(act)
	if !ok {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.ShowParticle(pos, part) })
}

// Player lifecycle/control
func (m *Manager) handlePlayerRespawn(act *pb.PlayerRespawnAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { _ = pl.Respawn() })
}
func (m *Manager) handlePlayerTransferAction(act *pb.PlayerTransferAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	addr := parseProtoAddress(act.Address)
	if addr == nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { _ = pl.Transfer(addr.String()) })
}
func (m *Manager) handlePlayerKnockBack(act *pb.PlayerKnockBackAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	src, ok := vec3FromProto(act.Source)
	if !ok {
		return
	}
	force := act.Force
	height := act.Height
	m.execMethod(id, func(pl *player.Player) { pl.KnockBack(src, force, height) })
}
func (m *Manager) handlePlayerSwingArm(act *pb.PlayerSwingArmAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.SwingArm() })
}
func (m *Manager) handlePlayerPunchAirAction(act *pb.PlayerPunchAirAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.PunchAir() })
}

// Player boss bar
func (m *Manager) handlePlayerSendBossBar(act *pb.PlayerSendBossBarAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		bar := bossbar.New(act.Text)
		if act.HealthPercentage != nil {
			h := float64(*act.HealthPercentage)
			if h < 0 {
				h = 0
			}
			if h > 1 {
				h = 1
			}
			bar = bar.WithHealthPercentage(h)
		}
		if act.Colour != nil {
			bar = bar.WithColour(convertBossBarColour(*act.Colour))
		}
		pl.SendBossBar(bar)
	})
}

func (m *Manager) handlePlayerRemoveBossBar(act *pb.PlayerRemoveBossBarAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.RemoveBossBar() })
}

// Player HUD
func (m *Manager) handlePlayerShowHudElement(act *pb.PlayerShowHudElementAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		if el, ok := convertHudElement(act.Element); ok {
			pl.ShowHudElement(el)
		}
	})
}

func (m *Manager) handlePlayerHideHudElement(act *pb.PlayerHideHudElementAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		if el, ok := convertHudElement(act.Element); ok {
			pl.HideHudElement(el)
		}
	})
}

// UI closers
func (m *Manager) handlePlayerCloseDialogue(act *pb.PlayerCloseDialogueAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.CloseDialogue() })
}

func (m *Manager) handlePlayerCloseForm(act *pb.PlayerCloseFormAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.CloseForm() })
}

// Signs & Lecterns
func (m *Manager) handlePlayerOpenSign(act *pb.PlayerOpenSignAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	if act.Position == nil {
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	m.execMethod(id, func(pl *player.Player) { pl.OpenSign(pos, act.FrontSide) })
}

func (m *Manager) handlePlayerEditSign(act *pb.PlayerEditSignAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	if act.Position == nil {
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	m.execMethod(id, func(pl *player.Player) { _ = pl.EditSign(pos, act.FrontText, act.BackText) })
}

func (m *Manager) handlePlayerTurnLecternPage(act *pb.PlayerTurnLecternPageAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	if act.Position == nil {
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	page := int(act.Page)
	m.execMethod(id, func(pl *player.Player) { _ = pl.TurnLecternPage(pos, page) })
}

// Entity visibility (players)
func (m *Manager) handlePlayerHidePlayer(act *pb.PlayerHidePlayerAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	targetID, err := uuid.Parse(act.TargetUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		for p := range m.srv.Players(nil) {
			if p.UUID() == targetID {
				pl.HideEntity(p)
				break
			}
		}
	})
}

func (m *Manager) handlePlayerShowPlayer(act *pb.PlayerShowPlayerAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	targetID, err := uuid.Parse(act.TargetUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		for p := range m.srv.Players(nil) {
			if p.UUID() == targetID {
				pl.ShowEntity(p)
				break
			}
		}
	})
}

// Debug shapes
func (m *Manager) handlePlayerRemoveAllDebugShapes(act *pb.PlayerRemoveAllDebugShapesAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) { pl.RemoveAllDebugShapes() })
}

// Interaction extras
func (m *Manager) handlePlayerOpenBlockContainer(act *pb.PlayerOpenBlockContainerAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	if act.Position == nil {
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	m.execMethod(id, func(pl *player.Player) { pl.OpenBlockContainer(pos, pl.Tx()) })
}

func (m *Manager) handlePlayerDropItem(act *pb.PlayerDropItemAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		var s item.Stack
		if act.Item != nil {
			if stack, ok := convertProtoItemStackValue(act.Item); ok {
				s = stack
			} else {
				return
			}
		} else {
			held, _ := pl.HeldItems()
			if held.Empty() {
				return
			}
			s = held
		}
		_ = pl.Drop(s)
	})
}

func (m *Manager) handlePlayerSetItemCooldown(act *pb.PlayerSetItemCooldownAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	if act.Item == nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		if stack, ok := convertProtoItemStackValue(act.Item); ok {
			d := time.Duration(act.DurationMs) * time.Millisecond
			pl.SetCooldown(stack.Item(), d)
		}
	})
}

// Converters
func convertBossBarColour(c pb.BossBarColour) bossbar.Colour {
	switch c {
	case pb.BossBarColour_BOSS_BAR_COLOUR_GREY:
		return bossbar.Grey()
	case pb.BossBarColour_BOSS_BAR_COLOUR_BLUE:
		return bossbar.Blue()
	case pb.BossBarColour_BOSS_BAR_COLOUR_RED:
		return bossbar.Red()
	case pb.BossBarColour_BOSS_BAR_COLOUR_GREEN:
		return bossbar.Green()
	case pb.BossBarColour_BOSS_BAR_COLOUR_YELLOW:
		return bossbar.Yellow()
	case pb.BossBarColour_BOSS_BAR_COLOUR_PURPLE:
		return bossbar.Purple()
	case pb.BossBarColour_BOSS_BAR_COLOUR_WHITE:
		return bossbar.White()
	default:
		return bossbar.Purple()
	}
}

func convertHudElement(e pb.HudElement) (hud.Element, bool) {
	switch e {
	case pb.HudElement_HUD_ELEMENT_PAPER_DOLL:
		return hud.PaperDoll(), true
	case pb.HudElement_HUD_ELEMENT_ARMOUR:
		return hud.Armour(), true
	case pb.HudElement_HUD_ELEMENT_TOOL_TIPS:
		return hud.ToolTips(), true
	case pb.HudElement_HUD_ELEMENT_TOUCH_CONTROLS:
		return hud.TouchControls(), true
	case pb.HudElement_HUD_ELEMENT_CROSSHAIR:
		return hud.Crosshair(), true
	case pb.HudElement_HUD_ELEMENT_HOT_BAR:
		return hud.HotBar(), true
	case pb.HudElement_HUD_ELEMENT_HEALTH:
		return hud.Health(), true
	case pb.HudElement_HUD_ELEMENT_PROGRESS_BAR:
		return hud.ProgressBar(), true
	case pb.HudElement_HUD_ELEMENT_HUNGER:
		return hud.Hunger(), true
	case pb.HudElement_HUD_ELEMENT_AIR_BUBBLES:
		return hud.AirBubbles(), true
	case pb.HudElement_HUD_ELEMENT_HORSE_HEALTH:
		return hud.HorseHealth(), true
	case pb.HudElement_HUD_ELEMENT_STATUS_EFFECTS:
		return hud.StatusEffects(), true
	case pb.HudElement_HUD_ELEMENT_ITEM_TEXT:
		return hud.ItemText(), true
	default:
		return hud.Element{}, false
	}
}

// Local no-op submittables for forms/dialogues.
type formMenuNoop struct{}

func (formMenuNoop) Submit(form.Submitter, form.Button, *world.Tx) {}

type formModalNoop struct {
	Yes form.Button
	No  form.Button
}

func (formModalNoop) Submit(form.Submitter, form.Button, *world.Tx) {}

type dialogueNoop struct{}

func (dialogueNoop) Submit(dialogue.Submitter, dialogue.Button, *world.Tx) {}

func resolveWorldEntity(pl *player.Player, ref *pb.EntityRef) world.Entity {
	if ref == nil || ref.Uuid == "" {
		return nil
	}
	if targetUUID, err := uuid.Parse(ref.Uuid); err == nil {
		for ent := range pl.Tx().Entities() { // TODO: optimize, direct lookup with df private handles map
			if ent.H().UUID() == targetUUID {
				return ent
			}
		}
	}
	return nil
}

// Player armour
func (m *Manager) handlePlayerSetArmour(act *pb.PlayerSetArmourAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		if act.Helmet != nil {
			if s, ok := convertProtoItemStackValue(act.Helmet); ok {
				pl.Armour().SetHelmet(s)
			} else {
				pl.Armour().SetHelmet(item.Stack{})
			}
		}
		if act.Chestplate != nil {
			if s, ok := convertProtoItemStackValue(act.Chestplate); ok {
				pl.Armour().SetChestplate(s)
			} else {
				pl.Armour().SetChestplate(item.Stack{})
			}
		}
		if act.Leggings != nil {
			if s, ok := convertProtoItemStackValue(act.Leggings); ok {
				pl.Armour().SetLeggings(s)
			} else {
				pl.Armour().SetLeggings(item.Stack{})
			}
		}
		if act.Boots != nil {
			if s, ok := convertProtoItemStackValue(act.Boots); ok {
				pl.Armour().SetBoots(s)
			} else {
				pl.Armour().SetBoots(item.Stack{})
			}
		}
	})
}

// Player scoreboard
func (m *Manager) handlePlayerSendScoreboard(act *pb.PlayerSendScoreboardAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		sb := scoreboard.New(act.Title)
		if act.Padding != nil && !*act.Padding {
			sb.RemovePadding()
		}
		if act.Descending != nil && *act.Descending {
			sb.SetDescending()
		}
		// Clamp to 15 lines as per Dragonfly's limit and set them deterministically without trailing newlines.
		max := min(len(act.Lines), 15)
		for i := range max {
			sb.Set(i, act.Lines[i])
		}
		pl.SendScoreboard(sb)
	})
}

func (m *Manager) handlePlayerRemoveScoreboard(act *pb.PlayerRemoveScoreboardAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		pl.RemoveScoreboard()
	})
}

// Player forms (show)
func (m *Manager) handlePlayerSendMenuForm(act *pb.PlayerSendMenuFormAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		sub := formMenuNoop{}
		menu := form.NewMenu(sub, act.Title)
		if act.Body != nil {
			menu = menu.WithBody(*act.Body)
		}
		btns := make([]form.Button, len(act.Buttons))
		for i := range act.Buttons {
			btns[i] = form.NewButton(act.Buttons[i], "")
		}
		menu = menu.WithButtons(btns...)
		pl.SendForm(menu)
	})
}

func (m *Manager) handlePlayerSendModalForm(act *pb.PlayerSendModalFormAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		sub := formModalNoop{Yes: form.NewButton(act.YesText, ""), No: form.NewButton(act.NoText, "")}
		modal := form.NewModal(sub, act.Title).WithBody(act.Body)
		pl.SendForm(modal)
	})
}

// Player dialogue (show)
func (m *Manager) handlePlayerSendDialogue(p *pluginProcess, correlationID string, act *pb.PlayerSendDialogueAction) {
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		m.sendActionError(p, correlationID, "invalid player_uuid")
		return
	}
	m.execMethod(id, func(pl *player.Player) {
		sub := dialogueNoop{}
		d := dialogue.New(sub, act.Title)
		if act.Body != nil {
			d = d.WithBody(*act.Body)
		}
		// Clamp to 6
		max := min(len(act.Buttons), 6)
		btns := make([]dialogue.Button, max)
		for i := range max {
			btns[i] = dialogue.Button{Text: act.Buttons[i]}
		}
		d = d.WithButtons(btns...)

		e := resolveWorldEntity(pl, act.Entity)
		if e == nil {
			m.sendActionError(p, correlationID, "entity not found")
			return
		}
		pl.SendDialogue(d, e)
	})
}

func (m *Manager) handleWorldSetTickRange(p *pluginProcess, correlationID string, act *pb.WorldSetTickRangeAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.TickRange < 0 {
		m.sendActionError(p, correlationID, "tick range must be non-negative")
		return
	}
	w.SetTickRange(int(act.TickRange))
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldPlaySound(p *pluginProcess, correlationID string, act *pb.WorldPlaySoundAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	pos, ok := vec3FromProto(act.Position)
	if !ok {
		m.sendActionError(p, correlationID, "invalid position")
		return
	}
	s := soundFromProto(act.Sound)
	<-w.Exec(func(tx *world.Tx) {
		tx.PlaySound(pos, s)
	})
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldAddParticle(p *pluginProcess, correlationID string, act *pb.WorldAddParticleAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	pos, ok := vec3FromProto(act.Position)
	if !ok {
		m.sendActionError(p, correlationID, "invalid position")
		return
	}
	part, ok := particleFromProto(act)
	if !ok {
		m.sendActionError(p, correlationID, "unknown particle")
		return
	}
	<-w.Exec(func(tx *world.Tx) {
		tx.AddParticle(pos, part)
	})
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldSetTime(p *pluginProcess, correlationID string, act *pb.WorldSetTimeAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	w.SetTime(int(act.Time))
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldStopTime(p *pluginProcess, correlationID string, act *pb.WorldStopTimeAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	w.StopTime()
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldStartTime(p *pluginProcess, correlationID string, act *pb.WorldStartTimeAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	w.StartTime()
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldSetSpawn(p *pluginProcess, correlationID string, act *pb.WorldSetSpawnAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Spawn == nil {
		m.sendActionError(p, correlationID, "missing spawn position")
		return
	}
	pos := cube.Pos{int(act.Spawn.X), int(act.Spawn.Y), int(act.Spawn.Z)}
	w.SetSpawn(pos)
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldQueryEntities(p *pluginProcess, correlationID string, act *pb.WorldQueryEntitiesAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	entities := make([]world.Entity, 0)
	<-w.Exec(func(tx *world.Tx) {
		for e := range tx.Entities() {
			entities = append(entities, e)
		}
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldEntities{WorldEntities: &pb.WorldEntitiesResult{
			World:    protoWorldRef(w),
			Entities: protoEntityRefs(entities),
		}},
	})
}

func (m *Manager) handleWorldQueryPlayers(p *pluginProcess, correlationID string, act *pb.WorldQueryPlayersAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	players := make([]world.Entity, 0)
	<-w.Exec(func(tx *world.Tx) {
		for pl := range tx.Players() {
			players = append(players, pl)
		}
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldPlayers{WorldPlayers: &pb.WorldPlayersResult{
			World:   protoWorldRef(w),
			Players: protoEntityRefs(players),
		}},
	})
}

func (m *Manager) handleWorldQueryEntitiesWithin(p *pluginProcess, correlationID string, act *pb.WorldQueryEntitiesWithinAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	box, ok := bboxFromProto(act.Box)
	if !ok {
		m.sendActionError(p, correlationID, "invalid bounding box")
		return
	}
	entities := make([]world.Entity, 0)
	<-w.Exec(func(tx *world.Tx) {
		for e := range tx.EntitiesWithin(box) {
			entities = append(entities, e)
		}
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldEntitiesWithin{WorldEntitiesWithin: &pb.WorldEntitiesWithinResult{
			World:    protoWorldRef(w),
			Box:      protoBBox(box),
			Entities: protoEntityRefs(entities),
		}},
	})
}

func (m *Manager) handleWorldQueryDefaultGameMode(p *pluginProcess, correlationID string, act *pb.WorldQueryDefaultGameModeAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	mode := w.DefaultGameMode()
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldDefaultGameMode{WorldDefaultGameMode: &pb.WorldDefaultGameModeResult{
			World:    protoWorldRef(w),
			GameMode: func() pb.GameMode { id, _ := world.GameModeID(mode); return pb.GameMode(id) }(),
		}},
	})
}

func (m *Manager) handleWorldQueryPlayerSpawn(p *pluginProcess, correlationID string, act *pb.WorldQueryPlayerSpawnAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.PlayerUuid == "" {
		m.sendActionError(p, correlationID, "missing player_uuid")
		return
	}
	id, err := uuid.Parse(act.PlayerUuid)
	if err != nil {
		m.sendActionError(p, correlationID, "invalid player_uuid")
		return
	}
	spawn := w.PlayerSpawn(id)
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldPlayerSpawn{WorldPlayerSpawn: &pb.WorldPlayerSpawnResult{
			World:      protoWorldRef(w),
			PlayerUuid: act.PlayerUuid,
			Spawn:      protoBlockPos(spawn),
		}},
	})
}

func (m *Manager) execMethod(id uuid.UUID, method func(pl *player.Player)) {
	if m.srv == nil {
		return
	}
	if handle, ok := m.srv.Player(id); ok {
		handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
			if pl, ok := e.(*player.Player); ok {
				method(pl)
			}
		})
	}
}

func (m *Manager) sendActionResult(p *pluginProcess, result *pb.ActionResult) {
	if p == nil || result == nil || result.CorrelationId == "" {
		return
	}
	p.queue(&pb.HostToPlugin{
		PluginId: p.id,
		Payload:  &pb.HostToPlugin_ActionResult{ActionResult: result},
	})
}

func (m *Manager) sendActionOK(p *pluginProcess, correlationID string) {
	if correlationID == "" {
		return
	}
	m.sendActionResult(p, &pb.ActionResult{CorrelationId: correlationID, Status: &pb.ActionStatus{Ok: true}})
}

func (m *Manager) sendActionError(p *pluginProcess, correlationID, msg string) {
	if correlationID == "" {
		return
	}
	m.sendActionResult(p, &pb.ActionResult{CorrelationId: correlationID, Status: &pb.ActionStatus{Ok: false, Error: &msg}})
}

func soundFromProto(s pb.Sound) world.Sound {
	switch s {
	case pb.Sound_ATTACK:
		return sound.Attack{Damage: true}
	case pb.Sound_DROWNING:
		return sound.Drowning{}
	case pb.Sound_BURNING:
		return sound.Burning{}
	case pb.Sound_FALL:
		return sound.Fall{Distance: 4}
	case pb.Sound_BURP:
		return sound.Burp{}
	case pb.Sound_POP:
		return sound.Pop{}
	case pb.Sound_EXPLOSION:
		return sound.Explosion{}
	case pb.Sound_THUNDER:
		return sound.Thunder{}
	case pb.Sound_LEVEL_UP:
		return sound.LevelUp{}
	case pb.Sound_EXPERIENCE:
		return sound.Experience{}
	case pb.Sound_FIREWORK_LAUNCH:
		return sound.FireworkLaunch{}
	case pb.Sound_FIREWORK_HUGE_BLAST:
		return sound.FireworkHugeBlast{}
	case pb.Sound_FIREWORK_BLAST:
		return sound.FireworkBlast{}
	case pb.Sound_FIREWORK_TWINKLE:
		return sound.FireworkTwinkle{}
	case pb.Sound_TELEPORT:
		return sound.Teleport{}
	case pb.Sound_ARROW_HIT:
		return sound.ArrowHit{}
	case pb.Sound_ITEM_BREAK:
		return sound.ItemBreak{}
	case pb.Sound_ITEM_THROW:
		return sound.ItemThrow{}
	case pb.Sound_TOTEM:
		return sound.Totem{}
	case pb.Sound_FIRE_EXTINGUISH:
		return sound.FireExtinguish{}
	default:
		return sound.Pop{}
	}
}

func particleFromProto(act *pb.WorldAddParticleAction) (world.Particle, bool) {
	return particleFromType(act.GetParticle(), act.Block, act.Face)
}

// particleFromType maps a particle enum plus optional block/face into a world.Particle.
func particleFromType(pt pb.ParticleType, blk *pb.BlockState, f *int32) (world.Particle, bool) {
	switch pt {
	case pb.ParticleType_PARTICLE_HUGE_EXPLOSION:
		return particle.HugeExplosion{}, true
	case pb.ParticleType_PARTICLE_ENDERMAN_TELEPORT:
		return particle.EndermanTeleport{}, true
	case pb.ParticleType_PARTICLE_SNOWBALL_POOF:
		return particle.SnowballPoof{}, true
	case pb.ParticleType_PARTICLE_EGG_SMASH:
		return particle.EggSmash{}, true
	case pb.ParticleType_PARTICLE_SPLASH:
		return particle.Splash{}, true
	case pb.ParticleType_PARTICLE_EFFECT:
		return particle.Effect{}, true
	case pb.ParticleType_PARTICLE_ENTITY_FLAME:
		return particle.EntityFlame{}, true
	case pb.ParticleType_PARTICLE_FLAME:
		return particle.Flame{}, true
	case pb.ParticleType_PARTICLE_DUST:
		return particle.Dust{}, true
	case pb.ParticleType_PARTICLE_BLOCK_FORCE_FIELD:
		return particle.BlockForceField{}, true
	case pb.ParticleType_PARTICLE_BONE_MEAL:
		return particle.BoneMeal{}, true
	case pb.ParticleType_PARTICLE_EVAPORATE:
		return particle.Evaporate{}, true
	case pb.ParticleType_PARTICLE_WATER_DRIP:
		return particle.WaterDrip{}, true
	case pb.ParticleType_PARTICLE_LAVA_DRIP:
		return particle.LavaDrip{}, true
	case pb.ParticleType_PARTICLE_LAVA:
		return particle.Lava{}, true
	case pb.ParticleType_PARTICLE_DUST_PLUME:
		return particle.DustPlume{}, true
	case pb.ParticleType_PARTICLE_BLOCK_BREAK:
		if b, ok := blockFromProto(blk); ok {
			return particle.BlockBreak{Block: b}, true
		}
	case pb.ParticleType_PARTICLE_PUNCH_BLOCK:
		if b, ok := blockFromProto(blk); ok {
			face := cube.FaceUp
			if f != nil {
				face = cube.Face(*f)
			}
			return particle.PunchBlock{Block: b, Face: face}, true
		}
	}
	return nil, false
}

func particleFromPlayerAction(act *pb.PlayerShowParticleAction) (world.Particle, bool) {
	return particleFromType(act.GetParticle(), act.Block, act.Face)
}

func playerTitleFromAction(act *pb.SendTitleAction) title.Title {
	t := title.New(act.Title)
	if act.Subtitle != nil && *act.Subtitle != "" {
		t = t.WithSubtitle(*act.Subtitle)
	}
	if act.FadeInMs != nil {
		t = t.WithFadeInDuration(time.Duration(*act.FadeInMs) * time.Millisecond)
	}
	if act.DurationMs != nil {
		t = t.WithDuration(time.Duration(*act.DurationMs) * time.Millisecond)
	}
	if act.FadeOutMs != nil {
		t = t.WithFadeOutDuration(time.Duration(*act.FadeOutMs) * time.Millisecond)
	}
	return t
}

// World query handlers

func (m *Manager) handleWorldQueryBlock(p *pluginProcess, correlationID string, act *pb.WorldQueryBlockAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var block world.Block
	<-w.Exec(func(tx *world.Tx) {
		block = tx.Block(pos)
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldBlock{WorldBlock: &pb.WorldBlockResult{
			World:    protoWorldRef(w),
			Position: act.Position,
			Block:    protoBlockState(block),
		}},
	})
}

func (m *Manager) handleWorldQueryBiome(p *pluginProcess, correlationID string, act *pb.WorldQueryBiomeAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var biome world.Biome
	<-w.Exec(func(tx *world.Tx) {
		biome = tx.Biome(pos)
	})
	biomeID := fmt.Sprintf("%d", biome.EncodeBiome())
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldBiome{WorldBiome: &pb.WorldBiomeResult{
			World:    protoWorldRef(w),
			Position: act.Position,
			BiomeId:  biomeID,
		}},
	})
}

func (m *Manager) handleWorldQueryLight(p *pluginProcess, correlationID string, act *pb.WorldQueryLightAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var lightLevel uint8
	<-w.Exec(func(tx *world.Tx) {
		lightLevel = tx.Light(pos)
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldLight{WorldLight: &pb.WorldLightResult{
			World:      protoWorldRef(w),
			Position:   act.Position,
			LightLevel: int32(lightLevel),
		}},
	})
}

func (m *Manager) handleWorldQuerySkyLight(p *pluginProcess, correlationID string, act *pb.WorldQuerySkyLightAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var skyLightLevel uint8
	<-w.Exec(func(tx *world.Tx) {
		skyLightLevel = tx.SkyLight(pos)
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldSkyLight{WorldSkyLight: &pb.WorldSkyLightResult{
			World:         protoWorldRef(w),
			Position:      act.Position,
			SkyLightLevel: int32(skyLightLevel),
		}},
	})
}

func (m *Manager) handleWorldQueryTemperature(p *pluginProcess, correlationID string, act *pb.WorldQueryTemperatureAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var temperature float64
	<-w.Exec(func(tx *world.Tx) {
		temperature = tx.Temperature(pos)
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldTemperature{WorldTemperature: &pb.WorldTemperatureResult{
			World:       protoWorldRef(w),
			Position:    act.Position,
			Temperature: temperature,
		}},
	})
}

func (m *Manager) handleWorldQueryHighestBlock(p *pluginProcess, correlationID string, act *pb.WorldQueryHighestBlockAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	var y int
	<-w.Exec(func(tx *world.Tx) {
		y = tx.HighestBlock(int(act.X), int(act.Z))
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldHighestBlock{WorldHighestBlock: &pb.WorldHighestBlockResult{
			World: protoWorldRef(w),
			X:     act.X,
			Z:     act.Z,
			Y:     int32(y),
		}},
	})
}

func (m *Manager) handleWorldQueryRainingAt(p *pluginProcess, correlationID string, act *pb.WorldQueryRainingAtAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var raining bool
	<-w.Exec(func(tx *world.Tx) {
		raining = tx.RainingAt(pos)
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldRainingAt{WorldRainingAt: &pb.WorldRainingAtResult{
			World:    protoWorldRef(w),
			Position: act.Position,
			Raining:  raining,
		}},
	})
}

func (m *Manager) handleWorldQuerySnowingAt(p *pluginProcess, correlationID string, act *pb.WorldQuerySnowingAtAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var snowing bool
	<-w.Exec(func(tx *world.Tx) {
		snowing = tx.SnowingAt(pos)
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldSnowingAt{WorldSnowingAt: &pb.WorldSnowingAtResult{
			World:    protoWorldRef(w),
			Position: act.Position,
			Snowing:  snowing,
		}},
	})
}

func (m *Manager) handleWorldQueryThunderingAt(p *pluginProcess, correlationID string, act *pb.WorldQueryThunderingAtAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var thundering bool
	<-w.Exec(func(tx *world.Tx) {
		thundering = tx.ThunderingAt(pos)
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldThunderingAt{WorldThunderingAt: &pb.WorldThunderingAtResult{
			World:      protoWorldRef(w),
			Position:   act.Position,
			Thundering: thundering,
		}},
	})
}

func (m *Manager) handleWorldQueryLiquid(p *pluginProcess, correlationID string, act *pb.WorldQueryLiquidAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var liquidState *pb.LiquidState
	<-w.Exec(func(tx *world.Tx) {
		if liq, ok := tx.Liquid(pos); ok {
			liquidState = protoLiquidState(liq)
		}
	})
	m.sendActionResult(p, &pb.ActionResult{
		CorrelationId: correlationID,
		Status:        &pb.ActionStatus{Ok: true},
		Result: &pb.ActionResult_WorldLiquid{WorldLiquid: &pb.WorldLiquidResult{
			World:    protoWorldRef(w),
			Position: act.Position,
			Liquid:   liquidState,
		}},
	})
}

// World mutation handlers

func (m *Manager) handleWorldSetBiome(p *pluginProcess, correlationID string, act *pb.WorldSetBiomeAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	if act.BiomeId == "" {
		m.sendActionError(p, correlationID, "missing biome_id")
		return
	}
	// Parse biome ID - can be numeric ID or canonical biome name
	var biome world.Biome
	var biomeID int
	if _, err := fmt.Sscanf(act.BiomeId, "%d", &biomeID); err == nil {
		var ok bool
		biome, ok = world.BiomeByID(biomeID)
		if !ok {
			m.sendActionError(p, correlationID, "unknown biome ID")
			return
		}
	} else {
		var ok bool
		biome, ok = world.BiomeByName(act.BiomeId)
		if !ok {
			m.sendActionError(p, correlationID, "unknown biome name")
			return
		}
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBiome(pos, biome)
	})
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldSetLiquid(p *pluginProcess, correlationID string, act *pb.WorldSetLiquidAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	var liquid world.Liquid
	if act.Liquid != nil && act.Liquid.Block != nil {
		if blk, ok := blockFromProto(act.Liquid.Block); ok {
			if liq, ok := blk.(world.Liquid); ok {
				liquid = liq
			} else {
				m.sendActionError(p, correlationID, "block is not a liquid")
				return
			}
		} else {
			m.sendActionError(p, correlationID, "unknown liquid block")
			return
		}
	}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetLiquid(pos, liquid)
	})
	m.sendActionOK(p, correlationID)
}

func (m *Manager) handleWorldScheduleBlockUpdate(p *pluginProcess, correlationID string, act *pb.WorldScheduleBlockUpdateAction) {
	w := m.worldFromRef(act.GetWorld())
	if w == nil {
		m.sendActionError(p, correlationID, "world not found")
		return
	}
	if act.Position == nil {
		m.sendActionError(p, correlationID, "missing position")
		return
	}
	if act.Block == nil {
		m.sendActionError(p, correlationID, "missing block")
		return
	}
	blk, ok := blockFromProto(act.Block)
	if !ok {
		m.sendActionError(p, correlationID, "unknown block")
		return
	}
	pos := cube.Pos{int(act.Position.X), int(act.Position.Y), int(act.Position.Z)}
	delay := time.Duration(act.DelayMs) * time.Millisecond
	<-w.Exec(func(tx *world.Tx) {
		tx.ScheduleBlockUpdate(pos, blk, delay)
	})
	m.sendActionOK(p, correlationID)
}
