package ssoc

type Options struct {
	name    string
	mode    string
	protect bool
}

func Protect(flag bool) func(*Options) {
	return func(opts *Options) {
		opts.protect = flag
	}
}

func Devel() func(*Options) {
	return func(opts *Options) {
		opts.mode = "devel"
	}
}

func Worker() func(*Options) {
	return func(opts *Options) {
		opts.mode = "worker"
	}
}
