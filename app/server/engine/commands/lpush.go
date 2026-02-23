package commands

import (
	"slices"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type lpush struct {
	key    string
	values []string
}

func parseLpush(command string, args []string) (*lpush, []byte) {
	if len(args) < 2 {
		return nil, rediserrors.NumArgs(command)
	}

	return &lpush{key: args[0], values: args[1:]}, nil
}

func (cmd *lpush) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: lists.List{}}
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return rediserrors.WrongType, nil
	}

	values := slices.Clone(cmd.values)
	slices.Reverse(values)
	values = append(values, l...)

	o.Value = values

	s.Set(ctx, cmd.key, o)

	return resp.Integer(len(values)), nil
}
