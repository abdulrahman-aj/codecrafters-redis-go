package commands

import (
	"slices"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/lists"
)

type lpush struct {
	key    string
	values []string
}

func ParseLpush(ctx *request.Context) (*lpush, error) {
	if len(ctx.Args) < 2 {
		return nil, errors.NumArgs(ctx)
	}

	return &lpush{key: ctx.Args[0], values: ctx.Args[1:]}, nil
}

func (cmd *lpush) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: lists.List{}}
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return nil, errors.WrongType
	}

	values := slices.Clone(cmd.values)
	slices.Reverse(values)
	values = append(values, l...)

	o.Value = values

	s.Set(ctx, cmd.key, o)

	return resp.Integer(len(values)), nil
}
