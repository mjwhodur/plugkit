// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT
package main

import (
	"fmt"
)

type clientImpl struct {
}

func (c *clientImpl) Handle(msgType string, _ []byte) {
	if msgType == "pong" {
		fmt.Println("pong received")
	} else {
		panic("Unknown message type: " + msgType)
	}
}
