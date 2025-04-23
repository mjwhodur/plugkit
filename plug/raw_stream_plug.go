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
	Handle(kind string, payload cbor.RawMessage)
	Mount(c *RawStreamPlug)
	CloseSignal()
}

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

func (p *RawStreamPlug) Shutdown() {
	p.cancel()
}
