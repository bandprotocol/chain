package logger

import (
	"os"

	"github.com/kyokomi/emoji"
	"github.com/tendermint/tendermint/libs/log"
)

// Logger is a wrapper around the Tendermint logger.
type Logger struct {
	logger log.Logger
}

// NewLogger creates a new instance of the Logger.
func NewLogger(level log.Option) *Logger {
	return &Logger{logger: log.NewFilter(log.NewTMLogger(os.Stdout), level)}
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logger.Debug(emoji.Sprintf(format, args...))
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.logger.Info(emoji.Sprintf(format, args...))
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.logger.Error(emoji.Sprintf(format, args...))
}

// With adds additional key-value pairs to the logger.
func (l *Logger) With(keyvals ...interface{}) *Logger {
	return &Logger{logger: l.logger.With(keyvals...)}
}
