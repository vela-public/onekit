package todo

type Option[T, E, U any] struct {
	Value T
	Error E
	Ok    bool
}

func (o *Option[T, E, U]) Result() *Result[T, E] {
	return &Result[T, E]{
		Value: o.Value,
		Error: o.Error,
		Ok:    o.Ok,
	}
}

func (o *Option[T, E, U]) Unwrap() (t T, ok bool) {
	if o.Ok {
		return o.Value, true
	}
	return o.Value, false
}

func (o *Option[T, E, U]) Err() E {
	return o.Error
}

func (o *Option[T, E, U]) Map(fn func(T) U) *Result[U, E] {
	if o.Ok {
		u := fn(o.Value)
		return &Result[U, E]{
			Value: u,
			Error: o.Error,
			Ok:    true,
		}
	}

	var u U
	return &Result[U, E]{
		Value: u,
		Error: o.Error,
		Ok:    false,
	}
}
