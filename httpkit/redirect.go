// Copyright (c) 2015-2023 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package httpkit

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var (
	// Since v2.8.0
	ErrAutoRedirectDisabled = errors.New("auto redirect is disabled")
)

type (
	// RedirectPolicy to regulate the redirects in the resty client.
	// Objects implementing the RedirectPolicy interface can be registered as
	//
	// Apply function should return nil to continue the redirect journey, otherwise
	// return error to stop the redirect.
	RedirectPolicy interface {
		Apply(req *http.Request, via []*http.Request) error
	}

	// The RedirectPolicyFunc type is an adapter to allow the use of ordinary functions as RedirectPolicy.
	// If f is a function with the appropriate signature, RedirectPolicyFunc(f) is a RedirectPolicy object that calls f.
	RedirectPolicyFunc func(*http.Request, []*http.Request) error
)

// Apply calls f(req, via).
func (f RedirectPolicyFunc) Apply(req *http.Request, via []*http.Request) error {
	return f(req, via)
}

// NoRedirectPolicy is used to disable redirects in the HTTP client
//
//	resty.SetRedirectPolicy(NoRedirectPolicy())
func NoRedirectPolicy() RedirectPolicy {
	return RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return ErrAutoRedirectDisabled
	})
}

// FlexibleRedirectPolicy is convenient method to create No of redirect policy for HTTP client.
//
//	resty.SetRedirectPolicy(FlexibleRedirectPolicy(20))
func FlexibleRedirectPolicy(noOfRedirect int) RedirectPolicy {
	return RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		if len(via) >= noOfRedirect {
			return fmt.Errorf("stopped after %d redirects", noOfRedirect)
		}
		checkHostAndAddHeaders(req, via[0])
		return nil
	})
}

// DomainCheckRedirectPolicy is convenient method to define domain name redirect rule in resty client.
// Redirect is allowed for only mentioned host in the policy.
//
//	resty.SetRedirectPolicy(DomainCheckRedirectPolicy("host1.com", "host2.org", "host3.net"))
func DomainCheckRedirectPolicy(hostnames ...string) RedirectPolicy {
	hosts := make(map[string]bool)
	for _, h := range hostnames {
		hosts[strings.ToLower(h)] = true
	}

	fn := RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		if ok := hosts[getHostname(req.URL.Host)]; !ok {
			return errors.New("redirect is not allowed as per DomainCheckRedirectPolicy")
		}

		return nil
	})

	return fn
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package Unexported methods
//_______________________________________________________________________

func getHostname(host string) (hostname string) {
	if strings.Index(host, ":") > 0 {
		host, _, _ = net.SplitHostPort(host)
	}
	hostname = strings.ToLower(host)
	return
}

// By default Golang will not redirect request headers
// after go throwing various discussion comments from thread
// https://github.com/golang/go/issues/4800
// Resty will add all the headers during a redirect for the same host
func checkHostAndAddHeaders(cur *http.Request, pre *http.Request) {
	curHostname := getHostname(cur.URL.Host)
	preHostname := getHostname(pre.URL.Host)
	if strings.EqualFold(curHostname, preHostname) {
		for key, val := range pre.Header {
			cur.Header[key] = val
		}
	} else { // only library User-Agent header is added
		cur.Header.Set(hdrUserAgentKey, hdrUserAgentValue)
	}
}
