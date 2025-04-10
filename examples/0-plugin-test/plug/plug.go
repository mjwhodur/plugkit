package main

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/examples/0-plugin-test/shared"
	"github.com/mjwhodur/plugkit/messages"
	"github.com/mjwhodur/plugkit/plug"
)

var p *plug.Plug

func main() {
	p = plug.New()
	p.HandleMessageType("ping", func(bytes []byte) (*messages.Result, codes.PluginExitReason, error) {
		var pm shared.Ping
		err := cbor.Unmarshal(bytes, &pm)
		if err != nil {
			return nil, codes.HostToPluginCommunicationError, err
		}

		return &messages.Result{
			Type: "pong",
			Value: &shared.Pong{
				Message: "Successful",
			},
		}, codes.OperationSuccess, nil
	})
	err := p.Main()
	if err != nil {
		return
	}

}
