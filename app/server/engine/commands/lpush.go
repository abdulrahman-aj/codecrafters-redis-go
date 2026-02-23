package commands

import (
	"slices"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
)

type lpush struct {
	key    string
	values []string
}

func parseLpush(command string, args []string) (*lpush, error) {
	if len(args) < 2 {
		return nil, errors.NumArgs(command)
	}

	return &lpush{key: args[0], values: args[1:]}, nil
}

func (cmd *lpush) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
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
