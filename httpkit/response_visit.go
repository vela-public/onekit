package httpkit

func NewRespE(r *Request, e error) *Response {
	return &Response{Request: r, Err: e}
}
