package commands

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type blpop struct {
	key        string
	hasTimeout bool
	timeout    time.Duration
}

func parseBlpop(command string, args []string) (*blpop, []byte) {
	if len(args) != 2 {
		return nil, rediserrors.NumArgs(command)
	}

	timeout, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return nil, rediserrors.TimeoutNotFloat
	}

	return &blpop{
		key:        args[0],
		hasTimeout: timeout != 0,
		timeout:    time.Duration(timeout * float64(time.Second)),
	}, nil
}

func (cmd *blpop) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	ctx.Dependencies[cmd.key] = true

	if cmd.hasTimeout {
		ctx.SetTimeout(cmd.timeout)
	}

	o, ok := s.Get(cmd.key)
	if !ok {
		if ctx.DeadlineExceeded() {
			return resp.NullArray, nil
		}

		return nil, ErrBlocked
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		if ctx.DeadlineExceeded() {
			return resp.NullArray, nil
		}

		return nil, ErrBlocked
	}

	o.Value = l[1:]
	s.Set(ctx, cmd.key, o)

	return resp.Array([]string{cmd.key, l[0]}), nil
}
