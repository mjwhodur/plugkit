package main

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/mjwhodur/plugkit/examples/3-rawstream-basic/shared"
	"github.com/mjwhodur/plugkit/helpers"
	"github.com/mjwhodur/plugkit/plug"
)

type StreamPlugExample struct {
	transport *plug.RawStreamPlug
}

func (s *StreamPlugExample) Handle(kind string, payload cbor.RawMessage) {
	switch kind {
	case "ping":
		var pingmsg shared.Ping
		e := cbor.Unmarshal(payload, &pingmsg)
		if e != nil {
			panic(e)
		}
		if pingmsg.ID == 3 {
			s.transport.Send("pong-3", helpers.MustRaw(&shared.Pong{Message: "Ending Pong"}))
			s.transport.Shutdown()
		} else {
			s.transport.Send("pong", helpers.MustRaw(&shared.Pong{Message: "Just a Pong with id < 3"}))

		}
	default:
		panic("unknown kind " + kind)
	}
}

func (s *StreamPlugExample) Mount(c *plug.RawStreamPlug) {
	s.transport = c
}

func (s *StreamPlugExample) CloseSignal() {
	// Pretend we're doing something
	// Normally we'd clean up the plugin implementation, close the connections and so on
}

func main() {
	impl := &StreamPlugExample{}
	streamPlug := plug.RawStreamPlug{PlugImpl: impl}
	streamPlug.Main()
}
