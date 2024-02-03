package irc

import (
	"fmt"
	"strings"
)

// Message represents a single message received from the Twitch IRC server with both the
// 'twitch.tv/commands' and 'twitch.tv/tags' capabilities enabled
type Message struct {
	// Raw is the original message in its plain-text format, stripped of the trailing
	// newline
	Raw string

	// Extra is a parsed representation of the '@'-prefixed, ';'-delimited string of
	// 'key=value' pairs that may appear before the message, separated from it by a
	// space: e.g. '@foo=bar;x=42;type= :tmi.twitch.tv CLEARCHAT #somechannel' will have
	// an Extra map of {"foo": "bar", "x": "42", "type": ""}
	Extra map[string]string

	// Prefix is the user and/or host information that precedes the data of the message,
	// stripped of the leading colon, e.g. 'tmi.twitch.tv',
	// 'somebody!somebody@somebody.tmi.twitch.tv'
	Prefix string

	// Type is the message type identifier that follows the prefix, e.g. 'USERSTATE',
	// 'CLEARCHAT', 'JOIN', 'PART', 'PING', etc.
	Type string

	// Params is a list of all space-delimited tokens that follow the type, up to (and
	// not including) the first value that's prefixed with a ':'
	Params []string

	// Body is the string that follows the first ':' appearing after the message type
	Body string
}

func parseMessage(s string) (*Message, error) {
	// All IRC messages are CRLF-delimited; trim any training carriage returns
	raw := strings.TrimSuffix(s, "\r")

	// If the message begins with '@', parse extra parameters
	extra, remainder, err := parseExtra(raw)
	if err != nil {
		return nil, err
	}

	// If the remainder of the message begins with ':', parse that prefix
	prefix, remainder := parsePrefix(remainder)

	// The next space-delimited token should be a message type, or a 3-digit numeric
	// string
	messageType, remainder, err := parseMessageType(remainder)
	if err != nil {
		return nil, err
	}

	// All subsequent space-delimited tokens up to the next ':' should be parsed as
	// parameters, and the remainder following the ':' is the message body
	params, body := parseParamsAndBody(remainder)

	return &Message{
		Raw:    raw,
		Extra:  extra,
		Prefix: prefix,
		Type:   messageType,
		Params: params,
		Body:   body,
	}, nil
}

func parseExtra(s string) (map[string]string, string, error) {
	// If the message doesn't begin with '@', there are no extra parameters
	extra := make(map[string]string)
	if s == "" || s[0] != '@' {
		return extra, s, nil
	}

	// Split the message into the '@'-prefixed extra parameters (e.g. '@foo=bar;baz=42')
	// and the remaining raw message after the next space
	semicolonDelimitedKeyValuePairs := s[1:]
	remainder := ""
	if spacePos := strings.IndexRune(s, ' '); spacePos > 0 {
		semicolonDelimitedKeyValuePairs = s[1:spacePos]
		remainder = s[spacePos+1:]
	}

	// Parse the extra parameters to a map (e.g. 'foo=bar;baz=42' =>
	// []string{'foo=bar', 'baz=42'})
	for _, keyValuePair := range strings.Split(semicolonDelimitedKeyValuePairs, ";") {
		equalsPos := strings.IndexRune(keyValuePair, '=')
		if equalsPos <= 0 {
			return nil, "", fmt.Errorf("found no equals sign in extra key-value pair '%s'", keyValuePair)
		}
		key := keyValuePair[:equalsPos]
		value := keyValuePair[equalsPos+1:]
		extra[key] = value
	}
	return extra, remainder, nil
}

func parsePrefix(s string) (string, string) {
	// If the message doesn't begin with ':', it has no prefix
	if s == "" || s[0] != ':' {
		return "", s
	}

	// Take the first space-delimited token as the prefix (stripping the leading ':'),
	// and retain the remainder after the next space
	prefix := s[1:]
	remainder := ""
	if spacePos := strings.IndexRune(s, ' '); spacePos > 0 {
		prefix = s[1:spacePos]
		remainder = s[spacePos+1:]
	}
	return prefix, remainder
}

func parseMessageType(s string) (string, string, error) {
	// Every message should have a type
	if s == "" {
		return "", "", fmt.Errorf("no message type token found")
	}

	// Take the first space-delimited token as the message type
	token := s
	remainder := ""
	if spacePos := strings.IndexRune(s, ' '); spacePos > 0 {
		token = s[:spacePos]
		remainder = s[spacePos+1:]
	}

	// Message type must be all-caps or a three-digit number
	numUpper := 0
	numNumeric := 0
	numNeither := 0
	for _, c := range token {
		if c >= 'A' && c <= 'Z' {
			numUpper++
		} else if c >= '0' && c <= '9' {
			numNumeric++
		} else {
			numNeither++
		}
	}
	isValid := numNeither == 0 && (numNumeric == 3 || numUpper > 0)
	if !isValid {
		return "", "", fmt.Errorf("message type is neither all-caps nor a three-digit numeric string")
	}
	return token, remainder, nil
}

func parseParamsAndBody(s string) ([]string, string) {
	// If the remainder of the message after the type is empty, there are no params and
	// no body
	if s == "" {
		return nil, ""
	}

	// If the remainder of the message contains a ':', it delimits the list of
	// parameters from the body
	spaceDelimitedParams := s
	body := ""
	if colonPos := strings.IndexRune(s, ':'); colonPos > 1 {
		spaceDelimitedParams = s[:colonPos-1]
		body = s[colonPos+1:]
	}
	return strings.Split(spaceDelimitedParams, " "), body
}
