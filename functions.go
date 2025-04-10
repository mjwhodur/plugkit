package plugkit

import (
	"github.com/mjwhodur/plugkit/options"
	"github.com/mjwhodur/plugkit/plug"
)

// NewPlug returns new empty plug
func NewPlug() *plug.Plug {
	return plug.New()
}

// CreateAndRunPlug creates functional-like plug and runs it
func CreateAndRunPlug(handlers *[]options.CommandHandler) error {
	p := plug.New()
	for _, h := range *handlers {
		p.HandleMessageType(h.Command, h.Handler)
	}
	return p.Main()
}

// func NewBiDiPlug()   {}
func NewClient()     {}
func NewBiDiClient() {}
