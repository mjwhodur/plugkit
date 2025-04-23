package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/client"
	"github.com/mjwhodur/plugkit/examples/3-rawstream-basic/shared"
	"github.com/mjwhodur/plugkit/helpers"
)

type SimpleStreamClient struct {
	c *client.RawStreamClient
}

func (s *SimpleStreamClient) Handle(kind string, _ *cbor.RawMessage) {
	switch kind {
	case "pong":
		fmt.Println("pong received")
	case "pong-3":
		fmt.Println("pong 3 received")
		s.c.Stop()
	default:
		panic("unknown kind " + kind)
	}
}

func (s *SimpleStreamClient) Mount(c *client.RawStreamClient) {
	s.c = c
}

func (s *SimpleStreamClient) CloseSignal() {
	// TODO implement me
	fmt.Println("close signal")
	time.Sleep(2 * time.Second)
}

func main() {
	wg := sync.WaitGroup{}

	s := client.NewRawStreamClient(&SimpleStreamClient{}, "./plugin")
	err := s.Start()
	if err != nil {
		panic(err)
	}
	wg.Add(1)
	go func() {
		s.Run()
		wg.Done()
	}()
	s.Send("ping", helpers.MustRaw(&shared.Ping{ID: 1}))
	s.Send("ping", helpers.MustRaw(&shared.Ping{ID: 2}))
	s.Send("ping", helpers.MustRaw(&shared.Ping{ID: 2}))
	time.Sleep(3 * time.Second)
	s.Stop()
	wg.Wait()
}
