// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

package plugkit

import (
	"github.com/mjwhodur/plugkit/options"
	"github.com/mjwhodur/plugkit/plug"
)

// NewPlug returns a new, empty SmartPlug instance.
//
// This is a shortcut for plug.New(), useful for clarity or abstraction
// when exposing PlugKit in higher-level packages.
func NewPlug() *plug.SmartPlug {
	return plug.New()
}

// CreateAndRunPlug builds and runs a SmartPlug instance using the provided list of handlers.
//
// It is designed for functional-style plugin declarations, where a slice of
// message-command handlers is passed in and automatically registered.
//
// The function will block until the SmartPlug finishes execution and return any error
// reported during its lifecycle.
func CreateAndRunPlug(handlers *[]options.CommandHandler) error {
	p := plug.New()
	for _, h := range *handlers {
		p.HandleMessageType(h.Command, h.Handler)
	}
	return p.Main()
}
