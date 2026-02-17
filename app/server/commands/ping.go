package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type ping struct{}

func ParsePing(ctx *request.Context) (*ping, error) {
	if len(ctx.Args) != 0 {
		return nil, errors.NumArgs(ctx)
	}

	return &ping{}, nil
}

func (cmd *ping) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	return resp.SimpleString("PONG"), nil
}
