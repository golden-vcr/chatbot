package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golden-vcr/auth"
	"github.com/golden-vcr/server-common/rmq"
)

type SayFunc func(s string) error

type Handler interface {
	Handle(command, args, userId, userDisplayName string) error
}

func NewHandler(ctx context.Context, authServiceClient auth.ServiceClient, say SayFunc, twitchEventsProducer rmq.Producer) Handler {
	return &handler{
		ctx:                  ctx,
		authServiceClient:    authServiceClient,
		say:                  say,
		twitchEventsProducer: twitchEventsProducer,
	}
}

type handler struct {
	ctx                  context.Context
	authServiceClient    auth.ServiceClient
	say                  func(s string) error
	twitchEventsProducer rmq.Producer
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
	if numPoints, err := strconv.Atoi(command); err == nil && numPoints > 0 {
		return h.handleNumericCommand(numPoints, args, userId, userDisplayName)
	}
	return fmt.Errorf("unrecognized command: %s", command)
}
