package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/streams"
)

type typeCmd struct {
	key string
}

func parseType(command string, args []string) (*typeCmd, error) {
	if len(args) != 1 {
		return nil, errors.NumArgs(command)
	}

	return &typeCmd{key: args[0]}, nil
}

func (cmd *typeCmd) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
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
		panic("unknown type?")
	}
}
