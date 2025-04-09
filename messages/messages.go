package messages

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
)

// Envelope is a CBOR message that encapsulates all other messages. It can be of
// any serializable type. It is required to set Type to a string. Type can be any string, of arbitrary length.
type Envelope struct {
	Version int             `cbor:"version"` // np. 1
	Type    string          `cbor:"type"`
	Raw     cbor.RawMessage `cbor:"data"`
}

// StopCommand - handling host-to-plug exit demand.
type StopCommand struct {
	Reason codes.PluginExitReason `cbor:"reason"`
}

// PluginFinish is a response from plug that has one of exit reasons. It is typically
type PluginFinish struct {
	Reason  codes.PluginExitReason `cbor:"reason"`
	Message string                 `cbor:"message"`
}

// MessageUnsupported is an umbrella message stating that incoming payload is unsupported.
// It is sent i.e. when a plug does not understand the command that was sent to it.
type MessageUnsupported struct {
}
