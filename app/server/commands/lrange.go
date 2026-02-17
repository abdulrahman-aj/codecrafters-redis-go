package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/lists"
)

type lrange struct {
	key   string
	start int
	end   int
}

func ParseLrange(ctx *request.Context) (*lrange, error) {
	if len(ctx.Args) != 3 {
		return nil, errors.NumArgs(ctx)
	}

	start, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return nil, errors.InvalidInteger
	}
	end, err := strconv.Atoi(ctx.Args[2])
	if err != nil {
		return nil, errors.InvalidInteger
	}

	return &lrange{key: ctx.Args[0], start: start, end: end}, nil
}

func (cmd *lrange) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.Array(nil), nil
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return nil, errors.WrongType
	}

	n := len(l)

	normalizeIndex := func(x int) int {
		if x < 0 {
			x += n
		}

		return max(0, x)
	}

	start := normalizeIndex(cmd.start)
	end := normalizeIndex(cmd.end)

	if start > end || start >= n || n == 0 {
		return resp.Array(nil), nil
	}

	return resp.Array(l[start:min(end+1, n)]), nil
}
