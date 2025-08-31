package config

import (
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

func (l LogLevel) String() string {
	return string(l)
}

func (l LogLevel) Zap() zap.AtomicLevel {
	switch strings.ToLower(strings.TrimSpace(l.String())) {
	case LogLevelDebug.String(), "trace":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case LogLevelInfo.String(), "information", "notice":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case LogLevelWarn.String(), "warning", "alert":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case LogLevelError.String(), "silent":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case LogLevelFatal.String(), "critical", "emergency":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	case LogLevelPanic.String():
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
}

func (c LogLevel) GORM() (level logger.LogLevel) {
	switch strings.ToLower(strings.TrimSpace(c.String())) {
	case LogLevelDebug.String(), "trace":
		return logger.Info
	case LogLevelInfo.String(), "information", "notice":
		return logger.Info
	case LogLevelWarn.String(), "warning", "alert":
		return logger.Warn
	case LogLevelError.String():
		return logger.Error
	case LogLevelFatal.String(), "critical", "emergency":
		return logger.Error
	case LogLevelPanic.String():
		return logger.Error
	case "silent":
		return logger.Silent
	default:
		return logger.Error
	}
}
