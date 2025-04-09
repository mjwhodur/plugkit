package codes

type MessageCode string

const (
	ExitMessage   MessageCode = "Exit"
	Unsupported   MessageCode = "Unsupported"
	FinishMessage MessageCode = "Finish"
)

type PluginExitReason int

const (
	OperationSuccess PluginExitReason = iota
	OperationError
	OperationTimeout
	OperationCancelledByClient
	OperationCancelledByPlugin
	OperationSucceededWithWarnings
)
