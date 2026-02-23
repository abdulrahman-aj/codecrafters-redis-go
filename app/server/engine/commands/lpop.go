package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
)

type lpop struct {
	key              string
	count            int
	respondWithArray bool
}

func parseLpop(command string, args []string) (*lpop, error) {
	if len(args) == 0 || len(args) > 2 {
		return nil, errors.NumArgs(command)
	}

	count := 1

	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(args[1])
		if err != nil || count < 0 {
			return nil, errors.MustBePositive
		}
	}

	return &lpop{
		key:              args[0],
		count:            count,
		respondWithArray: len(args) == 2,
	}, nil
}

func (cmd *lpop) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
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

	if cmd.respondWithArray { // hacky
		return resp.Array(ret), nil
	}

	return resp.BulkString(ret[0]), nil
}
