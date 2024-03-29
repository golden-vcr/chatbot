package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	case "friends":
		return h.handleFriends()
	case "alerts":
		return h.handleAlerts()
	case "tapes":
		return h.handleTapes()
	case "remix":
		return h.handleRemix()
	case "youtube":
		return h.handleYoutube()
	case "camera":
		return h.handleCamera()
	case "bc":
		return h.handleBc()
	case "uptime":
		return h.handleUptime()
	case "tape":
		return h.handleTape()
	case "balance":
		return h.handleBalance(userId, userDisplayName)
	}
	if strings.ToLower(command) == "prayerbear" {
		return h.handleNumericCommand(200, "prayerbear", userId, userDisplayName)
	}
	if strings.ToLower(command) == "standback" {
		return h.handleNumericCommand(300, "standback", userId, userDisplayName)
	}
	if command == "ghost" {
		message := command + " "
		if !strings.HasPrefix(args, "of ") {
			message += "of "
		}
		message += args
		return h.handleNumericCommand(200, message, userId, userDisplayName)
	}
	if command == "friend" {
		message := fmt.Sprintf("%s %s", command, args)
		return h.handleNumericCommand(200, message, userId, userDisplayName)
	}

	if numPoints, err := strconv.Atoi(command); err == nil && numPoints > 0 {
		return h.handleNumericCommand(numPoints, args, userId, userDisplayName)
	}
	return fmt.Errorf("unrecognized command: %s", command)
}
