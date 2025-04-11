// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT
package main

import (
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/examples/0-plugin-test/shared"
	"github.com/mjwhodur/plugkit/messages"
	"github.com/mjwhodur/plugkit/plug"
)

var p *plug.SmartPlug

func main() {
	p = plug.New()
	// p.HandleMessageType("ping", func(bytes []byte) (*messages.Result, codes.PluginExitReason, error) {
	//	var pm shared.Ping
	//	err := cbor.Unmarshal(bytes, &pm)
	//	if err != nil {
	//		return nil, codes.HostToPluginCommunicationError, err
	//	}
	//
	//	return &messages.Result{
	//		Type: "pong",
	//		Value: &shared.Pong{
	//			Message: "Successful",
	//		},
	//	}, codes.OperationSuccess, nil
	// })
	plug.HandleSmartPlugMessage(p, "ping", PingHandler)
	err := p.Main()
	if err != nil {
		return
	}

}

func PingHandler(p *shared.Ping) (*messages.Result, codes.PluginExitReason, error) {
	if p != nil {
		return &messages.Result{
			Type: "pong",
			Value: &shared.Pong{
				Message: "Successful",
			},
		}, codes.OperationSuccess, nil
	}
	return &messages.Result{
		Type: "pong",
		Value: &shared.Pong{
			Message: "Unsuccessful. p was nil.",
		},
	}, codes.OperationSuccess, nil
}
