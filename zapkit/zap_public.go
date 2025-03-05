package zapkit

import "go.uber.org/zap/zapcore"

func (l *Logger) Save(level zapcore.Level, v ...interface{}) {
	l.sugar.Log(level, v...)
}

func (l *Logger) Savef(level zapcore.Level, format string, v ...interface{}) {
	l.sugar.Logf(level, format, v...)
}

func (l *Logger) Debug(i ...interface{}) {
	if l.sugar == nil {
		return
	}

	l.sugar.Debug(i...)
}

func (l *Logger) Info(i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Info(i...)
}

func (l *Logger) Warn(i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Warn(i...)
}

func (l *Logger) Error(i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Error(i...)
}

func (l *Logger) Panic(i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Panic(i...)
}

func (l *Logger) Fatal(i ...interface{}) {
	if l.sugar == nil {
		return
	}

	l.sugar.Fatal(i...)
}

func (l *Logger) Trace(i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Debug(i...)
}

func (l *Logger) Debugf(s string, i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Debugf(s, i...)
}

func (l *Logger) Infof(s string, i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Infof(s, i...)
}

func (l *Logger) Warnf(s string, i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Warnf(s, i...)
}

func (l *Logger) Errorf(s string, i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Errorf(s, i...)
}

func (l *Logger) Panicf(s string, i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Panicf(s, i...)
}

func (l *Logger) Fatalf(s string, i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Fatalf(s, i...)
}

func (l *Logger) Tracef(s string, i ...interface{}) {
	if l.sugar == nil {
		return
	}
	l.sugar.Debugf(s, i...)
}
