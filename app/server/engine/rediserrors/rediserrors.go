package rediserrors

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func NumArgs(command string) []byte {
	return resp.SimpleError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", command))
}

func UknownCommand(command string) []byte {
	return resp.SimpleError(fmt.Sprintf("ERR unknown command '%s'", command))
}

var (
	InvalidInteger      = resp.SimpleError("ERR value is not an integer or out of range")
	SyntaxError         = resp.SimpleError("ERR syntax error")
	WrongType           = resp.SimpleError("WRONGTYPE Operation against a key holding the wrong kind of value")
	MustBePositive      = resp.SimpleError("ERR value is out of range, must be positive")
	TimeoutNotFloat     = resp.SimpleError("ERR timeout is not a float or out of range")
	TimeoutNotInt       = resp.SimpleError("ERR timeout is not an integer or out of range")
	TimeoutNegative     = resp.SimpleError("ERR timeout is negative")
	UnbalancedXread     = resp.SimpleError("ERR Unbalanced 'xread' list of streams: for each stream key an ID, '+', or '$' must be specified.")
	NestedTransaction   = resp.SimpleError("ERR MULTI calls can not be nested")
	ExecWithoutMulti    = resp.SimpleError("ERR EXEC without MULTI")
	DiscardWithoutMulti = resp.SimpleError("ERR DISCARD without MULTI")
	InvalidStreamID     = resp.SimpleError("ERR Invalid stream ID specified as stream command argument")
	XaddEqualOrSmaller  = resp.SimpleError("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	XaddZeroID          = resp.SimpleError("ERR The ID specified in XADD must be greater than 0-0")
)
