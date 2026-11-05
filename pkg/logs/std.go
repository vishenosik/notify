package logs

import (
	"fmt"
	"log"
	"log/slog"
	"strings"
)

type stdLogger struct {
	logger *slog.Logger
}

func NewStdLogger(logger *slog.Logger) *stdLogger {
	return &stdLogger{
		logger: logger,
	}
}

func (sl *stdLogger) Fatalf(format string, v ...any) {
	sl.logger.Error(fmt.Sprintf(format, v...))
}

func (sl *stdLogger) Printf(format string, v ...any) {
	sl.logger.Info(fmt.Sprintf(format, v...))
}

func (w *stdLogger) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	msg := strings.TrimSuffix(string(p), "\n")
	w.logger.Debug(msg)
	return len(p), nil
}

func RedirectStdLogger(logger *slog.Logger) {
	wrapper := &stdLogger{logger: logger}
	log.SetOutput(wrapper)
	// Optionally set flags to 0 to prevent std log from adding prefixes
	log.SetFlags(0)
}
