// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

package helpers

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// MustRaw serializes the given value into CBOR format and returns it as a RawMessage.
//
// This function panics if marshaling fails.
// Suitable for internal use where inputs are trusted.
func MustRaw(v any) cbor.RawMessage {
	// FIXME: This probably should have better naming...
	b, err := cbor.Marshal(v)
	if err != nil {
		panic(err) // FIXME: Handle error
	}
	return cbor.RawMessage(b)
}

// WrapHandler adapts a strongly-typed handler function into a generic handler
// that can be registered in PlugKit's message system.
//
// It expects input in the form of an `any` value, typically representing CBOR-decoded data.
// If the input is not in the expected strongly-typed form, it is re-encoded to CBOR and then
// decoded into the expected type. This provides compatibility with inputs that were
// decoded generically (e.g., into map[interface{}]interface{} or map[string]any).
//
// This pattern ensures type safety on the handler side, while remaining flexible for
// various decoding strategies used within the runtime.
//
// This function performs an indirect CBOR remarshal to transform dynamically-typed data
// into a statically typed value. While effective and safe, the mechanism is not idiomatic Go.
// It is recommended for internal use in PlugKit only, to simplify handler registration.
// Example:
//
//	func HandlePing(p shared.Ping) (bool, error) { ... }
//	client.HandleMessageType("ping", WrapHandler(HandlePing))
func WrapHandler[In any, Out any](fn func(In) (Out, error)) func(any) (any, error) {
	return func(v any) (any, error) {
		// It's an ugly fix - data sometimes is treated as map{}interface...
		data, err := cbor.Marshal(v)
		if err != nil {
			var zero Out
			return zero, fmt.Errorf("re-marshal to CBOR failed: %w", err)
		}

		var decoded In
		if err := cbor.Unmarshal(data, &decoded); err != nil {
			var zero Out
			return zero, fmt.Errorf("CBOR decode failed into %T: %w", decoded, err)
		}
		return fn(decoded)
	}
}

// DecodeAs attempts to cast a value of type any to a specific target type T.
//
// Returns the value and true if successful, or the zero value of T and false otherwise.
func DecodeAs[T any](value any) (T, bool) {
	v, ok := value.(T)
	return v, ok
}
