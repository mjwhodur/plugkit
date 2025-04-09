package plug

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/messages"
	"os"
)

// Plug is a most standard type of plug.
type Plug struct {
	Handlers map[string]func([]byte)
	decoder  *cbor.Decoder
	encoder  *cbor.Encoder
}

// New creates basic plug that supports basic options
func New() *Plug {
	h := &Plug{}
	h.Handlers = make(map[string]func([]byte))
	h.decoder = cbor.NewDecoder(os.Stdin)
	h.encoder = cbor.NewEncoder(os.Stdout)

	// FIXME: Fix message to correctly support cleanup and disposing
	h.Handlers["exit"] = func(p []byte) { os.Exit(0) }

	return h
}

// HandleMessageType registers a function that decodes a message of particular type to process it
func (h *Plug) HandleMessageType(name string, handler func([]byte)) {
	// FIXME: Make this function more generic - it needs to have nice interface
	h.Handlers[name] = handler
}

// Main runs the main loop for a plugin
func (h *Plug) Main() {
	// FIXME: Add handshake
	// FIXME: Add exit and possibly other signals
	for {
		var msg messages.Envelope
		if err := h.decoder.Decode(&msg); err != nil {
			fmt.Fprintln(os.Stderr, "decode error:", err)
			os.Exit(1)
		}
		if handler, ok := h.Handlers[msg.Type]; ok {
			handler(msg.Raw)
		} else {
			h.Respond(string(codes.Unsupported), &messages.MessageUnsupported{})
		}
	}
}

// Respond sends raw response to the plugin host. It encapsulates the information
// in an envelope
func (h *Plug) Respond(t string, v any) {
	// FIXME: Unhandled error here
	// FIXME: Test
	h.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    t,
		Raw:     plugkit.MustRaw(v),
	})

	// FIXME: No support for errors
}

// Finish sends confirmation message and exits the plugin
func (h *Plug) Finish(message string, code codes.PluginExitReason) {
	val := &messages.PluginFinish{
		Reason:  code,
		Message: message,
	}

	h.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    string(codes.FinishMessage),
		Raw:     plugkit.MustRaw(val),
	})
	//FIXME: It should not work like that... There has to be an exit code.
	os.Exit(0)
}
