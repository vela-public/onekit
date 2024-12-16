package httpkit

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// A Requester defines the bare minimum set of methods needed to make an HTTP request.
type Requester interface {
	// IsTLS should return true if the connection should be made using TLS
	IsTLS() bool

	Address() string

	// String should return the request as a string E.g:
	//   GET / HTTP/1.1\r\nHost:...
	String() string

	Connection() (net.Conn, error)

	// GetTimeout returns the timeout for a request
	GetTimeout() time.Duration
}

// Request is the main implementation of Requester. It gives you
// fine-grained control over just about everything to do with the
// request, but with the posibility of sane defaults.
type RawHttp struct {
	// TLS should be true if TLS should be used
	TLS bool

	// Method is the HTTP verb. E.g. GET
	Method string

	// Scheme is the protocol scheme. E.g. https
	Scheme string

	// Hostname is the hostname to connect to. E.g. localhost
	Hostname string

	// Port is the port to connect to. E.g. 80
	Port string

	// Path is the path to request. E.g. /security.txt
	Path string

	// Query is the query string of the path. E.g. q=searchterm&page=3
	Query string

	// Fragment is the bit after the '#'. E.g. pagesection
	Fragment string

	// Proto is the protocol specifier in the first line of the request.
	// E.g. HTTP/1.1
	Proto string

	// Headers is a slice of headers to send. E.g:
	//   []string{"Host: localhost", "Accept: text/plain"}
	Headers []string

	// Body is the 'POST' data to send. E.g:
	//   username=AzureDiamond&password=hunter2
	Body string

	// EOL is the string that should be used for line endings. E.g. \r\n
	EOL string

	// Deadline
	Timeout time.Duration

	NetConn net.Conn

	Peer string
}

// IsTLS returns true if TLS should be used
func (r RawHttp) IsTLS() bool {
	return r.TLS
}

func (r RawHttp) Connection() (net.Conn, error) {
	if r.NetConn == nil {
		return nil, ErrNotFound
	}

	return r.NetConn, nil
}

func (r RawHttp) Address() string {
	if len(r.Peer) != 0 {
		return r.Peer + ":" + r.Port
	}

	fnc := func(peer string, scheme string) string {
		peer = strings.TrimSpace(peer)
		if r.Scheme == "https" {
			return peer + ":443"
		}
		if r.Scheme == "http" {
			return peer + ":80"
		}
		return peer + ":80"
	}

	host := r.Host()
	peer, port, err := net.SplitHostPort(host)
	if err != nil {
		return fnc(peer, r.Scheme)
	}

	if len(port) == 0 {
		return fnc(peer, r.Scheme)
	}

	return peer + ":" + port
}

// Host returns the hostname:port pair to connect to
func (r RawHttp) Host() string {
	return r.Hostname + ":" + r.Port
}

// AddHeader adds a header to the *Request
func (r *RawHttp) AddHeader(h string) {
	r.Headers = append(r.Headers, h)
}

// Header finds and returns the value of a header on the request.
// An empty string is returned if no match is found.
func (r RawHttp) Header(search string) string {
	search = strings.ToLower(search)

	for _, header := range r.Headers {

		p := strings.SplitN(header, ":", 2)
		if len(p) != 2 {
			continue
		}

		if strings.ToLower(p[0]) == search {
			return strings.TrimSpace(p[1])
		}
	}
	return ""
}

// AutoSetHost adds a Host header to the request
// using the value of Request.Hostname
func (r *RawHttp) AutoSetHost() {
	n := len(r.Headers)
	for i := 0; i < n; i++ {
		item := r.Headers[i]
		p := strings.SplitN(item, ":", 2)
		if len(p) != 2 {
			continue
		}
		k := strings.ToLower(p[0])
		if k == "host" {
			return
		}
	}

	r.AddHeader(fmt.Sprintf("Host: %s", r.Hostname))
}

func (r *RawHttp) H(key, value string) {
	key = strings.ToLower(key)

	n := len(r.Headers)
	for i := 0; i < n; i++ {
		item := r.Headers[i]
		p := strings.SplitN(item, ":", 2)
		if len(p) != 2 {
			continue
		}
		k := strings.ToLower(p[0])
		if k == key {
			r.Headers[i] = fmt.Sprintf("%s: %s", key, value)
			return
		}
	}

	r.AddHeader(fmt.Sprintf("%s: %s", key, value))
}

// AutoSetContentLength adds a Content-Length header
// to the request with the length of Request.Body as the value
func (r *RawHttp) AutoSetContentLength() {
	if n := len(r.Body); n >= 0 {
		r.H("Content-Length", strconv.Itoa(n))
	}
}

// fullPath returns the path including query string and fragment
func (r RawHttp) fullPath() string {
	q := ""
	if r.Query != "" {
		q = "?" + r.Query
	}

	f := ""
	if r.Fragment != "" {
		f = "#" + r.Fragment
	}
	return r.Path + q + f
}

// URL forms and returns a complete URL for the request
func (r RawHttp) URL() string {
	return fmt.Sprintf(
		"%s://%s%s",
		r.Scheme,
		r.Host(),
		r.fullPath(),
	)
}

// RequestLine returns the request line. E.g. GET / HTTP/1.1
func (r RawHttp) RequestLine() string {
	return fmt.Sprintf("%s %s %s", r.Method, r.fullPath(), r.Proto)
}

// String returns a plain-text version of the request to be sent to the server
func (r RawHttp) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("%s%s", r.RequestLine(), r.EOL))

	for _, h := range r.Headers {
		b.WriteString(fmt.Sprintf("%s%s", h, r.EOL))
	}

	b.WriteString(r.EOL)

	b.WriteString(r.Body)

	return b.String()
}

// GetTimeout returns the timeout for a request
func (r RawHttp) GetTimeout() time.Duration {
	// default 30 seconds
	if r.Timeout == 0 {
		return time.Second * 30
	}
	return r.Timeout
}

// RawRequest is the most basic implementation of Requester. You should
// probably only use it if you're doing something *really* weird
type RawRequest struct {
	// TLS should be true if TLS should be used
	TLS bool

	// Hostname is the name of the host to connect to. E.g: localhost
	Hostname string

	// Port is the port to connect to. E.g.: 80
	Port string

	// Request is the actual message to send to the server. E.g:
	//   GET / HTTP/1.1\r\nHost:...
	Request string

	// Timeout for the request
	Timeout time.Duration
}

// IsTLS returns true if the connection should use TLS
func (r RawRequest) IsTLS() bool {
	return r.TLS
}

// Host returns the hostname:port pair
func (r RawRequest) Host() string {
	return r.Hostname + ":" + r.Port
}

// String returns the message to send to the server
func (r RawRequest) String() string {
	return r.Request
}

// GetTimeout returns the timeout for the request
func (r RawRequest) GetTimeout() time.Duration {
	// default 30 seconds
	if r.Timeout == 0 {
		return time.Second * 30
	}
	return r.Timeout
}
