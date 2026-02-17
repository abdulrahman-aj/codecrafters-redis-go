package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/streams"
)

type typeCmd struct {
	key string
}

func ParseType(ctx *request.Context) (*typeCmd, error) {
	if len(ctx.Args) != 1 {
		return nil, errors.NumArgs(ctx)
	}

	return &typeCmd{key: ctx.Args[0]}, nil
}

func (cmd *typeCmd) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
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
