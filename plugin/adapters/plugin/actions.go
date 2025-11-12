package plugin

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
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

func (m *Manager) execMethod(id uuid.UUID, method func(pl *player.Player)) {
	if handle, ok := m.srv.Player(id); ok {
		handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
			if pl, ok := e.(*player.Player); ok {
				method(pl)
			}
		})
	}
}
