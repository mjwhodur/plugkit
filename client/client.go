package client

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/messages"
	"io"
	"os"
	"os/exec"
	"sync"
)

type Client struct {
	encoder   *cbor.Encoder
	decoder   *cbor.Decoder
	breakChan chan bool
	command   string
	Handlers  map[string]func([]byte)
	Wg        *sync.WaitGroup
}

func (c *Client) DownloadAndStart() {
}

func (c *Client) StartLocal() {
	cmd := exec.Command(c.command)
	stdin, e1 := cmd.StdinPipe()
	stdout, e2 := cmd.StdoutPipe()
	if e1 != nil {
		fmt.Println("STDIN PIPE ERROR")
		panic(e1)
	}
	if e2 != nil {
		fmt.Println("STDOUT PIPE ERROR")
		panic(e2)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	c.encoder = cbor.NewEncoder(stdin)
	c.decoder = cbor.NewDecoder(stdout)
	if c.decoder == nil {
		panic("no decoder")
	}
	if c.encoder == nil {
		panic("no encoder")
	}

	c.Wg.Add(1)
	go c.loop()
}

func NewClient(name string) *Client {
	return &Client{
		command:  name,
		Handlers: make(map[string]func([]byte)),
		Wg:       &sync.WaitGroup{},
	}
}

func (c *Client) RunCommand(name string, v any) {
	fmt.Println("Sending message type", name)
	err := c.encoder.Encode(&messages.Envelope{
		Version: 0,
		Type:    name,
		//Raw:     pluginsdk.MustRaw(v),
	})
	if err != nil {
		fmt.Println("Error during encoding of the envelope")
		panic(err)
	}
}

func (c *Client) loop() {
	fmt.Println("Starting loop")
	// FIXME: Add handshake
	// FIXME: Add exit and possibly other signals
	for {
		var msg messages.Envelope
		if err := c.decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				c.Wg.Done()
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
			c.Wg.Done()
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
			// FIXME: Client Library does not support unsupported messages
			c.Respond(string(codes.Unsupported), &messages.MessageUnsupported{})
		}
	}
}

func (c *Client) Kill() {
	c.RunCommand("stopcommand", &messages.StopCommand{
		Reason: codes.OperationCancelledByClient,
	})
}

func (h *Client) HandleMessageType(name string, handler func([]byte)) {
	h.Handlers[name] = handler
}

func (h *Client) Respond(t string, v any) {
	// FIXME: Unhandled error here
	// FIXME: Test
	h.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    t,
		Raw:     plugkit.MustRaw(v),
	})

	// FIXME: No support for errors
}

func (h *Client) Init() {
	panic("implement me")
	//FIXME: Implement me
}

func (c *Client) SetCommand(command string) {
	c.command = command
}
