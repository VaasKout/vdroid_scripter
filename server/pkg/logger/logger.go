// Package logger ...
package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// LogLevel ...
type LogLevel byte

// Levels of logs
const (
	INFO  LogLevel = 0
	DEBUG LogLevel = 1
	WARN  LogLevel = 2
	ERROR LogLevel = 3
)

func toSlogLevel(level LogLevel) slog.Level {
	switch level {
	case DEBUG:
		return slog.LevelDebug
	case WARN:
		return slog.LevelWarn
	case ERROR:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Logger ...
type Logger struct {
	*slog.Logger
	requestID string
}

// Options ...
type Options struct {
	RequestID string
}

// New instance of Logger
func New(level LogLevel, isDev bool) *Logger {
	handler := tint.NewTextHandler(os.Stdout, &tint.Options{
		Level: toSlogLevel(level),
	})

	if !isDev {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: toSlogLevel(level),
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// SetOptions ...
func (l *Logger) SetOptions(opts *Options) *Logger {
	l.requestID = opts.RequestID

	return l
}

// GetRequestID ...
func (l *Logger) GetRequestID() string {
	return l.requestID
}

// Info ...
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, l.prepareArgs(args...)...)
}

// Debug ...
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, l.prepareArgs(args...)...)
}

// Warn ...
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, l.prepareArgs(args...)...)
}

// Error ...
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, l.prepareArgs(args...)...)
}

func (l *Logger) prepareArgs(args ...any) []any {
	if l.requestID != "" {
		args = append(args, String("request-id", l.requestID))
	}
	return args
}
