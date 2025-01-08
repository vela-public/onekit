package pipekit

func NewChain[T any]() *Chain[T] {
	return &Chain[T]{}
}

/*

	chains := &Chains{}

	chains.H(func() error {
		return nil
	})

*/
