package plugin

import (
	"strings"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
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
	for _, action := range batch.Actions {
		if action == nil {
			continue
		}
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
		case *pb.Action_WorldSetBlock:
			m.handleWorldSetBlock(p, correlationID, kind.WorldSetBlock)
		case *pb.Action_WorldPlaySound:
			m.handleWorldPlaySound(p, correlationID, kind.WorldPlaySound)
		case *pb.Action_WorldAddParticle:
			m.handleWorldAddParticle(p, correlationID, kind.WorldAddParticle)
		case *pb.Action_WorldQueryEntities:
			m.handleWorldQueryEntities(p, correlationID, kind.WorldQueryEntities)
		case *pb.Action_WorldQueryPlayers:
			m.handleWorldQueryPlayers(p, correlationID, kind.WorldQueryPlayers)
		case *pb.Action_WorldQueryEntitiesWithin:
			m.handleWorldQueryEntitiesWithin(p, correlationID, kind.WorldQueryEntitiesWithin)
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

func (m *Manager) handleWorldSetBlock(p *pluginProcess, correlationID string, act *pb.WorldSetBlockAction) {
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
	var blk world.Block
	var ok bool
	if act.Block != nil {
		blk, ok = blockFromProto(act.Block)
		if !ok {
			m.sendActionError(p, correlationID, "unknown block")
			return
		}
	}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, blk, nil)
	})
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
	switch act.GetParticle() {
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
		if act.Block != nil {
			if blk, ok := blockFromProto(act.Block); ok {
				return particle.BlockBreak{Block: blk}, true
			}
		}
	case pb.ParticleType_PARTICLE_PUNCH_BLOCK:
		if act.Block != nil {
			if blk, ok := blockFromProto(act.Block); ok {
				face := cube.Face(0)
				if act.Face != nil {
					face = cube.Face(*act.Face)
				}
				return particle.PunchBlock{Block: blk, Face: face}, true
			}
		}
	}
	return nil, false
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
