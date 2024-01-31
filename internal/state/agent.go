package state

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golden-vcr/chatbot"
	"github.com/golden-vcr/chatbot/internal/irc"
	"golang.org/x/exp/slog"
)

type Agent interface {
	Disconnect()
	Reinitialize(userAccessToken string, timeout time.Duration) error
	GetStatus() chatbot.Status
}

func NewAgent(ctx context.Context, logger *slog.Logger, channelName, botUsername string) Agent {
	return &agent{
		rootCtx:     ctx,
		logger:      logger,
		channelName: channelName,
		botUsername: botUsername,
	}
}

type agent struct {
	rootCtx     context.Context
	logger      *slog.Logger
	channelName string
	botUsername string

	conn irc.Conn
	bot  irc.Bot
	mu   sync.RWMutex
}

func (a *agent) Disconnect() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
		a.bot = nil
	}
}

func (a *agent) Reinitialize(userAccessToken string, timeout time.Duration) error {
	a.Disconnect()

	conn, err := irc.NewConn(a.rootCtx, irc.ConnOpts{
		Logger: irc.NewStructuredLogger(a.logger),
	})
	if err != nil {
		return err
	}
	b, err := irc.NewBot(a.rootCtx, conn, a.channelName, a.botUsername, userAccessToken)
	if err != nil {
		conn.Close()
		return err
	}

	// Block (with a timeout) until the bot has successfully connected to the IRC
	// server, authenticated, and joined the channel
	ctx, cancel := context.WithTimeout(a.rootCtx, timeout)
	defer cancel()
	ready := false
	for !ready {
		select {
		case <-ctx.Done():
			// If our context is done, kill the connection (abandoning the bot) and
			// return an error
			conn.Close()
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("bot failed to become ready after %0.2f seconds", timeout.Seconds())
			}
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// Periodically check the status of the bot: if it's encountered an error,
			// kill the connection (abandoning the bot) and abort with that error
			if err := b.GetLastError(); err != nil {
				return err
			}

			// If the bot is ready now, break out of the loop and finish initializing
			if b.GetStatus() == chatbot.StatusConnected {
				ready = true
			}

			// Otherwise, keep trying until we're ready, we hit an error, or the context
			// is canceled
		}
	}

	// If our bot successfully became ready, install it as the agent's new bot and
	// finish successfully
	a.mu.Lock()
	defer a.mu.Unlock()
	a.conn = conn
	a.bot = b
	return nil
}

func (a *agent) GetStatus() chatbot.Status {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.bot == nil {
		return chatbot.StatusDisconnected
	}
	return a.bot.GetStatus()
}
