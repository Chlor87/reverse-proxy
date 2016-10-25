package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"proxy/lib/proxy"
	"regexp"
	"strconv"
	"strings"
)

const port = ":8080"

var targets = []string{
//	"https://youtube.com",
//	"https://www.facebook.com",
//	"http://onet.pl/",
}

var imgReg = regexp.MustCompile(`<img[\s\S]*?src="([^"]+)`)

func getBody(res *http.Response) (body io.ReadCloser, err error) {
	switch res.Header.Get("content-encoding") {
	case "gzip":
		body, err = gzip.NewReader(res.Body)
	default:
		body = res.Body
	}
	return
}

func convertResponse(res *http.Response) (err error) {
	raw, err := getBody(res)
	res.Header.Del("content-encoding")
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(raw)
	if err != nil {
		return
	}

	err = res.Body.Close()
	if err != nil {
		return
	}

	asString := string(b)

	for _, v := range imgReg.FindAllStringSubmatch(asString, -1) {
		if len(v) > 0 {
			asString = strings.Replace(asString, v[1], "http://bgwall.net/wp-content/uploads/2014/01/scarlett-johansson-close-up-cute-wallpaper.jpg", -1)
		}
	}

	res.Body = ioutil.NopCloser(bytes.NewReader([]byte(asString)))
	res.ContentLength = int64(len(asString))
	res.Header.Set("content-length", strconv.Itoa(len(asString)))
	return nil
}

func main() {
	r := http.NewServeMux()
	p := proxy.New(targets)
	p.ConvertResponse = convertResponse
	r.Handle("/", p)

	log.Fatal(http.ListenAndServe(port, r))
}
