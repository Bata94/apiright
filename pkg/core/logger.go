package core

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	DebugLevel  = zapcore.DebugLevel
	InfoLevel   = zapcore.InfoLevel
	WarnLevel   = zapcore.WarnLevel
	ErrorLevel  = zapcore.ErrorLevel
	DPanicLevel = zapcore.DPanicLevel
	PanicLevel  = zapcore.PanicLevel
	FatalLevel  = zapcore.FatalLevel
)

type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	DPanic(msg string, fields ...any)
	Panic(msg string, fields ...any)
	Fatal(msg string, fields ...any)
	With(fields ...any) Logger
	Sync() error
}

type ZapLogger struct {
	logger *zap.Logger
}

func NewLogger(development bool) (*ZapLogger, error) {
	var cfg zap.Config
	if development {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		logger: logger,
	}, nil
}

func NewLoggerWithLevel(level string, development bool) (*ZapLogger, error) {
	var cfg zap.Config
	if development {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = zapcore.InfoLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		logger: logger,
	}, nil
}

func (l *ZapLogger) Debug(msg string, fields ...any) {
	if len(fields) == 0 {
		l.logger.Debug(msg)
	} else {
		l.logger.Debug(msg, l.fieldsToPairs(fields)...)
	}
}

func (l *ZapLogger) Info(msg string, fields ...any) {
	if len(fields) == 0 {
		l.logger.Info(msg)
	} else {
		l.logger.Info(msg, l.fieldsToPairs(fields)...)
	}
}

func (l *ZapLogger) Warn(msg string, fields ...any) {
	if len(fields) == 0 {
		l.logger.Warn(msg)
	} else {
		l.logger.Warn(msg, l.fieldsToPairs(fields)...)
	}
}

func (l *ZapLogger) Error(msg string, fields ...any) {
	if len(fields) == 0 {
		l.logger.Error(msg)
	} else {
		l.logger.Error(msg, l.fieldsToPairs(fields)...)
	}
}

func (l *ZapLogger) DPanic(msg string, fields ...any) {
	if len(fields) == 0 {
		l.logger.DPanic(msg)
	} else {
		l.logger.DPanic(msg, l.fieldsToPairs(fields)...)
	}
}

func (l *ZapLogger) Panic(msg string, fields ...any) {
	if len(fields) == 0 {
		l.logger.Panic(msg)
	} else {
		l.logger.Panic(msg, l.fieldsToPairs(fields)...)
	}
}

func (l *ZapLogger) Fatal(msg string, fields ...any) {
	if len(fields) == 0 {
		l.logger.Fatal(msg)
	} else {
		l.logger.Fatal(msg, l.fieldsToPairs(fields)...)
	}
}

func (l *ZapLogger) With(fields ...any) Logger {
	return &ZapLogger{
		logger: l.logger.With(l.fieldsToPairs(fields)...),
	}
}

func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

func (l *ZapLogger) Logger() *zap.Logger {
	return l.logger
}

func (l *ZapLogger) fieldsToPairs(fields []any) []zap.Field {
	pairs := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		pairs = append(pairs, zap.Any(key, fields[i+1]))
	}
	return pairs
}

func String(key string, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int32(key string, val int32) zap.Field {
	return zap.Int32(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Uint(key string, val uint) zap.Field {
	return zap.Uint(key, val)
}

func Uint32(key string, val uint32) zap.Field {
	return zap.Uint32(key, val)
}

func Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

func Float32(key string, val float32) zap.Field {
	return zap.Float32(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Duration(key string, val any) zap.Field {
	return zap.Any(key, val)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, val any) zap.Field {
	return zap.Any(key, val)
}

func Skip() zap.Field {
	return zap.Skip()
}

func NamedError(key string, err error) zap.Field {
	return zap.NamedError(key, err)
}

func Time(key string, val any) zap.Field {
	return zap.Any(key, val)
}

func Stack(key string) zap.Field {
	return zap.Stack(key)
}

func Strings(key string, val []string) zap.Field {
	return zap.Strings(key, val)
}

func Int64s(key string, val []int64) zap.Field {
	return zap.Int64s(key, val)
}
