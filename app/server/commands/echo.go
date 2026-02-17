package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type echo struct {
	arg string
}

func ParseEcho(ctx *request.Context) (*echo, error) {
	if len(ctx.Args) != 1 {
		return nil, errors.NumArgs(ctx)
	}
	return &echo{arg: ctx.Args[0]}, nil
}

func (cmd *echo) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	return resp.BulkString(cmd.arg), nil
}
