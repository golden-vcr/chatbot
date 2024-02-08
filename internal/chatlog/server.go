package chatlog

import (
	"context"
	"errors"

	"github.com/golden-vcr/chatbot/internal/irc"
	"github.com/golden-vcr/server-common/sse"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
)

type Server struct {
	mb         *eventBuffer
	eventsChan chan *Event
}

func NewServer(ctx context.Context, logger *slog.Logger, messagesChan <-chan *irc.Message) *Server {
	mb := newEventBuffer(128)
	eventsChan := make(chan *Event, 32)

	go func() {
		for message := range messagesChan {
			ev, err := EventFromMessage(message)
			if err != nil {
				if !errors.Is(err, ErrIgnored) {
					logger.Error("Failed to generate chatlog event from IRC message",
						"message", message,
						"error", err,
					)
				}
			} else {
				ev.eventStreamId = uuid.NewString()
				logger.Info("Propagating chatlog event", "chatlogEvent", ev)
				mb.push(ev)
				eventsChan <- ev
			}
		}
	}()

	return &Server{
		mb:         mb,
		eventsChan: eventsChan,
	}
}

func (s *Server) RegisterRoutes(ctx context.Context, r *mux.Router) {
	h := sse.NewHandler[*Event](ctx, s.eventsChan)
	h.ResolveEventId = func(ev *Event) string {
		return ev.eventStreamId
	}
	h.OnConnect = func(lastEventId string) []*Event {
		// If no Last-Event-ID is specified, just send an initial burst of the N most
		// recent events, up to a reasonable limit
		if lastEventId == "" {
			return s.mb.take(64)
		}

		// Otherwise, take all events from the buffer so we can scan for event ID
		events := s.mb.take(s.mb.size)

		// Find the index where our last-received event appears
		lastEventIndex := -1
		for i := 0; i < len(events); i++ {
			if events[i].eventStreamId == lastEventId {
				lastEventIndex = i
				break
			}
		}

		// If we recognize the last event, catch the client up by sending all events
		// that have been buffered since; otherwise send all known events, since the
		// last event must be so old that we don't remember it
		return events[lastEventIndex+1:]
	}

	r.Path("/chatlog").Methods("GET").Handler(h)
}

func (s *Server) EmitBotMessage(text string) {
	ev := &Event{
		Type: EventTypeAppend,
		Payload: &Payload{
			Append: &PayloadAppend{
				MessageId: uuid.NewString(),
				UserId:    "_BOT_",
				Username:  "_BOT_",
				Color:     "#FFFFFF",
				Text:      text,
				Emotes:    []EmoteDetails{},
			},
		},
		eventStreamId: uuid.NewString(),
	}
	s.mb.push(ev)
	s.eventsChan <- ev
}
