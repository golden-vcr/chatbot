package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golden-vcr/broadcasts"
)

func (h *handler) handleUptime() error {
	// GET /api/broadcasts/history to obtain data for the most recent stream
	url := "https://goldenvcr.com/api/broadcasts/history"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Set("n", "1")
	req.URL.RawQuery = q.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got response %d from broadcast history request", res.StatusCode)
	}

	// Decode the results and resolve the active broadcast and screening, if any
	var history broadcasts.History
	if err := json.NewDecoder(res.Body).Decode(&history); err != nil {
		return err
	}
	var broadcast *broadcasts.Broadcast
	if len(history.Broadcasts) > 0 && history.Broadcasts[0].EndedAt == nil {
		broadcast = &history.Broadcasts[0]
	}

	// Early-out if we're not screening a tape
	if broadcast == nil {
		return h.say("No broadcast is currently live.")
	}

	// Send a message indicating how long we've been live
	minutesElapsed := max(0, int(time.Since(broadcast.StartedAt).Minutes()))
	hourFigure := minutesElapsed / 60
	minuteFigure := minutesElapsed - (hourFigure * 60)
	readout := ""
	if hourFigure > 0 {
		readout = fmt.Sprintf("%dh%02dm", hourFigure, minuteFigure)
	} else {
		readout = fmt.Sprintf("%dm", minuteFigure)
	}
	return h.say(fmt.Sprintf("Broadcast %d has been live for %s.", broadcast.Id, readout))
}
