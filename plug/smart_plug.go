// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

// Package plug provides the default runtime for a PlugKit-compatible plugin.
//
// A SmartPlug is a minimal, one-shot command handler that receives a CBOR-encoded
// Envelope from the host via stdin, processes it using a registered handler,
// sends a response back to stdout, and terminates with a specific exit code.
//
// This simple model allows for sandboxed, transactional plugin operations
// using structured messaging.
package plug

import (
	"errors"
	"fmt"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/helpers"
	"github.com/mjwhodur/plugkit/messages"
)

// SmartPlug is a most standard type of plug. It supports one-off comand.
// SmartPlug is the core runtime structure for a PlugKit plugin.
//
// It supports registering handlers for specific message types and
// handles a single Envelope message per execution.
type SmartPlug struct {
	Handlers map[string]func([]byte) (result *messages.Result, exitReason codes.PluginExitReason, e error)
	decoder  *cbor.Decoder
	encoder  *cbor.Encoder
}

// New creates a new SmartPlug instance wired to stdin and stdout.
//
// A default "exit" message handler is registered, which allows the host
// to gracefully terminate the plugin if needed.
func New() *SmartPlug {
	h := &SmartPlug{}
	h.Handlers = make(map[string]func([]byte) (result *messages.Result, exitReason codes.PluginExitReason, e error))
	h.decoder = cbor.NewDecoder(os.Stdin)
	h.encoder = cbor.NewEncoder(os.Stdout)

	// FIXME: Fix message to correctly support cleanup and disposing
	h.Handlers["exit"] = func(_ []byte) (result *messages.Result, exitReason codes.PluginExitReason, e error) {
		return nil, codes.OperationCancelledByClient, nil
	}

	return h
}

// WrapSmartPlugTypedHandler adapts a strongly-typed plugin handler to the
// func([]byte) (*messages.Result, codes.PluginExitReason, error) form expected
// by SmartPlug.
//
// It performs CBOR decoding of the input and passes the resulting value to the user-defined handler.
func WrapSmartPlugTypedHandler[In any](
	fn func(In) (*messages.Result, codes.PluginExitReason, error),
) func([]byte) (*messages.Result, codes.PluginExitReason, error) {
	return func(raw []byte) (*messages.Result, codes.PluginExitReason, error) {
		var input In
		if err := cbor.Unmarshal(raw, &input); err != nil {
			return nil, codes.HostToPluginCommunicationError, fmt.Errorf("CBOR decode error: %w", err)
		}

		return fn(input)
	}
}

func HandleSmartPlugMessage[In any](
	s *SmartPlug,
	messageType string,
	handler func(In) (*messages.Result, codes.PluginExitReason, error),
) {
	s.HandleMessageType(messageType, WrapSmartPlugTypedHandler(handler))
}

// HandleMessageType registers a function to handle a given message type.
//
// The handler receives raw CBOR-encoded data from the Envelope and is
// responsible for decoding and processing it.
//
// The handler must return a messages.Result (or nil), a PluginExitReason,
// and an error (or nil). message.Result must contain status code, and value of the response.
func (h *SmartPlug) HandleMessageType(name string, handler func([]byte) (*messages.Result, codes.PluginExitReason, error)) {
	// FIXME: Make this function more generic - it needs to have nice interface
	h.Handlers[name] = handler
}

// Main runs the main routine of the plugin.
//
// It waits for a single incoming message, dispatches it to the appropriate handler,
// sends back any response, and terminates with the declared PluginExitReason.
//
// This function is designed for one-shot plugin invocations. It should be called
// from the plugin's main() function.
func (h *SmartPlug) Main() error {
	// FIXME: Add handshake
	// FIXME: Add exit and possibly other signals
	exitCode := codes.OperationSuccess

	var msg messages.Envelope
	if err := h.decoder.Decode(&msg); err != nil {
		h.Finish("Malformed message received", codes.HostToPluginCommunicationError)

	}

	if msg.Type == string(codes.Unsupported) {
		h.Finish("Unsupported message received from host", codes.PluginToHostCommunicationError)
	}

	if handler, ok := h.Handlers[msg.Type]; ok {
		// endMessage := ""
		resp, _, err := handler(msg.Raw) //FIXME: Doesn't propagate the exit code
		if err != nil {
			panic(err)
		}
		if resp != nil {
			h.Respond(resp)
		}
		// h.Finish(endMessage, exitCode)

	} else {
		err := h.encoder.Encode(messages.Envelope{
			Version: 1,
			Type:    string(codes.Unsupported),
			Raw:     helpers.MustRaw(&messages.MessageUnsupported{}),
		})
		if err != nil {
			return err
		}

	}

	switch exitCode {
	case codes.OperationSuccess:
		return nil

	default:
		// FIXME: Handle the exit codes to messages?
		return errors.New(string(rune(exitCode)))
	}
}

// Respond sends a typed message to the host.
//
// It wraps the payload into a CBOR-encoded Envelope and writes it to stdout.
// Panics if encoding fails.
//
// Should only be used from within message handlers.
func (h *SmartPlug) Respond(r *messages.Result) {
	err := h.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    string(codes.PluginResponse),
		Raw:     helpers.MustRaw(r),
	})
	if err != nil {
		panic(err)
	}
	// FIXME: No support for errors
	// FIXME: Panics need to be handled on host side!
}

// Finish sends a PluginFinish message and terminates the plugin process.
//
// The plugin will exit with the given PluginExitReason code.
func (h *SmartPlug) Finish(message string, code codes.PluginExitReason) {
	val := &messages.PluginFinish{
		Reason:  code,
		Message: message,
	}

	err := h.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    string(codes.FinishMessage),
		Raw:     helpers.MustRaw(val),
	})
	if err != nil {
		os.Exit(int(codes.OperationError))
	}
	os.Exit(int(code))
}
