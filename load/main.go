package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	count = flag.Int("count", 16, "client count")
	sleep = flag.Int("sleep", 200, "milliseconds to sleep")
	prime = flag.Int("prime", 10000, "calculate largest prime less than")
	bloat = flag.Int("bloat", 2, "mb of memory to consume")
	url   = flag.String("url", "http://app.kubecon-seattle-2018.josephburnett.com", "endpoint to get")
)

type client struct {
	lastResponse string
}

func main() {
	flag.Parse()
	if *count < 1 {
		panic("count must be at least 1")
	}
	resp, err := http.Get(*url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", string(body))
}
