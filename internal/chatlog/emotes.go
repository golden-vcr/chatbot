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

// parseEmotes parses the extra 'emotes' attribute from a PRIVMSG
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

// substituteEmotes reformats the body of a PRIVMSG so that each occurrence of an emote
// is replaced with '$0', '$1', etc., where the integer represents an index into an
// array of EmoteDetails (literal '$' will be escaped as '$$')
func substituteEmotes(body string, emoteInfos []emoteInfo) (string, []EmoteDetails, error) {
	// Early-out if the message contains no emotes
	if len(emoteInfos) == 0 {
		return body, []EmoteDetails{}, nil
	}

	// Resolve the name of each emote by identifying the corresponding span of text in
	// the message body, and ensure that all instances of the same emote show the same
	// name
	emoteNames := make([]string, 0, len(emoteInfos))
	for _, emoteInfo := range emoteInfos {
		emoteName := ""
		for i, span := range emoteInfo.spans {
			if span.start > span.end || span.start < 0 || span.start >= len(body) || span.end < 0 || span.end >= len(body) {
				return "", nil, fmt.Errorf("invalid span bounds %d-%d", span.start, span.end)
			}
			spanText := body[span.start : span.end+1]
			if strings.IndexRune(spanText, ' ') >= 0 {
				return "", nil, fmt.Errorf("span '%s' contains space", spanText)
			}
			if i == 0 {
				emoteName = spanText
			} else if spanText != emoteName {
				return "", nil, fmt.Errorf("name of emote %s was detected as '%s' in first span; '%s' subsequently", emoteInfo.id, emoteName, spanText)
			}
		}
		if emoteName == "" {
			return "", nil, fmt.Errorf("emote %s has no spans", emoteInfo.id)
		}
		emoteNames = append(emoteNames, emoteName)
	}

	// Substitute '$0', '$1', etc. for each instance of each emote, and build a list of
	// EmoteDetails that can be indexed with those values to resolve the full image URL
	// for the emote
	tokens := strings.Split(strings.ReplaceAll(body, "$", "$$"), " ")
	emotes := make([]EmoteDetails, 0, len(emoteNames))
	for emoteIndex, emoteName := range emoteNames {
		for tokenIndex := 0; tokenIndex < len(tokens); tokenIndex++ {
			if tokens[tokenIndex] == emoteName {
				tokens[tokenIndex] = fmt.Sprintf("$%d", emoteIndex)
			}
		}
		emoteId := emoteInfos[emoteIndex].id
		emotes = append(emotes, EmoteDetails{
			Name: emoteName,
			Url:  fmt.Sprintf("https://static-cdn.jtvnw.net/emoticons/v2/%s/default/dark/1.0", emoteId),
		})
	}
	return strings.Join(tokens, " "), emotes, nil
}
