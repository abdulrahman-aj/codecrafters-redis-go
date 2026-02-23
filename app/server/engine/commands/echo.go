package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type echo struct {
	arg string
}

func parseEcho(command string, args []string) (*echo, error) {
	if len(args) != 1 {
		return nil, errors.NumArgs(command)
	}
	return &echo{arg: args[0]}, nil
}

func (cmd *echo) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	return resp.BulkString(cmd.arg), nil
}
