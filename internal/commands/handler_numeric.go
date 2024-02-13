package commands

import (
	"context"
	"encoding/json"

	"github.com/golden-vcr/schemas/core"
	etwitch "github.com/golden-vcr/schemas/twitch-events"
)

func (h *handler) handleNumericCommand(numPoints int, args, userId, userDisplayName string) error {
	ev := etwitch.Event{
		Type: etwitch.EventTypeViewerRedeemedFunPoints,
		Viewer: &core.Viewer{
			TwitchUserId:      userId,
			TwitchDisplayName: userDisplayName,
		},
		Payload: &etwitch.Payload{
			ViewerRedeemedFunPoints: &etwitch.PayloadViewerRedeemedFunPoints{
				NumPoints: numPoints,
				Message:   args,
			},
		},
	}
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return h.twitchEventsProducer.Send(context.Background(), data)
}
