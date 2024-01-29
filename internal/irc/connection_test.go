package irc

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Connection(t *testing.T) {
	var c Conn = newMockConn()
	defer c.Close()

	receivedLines := make([]string, 0, 8)
	ch, err := c.Recv()
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for line := range ch {
			receivedLines = append(receivedLines, line)
		}
	}()

	err = c.Send("HELLO")
	assert.NoError(t, err)
	err = c.Send("PING")
	assert.NoError(t, err)
	err = c.Send("ECHO foo bar 42")
	assert.NoError(t, err)
	err = c.Send("PING")
	assert.NoError(t, err)
	err = c.Send("GOODBYE")
	assert.NoError(t, err)

	wg.Wait()
	assert.Equal(t, []string{
		"AHOY",
		"PONG",
		"ECHO foo bar 42",
		"PONG",
		"SAYONARA",
	}, receivedLines)
}

type mockConn struct {
	linesReceived chan string
	serverReplies chan string
	closed        bool
}

func newMockConn() *mockConn {
	c := &mockConn{
		linesReceived: make(chan string),
		serverReplies: make(chan string),
	}
	go func() {
		for line := range c.serverReplies {
			c.linesReceived <- line
			if line == "SAYONARA" {
				c.Close()
				break
			}
		}
	}()
	return c
}

func (c *mockConn) Close() {
	if !c.closed {
		c.closed = true
		close(c.linesReceived)
	}
}

func (c *mockConn) Recv() (<-chan string, error) {
	if c.closed {
		return nil, fmt.Errorf("already closed")
	}
	return c.linesReceived, nil
}

func (c *mockConn) Send(s string) error {
	if strings.HasPrefix(s, "ECHO") {
		c.serverReplies <- s
	} else if s == "PING" {
		c.serverReplies <- "PONG"
	} else if s == "HELLO" {
		c.serverReplies <- "AHOY"
	} else if s == "GOODBYE" {
		c.serverReplies <- "SAYONARA"
	}
	return nil
}

func (c *mockConn) Sendf(format string, a ...any) error {
	return c.Send(fmt.Sprintf(format, a...))
}
