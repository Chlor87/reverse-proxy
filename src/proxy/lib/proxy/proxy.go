package proxy

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	*httputil.ReverseProxy
	*Hooks
}

// transport uses RoundTripper for simple requests, Client for
// header location redirects (with cookiejar) and stores a reference to Proxy.Hooks,
// as RoundTrip is the only place, where Response can be modified.
type transport struct {
	http.RoundTripper
	*http.Client
	*Hooks
}

// RoundTrip calls default transport's RoundTripper, obtains the response and
// allows modifications via Hooks passed do Proxy. If the response contains
// location header, it will follow redirects untill it reaches destination.
func (t *transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	res, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return
	}

	res, err = t.follow(res, req)
	if err != nil {
		return
	}

	t.Hooks.OnResponseHeader(res)
	err = t.Hooks.OnResponseBody(res)
	return
}

// folow gets called by RoundTrip. It calls itself until no location header
// is returned. If no location headrer is present, this is a noop.
// @todo make it configurable as a flag or a policy.
func (t *transport) follow(res *http.Response, req *http.Request) (final *http.Response, err error) {
	loc := res.Header.Get("location")
	if loc == "" {
		final = res
		return
	}

	req, err = http.NewRequest(req.Method, loc, res.Body)
	if err != nil {
		return
	}

	final, err = t.Client.Do(req)
	if err != nil {
		return
	}
	return t.follow(final, req)
}

func New(target string) *Proxy {
	targetURI, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	rp := httputil.NewSingleHostReverseProxy(targetURI)
	oldDirector := rp.Director
	rp.Director = func(req *http.Request) {
		req.Host = targetURI.Host
		oldDirector(req)
		log.Printf("%s\t%s", req.Method, req.URL.String())
	}

	p := &Proxy{ReverseProxy: rp, Hooks: &Hooks{}}
	jar, _ := cookiejar.New(nil)
	t := &transport{http.DefaultTransport, &http.Client{Jar: jar}, p.Hooks}
	p.ReverseProxy.Transport = t
	return p

}
