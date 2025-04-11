// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

// Package client provides a standard PlugKit host implementation.
// It allows launching and communicating with external plugin processes
// using CBOR-encoded messages over standard input and output.
package client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/helpers"
	"github.com/mjwhodur/plugkit/messages"
)

// SmartPlugClient represents a PlugKit host instance.
// It manages the lifecycle and communication of a plugin process.
//
// Communication is handled via stdin/stdout pipes using CBOR encoding.
// Messages are exchanged as Envelope structures with a defined message type and payload.
type SmartPlugClient struct {
	encoder  *cbor.Encoder
	decoder  *cbor.Decoder
	command  string
	Handlers map[string]func(any) (any, error)
	isReady  bool
}

// StartLocal starts the plugin process using the provided command.
//
// It sets up CBOR encoding/decoding over stdin/stdout pipes.
// Returns an error if the plugin cannot be started or the communication setup fails.
func (c *SmartPlugClient) StartLocal() error {

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

// NewSmartClient creates a new SmartPlugClient instance with the given plugin command.
// The plugin is not started automatically — use StartLocal() to launch it.
func NewSmartClient(name string) *SmartPlugClient {
	isReady := name != ""

	return &SmartPlugClient{
		command:  name,
		Handlers: make(map[string]func(any) (any, error)),
		isReady:  isReady,
	}
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
func (c *SmartPlugClient) RunCommand(name codes.MessageCode, v any) (codes.PluginExitReason, any, error) {
	if !c.isReady {
		return codes.PlugNotStarted, nil, errors.New("client is not ready")
	}
	err := c.encoder.Encode(&messages.Envelope{
		Version: 0,
		Type:    string(name),
		Raw:     helpers.MustRaw(v),
	})
	if err != nil {
		fmt.Println("Error during encoding of the envelope")
		return codes.PlugCrashed, nil, err
	}
	var msg messages.Envelope
	if err := c.decoder.Decode(&msg); err != nil {
		if err == io.EOF {
			// FIXME: Log Error?
			fmt.Println("Plugin finished prematurely - broken pipe")
			return codes.PlugCrashed, nil, err
		}
		fmt.Fprintln(os.Stderr, "decode error:", err)
		return codes.PluginToHostCommunicationError, nil, err
	}
	if msg.Type == string(codes.FinishMessage) {
		fmt.Println("Plugin finished its job")
		fmt.Println("Cleaning up")
		var fin *messages.PluginFinish
		err := cbor.Unmarshal(msg.Raw, &fin)
		if err != nil {
			return codes.PluginToHostCommunicationError, nil, err
		}
		return fin.Reason, nil, errors.New(fin.Message)
	}
	if msg.Type == string(codes.Unsupported) {
		return codes.CommandInvokedCannotExecute, nil, errors.New("unsupported message type")
	}
	if msg.Type == string(codes.PluginResponse) {
		var result messages.Result
		if e := cbor.Unmarshal(msg.Raw, &result); e != nil {
			panic(e)
		}
		if handler, ok := c.Handlers[result.Type]; ok {
			if result.Value == nil {
				fmt.Println("Received message with no raw payload")
			}
			res, e := handler(result.Value)
			return codes.OperationSuccess, res, e
		}
	}

	e := c.respond(codes.Unsupported, &messages.MessageUnsupported{})
	if e != nil {
		// FIXME: Maybe too vague?
		return codes.PlugCrashed, nil, errors.New("plug crashed " + err.Error())
	}
	return codes.OperationError, nil, errors.New("unsupported response message type")

}

// HandleMessageType registers a handler for a specific incoming message type.
//
// Handlers are called when a message of the given type is received from the plugin.
func (c *SmartPlugClient) HandleMessageType(name string, handler func(any) (any, error)) {
	c.Handlers[name] = handler
}

type WrappedMessageHandler interface {
	HandleMessageType(name string, handler func(any) (any, error))
}

// HandleMessage is a generic convenience wrapper for HandleMessageType.
// It allows registering a strongly-typed handler function without manual wrapping.
//
// The handler should have the form: func(In) (Out, error), where In is the expected input type
// (after CBOR decoding), and Out is the return type. Internally, the handler is wrapped
// using WrapHandler to conform to the PlugKit internal interface.
func HandleMessage[In any, Out any](c WrappedMessageHandler, name string, handler func(In) (Out, error)) {
	c.HandleMessageType(name, helpers.WrapHandler(handler))
}

// RespondRaw sends a raw response to the plugin with a specified type and payload.
//
// This method bypasses the MessageCode abstraction and can be used for ad-hoc messages.
func (c *SmartPlugClient) RespondRaw(t string, v any) error {
	// FIXME: Lacking test?
	err := c.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    t,
		Raw:     helpers.MustRaw(v),
	})

	return err

}

// respond is a helper method that sends a typed message to the plugin.
//
// It wraps the message code and value into an Envelope.
// This is the preferred way to respond using predefined MessageCode values.
func (c *SmartPlugClient) respond(messageCode codes.MessageCode, v any) error {
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
func (c *SmartPlugClient) SetCommand(command string) {
	c.command = command
}
