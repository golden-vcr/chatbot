package chatlog

import (
	"fmt"
	"strconv"
	"strings"
)

// emoteInfo is the parsed representation of the 'emotes' attribute included in PRIVMSG
// messages from Twitch IRC
type emoteInfo struct {
	id    string
	spans []emoteSpan
}

// emoteSpan represents a single occurrence of an emote by its start and end position in
// the body of the PRIVMSG
type emoteSpan struct {
	start int
	end   int
}

// parseEmotes parses the extra 'emotes' attribute form a PRIVMSG
func parseEmotes(emotes string) ([]emoteInfo, error) {
	// Early-out if input attribute is empty
	if emotes == "" {
		return []emoteInfo{}, nil
	}

	// The extra 'emotes' attribute in Twitch IRC messages should contain a
	// '/'-delimited list of emotes, where each emote takes the form
	// '<emote-id>:<start>-<end>,<start>-<end>,...'
	emoteTokens := strings.Split(emotes, "/")
	emoteInfos := make([]emoteInfo, 0, len(emoteTokens))
	for _, emoteToken := range emoteTokens {
		// For each emote, split on the first colon to separate ID from spans
		colonPos := strings.IndexRune(emoteToken, ':')
		if colonPos <= 0 || colonPos+1 >= len(emoteToken) {
			return nil, fmt.Errorf("emote attribute is not delimited by a colon")
		}
		emoteId := emoteToken[:colonPos]
		spanTokens := strings.Split(emoteToken[colonPos+1:], ",")
		if len(spanTokens) == 0 {
			return nil, fmt.Errorf("emote attribute has invalid spans")
		}

		// Parse each span token, which should take the form '%d-%d'
		spans := make([]emoteSpan, 0, len(spanTokens))
		for _, spanToken := range spanTokens {
			dashPos := strings.IndexRune(spanToken, '-')
			if dashPos <= 0 || dashPos+1 >= len(spanToken) {
				return nil, fmt.Errorf("span in emote attribute is not delimited by a hyphen")
			}
			start, err := strconv.Atoi(spanToken[:dashPos])
			if err != nil {
				return nil, fmt.Errorf("span in emote attribute has non-numeric start")
			}
			end, err := strconv.Atoi(spanToken[dashPos+1:])
			if err != nil {
				return nil, fmt.Errorf("span in emote attribute has non-numeric end")
			}
			spans = append(spans, emoteSpan{
				start: start,
				end:   end,
			})
		}
		emoteInfos = append(emoteInfos, emoteInfo{
			id:    emoteId,
			spans: spans,
		})
	}
	return emoteInfos, nil
}
