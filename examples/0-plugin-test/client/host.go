// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT
package main

import (
	"fmt"
	"os"

	"github.com/mjwhodur/plugkit/helpers"

	"github.com/mjwhodur/plugkit/client"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/examples/0-plugin-test/shared"
)

func main() {
	fmt.Println("Hello, client")

	// Assume the plug is built and that is the correct name (The makefile takes care of that...)
	c := client.NewSmartClient("./plugin")

	// Plug always responds with Pong type for every Ping message we send to it
	c.HandleMessageType("pong", helpers.WrapHandler(PongHandler))

	// or...
	client.HandleMessage(c, "pong", PongHandler)

	// When using more sophisticated plugs, more handlers may be required, i.e.
	// plug can respond with different structures.

	// Execute the plug. Plug will listen to new messages.
	err := c.StartLocal()
	if err != nil {
		return
	}

	// Actual "Command" execution.
	reason, v, e := c.RunCommand("ping", &shared.Ping{})
	// We sent "ping" message type (of value of empty Ping struct) to the plug. The plug shall respond
	// with the "pong" type of message (that is handled by PongHandler).

	// Let's check, if the operation is a success. You can handle different
	// exit codes (type PluginExitReason).
	if reason != codes.OperationSuccess {
		panic("Unsuccessful plugin execution")
	}
	// Plug may finish its job with different code. It doesn't necessarily mean
	// that plug broke. Let's say plug downloads something and the download failed.
	// It may actually return execution as failure.

	// Plugin operation could be marked as success - so execution was successful
	// but we still may not be happy from the answer :)
	if e != nil {
		panic(e)
	}

	// Let's check, whether correct handler (PongHandler) was run.
	// Since v (response from the handler) is bool, it is what we expect.
	_, ok := helpers.DecodeAs[bool](v)
	if !ok {
		os.Exit(1)
	}

}

func PongHandler(b *shared.Pong) (bool, error) {
	fmt.Println("Received PongHandler: ", b.Message)
	return true, nil
}
