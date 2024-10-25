package strkit

type TrimFunc func(string) string

type Trim struct {
	Mask   string
	Text   string
	Handle []TrimFunc
}

func (t *Trim) ToMask() string {
	n := len(t.Handle)
	if n == 0 {
		return t.Text
	}

	if len(t.Text) == 0 {
		return ""
	}

	if len(t.Mask) > 0 {
		return t.Mask
	}

	mark := t.Text
	for i := 0; i < n; i++ {
		mark = t.Handle[i](mark)
	}

	t.Mask = mark
	return mark
}

func NewTrim(text string, options ...TrimFunc) *Trim {
	return &Trim{Text: text, Handle: options}
}
