package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

const uri = "http://localhost:8080/local"

func request(inputChan chan string, resChan chan time.Time, errChan chan error) {
	for target := range inputChan {
		res, err := http.Get(target)
		if err != nil {
			errChan <- err
		} else {
			resChan <- time.Now()
			res.Body.Close()
		}
	}
}

type counter struct {
	*sync.Mutex
	count int
}

func (c *counter) Inc() {
	c.Lock()
	c.count++
	c.Unlock()
}

func (c *counter) Clear() {
	c.Lock()
	c.count = 0
	c.Unlock()
}

func (c *counter) Get() int {
	return c.count
}

func NewCounter() *counter {
	return &counter{Mutex: &sync.Mutex{}}
}

func main() {
	inputChan := make(chan string)
	resChan := make(chan time.Time)
	errChan := make(chan error)

	c := NewCounter()

	for i := 0; i < 1000; i++ {
		go request(inputChan, resChan, errChan)
	}

	go func() {
		for {
			inputChan <- uri
		}
	}()

	go func() {
		for range time.Tick(time.Second) {
			fmt.Printf("RPS: %d\r", c.Get())
			c.Clear()
		}
	}()

	for {
		select {
		case err := <-errChan:
			panic(err)
		case <-resChan:
			c.Inc()
		}
	}

}
