package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

const port = ":9999"

var file []byte

func init() {
	var err error
	file, err = ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		cp := bytes.NewReader(file)
		_, err := cp.WriteTo(res)
		if err != nil {
			res.WriteHeader(http.StatusServiceUnavailable)
		}
	})
	log.Println("Echo server listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
