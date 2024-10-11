package todo

// Result 模拟 Rust 的 Result 类型
type Result[T, E any] struct {
	Value T
	Error E
	Ok    bool
}

// Ok 返回表示成功的 Result
func Ok[T, E, U any](value T) Result[T, E] {
	return Result[T, E]{Value: value, Ok: true}
}

// Err 返回表示错误的 Result
func Err[T, E any](err E) Result[T, E] {
	return Result[T, E]{Error: err, Ok: false}
}

// Unwrap 返回成功的值，如果 Result 是错误则引发 panic
func (r Result[T, E]) Unwrap() (t T, ok bool) {
	if !r.Ok {
		return r.Value, false
		//panic("called `Unwrap` on an `Err` value")
	}
	return r.Value, true
}

// UnwrapErr 返回错误值，如果 Result 是成功则引发 panic
func (r Result[T, E]) Err() E {
	if r.Ok {
		panic("called `UnwrapErr` on an `Ok` value")
	}
	return r.Error
}

// Expect 返回成功的值，否则引发带有消息的 panic
func (r Result[T, E]) Expect(fn func(E)) T {
	if !r.Ok {
		fn(r.Error)
	}

	return r.Value
}

func Then[T, E, U any](r Result[T, E], fn func(T) U) Result[U, E] {
	if r.Ok {
		// 当 r 是成功状态时，应用函数 f 并构造新的成功 Result
		return Result[U, E]{Value: fn(r.Value), Ok: true}
	}
	// 当 r 是错误状态时，直接返回新的错误 Result（复制原错误）
	return Result[U, E]{Error: r.Error, Ok: false}
}
