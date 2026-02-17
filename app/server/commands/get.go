package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type get struct {
	key string
}

func ParseGet(ctx *request.Context) (*get, error) {
	if len(ctx.Args) != 1 {
		return nil, errors.NumArgs(ctx)
	}

	return &get{key: ctx.Args[0]}, nil
}

func (cmd *get) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.NullBulkString, nil
	}

	valueStr, ok := o.Value.(string)
	if !ok {
		return nil, errors.WrongType
	}

	return resp.BulkString(valueStr), nil
}
