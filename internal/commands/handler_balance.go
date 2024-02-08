package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golden-vcr/auth"
)

func (h *handler) handleBalance(userId, userDisplayName string) error {
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
