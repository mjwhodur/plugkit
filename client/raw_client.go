// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

// Package client provides a low-level PlugKit host implementation.
// The RawClient allows for direct handling of plugin communication
// by exposing decoded CBOR messages and letting the user define custom message logic
package client

import (
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/helpers"
	"github.com/mjwhodur/plugkit/messages"
)

// RawClientImpl defines the interface that must be implemented by users of RawClient.
// It is responsible for handling incoming messages manually.
type RawClientImpl interface {
	Handle(responseType string, payload []byte)
}

// RawClient provides a minimalistic PlugKit host implementation.
// Unlike SmartPlugClient, it does not automatically decode or route messages by type.
//
// Instead, it allows the user to implement their own message handler via RawClientImpl.
// This enables advanced use cases or full control over the plugin protocol.
//
// Communication occurs over CBOR-encoded envelopes using stdin and stdout pipes.
type RawClient struct {
	encoder *cbor.Encoder
	decoder *cbor.Decoder
	command string
	Impl    RawClientImpl
	isReady bool
}

// StartLocal starts the plugin process using the configured command.
//
// It establishes CBOR-based communication over stdin/stdout pipes.
// The method must be called before sending any commands to the plugin.
func (c *RawClient) StartLocal() error {
	if c.command == "" {
		return errors.New("command executable is required")
	}

	cmd := exec.Command(c.command) // #nosec G204
	stdin, err := cmd.StdinPipe()

	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if e := cmd.Start(); e != nil {
		return e
	}

	c.encoder = cbor.NewEncoder(stdin)
	c.decoder = cbor.NewDecoder(stdout)

	if c.decoder == nil {
		return errors.New("failed to create decoder")
	}
	if c.encoder == nil {
		return errors.New("failed to create encoder")
	}

	return nil
}

// RunCommand sends a command (message) to the plugin and waits for its response.
//
// The message is wrapped in an Envelope with the given message code and payload.
// Based on the plugin's response, this method returns an appropriate PluginExitReason,
// a handler result (if any), and an error.
//
// The second return value (of type any) represents the value returned by a matching message
// handler registered via HandleMessageType. If the plugin responds with a recognized message type
// and the corresponding handler executes successfully, its return value is passed back to the caller.
//
// If no matching handler is found or the handler fails, the returned result will be nil.
// In general, it's recommended that plugins return a single, well-defined response type,
// or exit with an appropriate error code when things go wrong. Supporting multiple possible
// response types is possible, but may require custom decoding logic on the client side.
func (c *RawClient) RunCommand(name codes.MessageCode, v any) (codes.PluginExitReason, any, error) {
	if !c.isReady {
		return codes.PlugNotStarted, nil, errors.New("client is not ready")
	}
	err := c.encoder.Encode(&messages.Envelope{
		Version: 0,
		Type:    string(name),
		Raw:     helpers.MustRaw(v),
	})
	if err != nil {
		return codes.PlugCrashed, nil, err
	}
	var envelope messages.Envelope
	if err := c.decoder.Decode(&envelope); err != nil {
		if err == io.EOF {
			return codes.PlugCrashed, nil, err
		}
		return codes.PluginToHostCommunicationError, nil, err
	}
	fmt.Println(envelope.Type)
	if envelope.Type == string(codes.FinishMessage) {
		var fin *messages.PluginFinish
		err := cbor.Unmarshal(envelope.Raw, &fin)
		if err != nil {
			return codes.PluginToHostCommunicationError, nil, err
		}
		return fin.Reason, nil, errors.New(fin.Message)
	}
	if envelope.Type == string(codes.Unsupported) {
		return codes.CommandInvokedCannotExecute, nil, errors.New("unsupported message type")
	}
	if envelope.Type == string(codes.PluginResponse) {
		var result messages.Result

		if e := cbor.Unmarshal(envelope.Raw, &result); e != nil {
			panic(e)
		}
		val, _ := cbor.Marshal(result.Value)
		c.Impl.Handle(result.Type, val)
		return codes.OperationSuccess, nil, nil
	}

	e := c.respond(codes.Unsupported, &messages.MessageUnsupported{})
	if e != nil {
		// FIXME: Maybe too vague?
		return codes.PlugCrashed, nil, e
	}
	return codes.PluginToHostCommunicationError, nil, errors.New("unsupported response message type")

}

// respond sends a response message to the plugin using a predefined MessageCode.
//
// This helper wraps the provided payload into a PlugKit Envelope and sends it over stdout.
func (c *RawClient) respond(messageCode codes.MessageCode, v any) error {
	// FIXME: Lacking test?
	err := c.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    string(messageCode),
		Raw:     helpers.MustRaw(v),
	})

	return err
}

// SetCommand sets the executable path or name of the plugin binary.
// This must be set before calling StartLocal().
func (c *RawClient) SetCommand(command string) {
	c.command = command
}

// NewRawClient returns plugin with impl implementation
func NewRawClient(name string, impl RawClientImpl) *RawClient {
	isReady := name != ""

	return &RawClient{
		isReady: isReady,
		command: name,
		Impl:    impl,
	}
}
