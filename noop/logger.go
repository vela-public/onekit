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
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), os.Stdout, zapcore.DebugLevel)

	return &Logger{
		sugar: zap.New(core).WithOptions().Sugar(),
	}
}
