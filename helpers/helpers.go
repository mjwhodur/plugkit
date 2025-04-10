// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

package helpers

import "github.com/fxamacker/cbor/v2"

func MustRaw(v any) cbor.RawMessage {
	// FIXME: This probably should have better naming...
	b, err := cbor.Marshal(v)
	if err != nil {
		panic(err) // FIXME: Handle error
	}
	return cbor.RawMessage(b)
}
