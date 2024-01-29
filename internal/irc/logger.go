package irc

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Logger interface {
	LogSend(s string)
	LogRecv(s string)
	LogError(err error)
}

func NewLogger(stream io.Writer) Logger {
	if stream == nil {
		stream = os.Stdout
	}
	return &streamLogger{
		w: stream,
	}
}

type streamLogger struct {
	w io.Writer
}

func (l *streamLogger) LogSend(s string) {
	fmt.Fprintf(l.w, "> %s\n", redactSend(s))
}

func (l *streamLogger) LogRecv(s string) {
	fmt.Fprintf(l.w, "< %s\n", s)
}

func (l *streamLogger) LogError(err error) {
	if err != nil {
		fmt.Fprintf(l.w, "ERROR: %v\n", err)
	}
}

func redactSend(s string) string {
	for _, prefix := range []string{"PASS oauth:", "PASS "} {
		if strings.HasPrefix(s, prefix) {
			return prefix + "<REDACTED>"
		}
	}
	return s
}
