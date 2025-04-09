package messages

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
)

type Envelope struct {
	Version int             `cbor:"version"` // np. 1
	Type    string          `cbor:"type"`
	Raw     cbor.RawMessage `cbor:"data"`
}

type StopCommand struct {
	Reason codes.PluginExitReason `cbor:"reason"`
}

type PluginFinish struct {
	Reason  codes.PluginExitReason `cbor:"reason"`
	Message string                 `cbor:"message"`
}
type MessageUnsupported struct {
}
