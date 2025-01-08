package noop

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	INFO LogLevel = iota
	WARN
	ERROR
)

// Logger 定义日志结构
type Logger struct {
	skip int
}

func (l *Logger) Debug(i ...interface{}) {
}

func (l *Logger) Info(i ...interface{}) {
}

func (l *Logger) Warn(i ...interface{}) {
}

func (l *Logger) Error(i ...interface{}) {
}

func (l *Logger) Panic(i ...interface{}) {
}

func (l *Logger) Fatal(i ...interface{}) {
}

func (l *Logger) Trace(i ...interface{}) {
}

func (l *Logger) Panicf(s string, i ...interface{}) {

}

func (l *Logger) Fatalf(s string, i ...interface{}) {
}

func (l *Logger) Tracef(s string, i ...interface{}) {

}

// NewLogger 创建新的日志实例
func NewLogger(skip int) *Logger {
	return &Logger{
		skip: 1 + skip,
	}
}

// Log 打印日志
func (l *Logger) Log(level LogLevel, message string) {

	// 获取调用的文件名和行号
	_, file, line, ok := runtime.Caller(l.skip)
	if !ok {
		file = "unknown"
		line = 0
	}

	// 获取当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// 格式化日志信息
	logMessage := fmt.Sprintf("%s [%s] %s:%d %s", currentTime, levelToString(level), file, line, message)
	fmt.Fprintln(os.Stdout, logMessage)
}

// levelToString 将日志级别转换为字符串
func levelToString(level LogLevel) string {
	switch level {
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) Infof(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	l.Log(INFO, message)
}

func (l *Logger) Warnf(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	l.Log(WARN, message)
}

func (l *Logger) Errorf(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	l.Log(ERROR, message)
}

func (l *Logger) Debugf(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	l.Log(INFO, message)
}
