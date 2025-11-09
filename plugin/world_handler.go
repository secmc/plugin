package plugin

import (
	"github.com/df-mc/dragonfly/plugin/proto"
	"github.com/df-mc/dragonfly/server/world"
)

type WorldHandler struct {
	world.NopHandler
	mgr *Manager
}

func (h *WorldHandler) HandleClose(tx *world.Tx) {
	if h.mgr == nil || tx == nil {
		return
	}
	evt := &proto.EventEnvelope{
		EventID:    generateEventID(),
		Type:       "WORLD_CLOSE",
		WorldClose: &proto.WorldCloseEvent{},
	}
	h.mgr.broadcastEvent(evt.Type, evt)
}
