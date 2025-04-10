package plug

import (
	"errors"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/helpers"
	"github.com/mjwhodur/plugkit/messages"
)

// Plug is a most standard type of plug.
type Plug struct {
	Handlers map[string]func([]byte) (result *messages.Result, exitReason codes.PluginExitReason, e error)
	decoder  *cbor.Decoder
	encoder  *cbor.Encoder
}

// New creates basic plug that supports basic options
func New() *Plug {
	h := &Plug{}
	h.Handlers = make(map[string]func([]byte) (result *messages.Result, exitReason codes.PluginExitReason, e error))
	h.decoder = cbor.NewDecoder(os.Stdin)
	h.encoder = cbor.NewEncoder(os.Stdout)

	// FIXME: Fix message to correctly support cleanup and disposing
	h.Handlers["exit"] = func(_ []byte) (result *messages.Result, exitReason codes.PluginExitReason, e error) {
		return nil, codes.OperationCancelledByClient, nil
	}

	return h
}

// HandleMessageType registers a function that decodes a message of particular type to process it
func (h *Plug) HandleMessageType(name string, handler func([]byte) (*messages.Result, codes.PluginExitReason, error)) {
	// FIXME: Make this function more generic - it needs to have nice interface
	h.Handlers[name] = handler
}

// Main runs the main loop for a plugin
func (h *Plug) Main() error {
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
		endMessage := ""
		resp, exitCode, err := handler(msg.Raw)
		if err != nil {
			endMessage = err.Error()
		}
		if resp != nil {
			h.Respond(resp.Type, resp.Value)
		}
		h.Finish(endMessage, exitCode)

	} else {
		h.Respond(string(codes.Unsupported), &messages.MessageUnsupported{})
		h.Finish("Unsupported message send to host", codes.HostToPluginCommunicationError)

	}

	switch exitCode {
	case codes.OperationSuccess:
		return nil

	default:
		// FIXME: Handle the exit codes to messages?
		return errors.New(string(rune(exitCode)))
	}
}

// Respond sends raw response to the plugin host. It encapsulates the information
// in an envelope
func (h *Plug) Respond(t string, v any) {
	err := h.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    t,
		Raw:     helpers.MustRaw(v),
	})
	panic(err)
	// FIXME: No support for errors
	// FIXME: Panics need to be handled on host side!
}

// Finish sends confirmation message and exits the plugin
func (h *Plug) Finish(message string, code codes.PluginExitReason) {
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
