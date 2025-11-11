package handlers

import (
	"github.com/df-mc/dragonfly/server/world"
	pb "github.com/secmc/plugin/proto/generated"
)

var _ world.Handler = (*WorldHandler)(nil)

type EventBroadcaster interface {
	BroadcastEvent(evt *pb.EventEnvelope)
	GenerateEventID() string
}

type WorldHandler struct {
	world.NopHandler
	broadcaster EventBroadcaster
}

func NewWorldHandler(broadcaster EventBroadcaster) world.Handler {
	return &WorldHandler{broadcaster: broadcaster}
}

func (h *WorldHandler) HandleClose(tx *world.Tx) {
	if h.broadcaster == nil || tx == nil {
		return
	}
	evt := &pb.EventEnvelope{
		EventId: h.broadcaster.GenerateEventID(),
		Type:    pb.EventType_WORLD_CLOSE,
		Payload: &pb.EventEnvelope_WorldClose{
			WorldClose: &pb.WorldCloseEvent{},
		},
	}
	h.broadcaster.BroadcastEvent(evt)
}
