package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type echo struct {
	arg string
}

func parseEcho(command string, args []string) (*echo, []byte) {
	if len(args) != 1 {
		return nil, rediserrors.NumArgs(command)
	}
	return &echo{arg: args[0]}, nil
}

func (cmd *echo) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	return resp.BulkString(cmd.arg), nil
}
