package plug

import (
	"fmt"
	"os"

	"github.com/mjwhodur/plugkit/helpers"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/messages"
)

// RawPlugImpl is the interface that every raw plug implementation must satisfy.
//
// Handle receives the raw CBOR payload extracted from the envelope and returns:
// - a message code indicating the result (e.g., "ok", "unsupported", custom-defined, etc.),
// - a CBOR-encoded response payload to be sent back to the host,
// - or an error, which will cause the RawPlug to send a operation failure response to the host.
//
// If an error is returned from Handle, the plug is considered to have failed the request.
// In all other cases, the message is treated as successfully handled,
// even if the operation type was unknown or invalid — it's up to the plugin to decide how to respond.
// This is intentional, as RawPlug provides no automatic validation or dispatching — full control is left to the implementer.
//
// Mount is called once at startup and provides the plugin with access to its host context,
// which can be used to configure or initialize internal state.
type RawPlugImpl interface {
	Handle(kind string, payload cbor.RawMessage) (messageCode string, response cbor.RawMessage, err error)
	Mount(c *RawPlug)
}

// RawPlug provides a low-level plugin host that communicates over stdin and stdout using CBOR encoding.
// It reads and writes Envelope messages, and delegates the handling of payloads to the user-defined RawPlugImpl.
// RawPlug is the most minimal building block for creating plugins with custom protocols or structure.
type RawPlug struct {
	PlugImpl RawPlugImpl
	decoder  *cbor.Decoder
	encoder  *cbor.Encoder
}

// NewRawPlug creates a new RawPlug with the given user-defined implementation.
func NewRawPlug(impl RawPlugImpl) *RawPlug {
	return &RawPlug{
		PlugImpl: impl,
	}
}

// Main starts the main loop of the RawPlug.
// It reads a single Envelope from stdin, passes its raw CBOR payload to the user-defined implementation,
// and writes a response Envelope to stdout.
// If decoding fails, an appropriate error message is sent back immediately.
func (p *RawPlug) Main() error {
	p.PlugImpl.Mount(p)
	p.decoder = cbor.NewDecoder(os.Stdin)
	p.encoder = cbor.NewEncoder(os.Stdout)
	var msg messages.Envelope
	if err := p.decoder.Decode(&msg); err != nil {
		err := p.encoder.Encode(messages.Envelope{
			Version: 1,
			Type:    string(codes.PayloadMalformed),
			Raw:     helpers.MustRaw(&messages.MessageUnsupported{}),
		})
		if err != nil {
			panic(err)
		}

	}
	// Pass the raw payload to the implementation.
	msgCode, res, err := p.PlugImpl.Handle(msg.Type, msg.Raw)
	if err != nil {
		// Return a handling error with the error as payload.
		p.respondError(string(codes.HandlingError), helpers.MustRaw(err))
		return err
	}

	// Send the response with the provided message code and payload.
	p.respond(msgCode, res)
	return nil
}

// respond sends an Envelope with the success message code and CBOR payload to stdout.
func (p *RawPlug) respond(messageCode string, payload cbor.RawMessage) {
	// FIXME: message code name needs to be fixed
	res := &messages.Result{ExitCode: codes.OperationSuccess, Type: messageCode, Value: payload}
	// data, e := cbor.Marshal(res)
	// if e != nil {
	//	panic(e)
	//}
	err := p.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    string(codes.PluginResponse),
		Raw:     helpers.MustRaw(res),
	})
	if err != nil {
		// If we can't write the response, panic — plugin cannot recover.
		fmt.Println(err)
	}
}

// respondError sends an Envelope with the failure message code and CBOR payload to stdout.
func (p *RawPlug) respondError(messageCode string, payload cbor.RawMessage) {
	res := &messages.Result{ExitCode: codes.OperationError, Type: messageCode, Value: payload}
	// data, e := cbor.Marshal(res)
	// if e != nil {
	//	panic(e)
	//}
	err := p.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    string(codes.PluginResponse),
		Raw:     helpers.MustRaw(res),
	})
	if err != nil {
		// If we can't write the response, panic — plugin cannot recover.
		fmt.Println(err)
	}
}

// FIXME: Hide private functions?
