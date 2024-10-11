package noop

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type Logger struct {
	sugar *zap.SugaredLogger
}

func (l *Logger) Errorf(s string, a ...any) {
	l.sugar.Errorf(s, a...)
}

func (l *Logger) Warnf(s string, a ...any) {
	l.sugar.Warnf(s, a...)
}

func (l *Logger) Debugf(s string, a ...any) {
	l.sugar.Debugf(s, a...)
}

func (l *Logger) Infof(s string, a ...any) {
	l.sugar.Infof(s, a...)
}

func NewLogger() *Logger {
	encoder := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), os.Stdout, zapcore.DebugLevel)
	return &Logger{
		sugar: zap.New(core).WithOptions().Sugar(),
	}
}
