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
	able := zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
		return lv >= zapcore.DebugLevel
	})

	sync := zapcore.AddSync(os.Stdout)
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(cfg)
	core := zapcore.NewCore(encoder, sync, able)
	sugar := zap.New(core, zap.AddCallerSkip(1), zap.AddCaller()).Sugar()
	return &Logger{
		sugar: sugar,
	}
}
