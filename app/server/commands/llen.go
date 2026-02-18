package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/lists"
)

type llen struct {
	key string
}

func parseLlen(command string, args []string) (*llen, error) {
	if len(args) != 1 {
		return nil, errors.NumArgs(command)
	}

	return &llen{key: args[0]}, nil
}

func (cmd *llen) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.Integer(0), nil
	}

	l, ok := o.Value.(lists.List)
	if !ok {
		return nil, errors.WrongType
	}

	return resp.Integer(len(l)), nil
}
