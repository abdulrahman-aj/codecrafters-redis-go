package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type multi struct{}

func parseMulti(command string, args []string) (*multi, []byte) {
	if len(args) != 0 {
		return nil, rediserrors.NumArgs(command)
	}

	return &multi{}, nil
}

func (cmd *multi) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	if ctx.Connection.InsideTx {
		return rediserrors.NestedTransaction, nil
	}

	ctx.Connection.InsideTx = true

	return resp.SimpleString("OK"), nil
}
