package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type rpush struct {
	key    string
	values []string
}

func parseRpush(command string, args []string) (*rpush, []byte) {
	if len(args) < 2 {
		return nil, rediserrors.NumArgs(command)
	}

	return &rpush{key: args[0], values: args[1:]}, nil
}

func (cmd *rpush) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: lists.List{}}
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return rediserrors.WrongType, nil
	}

	l = append(l, cmd.values...)
	o.Value = l

	s.Set(ctx, cmd.key, o)

	return resp.Integer(len(l)), nil
}
