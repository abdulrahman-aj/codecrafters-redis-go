package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type multi struct{}

func parseMulti(command string, args []string) (*multi, []byte) {
	if len(args) != 0 {
		return nil, rediserrors.NumArgs(command)
	}

	return &multi{}, nil
}

func (cmd *multi) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	if ctx.Conn.InsideTx {
		return rediserrors.NestedTransaction, nil
	}

	ctx.Conn.InsideTx = true

	return resp.SimpleString("OK"), nil
}
