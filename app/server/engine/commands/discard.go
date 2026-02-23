package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type discard struct{}

func parseDiscard(command string, args []string) (*discard, error) {
	if len(args) != 0 {
		return nil, errors.NumArgs(command)
	}

	return &discard{}, nil
}

func (cmd *discard) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	if !ctx.Connection.InsideTx {
		return nil, errors.DiscardWithoutMulti
	}

	ctx.Connection.InsideTx = false
	ctx.Connection.TxCommands = nil

	return resp.SimpleString("OK"), nil
}
