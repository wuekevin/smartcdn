package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var global *zap.Logger

// Init initializes the global structured logger.
// Log level is controlled by LOG_LEVEL env var (debug/info/warn/error); defaults to info.
func Init() error {
	return InitWithLevel(os.Getenv("LOG_LEVEL"))
}

// InitWithLevel initializes the global logger at the given level string.
func InitWithLevel(levelStr string) error {
	level := parseLevel(levelStr)

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "module",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)

	global = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(s) {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// L returns the global logger, initializing with defaults if needed.
func L() *zap.Logger {
	if global == nil {
		_ = Init()
	}
	return global
}

// Named returns a named logger (sets the "module" field).
func Named(name string) *zap.Logger {
	return L().Named(name)
}

// With returns a logger with extra fields pre-attached.
func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}

// Sync flushes buffered log entries.
func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}
