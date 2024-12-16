package httpkit

import (
	"bufio"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Raw(text string) *Response {
	reader := bufio.NewReader(strings.NewReader(text))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return NewRespE(nil, err)
	}

	cli := New()
	r := cli.NewRequest()
	r.RawRequest = req

	resp, err := cli.execute(r)
	if err != nil {
		return NewRespE(r, err)
	}

	return resp
}

func FromRaw(raw string, options ...func(*RawHttp)) (*RawHttp, error) {

	hp := NewRawHTTP()
	_, err := hp.Parse(cast.S2B(raw))
	if err != nil {
		return nil, err
	}

	r := &RawHttp{}
	n := len(options)
	if n > 0 {
		for i := 0; i < n; i++ {
			options[i](r)
		}
	}

	r.Hostname = string(hp.Host())
	r.Path = hp.Path
	r.Method = hp.Method
	r.Proto = hp.Version
	r.EOL = "\r\n"
	n = len(hp.Headers)
	for i := 0; i < n; i++ {
		h := hp.Headers[i]
		name := cast.B2S(h.Name)

		if len(name) == 0 {
			break
		}

		if strings.ToLower(name) == "content-length" {
			continue
		}

		val := cast.B2S(h.Value)
		r.Headers = append(r.Headers, fmt.Sprintf("%s: %s", name, val))
	}

	r.Body = cast.B2S(hp.Body)
	r.AutoSetContentLength()

	return r, nil
}

// FromURL returns a *Request for a given method and URL and any
// error that occured during parsing the URL. Sane defaults are
// set for all of *Request's fields.
func FromURL(method, rawurl string) (*RawHttp, error) {
	r := &RawHttp{}

	u, err := url.Parse(rawurl)
	if err != nil {
		return r, err
	}

	// url.Parse() tends to mess with the path, so we need to
	// try and fix that.
	schemeEtc := strings.SplitN(rawurl, "//", 2)
	if len(schemeEtc) != 2 {
		return nil, fmt.Errorf("invalid url: %s", rawurl)
	}

	pathEtc := strings.SplitN(schemeEtc[1], "/", 2)
	path := "/"
	if len(pathEtc) == 2 {
		// Remove any query string or fragment
		path = "/" + pathEtc[1]
		noQuery := strings.Split(path, "?")
		noFragment := strings.Split(noQuery[0], "#")
		path = noFragment[0]
	}

	r.TLS = u.Scheme == "https"
	r.Method = method
	r.Scheme = u.Scheme
	r.Hostname = u.Hostname()
	r.Port = u.Port()
	r.Path = path
	r.Query = u.RawQuery
	r.Fragment = u.Fragment
	r.Proto = "HTTP/1.1"
	r.EOL = "\r\n"
	r.Timeout = time.Second * 30

	if r.Path == "" {
		r.Path = "/"
	}

	if r.Port == "" {
		if r.TLS {
			r.Port = "443"
		} else {
			r.Port = "80"
		}
	}

	return r, nil

}
