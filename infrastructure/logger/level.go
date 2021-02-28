package logger

import "strings"

// Level is the level at which a logger is configured. All messages sent
// to a level which is below the current level are filtered.
type Level uint32

// Level constants.
const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
	LevelOff
)

// levelStrs defines the human-readable names for each logging level.
var levelStrs = [...]string{"TRC", "DBG", "INF", "WRN", "ERR", "CRT", "OFF"}

// LevelFromString returns a level based on the input string s. If the input
// can't be interpreted as a valid log level, the info level and false is
// returned.
func LevelFromString(s string) (l Level, ok bool) {
	switch strings.ToLower(s) {
	case "trace", "trc":
		return LevelTrace, true
	case "debug", "dbg":
		return LevelDebug, true
	case "info", "inf":
		return LevelInfo, true
	case "warn", "wrn":
		return LevelWarn, true
	case "error", "err":
		return LevelError, true
	case "critical", "crt":
		return LevelCritical, true
	case "off":
		return LevelOff, true
	default:
		return LevelInfo, false
	}
}

// String returns the tag of the logger used in log messages, or "OFF" if
// the level will not produce any log output.
func (l Level) String() string {
	if l >= LevelOff {
		return "OFF"
	}
	return levelStrs[l]
}
