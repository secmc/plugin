package plugin

import (
	pb "github.com/df-mc/dragonfly/plugin/proto/generated"
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
	evt := &pb.EventEnvelope{
		EventId: generateEventID(),
		Type:    "WORLD_CLOSE",
		Payload: &pb.EventEnvelope_WorldClose{
			WorldClose: &pb.WorldCloseEvent{},
		},
	}
	h.mgr.broadcastEvent(evt)
}
