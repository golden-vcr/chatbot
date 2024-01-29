package irc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golden-vcr/chatbot"
)

type Bot interface {
	GetStatus() chatbot.Status
	GetLastError() error
	GetLastPingTime() time.Time
}

func NewBot(ctx context.Context, conn Conn, channelName, username, userAccessToken string) (Bot, error) {
	lines, err := conn.Recv()
	if err != nil {
		return nil, err
	}

	b := &bot{
		conn:        conn,
		channel:     fmt.Sprintf("#%s", channelName),
		nick:        strings.ToLower(username),
		accessToken: userAccessToken,
	}
	go func() {
		for s := range lines {
			if err := b.handle(s); err != nil {
				b.fail(err)
				return
			}
		}
	}()

	if err := b.init(); err != nil {
		return nil, err
	}

	return b, nil
}

type bot struct {
	conn        Conn
	channel     string
	nick        string
	accessToken string

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

func (b *bot) handle(s string) error {
	// We should always respond to a PING message by immediately replying with a PONG
	if strings.HasPrefix(s, "PING ") {
		if err := b.conn.Send(strings.Replace(s, "PING ", "PONG ", 1)); err != nil {
			return err
		} else {
			b.lastPingTime = time.Now()
		}
		return nil
	}

	// If we get a NOTICE telling us our login failed, abort
	if !strings.Contains(s, "PRIVMSG") && strings.Contains(s, "NOTICE") && strings.Contains(s, "Login authentication failed") {
		return fmt.Errorf("Login authentication failed")
	}

	// If we're still in the init stage, we need to send a JOIN message, but only once
	// user login is complete
	hasSentJoin := b.gotCapAck && b.gotGlobalUserState

	// Check for acknowledgement of our CAP REQ and PASS/NICK messages
	if !b.gotCapAck && strings.Contains(s, "CAP * ACK :twitch.tv/commands twitch.tv/tags") {
		b.gotCapAck = true
	}
	if !b.gotGlobalUserState && strings.Contains(s, "GLOBALUSERSTATE") {
		b.gotGlobalUserState = true
	}

	// If we've just now gotten both CAP * ACK and GLOBALUSERSTATE, send a JOIN message
	// to join our desired channel
	if !hasSentJoin && b.gotCapAck && b.gotGlobalUserState {
		if err := b.conn.Sendf("JOIN %s", b.channel); err != nil {
			return err
		}
	}

	// If we're receiving a ROOMSTATE message for the channel we wanted to join, we've
	// successfully joined that channel
	if !b.gotRoomState && strings.Contains(s, fmt.Sprintf("ROOMSTATE %s", b.channel)) {
		b.gotRoomState = true
	}

	// If we've reached this point, then the message has been handled without error
	return nil
}

func (b *bot) fail(err error) {
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
