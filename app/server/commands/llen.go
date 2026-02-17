package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/lists"
)

type llen struct {
	key string
}

func ParseLlen(ctx *request.Context) (*llen, error) {
	if len(ctx.Args) != 1 {
		return nil, errors.NumArgs(ctx)
	}

	return &llen{key: ctx.Args[0]}, nil
}

func (cmd *llen) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.Integer(0), nil
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return nil, errors.WrongType
	}

	return resp.Integer(len(l)), nil
}
