package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type get struct {
	key string
}

func parseGet(command string, args []string) (*get, error) {
	if len(args) != 1 {
		return nil, errors.NumArgs(command)
	}

	return &get{key: args[0]}, nil
}

func (cmd *get) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.NullBulkString, nil
	}

	valueStr, ok := o.Value.(string)
	if !ok {
		return nil, errors.WrongType
	}

	return resp.BulkString(valueStr), nil
}
