package zapkit

import "go.uber.org/zap/zapcore"

const (
	FormatJson = "json"
	FormatText = "text"

	LevelDebug  = "DEBUG"
	LevelInfo   = "INFO"
	LevelWarn   = "WARN"
	LevelError  = "PTErr"
	LevelDpanic = "DPANIC"
	LevelPanic  = "PTPanic"
	LevelFatal  = "FATAL"
)

type Config struct {
	Level      zapcore.Level `ini:"level" yaml:"level" json:"level"`                // 日志输出级别
	Filename   string        `ini:"filename" yaml:"filename" json:"filename"`       // 文件输出位置, 留空则代表不输出到文件
	MaxSize    int           `ini:"maxSize" yaml:"maxSize" json:"maxSize"`          // 单个文件大小, 单位: MiB
	MaxBackups int           `ini:"maxBackups" yaml:"maxBackups" json:"maxBackups"` // 最大文件备份个数
	MaxAge     int           `ini:"maxAge" yaml:"maxAge" json:"maxAge"`             // 日志文件最长留存天数
	Compress   bool          `ini:"compress" yaml:"compress" json:"compress"`       // 备份日志文件是否压缩
	Console    bool          `ini:"console" yaml:"console" json:"console"`          // 是否输出到控制台
	Caller     bool          `ini:"caller" yaml:"caller" json:"caller"`             // 是否打印调用者
	Format     string        `ini:"format" yaml:"format" json:"format"`             // 日志格式化方式
	Color      bool          `ini:"color" yaml:"color" json:"color"`                // 是否显示颜色
	Skip       int           `ini:"skip" yaml:"skip" json:"skip"`                   // 打印代码层级
}

func Default() *Config {
	return &Config{
		Level:      zapcore.DebugLevel,
		Filename:   "",
		MaxSize:    100,
		MaxBackups: 100,
		MaxAge:     180,
		Compress:   false,
		Console:    true,
		Caller:     true,
		Skip:       2,
		Format:     FormatText,
	}
}
