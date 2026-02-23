package commands

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type Command interface {
	Exec(ctx *context.Request, s *store.Store) ([]byte, error)
}

func Parse(ctx *context.Request, command string, args []string) (Command, []byte) {
	switch command {
	case "multi":
		return parseMulti(command, args)
	case "exec":
		return parseExec(command, args)
	case "discard":
		return parseDiscard(command, args)
	}

	if ctx.Connection.InsideTx {
		return &txQueue{Command: command, Args: args}, nil
	}

	switch command {
	case "ping":
		return parsePing(command, args)
	case "echo":
		return parseEcho(command, args)
	case "set":
		return parseSet(command, args)
	case "get":
		return parseGet(command, args)
	case "rpush":
		return parseRpush(command, args)
	case "lpush":
		return parseLpush(command, args)
	case "lrange":
		return parseLrange(command, args)
	case "llen":
		return parseLlen(command, args)
	case "lpop":
		return parseLpop(command, args)
	case "blpop":
		return parseBlpop(command, args)
	case "type":
		return parseType(command, args)
	case "xadd":
		return parseXadd(command, args)
	case "xrange":
		return parseXrange(command, args)
	case "xread":
		return parseXread(command, args)
	case "incr":
		return parseIncr(command, args)
	case "info":
		return parseInfo(command, args)
	default:
		return nil, rediserrors.UknownCommand(command)
	}
}

type txQueue struct {
	Command string
	Args    []string
}

func (cmd *txQueue) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	txCommand := context.TxCommand{Command: cmd.Command, Args: cmd.Args}
	ctx.Connection.TxCommands = append(ctx.Connection.TxCommands, txCommand)
	return resp.SimpleString("QUEUED"), nil
}

// A special error value that indicates whether an operation
// is waiting on something or not. e.g: BLPOP
// Note: commands inside a transaction should never return ErrBlocked
var ErrBlocked = errors.New("command blocked waiting on something")
