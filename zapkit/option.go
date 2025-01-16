package zapkit

func Console() func(option *Config) {
	return func(opt *Config) {
		opt.Console = true
	}
}

func Caller(skip int, flag bool) func(*Config) {
	return func(opt *Config) {
		opt.Caller = flag
		opt.Skip = skip
	}
}

func File(file string) func(*Config) {
	return func(opt *Config) {
		opt.Filename = file
	}
}

func Max(size, age, backup int) func(*Config) {
	return func(opt *Config) {
		opt.MaxSize = size
		opt.MaxAge = age
		opt.MaxBackups = backup
	}
}

func Compress(flag bool) func(*Config) {
	return func(opt *Config) {
		opt.Compress = flag
	}
}
