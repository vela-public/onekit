package zapkit

import (
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/todo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

type Logger struct {
	cfg   *Config
	out   *lumberjack.Logger
	core  zapcore.Core
	sugar *zap.SugaredLogger
}

func (l *Logger) Close() error {
	if l.out != nil {
		return l.out.Close()
	}
	return nil
}

func (l *Logger) Skip(n int) layer.LoggerType {
	return &Logger{
		cfg:   l.cfg,
		sugar: zap.New(l.core, zap.AddCallerSkip(n), zap.WithCaller(true)).Sugar(),
	}
}

func (l *Logger) Apply(cfg *Config) {
	old := &Logger{
		cfg:   l.cfg,
		core:  l.core,
		sugar: l.sugar,
		out:   l.out,
	}

	l.cfg = cfg
	//必要函数和等级要求
	encode := func(color bool) zapcore.Encoder {
		c := zap.NewProductionEncoderConfig()
		c.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(time.DateTime))
		}

		if color {
			c.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		return zapcore.NewConsoleEncoder(c)
	}(false)
	enable := zap.LevelEnablerFunc(func(v zapcore.Level) bool {
		return v >= cfg.Level
	})

	//输出到前台
	var console zapcore.Core
	if cfg.Console {
		console = zapcore.NewCore(encode, zapcore.AddSync(os.Stderr), enable)
	}

	//输出到文件
	if cfg.Filename != "" {
		w := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   cfg.Compress,
		}

		f := zapcore.NewCore(encode, zapcore.AddSync(w), enable)
		l.core = todo.IF(cfg.Console, zapcore.NewTee(f, console), console)
		l.out = w
	} else {
		l.core = console
	}

	if cfg.Caller {
		l.sugar = zap.New(l.core, zap.AddCallerSkip(cfg.Skip), zap.WithCaller(true)).Sugar()
	} else {
		l.sugar = zap.New(l.core).Sugar()
	}

	old.Close()
}

func Zap(level zapcore.Level, setting ...func(*Config)) *Logger {
	cfg := Default()
	for _, fn := range setting {
		fn(cfg)
	}

	cfg.Level = level
	log := &Logger{}
	log.Apply(cfg)

	return log
}

func Error(options ...func(*Config)) *Logger {
	return Zap(zapcore.ErrorLevel, options...)
}

func Info(options ...func(*Config)) *Logger {
	return Zap(zapcore.InfoLevel, options...)
}

func Debug(options ...func(*Config)) *Logger {
	return Zap(zapcore.DebugLevel, options...)
}
