package plug

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/helpers"
	"github.com/mjwhodur/plugkit/messages"
)

// RawStreamPlugImpl is the interface that every raw plug implementation must satisfy.
//
// Handle receives the raw CBOR payload extracted from the envelope and returns:
// - a message code indicating the result (e.g., "ok", "unsupported", custom-defined, etc.),
// - a CBOR-encoded response payload to be sent back to the host,
// - or an error, which will cause the RawPlug to send a operation failure response to the host.
//
// If an error is returned from Handle, the plug is considered to have failed the request.
// In all other cases, the message is treated as successfully handled,
// even if the operation type was unknown or invalid — it's up to the plugin to decide how to respond.
// This is intentional, as RawStreamPlug provides no automatic validation or dispatching — full control is left to the implementer.
//
// Mount is called once at startup and provides the plugin with access to its host context,
// which can be used to configure or initialize internal state.
// CloseSignal is called when the plug receives external shutdown signal (i.e. OS, plug host etc).
type RawStreamPlugImpl interface {
	Handle(kind string, payload cbor.RawMessage)
	Mount(c *RawStreamPlug)
	CloseSignal()
}

// RawStreamPlug is a low-level CBOR-based plugin communication framework.
//
// It provides raw input/output streams without automatic validation or message dispatching.
// The plugin implementation (RawStreamPlugImpl) is fully responsible for interpreting incoming messages
// and producing appropriate responses.
//
// RawStreamPlug handles system signals (e.g., SIGINT, SIGTERM) and supports graceful shutdown via internal signals.
// Each incoming message is processed asynchronously in its own goroutine.
// Ordering of responses is not guaranteed and must be handled by the plugin if needed.
//
// Usage:
//   - Initialize RawStreamPlug with a RawStreamPlugImpl implementation.
//   - Call Main() to start the event loop.
//   - Call Shutdown() to terminate the plug from the implementation.
type RawStreamPlug struct {
	PlugImpl RawStreamPlugImpl
	decoder  *cbor.Decoder
	encoder  *cbor.Encoder
	wg       *sync.WaitGroup
	ossig    context.Context
	implsig  context.Context
	cancel   context.CancelFunc
	osstop   context.CancelFunc
}

func NewRawStreamPlug(impl RawStreamPlugImpl) *RawStreamPlug {
	return &RawStreamPlug{
		PlugImpl: impl,
	}
}

// Main starts the main loop of the RawStreamPlug.
func (p *RawStreamPlug) Main() {
	p.PlugImpl.Mount(p)
	p.decoder = cbor.NewDecoder(os.Stdin)
	p.encoder = cbor.NewEncoder(os.Stdout)
	p.wg = &sync.WaitGroup{}
	p.ossig, p.osstop = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	p.implsig, p.cancel = context.WithCancel(context.Background())

	p.wg.Add(1)
	go p.Loop()
	p.wg.Wait()
}

// Send sends an Envelope with the message code and CBOR payload to stdout.
func (p *RawStreamPlug) Send(messageCode string, payload cbor.RawMessage) {
	err := p.encoder.Encode(messages.Envelope{
		Version: 1,
		Type:    messageCode,
		Raw:     payload,
	})
	if err != nil {
		// If we can't write the response, panic — plugin cannot recover.
		panic(err)
	}
}

// Loop contains the main logic of the RawStreamPlug. It takes care of decoding the incoming
// CBOR payload and sends it to handler asynchronously.
// Handler must decode the type of the message and respond accordingly. This plug type does
// not guarantee the order of incoming and outgoing messages.
// It is up to implementer to handle logic.
func (p *RawStreamPlug) Loop() {

loop:
	for {
		select {
		case <-p.ossig.Done():

			p.wg.Add(1)
			go func() {
				p.PlugImpl.CloseSignal()
				p.osstop()
				p.wg.Done()
			}()
			break loop
		case <-p.implsig.Done():
			p.wg.Done()
			break loop
		default:
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

			p.wg.Add(1)
			go p.responseWrapper(msg)

		}
	}

}

func (p *RawStreamPlug) responseWrapper(msg messages.Envelope) {
	p.PlugImpl.Handle(msg.Type, msg.Raw)
	p.wg.Done()

}

// Shutdown sends signal to shut down the plug. As the plug can be long living, it has to have a control mechanism
// to shut down the plug from the implementation.
func (p *RawStreamPlug) Shutdown() {
	p.cancel()
}
