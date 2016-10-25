package proxy

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
)

type ResponseConverter func(*http.Response) error

type Proxy struct {
	*httputil.ReverseProxy
	ConvertResponse ResponseConverter
	origHost        string
}

type transport struct {
	http.RoundTripper
	*http.Client
	ConvertResponse *ResponseConverter
}

func (t *transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	res, err = t.RoundTripper.RoundTrip(req)
	if loc := res.Header.Get("location"); loc != "" {
		res, err = t.follow(req.Method, loc, res.Body)
		if err != nil {
			return
		}
	}
	if *t.ConvertResponse != nil {
		err = (*t.ConvertResponse)(res)
	}
	return
}

func (t *transport) follow(method, target string, body io.Reader) (res *http.Response, err error) {
	req, err := http.NewRequest(method, target, body)
	if err != nil {
		return
	}
	res, err = t.Do(req)
	if err != nil {
		return
	}
	if loc := res.Header.Get("location"); loc != "" {
		return t.follow(req.Method, loc, res.Body)
	}
	return
}

func New(targets []string) *Proxy {
	target, err := url.Parse(targets[rand.Int()%len(targets)])
	if err != nil {
		panic(err)
	}

	rp := httputil.NewSingleHostReverseProxy(target)
	oldDirector := rp.Director
	rp.Director = func(req *http.Request) {
		req.Host = target.Host
		oldDirector(req)
		log.Printf("%s\t%s", req.Method, req.URL.String())
	}

	p := &Proxy{ReverseProxy: rp}
	jar, _ := cookiejar.New(nil)
	t := &transport{http.DefaultTransport, &http.Client{Jar: jar}, &p.ConvertResponse}
	p.ReverseProxy.Transport = t
	return p

}
