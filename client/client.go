// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

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

// Client is a standard plug host. It supports basic operations.
type Client struct {
	encoder  *cbor.Encoder
	decoder  *cbor.Decoder
	command  string
	Handlers map[string]func([]byte)
	isReady  bool
}

// StartLocal starts local plug
func (c *Client) StartLocal() error {

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

// NewClient returns a pre-setup simple client for a plug.
func NewClient(name string) *Client {
	isReady := name != ""

	return &Client{
		command:  name,
		Handlers: make(map[string]func([]byte)),
		isReady:  isReady,
	}
}

// RunCommand sends arbitrary message to a plug. Plug must understand the message.
func (c *Client) RunCommand(name codes.MessageCode, v any) (codes.PluginExitReason, error) {
	if !c.isReady {
		return codes.PlugNotStarted, errors.New("client is not ready")
	}
	err := c.encoder.Encode(&messages.Envelope{
		Version: 0,
		Type:    string(name),
		Raw:     helpers.MustRaw(v),
	})
	if err != nil {
		fmt.Println("Error during encoding of the envelope")
		return codes.PlugCrashed, err
	}
	var msg messages.Envelope
	if err := c.decoder.Decode(&msg); err != nil {
		if err == io.EOF {
			// FIXME: Log Error?
			fmt.Println("Plugin finished prematurely - broken pipe")
		}
		fmt.Fprintln(os.Stderr, "decode error:", err)
		return codes.PlugCrashed, err
	}
	if msg.Type == string(codes.FinishMessage) {
		fmt.Println("Plugin finished its job")
		fmt.Println("Cleaning up")
		var fin *messages.PluginFinish
		err := cbor.Unmarshal(msg.Raw, &fin)
		if err != nil {
			return codes.PluginToHostCommunicationError, err
		}
		return fin.Reason, errors.New(fin.Message)
	}
	if msg.Type == string(codes.Unsupported) {
		return codes.CommandInvokedCannotExecute, errors.New("unsupported message type")
	}
	if handler, ok := c.Handlers[msg.Type]; ok {
		if msg.Raw == nil {
			fmt.Println("Received message with no raw payload")
		}
		handler(msg.Raw)
	} else {
		err := c.respond(codes.Unsupported, &messages.MessageUnsupported{})
		if err != nil {
			// FIXME: Maybe too vague?
			return codes.PlugCrashed, errors.New("plug crashed " + err.Error())
		}
		return codes.OperationError, errors.New("unsupported response message type")
	}
	return codes.OperationSuccess, nil
}

// Kill demands graceful shutdown of a plug.
func (c *Client) Kill() error {
	reason, _ := c.RunCommand(codes.ExitMessage, &messages.StopCommand{
		Reason: codes.OperationCancelledByClient,
	})
	if reason != codes.OperationSuccess {
		return errors.New("plug crashed " + reason.String())
	}
	return nil
	// FIXME: Probably not required?
}

// HandleMessageType adds handler for incoming messages from the plug to the host.
func (c *Client) HandleMessageType(name string, handler func([]byte)) {
	c.Handlers[name] = handler
}

// RespondRaw is a basic function for sending any message from host to a plug.
func (c *Client) RespondRaw(t string, v any) error {
	// FIXME: Lacking test?
	err := c.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    t,
		Raw:     helpers.MustRaw(v),
	})

	return err

}

// respond is a function for sending message with a particular type from host to a plug.
// Its aim is to help the developer creating message codes. I.e.:
// ```
// const (
//
//	Ping code.MessageCode = "Ping"
//	Pong = "Pong"
//
// )
//
// client.Respond(Ping, &PingMsg{})
// ```
func (c *Client) respond(messageCode codes.MessageCode, v any) error {
	// FIXME: Lacking test?
	err := c.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    string(messageCode),
		Raw:     helpers.MustRaw(v),
	})

	return err
}

// SetCommand sets plug executable name
func (c *Client) SetCommand(command string) {
	c.command = command
}
