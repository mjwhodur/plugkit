package plugkit

import "github.com/fxamacker/cbor/v2"

// MustRaw encodes arbitrary Go value into a CBOR message.
func MustRaw(v any) cbor.RawMessage {
	// FIXME: This probably should have better naming...
	b, err := cbor.Marshal(v)
	if err != nil {
		panic(err) // FIXME: Handle error
	}
	return cbor.RawMessage(b)
}
