package errors

import "errors"

// A special error value that indicates whether an operation
// is waiting on something or not. e.g: BLPOP
// Note: commands inside a transaction should never return errors.Blocked
var Blocked = errors.New("operation blocked")

var (
	As  = errors.As
	Is  = errors.Is
	New = errors.New
)
