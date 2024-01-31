package irc

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// Conn is a low-level interface representing a connection to an IRC server. It simply
// handles TCP connection state and sending/receiving IRC messages in plain-text format:
// it does not concern itself with parsing those messages, authenticating itself as a
// particular user, joining channels, etc.
type Conn interface {
	Close()
	Recv() (<-chan string, error)
	Send(s string) error
	Sendf(format string, a ...any) error
}

// ConnOpts is the set of options used to configure a connection to an IRC server
type ConnOpts struct {
	Server string
	Dial   DialFunc
	Logger Logger
}

// DialFunc is a function that establishes a TCP connection to the given server. If the
// provided context is canceled during the connection attempt, the connection attempt
// should be aborted. Once the connection is established, ctx is no longer relevant:
// canceling it will NOT result in an automatic disconnect.
type DialFunc func(ctx context.Context, server string) (net.Conn, error)

func NewConn(ctx context.Context, opts ConnOpts) (Conn, error) {
	// Prepare default options if not explicitly specified
	if opts.Server == "" {
		opts.Server = "irc.chat.twitch.tv:6697"
	}
	if opts.Dial == nil {
		opts.Dial = func(ctx context.Context, server string) (net.Conn, error) {
			d := tls.Dialer{}
			return d.DialContext(ctx, "tcp", server)
		}
	}
	if opts.Logger == nil {
		opts.Logger = NewStreamLogger(os.Stdout)
	}

	// Initiate a TCP connection to the IRC server, giving us a bidirectional stream
	// which we can write to in order to send '\n'-delimited IRC messages, and which we
	// can read from in order to receive '\n'-delimited IRC messages
	tcpConn, err := opts.Dial(ctx, opts.Server)
	if err != nil {
		return nil, err
	}

	// Initialize an irc.Conn that will allow us to read from and write to our
	// connection in the form of plain-text IRC messages
	c := &conn{
		Conn:         tcpConn,
		reader:       bufio.NewReader(tcpConn),
		logger:       opts.Logger,
		messagesChan: make(chan string),
	}

	// Run a goroutine that will await new messages from the server for as long as our
	// connection is valid, sending them into messagesChan as they're received, and
	// closing the channel if any error occurs
	go func() {
		for {
			// Block until the next line is available
			s, err := c.reader.ReadString('\n')
			if err != nil {
				// If we failed to read, abort
				if !errors.Is(err, net.ErrClosed) {
					c.logger.LogError(fmt.Errorf("error reading from connection to IRC server: %w", err))
				}
				c.close()
				return
			}

			// The read succeeded, so handle the message by logging it and then writing
			// its raw text to our channel
			line := strings.TrimSuffix(s, "\n")
			c.logger.LogRecv(line)
			c.messagesChan <- line
		}
	}()

	// Run a goroutine that will close our connection when the parent context is
	// canceled
	go func() {
		<-ctx.Done()
		c.close()
	}()

	return c, nil
}

type conn struct {
	net.Conn
	reader       *bufio.Reader
	logger       Logger
	messagesChan chan string

	closed bool
	mu     sync.RWMutex
}

func (c *conn) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		err := c.Conn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			c.logger.LogError(fmt.Errorf("error closing connection to IRC server: %w", err))
		}
		close(c.messagesChan)
		c.closed = true
	}
}

func (c *conn) Close() {
	c.close()
}

func (c *conn) Recv() (<-chan string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("IRC client connection already closed")
	}
	return c.messagesChan, nil
}

func (c *conn) Send(s string) error {
	line := strings.TrimSuffix(s, "\n")
	c.logger.LogSend(line)
	if _, err := c.Write([]byte(line + "\n")); err != nil {
		return err
	}
	return nil
}

func (c *conn) Sendf(format string, a ...any) error {
	s := fmt.Sprintf(format, a...)
	return c.Send(s)
}
