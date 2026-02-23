package commands

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/streams"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type typeCmd struct {
	key string
}

func parseType(command string, args []string) (*typeCmd, []byte) {
	if len(args) != 1 {
		return nil, rediserrors.NumArgs(command)
	}

	return &typeCmd{key: args[0]}, nil
}

func (cmd *typeCmd) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		return resp.SimpleString("none"), nil
	}

	switch o.Value.(type) {
	case string:
		return resp.SimpleString("string"), nil
	case lists.List:
		return resp.SimpleString("list"), nil
	case streams.Stream:
		return resp.SimpleString("stream"), nil
	default:
		return nil, errors.New("unknown data type for command 'type'")
	}
}
