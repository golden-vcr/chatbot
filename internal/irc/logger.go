package irc

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

type Logger interface {
	LogSend(s string)
	LogRecv(s string)
	LogError(err error)
}

func NewStreamLogger(stream io.Writer) Logger {
	if stream == nil {
		stream = os.Stdout
	}
	return &streamLogger{
		w: stream,
	}
}

func NewStructuredLogger(logger *slog.Logger) Logger {
	if logger == nil {
		logger = slog.Default()
	}
	return &structuredLogger{
		logger: logger,
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

type structuredLogger struct {
	logger *slog.Logger
}

func (l *structuredLogger) LogSend(s string) {
	l.logger.Info("Sending IRC message", "message", redactSend(s))
}

func (l *structuredLogger) LogRecv(s string) {
	l.logger.Info("Received IRC message", "message", s)
}

func (l *structuredLogger) LogError(err error) {
	if err != nil {
		l.logger.Error("IRC client error", "error", err)
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
