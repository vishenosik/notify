// Package logs provides a structured logging wrapper around slog.
// It supports two build modes: development (dev_log) and production.
// The package integrates with traceid for request tracing and allows
// configuration of log levels and output destinations.
package logs

import (
	"log/slog"
	"os"

	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/traceid"
	"github.com/golang-cz/devslog"
)

var LogSchema = httplog.SchemaECS

type EnvChecker interface {
	IsDev() bool
	String() string
}

// LoggerConfig holds the configuration parameters for initializing the logger.
type LoggerConfig struct {
	// Name of the application writing the logs.
	AppName string
	// Environment identifier (e.g., "production", "develop").
	Env EnvChecker
	// Enable debug level logs.
	EnableDebug bool
}

func NewLogger(conf LoggerConfig) *slog.Logger {
	var handler slog.Handler

	level := slog.LevelInfo
	if conf.EnableDebug {
		level = slog.LevelDebug
	}

	if conf.Env.IsDev() {
		handler = dev(level)
	} else {
		handler = prod(level)
	}

	logger := slog.New(
		traceid.LogHandler(handler),
	)

	if !conf.Env.IsDev() {
		return logger.With(
			slog.String("app", conf.AppName),
			slog.String("env", conf.Env.String()),
		)
	}

	return logger
}

// newLogger creates new instance of slog.Logger for dev environment.
// Logs are formatted as dev logs and written into os.Stdout.
// Every message goes with trace ID.
// If a flag `enable--log-file` is provided logs are also written into rotated files.
func dev(level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{
		ReplaceAttr: LogSchema.Concise(true).ReplaceAttr,
		Level:       level,
	}

	return devslog.NewHandler(os.Stdout, &devslog.Options{
		SortKeys:           true,
		MaxErrorStackTrace: 5,
		MaxSlicePrintSize:  20,
		HandlerOptions:     opts,
	})
}

func prod(level slog.Level) slog.Handler {
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: LogSchema.ReplaceAttr,
		Level:       level,
	})
}
