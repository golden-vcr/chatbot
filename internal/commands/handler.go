package commands

import (
	"context"
	"fmt"

	"github.com/golden-vcr/auth"
)

type SayFunc func(s string) error

type Handler interface {
	Handle(command, args, userId, userDisplayName string) error
}

func NewHandler(ctx context.Context, authServiceClient auth.ServiceClient, say SayFunc) Handler {
	return &handler{
		ctx:               ctx,
		authServiceClient: authServiceClient,
		say:               say,
	}
}

type handler struct {
	ctx               context.Context
	authServiceClient auth.ServiceClient
	say               func(s string) error
}

func (h *handler) Handle(command, args, userId, userDisplayName string) error {
	switch command {
	case "ghosts":
		return h.handleGhosts()
	case "tapes":
		return h.handleTapes()
	case "youtube":
		return h.handleYoutube()
	case "balance":
		return h.handleBalance(userId, userDisplayName)
	}
	return fmt.Errorf("unrecognized command: %s", command)
}
