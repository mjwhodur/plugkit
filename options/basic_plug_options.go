package options

import (
	"github.com/mjwhodur/plugkit/codes"
	"github.com/mjwhodur/plugkit/messages"
)

type CommandHandler struct {
	Command string
	Handler func(p []byte) (*messages.Result, codes.PluginExitReason, error)
}
