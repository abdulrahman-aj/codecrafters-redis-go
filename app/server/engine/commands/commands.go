package commands

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type Command interface {
	Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error)
}

func Parse(ctx *types.RequestCtx, command string, args []string) (Command, []byte) {
	switch command {
	case "multi":
		return parseMulti(command, args)
	case "exec":
		return parseExec(command, args)
	case "discard":
		return parseDiscard(command, args)
	}

	if ctx.Conn.InsideTx {
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
	case "replconf":
		return parseReplConf(command, args)
	default:
		return nil, rediserrors.UknownCommand(command)
	}
}

type txQueue struct {
	Command string
	Args    []string
}

func (cmd *txQueue) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	txCommand := types.TxCommand{Command: cmd.Command, Args: cmd.Args}
	ctx.Conn.TxCommands = append(ctx.Conn.TxCommands, txCommand)
	return resp.SimpleString("QUEUED"), nil
}

// A special error value that indicates whether an operation
// is waiting on something or not. e.g: BLPOP
// Note: commands inside a transaction should never return ErrBlocked
var ErrBlocked = errors.New("command blocked waiting on something")
