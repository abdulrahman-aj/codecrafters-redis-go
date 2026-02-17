package commands

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/lists"
)

type blpop struct {
	key string
}

func ParseBlpop(ctx *request.Context) (*blpop, error) {

	if len(ctx.Args) != 2 {
		return nil, errors.NumArgs(ctx)
	}

	cmd := &blpop{key: ctx.Args[0]}

	timeout, err := strconv.ParseFloat(ctx.Args[1], 64)
	if err != nil {
		return nil, errors.TimeoutNotFloat
	}

	if timeout != 0 {
		ctx.SetTimeout(time.Duration(timeout * float64(time.Second)))
	}

	ctx.Dependencies[cmd.key] = true

	return cmd, nil
}

func (cmd *blpop) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	if ctx.IsExpired() {
		return resp.NullArray, nil
	}

	o, ok := s.Get(cmd.key)
	if !ok {
		return nil, errors.Blocked
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return nil, errors.Blocked
	}

	o.Value = l[1:]
	s.Set(ctx, cmd.key, o)

	return resp.Array([]string{cmd.key, l[0]}), nil
}
