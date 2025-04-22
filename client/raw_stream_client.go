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

type RawStreamClientImpl interface {
	Handle(kind string, payload *cbor.RawMessage) (messageCode string, response cbor.RawMessage, err error)
	Mount(c *RawStreamClient)
	CloseSignal()
}
type RawStreamClient struct {
	Impl    RawStreamClientImpl
	encoder *cbor.Encoder
	decoder *cbor.Decoder
	command string
	wg      *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewRawStreamClient(impl RawStreamClientImpl) *RawStreamClient {
	return &RawStreamClient{
		Impl: impl,
	}
}

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

func (c *RawStreamClient) Run() {
	c.Impl.Mount(c)
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel

	c.wg.Add(1)
	go c.loop()
	c.wg.Wait()
}

func (c *RawStreamClient) Stop() {
	c.cancel()
}

func (c *RawStreamClient) loop() {
loop:
	for {
		select {
		case <-c.ctx.Done():

			c.wg.Add(1)
			go c.Impl.CloseSignal()

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

			}

			c.wg.Add(1)
			go c.ResponseWrapper(msg)

		}
	}
}

func (c *RawStreamClient) ResponseWrapper(msg messages.Envelope) {
	msgCode, res, err := c.Impl.Handle(msg.Type, &msg.Raw)
	if err != nil {
		// Return a handling error with the error as payload.
		c.Respond(string(codes.HandlingError), helpers.MustRaw(err))

	}

	// Send the response with the provided message code and payload.
	c.Respond(msgCode, res)
	c.wg.Done()
}

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
