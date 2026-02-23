package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type get struct {
	key string
}

func parseGet(command string, args []string) (*get, []byte) {
	if len(args) != 1 {
		return nil, rediserrors.NumArgs(command)
	}

	return &get{key: args[0]}, nil
}

func (cmd *get) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.NullBulkString, nil
	}

	valueStr, ok := o.Value.(string)
	if !ok {
		return rediserrors.WrongType, nil
	}

	return resp.BulkString(valueStr), nil
}
