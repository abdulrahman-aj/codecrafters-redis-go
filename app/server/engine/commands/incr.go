package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type incr struct {
	key string
}

func parseIncr(command string, args []string) (*incr, []byte) {
	if len(args) != 1 {
		return nil, rediserrors.NumArgs(command)
	}

	return &incr{key: args[0]}, nil
}

func (cmd *incr) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: "0"}
	}

	asStr, ok := o.Value.(string)
	if !ok {
		return rediserrors.WrongType, nil
	}

	asInt, err := strconv.Atoi(asStr)
	if err != nil {
		return rediserrors.InvalidInteger, nil
	}

	asInt++

	o.Value = strconv.Itoa(asInt)
	s.Set(ctx, cmd.key, o)

	return resp.Integer(asInt), nil
}
