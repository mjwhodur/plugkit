// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT
package main

import (
	"github.com/mjwhodur/plugkit/codes"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/examples/2-rawclient-rawplug-test-basic/shared"
	"github.com/mjwhodur/plugkit/plug"
)

type rawplug struct {
	c *plug.RawPlug
}

func (r *rawplug) Handle(kind string, payload cbor.RawMessage) (messageCode string, response cbor.RawMessage, err error) {
	data, _ := cbor.Marshal(shared.Pong{Message: "PONG FROM RAW PLUG"})
	if kind == "ping" {
		ping := &shared.Ping{}
		_ = cbor.Unmarshal(payload, &ping)
		// if e != nil {
		//	panic(e)
		//
		//}

		// if er != nil {
		//	panic("error marshaling pong")
		//}
		// r.c.respond("pong", data)

		return "pong", data, nil
	}
	return string(codes.Unsupported), data, nil

}

func (r *rawplug) Mount(p *plug.RawPlug) {
	r.c = p
}

func main() {

	p := plug.NewRawPlug(&rawplug{})
	err := p.Main()
	if err != nil {
		return
	}

}
