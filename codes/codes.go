// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

package codes

// MessageCode represents a symbolic identifier used for communication between
// the plugin and the host within the PlugKit framework.
//
// This type is open to extension — developers may define custom message codes
// as needed for their own application-specific communication.
//
// However, a set of predefined message codes is reserved by PlugKit for internal use
// (e.g., "PLUGKIT_Exit", "PLUGKIT_Unsupported"). These must not be redefined or
// misused, as doing so may interfere with the behavior of the PlugKit runtime.
//
// It is recommended to:
//
//   - Avoid redefining or repurposing PlugKit-reserved codes.
//   - Use unique message code prefixes for custom messages (e.g., "X_", "APP_", etc.).
//   - Check for name collisions if writing plugins meant to work with generic hosts.
type MessageCode string

const (
	// ExitMessage indicates that the host intends plug to exit or shut down.
	ExitMessage MessageCode = "PLUGKIT_Exit"

	// Unsupported indicates that the received request or feature is not supported
	// by the plugin or host.
	Unsupported MessageCode = "PLUGKIT_Unsupported"

	// FinishMessage signals the successful completion of a task by the plugin.
	FinishMessage MessageCode = "PLUGKIT_Finish"

	// PluginResponse was intended to indicate a direct response from the plugin
	// to a request, but is currently unused. FIXME: Consider removing or implementing it.
	PluginResponse MessageCode = "PLUGKIT_Response"
)

// PluginExitReason defines standard exit codes used by plugins in the PlugKit system.
// Some values correspond to standard POSIX `sysexits.h` codes, while others are custom,
// PlugKit-specific statuses.
//
// Values in the range 0–64 are reserved for PlugKit-defined statuses and may be extended
// for custom application-level signaling.
//
// Values 65 and above are aligned with POSIX exit codes where applicable.
//
// It is encouraged to use the predefined constants to improve cross-plugin consistency.
// If you define your own exit reasons, avoid overlapping with known POSIX values
// and document their purpose clearly.
type PluginExitReason int

const (
	// PlugKit-defined exit reasons
	OperationSuccess               PluginExitReason = iota // Operation completed successfully (equivalent to code 0)
	OperationError                                         // Generic error during operation (like code 1)
	MisuseOfShellBuiltins                                  // Misuse of command syntax or reserved functionality
	OperationTimeout                                       // The operation took too long and timed out
	OperationCancelledByClient                             // The operation was cancelled by the client
	OperationCancelledByPlugin                             // The operation was cancelled internally by the plugin
	OperationSucceededWithWarnings                         // Operation finished, but with warnings
	PluginToHostCommunicationError                         // Failure in plugin → host communication
	HostToPluginCommunicationError                         // Failure in host → plugin communication
	PlugNotStarted                                         // Plugin was not started when expected
	PlugCrashed                                            // Plugin crashed or exited abnormally

	/// POSIX-aligned exit codes (based on sysexits.h)

	DataFormatError           = 65 // EX_DATAERR: The input data was incorrect in some way
	ErrNoInput                = 66 // EX_NOINPUT: Cannot open input
	ErrServiceUnavailable     = 69 // EX_UNAVAILABLE: Service unavailable
	InternalSoftwareError     = 70 // EX_SOFTWARE: Internal software error
	ErrOsError                = 71 // EX_OSERR: OS-level error
	ErrCriticalOsFileMissing  = 72 // EX_OSFILE: Critical file missing
	ErrCannotCreateOutputFile = 73 // EX_CANTCREAT: Cannot create output
	ErrIoError                = 74 // EX_IOERR: Input/output error
	TemporaryFailure          = 75 // EX_TEMPFAIL: Temporary failure, retry later
	RemoteErrorInProtocol     = 76 // EX_PROTOCOL: Remote protocol error
	PermissionDenied          = 77 // EX_NOPERM: Permission denied
	ConfigurationError        = 78 // EX_CONFIG: Configuration error

	CommandInvokedCannotExecute = 126 // Command found but is not executable
	CommandNotFound             = 127 // Command not found
	InvalidArgumentToExit       = 129 // Exit called with invalid argument
	ExitStatusOutOfRange        = 255 // Maximum exit code range exceeded
)

// String returns a human-readable name for the PluginExitReason code.
// If the code is not recognized, it returns "UnknownExternalStatusCode".
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
		return "UnknownExternalStatusCode"
	}
}
