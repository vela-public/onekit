package lua

type Options2[T any] struct {
	CallStackSize       int
	RegistrySize        int
	SkipOpenLibs        bool
	MinimizeStackMemory bool
	RegistryGrowStep    int
	RegistryMaxSize     int
	Exdata              T
}

func (opt *Options2[T]) Unwrap() *Options {
	return &Options{
		CallStackSize:       opt.CallStackSize,
		RegistrySize:        opt.RegistrySize,
		SkipOpenLibs:        opt.SkipOpenLibs,
		MinimizeStackMemory: opt.MinimizeStackMemory,
		RegistryGrowStep:    opt.RegistryGrowStep,
		RegistryMaxSize:     opt.RegistryMaxSize,
		Exdata:              opt.Exdata,
	}
}
