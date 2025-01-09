package ssoc

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/problem"
	"io"
	"net"
	"net/http"
)

type NoopTransportRouter struct {
}

func (n NoopTransportRouter) Error() error {
	return fmt.Errorf("noop transport router")
}

func (n NoopTransportRouter) GET(path string, handle fasthttp.RequestHandler) error {
	return n.Error()
}

func (n NoopTransportRouter) POST(path string, handle fasthttp.RequestHandler) error {
	return n.Error()
}

func (n NoopTransportRouter) DELETE(path string, handle fasthttp.RequestHandler) error {
	return n.Error()
}

func (n NoopTransportRouter) PUT(path string, handle fasthttp.RequestHandler) error {
	return n.Error()
}

func (n NoopTransportRouter) Handle(method, path string, handle fasthttp.RequestHandler) error {
	return n.Error()
}

func (n NoopTransportRouter) Bad(ctx *fasthttp.RequestCtx, code int, opt ...func(*problem.Problem)) {
}

func (n NoopTransportRouter) Then(f func(*fasthttp.RequestCtx) error) func(*fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		return
	}
}

type NoopTransport struct{}

func (NoopTransport) Broker() (net.IP, int) { return nil, 0 }

func (NoopTransport) R() layer.RouterType {
	return NoopTransportRouter{}
}

func (n NoopTransport) Node() string {
	return ""
}

func (n NoopTransport) Tags() []string {
	return nil
}

func (n NoopTransport) Doer(prefix string) (layer.Doer, error) {
	return nil, fmt.Errorf("noop transport")
}

func (n NoopTransport) Oneway(path string, reader io.Reader, header http.Header) error {
	return fmt.Errorf("noop transport")
}

func (n NoopTransport) Fetch(path string, reader io.Reader, header http.Header) (*http.Response, error) {
	return nil, fmt.Errorf("noop transport")
}

func (n NoopTransport) JSON(path string, data interface{}, result interface{}) error {
	return fmt.Errorf("noop transport")
}

func (n NoopTransport) Push(path string, data interface{}) error {
	return fmt.Errorf("noop transport")
}

func (n NoopTransport) OnConnect(name string, todo func() error) {

}

func (n NoopTransport) Stream(ctx context.Context, s string, header http.Header) (*websocket.Conn, error) {
	return nil, fmt.Errorf("noop transport")
}

func (n NoopTransport) Attachment(name string) (layer.Attachment, error) {
	return nil, fmt.Errorf("noop transport")
}
