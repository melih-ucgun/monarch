package core

import (
	"context"
	"io"
	"log/slog"
	// Self import is tricky, usually we simply use the package pterm usage replacement
	// But since this is package core, I don't need to import core.
)

type DefaultLogger struct {
	level   LogLevel
	handler *slog.Logger
	ui      UI
}

func NewDefaultLogger(ui UI, output io.Writer, level LogLevel) *DefaultLogger {
	var slogLevel slog.Level
	switch level {
	case LevelTrace, LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: slogLevel,
	}))

	return &DefaultLogger{
		level:   level,
		handler: handler,
		ui:      ui,
	}
}

func (l *DefaultLogger) Trace(msg string, args ...any) {
	if l.level <= LevelTrace {
		l.ui.Debug("TRACE: " + msg)
		l.handler.Debug(msg, args...)
	}
}

func (l *DefaultLogger) Debug(msg string, args ...any) {
	if l.level <= LevelDebug {
		l.ui.Debug(msg)
		l.handler.Debug(msg, args...)
	}
}

func (l *DefaultLogger) Info(msg string, args ...any) {
	if l.level <= LevelInfo {
		l.ui.Info(msg)
		l.handler.Info(msg, args...)
	}
}

func (l *DefaultLogger) Warn(msg string, args ...any) {
	if l.level <= LevelWarn {
		l.ui.Warning(msg)
		l.handler.Warn(msg, args...)
	}
}

func (l *DefaultLogger) Error(msg string, args ...any) {
	if l.level <= LevelError {
		l.ui.Error(msg)
		l.handler.Error(msg, args...)
	}
}

func (l *DefaultLogger) With(args ...any) Logger {
	return &DefaultLogger{
		level:   l.level,
		handler: l.handler.With(args...),
		ui:      l.ui,
	}
}

func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// Dummy context to satisfy slogans
var ctx = context.Background()
