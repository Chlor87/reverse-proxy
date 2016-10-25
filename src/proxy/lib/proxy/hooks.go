package proxy

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type ModResponseHeader func(*http.Header)
type ModResponseBody func([]byte) []byte

type Hooks struct {
	ModResponseHeader ModResponseHeader
	ModResponseBody   ModResponseBody
}

// OnResponse is a noop, if ModResponseHeader userland func is not implemented
func (h *Hooks) OnResponseHeader(res *http.Response) {
	if h.ModResponseHeader == nil {
		return
	}
	h.ModResponseHeader(&res.Header)
}

// OnResponseBody calls userland func ModResponseBody, if not implemented, this
// is a noop. If it is, OnResponseBody unsets content-encoding header, and sets
// proper content-length.
func (h *Hooks) OnResponseBody(res *http.Response) (err error) {

	if h.ModResponseBody == nil {
		return nil
	}

	raw, err := getBody(res)
	if err != nil {
		return
	}

	res.Header.Del("content-encoding")

	b, err := ioutil.ReadAll(raw)
	if err != nil {
		return
	}

	err = res.Body.Close()
	if err != nil {
		return
	}

	b = h.ModResponseBody(b)
	if err != nil {
		return
	}

	cLen := len(b)

	res.Body = ioutil.NopCloser(bytes.NewReader(b))
	res.ContentLength = int64(cLen)
	res.Header.Set("content-length", strconv.Itoa(cLen))

	return

}

// getBody assures, that the stream it returns is decoded properly,
// currently supports gzip encoding.
func getBody(res *http.Response) (body io.ReadCloser, err error) {
	switch res.Header.Get("content-encoding") {
	case "gzip":
		body, err = gzip.NewReader(res.Body)
	default:
		body = res.Body
	}
	return
}
