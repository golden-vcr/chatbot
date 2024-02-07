package chatlog

import (
	"context"
	"errors"
	"net/http"

	"github.com/golden-vcr/chatbot/internal/irc"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
)

type Server struct {
	mb *messageBuffer
}

func NewServer(ctx context.Context, logger *slog.Logger, messagesChan <-chan *irc.Message) *Server {
	mb := newMessageBuffer(128)

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
				logger.Info("Propagating chatlog event", "chatlogEvent", ev)
				switch ev.Type {
				case EventTypeAppend:
					mb.append(ev.Payload.Append)
				case EventTypeDelete:
					mb.delete(ev.Payload.Delete.MessageId)
				case EventTypeBan:
					mb.ban(ev.Payload.Ban.UserId)
				case EventTypeClear:
					mb.clear()
				}
			}
		}
	}()

	return &Server{
		mb: mb,
	}
}

func (s *Server) RegisterRoutes(r *mux.Router) {
	r.Path("/chatlog").Methods("GET").HandlerFunc(s.handleGetChatlog)
}

func (s *Server) handleGetChatlog(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "NYI", http.StatusInternalServerError)
}
