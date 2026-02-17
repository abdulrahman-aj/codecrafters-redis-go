package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/lists"
)

type lpop struct {
	key   string
	count int
}

func ParseLpop(ctx *request.Context) (*lpop, error) {
	if len(ctx.Args) == 0 || len(ctx.Args) > 2 {
		return nil, errors.NumArgs(ctx)
	}

	count := 1

	if len(ctx.Args) == 2 {
		var err error
		count, err = strconv.Atoi(ctx.Args[1])
		if err != nil || count < 0 {
			return nil, errors.MustBePositive
		}
	}

	return &lpop{key: ctx.Args[0], count: count}, nil
}

func (cmd *lpop) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.NullBulkString, nil
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return nil, errors.WrongType
	}

	count := min(cmd.count, len(l))
	ret := l[:count]
	o.Value = l[count:]

	s.Set(ctx, cmd.key, o)

	if len(ctx.Args) == 1 { // hacky
		return resp.BulkString(ret[0]), nil
	}

	return resp.Array(ret), nil
}
