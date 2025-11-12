package plugin

import (
	"strings"
	"time"

	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	pb "github.com/secmc/plugin/proto/generated"
)

func (m *Manager) applyActions(p *pluginProcess, batch *pb.ActionBatch) {
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
		case *pb.Action_GiveItem:
			m.handleGiveItem(kind.GiveItem)
		case *pb.Action_ClearInventory:
			m.handleClearInventory(kind.ClearInventory)
		case *pb.Action_SetHeldItem:
			m.handleSetHeldItem(kind.SetHeldItem)
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
		}
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

func (m *Manager) execMethod(id uuid.UUID, method func(pl *player.Player)) {
	if handle, ok := m.srv.Player(id); ok {
		handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
			if pl, ok := e.(*player.Player); ok {
				method(pl)
			}
		})
	}
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
