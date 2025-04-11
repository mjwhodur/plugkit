// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

package options

import (
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/messages"
)

// CommandHandler binds a command name (message type) to its handler function.
//
// The handler receives raw CBOR-encoded data from the host, decodes it,
// and returns a result along with an appropriate exit code and optional error.
//
// This structure is used by plugkit.CreateAndRunPlug() to register handlers declaratively.
type CommandHandler struct {
	Command string
	Handler func(p []byte) (*messages.Result, codes.PluginExitReason, error)
}
