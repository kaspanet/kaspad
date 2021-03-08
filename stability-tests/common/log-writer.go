package common

import (
	"strings"

	"github.com/kaspanet/kaspad/infrastructure/logger"
)

// LogWriter writes to the given log with the given log level and prefix
type LogWriter struct {
	log    *logger.Logger
	level  logger.Level
	prefix string
}

func (clw LogWriter) Write(p []byte) (n int, err error) {
	logWithoutNewLine := strings.TrimSuffix(string(p), "\n")
	clw.log.Writef(clw.level, "%s: %s", clw.prefix, logWithoutNewLine)
	return len(p), nil
}

// NewLogWriter returns a new LogWriter that forwards to `log` all data written to it using at `level` level
func NewLogWriter(log *logger.Logger, level logger.Level, prefix string) LogWriter {
	return LogWriter{
		log:    log,
		level:  level,
		prefix: prefix,
	}
}
