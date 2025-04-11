// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

package messages

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
)

// Envelope represents a generic message wrapper used for communication between
// the host and the plugin in the PlugKit system.
//
// The actual message content is encoded in the Raw field, which contains CBOR-encoded data.
// This field must be decoded by the recipient according to the message Type.
//
// Envelope is used purely for transporting typed messages — the interpretation
// of Raw depends on Type and is done in the application logic.
type Envelope struct {
	Version int             `cbor:"version"` // Protocol version (e.g., 1)
	Type    string          `cbor:"type"`    // Message type identifier
	Raw     cbor.RawMessage `cbor:"data"`    // CBOR-encoded payload (must be decoded manually)
}

// Result represents the outcome of a function or command executed by the plugin.
//
// The Type field indicates the type of result (e.g., "number", "text", or a custom identifier),
// and the Value field holds the decoded value of the result.
//
// This structure is used when the plugin returns actual computation results,
// rather than status or control messages.
type Result struct {
	Type     string                 `cbor:"type"` // Type of result, user-defined
	ExitCode codes.PluginExitReason `cbor:"exitCode"`
	Value    any                    // Decoded result value
}

// StopCommand is sent from the host to the plugin to request a graceful shutdown.
//
// The Reason field provides an exit reason, typically codes.OperationCancelledByClient.
type StopCommand struct {
	Reason codes.PluginExitReason `cbor:"reason"`
}

// PluginFinish is sent from the plugin to the host to indicate the plugin has completed
// its work and is shutting down.
//
// The Reason field specifies why the plugin is exiting.
// The Message field may contain a human-readable explanation or summary.
type PluginFinish struct {
	Reason  codes.PluginExitReason `cbor:"reason"`
	Message string                 `cbor:"message"`
}

// MessageUnsupported is sent when the plugin receives a message it cannot handle.
//
// This type indicates that the message type was unknown, unimplemented, or invalid
// from the perspective of the receiving party (usually the plugin).
type MessageUnsupported struct {
}
