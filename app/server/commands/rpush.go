package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/lists"
)

type rpush struct {
	key    string
	values []string
}

func ParseRpush(ctx *request.Context) (*rpush, error) {
	if len(ctx.Args) < 2 {
		return nil, errors.NumArgs(ctx)
	}

	return &rpush{key: ctx.Args[0], values: ctx.Args[1:]}, nil
}

func (cmd *rpush) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: lists.List{}}
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return nil, errors.WrongType
	}

	l = append(l, cmd.values...)
	o.Value = l

	s.Set(ctx, cmd.key, o)

	return resp.Integer(len(l)), nil
}
