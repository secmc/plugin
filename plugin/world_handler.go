package plugin

import (
	"github.com/df-mc/dragonfly/server/world"
	pb "github.com/secmc/plugin/plugin/proto/generated"
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
		EventId: h.mgr.generateEventID(),
		Type:    "WORLD_CLOSE",
		Payload: &pb.EventEnvelope_WorldClose{
			WorldClose: &pb.WorldCloseEvent{},
		},
	}
	h.mgr.broadcastEvent(evt)
}
