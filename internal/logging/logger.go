package logging

import (
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

var (
	// Logger is the global logger instance
	Logger *slog.Logger
)

type Options struct {
	JSON    bool
	Quiet   bool
	Verbose bool
	Output  io.Writer
}

// InitLogger initializes the logger with the specified level
func InitLogger(opts Options) {
	var logLevel slog.Level
	switch {
	case opts.Quiet:
		logLevel = slog.LevelError
	case opts.Verbose:
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(logLevel)
	if opts.JSON {
		Logger = slog.New(slog.NewJSONHandler(opts.Output, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: logLevel == slog.LevelDebug,
		}))
	} else {
		Logger = slog.New(tint.NewHandler(opts.Output, &tint.Options{
			Level:      logLevel,
			TimeFormat: time.RFC3339,
			AddSource:  logLevel == slog.LevelDebug,
		}))
	}
}
