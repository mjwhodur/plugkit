// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT
package client

import (
	"context"
	"errors"
	"os/exec"
	"sync"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/helpers"
	"github.com/mjwhodur/plugkit/messages"
)

// RawStreamClientImpl must be implemented by consumers of RawStreamClient.
// It defines the logic for handling incoming messages and sending responses.
//
// - Handle is invoked for each incoming message and must return the message type and CBOR payload of the response.
// - Mount is called before the communication loop starts.
// - CloseSignal is triggered when the stream is closing.
type RawStreamClientImpl interface {
	Handle(kind string, payload *cbor.RawMessage) (messageCode string, response cbor.RawMessage, err error)
	Mount(c *RawStreamClient)
	CloseSignal()
}

// RawStreamClient provides a streaming PlugKit host implementation.
//
// It launches the plugin process and maintains an open communication loop,
// continuously decoding incoming CBOR messages and dispatching them to the provided handler.
//
// This structure is well-suited for long-running plugins with complex protocols or event-based logic.
type RawStreamClient struct {
	Impl    RawStreamClientImpl
	encoder *cbor.Encoder
	decoder *cbor.Decoder
	command string
	wg      *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewRawStreamClient constructs a new RawStreamClient with the given handler implementation.
//
// The plugin process is not started automatically — use Start() and then Run() to begin communication.
func NewRawStreamClient(impl RawStreamClientImpl, command string) *RawStreamClient {
	return &RawStreamClient{
		Impl:    impl,
		command: command,
	}
}

// Start initializes and starts the plugin process using the configured command.
//
// It sets up CBOR encoders and decoders for stdin/stdout communication.
// Must be called before Run().
func (c *RawStreamClient) Start() error {
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
	c.wg = &sync.WaitGroup{}
	return nil
}

// Run begins the main communication loop with the plugin.
//
// This method blocks until Stop() is called or the stream ends. It spawns a background loop
// to decode messages and invokes the handler logic for each received message.
func (c *RawStreamClient) Run() {
	c.Impl.Mount(c)
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel

	c.wg.Add(1)
	go c.loop()
	c.wg.Wait()
}

// Stop signals the RawStreamClient to stop receiving messages.
//
// It cancels the internal context and allows the loop to exit gracefully.
func (c *RawStreamClient) Stop() {
	c.cancel()
}

// loop is the internal message receive loop.
//
// It continuously decodes CBOR messages from the plugin and dispatches them
// via the ResponseWrapper for asynchronous handling.
func (c *RawStreamClient) loop() {
loop:
	for {
		select {
		case <-c.ctx.Done():

			// FIXME: It may crash. It should be blocking (somehow - or expecting a value?)
			c.Impl.CloseSignal()

			break loop

		default:

			var msg messages.Envelope
			if err := c.decoder.Decode(&msg); err != nil {
				err := c.encoder.Encode(messages.Envelope{
					Version: 1,
					Type:    string(codes.PayloadMalformed),
					Raw:     helpers.MustRaw(&messages.MessageUnsupported{}),
				})
				if err != nil {
					panic(err)
				}
				continue
			}

			c.wg.Add(1)
			go c.ResponseWrapper(msg)

		}
	}
}

// ResponseWrapper wraps a single message and processes it via the implementation's Handle method.
//
// It constructs a response and sends it back to the plugin.
// This function is run as a goroutine for each message.
func (c *RawStreamClient) ResponseWrapper(msg messages.Envelope) {
	msgCode, res, err := c.Impl.Handle(msg.Type, &msg.Raw)
	if err != nil {
		// Return a handling error with the error as payload.
		c.Respond(string(codes.HandlingError), helpers.MustRaw(err))
		c.wg.Done()
		return
	}

	// Send the response with the provided message code and payload.
	c.Respond(msgCode, res)
	c.wg.Done()
}

// Respond sends a response message back to the plugin.
//
// The message type and CBOR payload must be specified explicitly.
func (c *RawStreamClient) Respond(messageCode string, payload cbor.RawMessage) {
	err := c.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    messageCode,
		Raw:     payload,
	})
	if err != nil {
		// If we can't write the response, panic — plugin cannot recover.
		panic(err)
	}
}
