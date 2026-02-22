package config

import "log/slog"

type LogLevel = string

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

type Leveler struct {
	level LogLevel
}

var _ slog.Leveler = (*Leveler)(nil)

func NewLeveler(level LogLevel) *Leveler {
	return &Leveler{level: level}
}

func (l *Leveler) Level() slog.Level {
	switch l.level {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo // default to info if invalid level is provided
	}
}
