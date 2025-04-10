package client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

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
	wg       *sync.WaitGroup
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

	c.wg.Add(1)
	go c.loop()
	return nil
}

// NewClient returns a pre-setup simple client for a plug.
func NewClient(name string) *Client {
	isReady := name != ""

	return &Client{
		command:  name,
		Handlers: make(map[string]func([]byte)),
		wg:       &sync.WaitGroup{},
		isReady:  isReady,
	}
}

// RunCommand sends arbitrary message to a plug. Plug must understand the message.
func (c *Client) RunCommand(name codes.MessageCode, v any) error {
	if !c.isReady {
		return errors.New("client is not ready")
	}
	err := c.encoder.Encode(&messages.Envelope{
		Version: 0,
		Type:    string(name),
		Raw:     helpers.MustRaw(v),
	})
	if err != nil {
		fmt.Println("Error during encoding of the envelope")
		return err
	}
	return nil
}

// Main loop for the plug
func (c *Client) loop() {
	fmt.Println("Starting loop")
	// FIXME: Add handshake
	// FIXME: Add exit and possibly other signals
	for {
		var msg messages.Envelope
		if err := c.decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				c.wg.Done()
				fmt.Println("Plugin finished prematurely - broken pipe")
				break
			}
			fmt.Fprintln(os.Stderr, "decode error:", err)
			os.Exit(1)
		}
		if msg.Type == string(codes.FinishMessage) {
			//FIXME: Something is missing here.
			//FIXME: Handling of errors
			fmt.Println("Plugin finished its job")
			fmt.Println("Cleaning up")
			c.wg.Done()
			break
		}
		if msg.Type == string(codes.Unsupported) {
			panic("unsupported message")
			// FIXME: There is probably better way to respond to that situation...
		}
		if handler, ok := c.Handlers[msg.Type]; ok {
			if msg.Raw == nil {
				fmt.Println("Received message with no raw payload")
			}
			handler(msg.Raw)
		} else {
			err := c.Respond(codes.Unsupported, &messages.MessageUnsupported{})
			if err != nil {
				// FIXME: If response fails it probably means that plugin died unexpectedly.
				panic(err)
			}
		}
	}
}

// Kill demands graceful shutdown of a plug.
func (c *Client) Kill() error {
	err := c.RunCommand(codes.ExitMessage, &messages.StopCommand{
		Reason: codes.OperationCancelledByClient,
	})
	if err != nil {
		return err
	}
	return nil
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

// Respond is a function for sending message with a particular type from host to a plug.
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
func (c *Client) Respond(messageCode codes.MessageCode, v any) error {
	// FIXME: Lacking test?
	err := c.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    string(messageCode),
		Raw:     helpers.MustRaw(v),
	})

	return err

}

func (c *Client) Init() {
	panic("implement me")
	//FIXME: Implement me
}

// SetCommand sets plug executable name
func (c *Client) SetCommand(command string) {
	c.command = command
}

// Wait blocks until client-plug connection terminates
func (c *Client) Wait() {
	c.wg.Wait()
}
