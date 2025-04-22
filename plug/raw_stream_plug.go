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

type RawStreamPlugImpl interface {
	Handle(kind string, payload *cbor.RawMessage) (messageCode string, response cbor.RawMessage, err error)
	Mount(c *RawStreamPlug)
	CloseSignal()
}

type RawStreamPlug struct {
	PlugImpl RawStreamPlugImpl
	decoder  *cbor.Decoder
	encoder  *cbor.Encoder
	wg       *sync.WaitGroup
	ctx      context.Context
}

func NewRawStreamPlug(impl RawStreamPlugImpl) *RawStreamPlug {
	return &RawStreamPlug{
		PlugImpl: impl,
	}
}

func (p *RawStreamPlug) Main() {
	p.PlugImpl.Mount(p)
	p.decoder = cbor.NewDecoder(os.Stdin)
	p.encoder = cbor.NewEncoder(os.Stdout)
	p.wg = &sync.WaitGroup{}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	p.ctx = ctx

	p.wg.Add(1)
	go p.Loop()
	p.wg.Wait()
}

func (p *RawStreamPlug) Respond(messageCode string, payload cbor.RawMessage) {
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

func (p *RawStreamPlug) Loop() {

loop:
	for {
		select {
		case <-p.ctx.Done():

			p.wg.Add(1)
			go p.PlugImpl.CloseSignal()
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
	msgCode, res, err := p.PlugImpl.Handle(msg.Type, &msg.Raw)
	if err != nil {
		// Return a handling error with the error as payload.
		p.Respond(string(codes.HandlingError), helpers.MustRaw(err))

	}

	// Send the response with the provided message code and payload.
	p.Respond(msgCode, res)
	p.wg.Done()

}
