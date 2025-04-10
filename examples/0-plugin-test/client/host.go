package main

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/client"
	"github.com/mjwhodur/plugkit/examples/0-plugin-test/shared"
)

func main() {
	fmt.Println("Hello, client")

	c := client.NewClient("./plugin")
	c.HandleMessageType("pong", Pong)
	err := c.StartLocal()
	if err != nil {
		return
	}
	c.RunCommand("ping", &shared.Ping{}) //nolint:errcheck
	c.Wait()
}

func Pong(b []byte) {
	var p shared.Pong
	err := cbor.Unmarshal(b, &p)
	if err != nil {
		fmt.Println("Error during unmarshalling Pong")
		panic(err)
	}
	fmt.Println("Received Pong: ", p.Message)
	fmt.Println(p.Message)
}
