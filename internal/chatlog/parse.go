package chatlog

import (
	"errors"
	"fmt"

	"github.com/golden-vcr/chatbot/internal/irc"
)

var ErrIgnored = errors.New("message ignored")

func EventFromMessage(message *irc.Message) (*Event, error) {
	switch message.Type {
	case "PRIVMSG":
		// PRIVMSG indicates that a user has sent a message in chat
		return eventFromPrivmsg(message)
	case "CLEARMSG":
		// CLEARMSG indicates that a mod has deleted a single message by ID
		return eventFromClearmsg(message)
	case "CLEARCHAT":
		// CLEARCHAT either clears the entire log or deletes all messages for a single
		// user, depending on whether the 'target-user-id' attribute is set
		return eventFromClearchat(message)
	}
	return nil, ErrIgnored
}

func eventFromPrivmsg(message *irc.Message) (*Event, error) {
	messageId := message.Extra["id"]
	if messageId == "" {
		return nil, fmt.Errorf("missing extra attribute 'id'")
	}

	userId := message.Extra["user-id"]
	if userId == "" {
		return nil, fmt.Errorf("missing extra attribute 'user-id'")
	}

	username := message.Extra["display-name"]
	if username == "" {
		return nil, fmt.Errorf("missing extra attribute 'display-name'")
	}

	color := message.Extra["color"]
	if color == "" {
		color = "#FFFFFF"
	}

	emoteInfos, err := parseEmotes(message.Extra["emotes"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse extra attribute 'emotes': %w", err)
	}
	text, emotes, err := substituteEmotes(message.Body, emoteInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to substitute emotes: %w", err)
	}

	return &Event{
		Type: EventTypeAppend,
		Payload: &Payload{
			Append: &PayloadAppend{
				MessageId: messageId,
				UserId:    userId,
				Username:  username,
				Color:     color,
				Text:      text,
				Emotes:    emotes,
			},
		},
	}, nil
}

func eventFromClearmsg(msg *irc.Message) (*Event, error) {
	messageId := msg.Extra["target-msg-id"]
	if messageId == "" {
		return nil, fmt.Errorf("missing extra attribute 'target-msg-id'")
	}
	return &Event{
		Type: EventTypeDelete,
		Payload: &Payload{
			Delete: &PayloadDelete{
				MessageId: messageId,
			},
		},
	}, nil
}

func eventFromClearchat(msg *irc.Message) (*Event, error) {
	userId := msg.Extra["target-user-id"]
	if userId != "" {
		return &Event{
			Type: EventTypeBan,
			Payload: &Payload{
				Ban: &PayloadBan{
					UserId: userId,
				},
			},
		}, nil
	}
	return &Event{
		Type: EventTypeClear,
	}, nil
}
