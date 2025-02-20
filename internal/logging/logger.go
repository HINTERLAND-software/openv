package logging

import (
	"log/slog"
	"os"
)

var (
	// Logger is the global logger instance
	Logger *slog.Logger
)

// LogLevel represents the logging level
type LogLevel string

const (
	// Debug level for verbose output
	Debug LogLevel = "debug"
	// Info level for normal output
	Info LogLevel = "info"
	// Warn level for warning messages
	Warn LogLevel = "warn"
	// Error level for error messages
	Error LogLevel = "error"
)

// InitLogger initializes the logger with the specified level
func InitLogger(level LogLevel) {
	var logLevel slog.Level
	switch level {
	case Debug:
		logLevel = slog.LevelDebug
	case Info:
		logLevel = slog.LevelInfo
	case Warn:
		logLevel = slog.LevelWarn
	case Error:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	Logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(Logger)
}
