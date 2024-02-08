package commands

import (
	"context"
	"fmt"
)

type SayFunc func(s string) error

type Handler interface {
	Handle(command string, args string) error
}

func NewHandler(ctx context.Context, say SayFunc) Handler {
	return &handler{
		ctx: ctx,
		say: say,
	}
}

type handler struct {
	ctx context.Context
	say func(s string) error
}

func (h *handler) Handle(command string, args string) error {
	switch command {
	case "speak":
		return h.say("woof woof")
	}
	return fmt.Errorf("unrecognized command: %s", command)
}
