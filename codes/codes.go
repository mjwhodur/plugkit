package codes

type MessageCode string

const (
	ExitMessage    MessageCode = "PLUGKIT_Exit"
	Unsupported    MessageCode = "PLUGKIT_Unsupported"
	FinishMessage  MessageCode = "PLUGKIT_Finish"
	PluginResponse MessageCode = "PLUGKIT_Response"
)

type PluginExitReason int

const (
	OperationSuccess PluginExitReason = iota
	OperationError
	MisuseOfShellBuiltins
	OperationTimeout
	OperationCancelledByClient
	OperationCancelledByPlugin
	OperationSucceededWithWarnings
	PluginToHostCommunicationError
	HostToPluginCommunicationError

	DataFormatError           = 65
	ErrNoInput                = 66
	ErrServiceUnavailable     = 69
	InternalSoftwareError     = 70
	ErrOsError                = 71
	ErrCriticalOsFileMissing  = 72
	ErrCannotCreateOutputFile = 73
	ErrIoError                = 74
	TemporaryFailure          = 75
	RemoteErrorInProtocol     = 76
	PermissionDenied          = 77
	ConfigurationError        = 78

	CommandInvokedCannotExecute = 126
	CommandNotFound             = 127
	InvalidArgumentToExit       = 129
	ExitStatusOutOfRange        = 255
)

func (r PluginExitReason) String() string {
	switch r {
	case OperationSuccess:
		return "OperationSuccess"
	case OperationError:
		return "OperationError"
	case MisuseOfShellBuiltins:
		return "MisuseOfShellBuiltins"
	case OperationTimeout:
		return "OperationTimeout"
	case OperationCancelledByClient:
		return "OperationCancelledByClient"
	case OperationCancelledByPlugin:
		return "OperationCancelledByPlugin"
	case OperationSucceededWithWarnings:
		return "OperationSucceededWithWarnings"
	case PluginToHostCommunicationError:
		return "PluginToHostCommunicationError"
	case HostToPluginCommunicationError:
		return "HostToPluginCommunicationError"
	case DataFormatError:
		return "DataFormatError"
	case ErrNoInput:
		return "ErrNoInput"
	case ErrServiceUnavailable:
		return "ErrServiceUnavailable"
	case InternalSoftwareError:
		return "InternalSoftwareError"
	case ErrOsError:
		return "ErrOsError"
	case ErrCriticalOsFileMissing:
		return "ErrCriticalOsFileMissing"
	case ErrCannotCreateOutputFile:
		return "ErrCannotCreateOutputFile"
	case ErrIoError:
		return "ErrIoError"
	case TemporaryFailure:
		return "TemporaryFailure"
	case RemoteErrorInProtocol:
		return "RemoteErrorInProtocol"
	case PermissionDenied:
		return "PermissionDenied"
	case ConfigurationError:
		return "ConfigurationError"
	case CommandInvokedCannotExecute:
		return "CommandInvokedCannotExecute"
	case CommandNotFound:
		return "CommandNotFound"
	case InvalidArgumentToExit:
		return "InvalidArgumentToExit"
	case ExitStatusOutOfRange:
		return "ExitStatusOutOfRange"
	default:
		panic("unhandled default case")
	}
}
