package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type llen struct {
	key string
}

func parseLlen(command string, args []string) (*llen, []byte) {
	if len(args) != 1 {
		return nil, rediserrors.NumArgs(command)
	}

	return &llen{key: args[0]}, nil
}

func (cmd *llen) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.Integer(0), nil
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return rediserrors.WrongType, nil
	}

	return resp.Integer(len(l)), nil
}
