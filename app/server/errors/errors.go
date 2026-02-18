package errors

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

// Errors the client will see
type ClientError struct {
	msg string
}

func newClientError(msg string) *ClientError {
	return &ClientError{msg: msg}
}

func (e *ClientError) Error() string { return e.msg }

func (e *ClientError) SerializeToResp() []byte { return resp.SimpleError(e.msg) }

func NumArgs(command string) error {
	return newClientError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", command))
}

func UknownCommand(command string) error {
	return newClientError(fmt.Sprintf("ERR unknown command '%s'", command))
}

var (
	InvalidInteger    = newClientError("ERR value is not an integer or out of range")
	SyntaxError       = newClientError("ERR syntax error")
	WrongType         = newClientError("WRONGTYPE Operation against a key holding the wrong kind of value")
	MustBePositive    = newClientError("ERR value is out of range, must be positive")
	TimeoutNotFloat   = newClientError("ERR timeout is not a float or out of range")
	TimeoutNotInt     = newClientError("ERR timeout is not an integer or out of range")
	TimeoutNegative   = newClientError("ERR timeout is negative")
	UnbalancedXread   = newClientError("ERR Unbalanced 'xread' list of streams: for each stream key an ID, '+', or '$' must be specified.")
	NestedTransaction = newClientError("ERR MULTI calls can not be nested")
	ExecWithoutMulti  = newClientError("ERR EXEC without MULTI")
)

// A special error value that indicates whether an operation
// is waiting on something or not. e.g: BLPOP
// Note: commands inside a transaction should never return errors.Blocked
var Blocked = errors.New("operation blocked")

var (
	As  = errors.As
	Is  = errors.Is
	New = errors.New
)
