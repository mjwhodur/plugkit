package main

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/examples/0-plugin-test/shared"
	"github.com/mjwhodur/plugkit/plug"
)

var p *plug.BasicHandler

func main() {
	p = plug.New()
	p.HandleMessageType("ping", func(bytes []byte) {
		var pm shared.Ping
		err := cbor.Unmarshal(bytes, &pm)
		if err != nil {
			fmt.Println("Error during unmarshalling Ping")
			panic(err)
		}
		p.Respond("pong", &shared.Pong{
			Message: "Successful",
		})
		p.Finish("Operation complete", codes.OperationSuccess)
	})
	plug.Main(p)

}
