package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golden-vcr/broadcasts"
)

func (h *handler) handleTape() error {
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
	var screening *broadcasts.Screening
	if len(history.Broadcasts) > 0 && history.Broadcasts[0].EndedAt == nil {
		broadcast = &history.Broadcasts[0]
		if len(broadcast.Screenings) > 0 && broadcast.Screenings[len(broadcast.Screenings)-1].EndedAt == nil {
			screening = &broadcast.Screenings[len(broadcast.Screenings)-1]
		}
	}

	// Early-out if we're not screening a tape
	if broadcast == nil {
		return h.say("No broadcast is currently live.")
	}
	if screening == nil {
		return h.say("No tape is currently being screened.")
	}

	// Request the full details of the tape we're currently screening
	url = fmt.Sprintf("https://goldenvcr.com/api/tapes/catalog/%d", screening.TapeId)
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got response %d from catalog request", res.StatusCode)
	}
	type fields struct {
		Id      int    `json:"id"`
		Title   string `json:"title"`
		Year    int    `json:"year"`
		Runtime int    `json:"runtime"`
	}
	var f fields
	if err := json.NewDecoder(res.Body).Decode(&f); err != nil {
		return err
	}

	// Send a message describing the current tape
	desc := ""
	if f.Year > 0 || f.Runtime > 0 {
		desc += " ("
		if f.Year > 0 {
			desc += fmt.Sprintf("%d", f.Year)
			if f.Runtime > 0 {
				desc += ", "
			}
		}
		if f.Runtime > 0 {
			desc += fmt.Sprintf("%dm", f.Runtime)
		}
		desc += ")"
	}
	minutesElapsed := max(0, int(time.Since(screening.StartedAt).Minutes()))
	tapeUrl := fmt.Sprintf("https://goldenvcr.com/tapes/%d", screening.TapeId)
	return h.say(fmt.Sprintf("The current tape is #%d: «%s»%s. It's been screened for %dm so far. %s", f.Id, f.Title, desc, minutesElapsed, tapeUrl))
}
