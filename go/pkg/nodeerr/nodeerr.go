package nodeerr

type ErrorCode int

const (
	Timeout            ErrorCode = 0  // "Timed out"
	NodeNotFound       ErrorCode = 1  // "Node not found"
	NotSupported       ErrorCode = 10 // "Operation not supported"
	Unavailable        ErrorCode = 11 // "Temporary unavailable"
	Malformed          ErrorCode = 12 // "Malformed request"
	Crashed            ErrorCode = 13 // "Crashed"
	Aborted            ErrorCode = 14 // "Aborted"
	KeyNotExist        ErrorCode = 20 // "Key does not exist"
	KeyAlreadyExist    ErrorCode = 21 // "Key already exist"
	PreconditionFailed ErrorCode = 22 // "Precondition failed"
	TXConflict         ErrorCode = 30 // "Transaction conflict"
)

func (e ErrorCode) String() string {
	switch e {
	case Timeout:
		return "Timed out"
	case NodeNotFound:
		return "Node not found"
	case NotSupported:
		return "Operation not supported"
	case Unavailable:
		return "Temporary unavailable"
	case Malformed:
		return "Malformed request"
	case Crashed:
		return "Crashed"
	case Aborted:
		return "Aborted"
	case KeyNotExist:
		return "Key does not exist"
	case KeyAlreadyExist:
		return "Key already exist"
	case PreconditionFailed:
		return "Precondition failed"
	case TXConflict:
		return "Transaction conflict"
	default:
		panic("unreachable")
	}
}
