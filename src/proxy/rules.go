package main

type Rules map[string]string

var rules = Rules{
	"/onet.pl": "http://www.onet.pl/",
	"/local":   "http://localhost:9999",
}
