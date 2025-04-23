// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT
package main

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/examples/2-rawclient-rawplug-test-basic/shared"
)

type clientImpl struct {
}

func (c *clientImpl) Handle(msgType string, payload []byte) {

	if msgType == "pong" {
		fmt.Println("pong received... allegedly")

		var pong shared.Pong
		e := cbor.Unmarshal(payload, &pong)
		if e != nil {
			fmt.Println("unmarshall exit code: ", e.Error())
		}
		fmt.Println("pong received: ", pong.Message)

	} else {
		fmt.Println("unknown msg type: ", msgType)
	}

	// if msgType == "pong" {
	//
	// } else {
	//	panic("Unknown message type: " + msgType)
	//}
}
