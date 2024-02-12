package irc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golden-vcr/auth"
	"github.com/golden-vcr/chatbot"
	"github.com/golden-vcr/chatbot/internal/commands"
)

var ErrReceivedReconnect = errors.New("received RECONNECT message from Twitch IRC server")

type Bot interface {
	GetStatus() chatbot.Status
	GetLastError() error
	GetLastPingTime() time.Time
}

func NewBot(ctx context.Context, conn Conn, channelName, username, userAccessToken string, messagesChan chan<- *Message, emitBotMessage func(string), authServiceClient auth.ServiceClient) (Bot, error) {
	lines, err := conn.Recv()
	if err != nil {
		return nil, err
	}

	b := &bot{
		conn:        conn,
		channel:     fmt.Sprintf("#%s", channelName),
		nick:        strings.ToLower(username),
		accessToken: userAccessToken,
		commandHandler: commands.NewHandler(ctx, authServiceClient, func(s string) error {
			if err := conn.Sendf("PRIVMSG #%s :%s", channelName, s); err != nil {
				return err
			}
			emitBotMessage(s)
			return nil
		}),
		signalError: func(err error) {
			emitBotMessage(fmt.Sprintf("ERROR: %s", err))
		},
	}
	go func() {
		for s := range lines {
			message, err := b.handle(s)
			if err != nil {
				b.fail(err)
				return
			}
			messagesChan <- message
		}
	}()

	if err := b.init(); err != nil {
		return nil, err
	}

	return b, nil
}

type bot struct {
	conn           Conn
	channel        string
	nick           string
	accessToken    string
	commandHandler commands.Handler
	signalError    func(err error)

	err          error
	lastPingTime time.Time

	gotCapAck          bool
	gotGlobalUserState bool
	gotRoomState       bool
}

func (b *bot) init() error {
	// Kick off process of initiating an IRC connection to our server
	if err := b.conn.Send("CAP REQ :twitch.tv/commands twitch.tv/tags"); err != nil {
		return err
	}
	if err := b.conn.Sendf("PASS oauth:%s", b.accessToken); err != nil {
		return err
	}
	if err := b.conn.Sendf("NICK %s", b.nick); err != nil {
		return err
	}
	return nil
}

func (b *bot) handle(s string) (*Message, error) {
	// Parse the incoming IRC message from plain-text
	m, err := parseMessage(s)
	if err != nil {
		return nil, err
	}

	// If we're still in the init stage, we need to send a JOIN message, but only once
	// user login is complete
	hasSentJoin := b.gotCapAck && b.gotGlobalUserState

	switch m.Type {
	// We should always respond to a PING message by immediately replying with a PONG
	case "PING":
		pong := strings.Replace(m.Raw, "PING ", "PONG ", 1)
		if err := b.conn.Send(pong); err != nil {
			return m, err
		}
		b.lastPingTime = time.Now()
		return m, nil

	// If we get a RECONNECT message, the connection is being closed server-side for
	// maintenance reasons and we should attempt to reconnect
	case "RECONNECT":
		return m, ErrReceivedReconnect

	// If we get a NOTICE telling us our login failed, abort
	case "NOTICE":
		if m.Body == "Login authentication failed" {
			return m, fmt.Errorf("Login authentication failed")
		}
		// All other NOTICE message should be ignored
		return m, nil

	// If we get a CAP * ACK matching the capabilities we requested, note it, and send a
	// JOIN message as soon as we have both CAP ACK and GLOBALUSERSTATE
	case "CAP":
		if !b.gotCapAck && includes(m.Params, "ACK") && m.Body == "twitch.tv/commands twitch.tv/tags" {
			b.gotCapAck = true
			if !hasSentJoin && b.gotGlobalUserState {
				return m, b.sendJoin()
			}
		}
		return m, nil

	// If we get a GLOBALUSERSTATE after a CAP * ACK, we're ready to join the channel
	case "GLOBALUSERSTATE":
		if !b.gotGlobalUserState {
			b.gotGlobalUserState = true
			if !hasSentJoin && b.gotCapAck {
				return m, b.sendJoin()
			}
		}
		return m, nil

	// If we're receiving a ROOMSTATE message for the channel we wanted to join, we've
	// successfully joined that channel
	case "ROOMSTATE":
		if includes(m.Params, b.channel) {
			b.gotRoomState = true
		}
		return m, nil

	// If we get a PRIVMSG prefixed with '!', attempt to parse it as a command
	case "PRIVMSG":
		if includes(m.Params, b.channel) && len(m.Body) > 1 && m.Body[0] == '!' {
			command := m.Body[1:]
			args := ""
			if spacePos := strings.IndexRune(m.Body, ' '); spacePos >= 2 {
				command = m.Body[1:spacePos]
				args = m.Body[spacePos+1:]
			}

			userId := m.Extra["user-id"]
			displayName := m.Extra["display-name"]
			if userId != "" && displayName != "" {
				go func() {
					if err := b.commandHandler.Handle(command, args, userId, displayName); err != nil {
						b.conn.Sendf("PRIVMSG %s :%s", b.channel, err)
					}
				}()
			}
		}
	}

	// All message types not explicitly handled are considered OK
	return m, nil
}

func (b *bot) sendJoin() error {
	return b.conn.Sendf("JOIN %s", b.channel)
}

func (b *bot) fail(err error) {
	b.signalError(err)
	b.err = err
}

func (b *bot) GetStatus() chatbot.Status {
	if b.err != nil {
		return chatbot.StatusDisconnected
	}
	if b.gotCapAck && b.gotGlobalUserState && b.gotRoomState {
		return chatbot.StatusConnected
	}
	return chatbot.StatusConnecting
}

func (b *bot) GetLastError() error {
	return b.err
}

func (b *bot) GetLastPingTime() time.Time {
	return b.lastPingTime
}

func includes(params []string, s string) bool {
	for _, p := range params {
		if p == s {
			return true
		}
	}
	return false
}
