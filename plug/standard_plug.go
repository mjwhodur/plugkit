package plug

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/messages"
	"os"
)

type BasicHandler struct {
	Handlers map[string]func([]byte)
	decoder  *cbor.Decoder
	encoder  *cbor.Encoder
}

func New() *BasicHandler {
	h := &BasicHandler{}
	h.Handlers = make(map[string]func([]byte))
	h.decoder = cbor.NewDecoder(os.Stdin)
	h.encoder = cbor.NewEncoder(os.Stdout)

	// FIXME: Something is off here :D
	h.Handlers["exit"] = func(p []byte) { os.Exit(0) }

	return h
}

func (h *BasicHandler) HandleMessageType(name string, handler func([]byte)) {
	h.Handlers[name] = handler
}

func Main(h *BasicHandler) {
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
			h.Respond("unsupported", &messages.MessageUnsupported{})
		}
	}
}

func (h *BasicHandler) Respond(t string, v any) {
	// FIXME: Unhandled error here
	// FIXME: Test
	h.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    t,
		Raw:     plugkit.MustRaw(v),
	})

	// FIXME: No support for errors
}

func (h *BasicHandler) Finish(message string, code codes.PluginExitReason) {
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
