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

type RawClientImpl interface {
	Handle(msgType string, payload []byte)
}
type RawClient struct {
	encoder *cbor.Encoder
	decoder *cbor.Decoder
	command string
	Impl    RawClientImpl
	isReady bool
}

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
		c.Impl.Handle(msg.Type, msg.Raw)
	}

	e := c.respond(codes.Unsupported, &messages.MessageUnsupported{})
	if e != nil {
		// FIXME: Maybe too vague?
		return codes.PlugCrashed, nil, errors.New("plug crashed " + err.Error())
	}
	return codes.OperationError, nil, errors.New("unsupported response message type")

}

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
