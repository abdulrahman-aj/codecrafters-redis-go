package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type lrange struct {
	key   string
	start int
	end   int
}

func parseLrange(command string, args []string) (*lrange, []byte) {
	if len(args) != 3 {
		return nil, rediserrors.NumArgs(command)
	}

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, rediserrors.InvalidInteger
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, rediserrors.InvalidInteger
	}

	return &lrange{key: args[0], start: start, end: end}, nil
}

func (cmd *lrange) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.Array(nil), nil
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return rediserrors.WrongType, nil
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
