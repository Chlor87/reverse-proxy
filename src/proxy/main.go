package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"proxy/lib/proxy"
	"regexp"
	"strings"
)

const port = ":8080"

var imgReg = regexp.MustCompile(`<img[\s\S]*?src="([^"]+)`)

// convertResponse replaces all img's srcs' with Scarlet's pic
func convertResponse(body []byte) []byte {

	s := string(body)

	for _, v := range imgReg.FindAllStringSubmatch(s, -1) {
		if len(v) > 0 {
			s = strings.Replace(s, v[1], "http://bgwall.net/wp-content/uploads/2014/01/scarlett-johansson-close-up-cute-wallpaper.jpg", -1)
		}
	}

	return []byte(s)

}

func main() {
	r := http.NewServeMux()

	r.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		scarletify := strings.Contains(req.RequestURI, "/scarlet")
		req.RequestURI = strings.Replace(req.RequestURI, "/scarlet", "", -1)

		dst, ok := rules[req.RequestURI]
		if !ok {
			res.WriteHeader(http.StatusPaymentRequired)
			fmt.Fprint(res, "<h1>402 $$$</h1>")
			return
		}
		parsed, err := url.Parse(dst)
		if err != nil {
			res.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		req.URL.Host = parsed.Host
		req.URL.Path = ""

		p := proxy.New(dst)
		if scarletify {
			p.Hooks.ModResponseBody = convertResponse
		}

		p.ServeHTTP(res, req)

	})

	fmt.Println("Proxy listening on", port)

	log.Fatal(http.ListenAndServe(port, r))
}
