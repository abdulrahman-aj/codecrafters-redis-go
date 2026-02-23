package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type discard struct{}

func parseDiscard(command string, args []string) (*discard, []byte) {
	if len(args) != 0 {
		return nil, rediserrors.NumArgs(command)
	}

	return &discard{}, nil
}

func (cmd *discard) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	if !ctx.Conn.InsideTx {
		return rediserrors.DiscardWithoutMulti, nil
	}

	ctx.Conn.InsideTx = false
	ctx.Conn.TxCommands = nil

	return resp.SimpleString("OK"), nil
}
