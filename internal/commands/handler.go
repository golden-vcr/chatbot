package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	case "speak":
		return h.say("woof woof")
	case "balance":
		accessToken, err := h.authServiceClient.RequestServiceToken(h.ctx, auth.ServiceTokenRequest{
			Service: "chatbot",
			User: auth.UserDetails{
				Id:          userId,
				Login:       strings.ToLower(userDisplayName),
				DisplayName: userDisplayName,
			},
		})
		if err != nil {
			return err
		}
		url := "https://goldenvcr.com/api/ledger/balance"
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", accessToken))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("got response %d from ledger balance request", res.StatusCode)
		}
		type fields struct {
			AvailablePoints int `json:"availablePoints"`
		}
		var f fields
		if err := json.NewDecoder(res.Body).Decode(&f); err != nil {
			return err
		}
		return h.say(fmt.Sprintf("@%s You have %d fun points available.", userDisplayName, f.AvailablePoints))
	}
	return fmt.Errorf("unrecognized command: %s", command)
}
