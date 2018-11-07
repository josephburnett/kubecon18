package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
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
	err          error
}

func (c *client) start(stopCh <-chan struct{}) {
	tickerCh := time.NewTicker(time.Second).C
	for {
		select {
		case <-tickerCh:
			urlWithParams := fmt.Sprintf("%v?sleep=%v&prime=%v&bloat=%v",
				*url, *sleep, *prime, *bloat)
			resp, err := http.Get(urlWithParams)
			if err != nil {
				c.err = err
				c.lastResponse = err.Error()
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.err = err
				c.lastResponse = err.Error()
				resp.Body.Close()
				continue
			}
			resp.Body.Close()
			c.err = nil
			c.lastResponse = strings.TrimSpace(string(body))
		case <-stopCh:
			return
		}
	}
}

func main() {
	flag.Parse()
	if *count < 1 {
		panic("count must be at least 1")
	}
	stopCh := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		close(stopCh)
	}()
	clients := make([]*client, *count)
	for i := 0; i < *count; i++ {
		c := &client{}
		go c.start((<-chan struct{})(stopCh))
		clients[i] = c
	}
	tickerCh := time.NewTicker(time.Second).C
	for {
		select {
		case <-tickerCh:
			for i, client := range clients {
				fmt.Printf("%v: %v\n", i, client.lastResponse)
			}
			fmt.Printf("\n\n\n")
		case <-stopCh:
			os.Exit(0)
		}
	}
}
