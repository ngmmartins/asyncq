package util

import "log/slog"

func ParseLogLevel(logLevel string) slog.Leveler {
	switch logLevel {
	case "Debug":
		return slog.LevelDebug
	case "Info":
		return slog.LevelInfo
	case "Warn":
		return slog.LevelWarn
	case "Error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
