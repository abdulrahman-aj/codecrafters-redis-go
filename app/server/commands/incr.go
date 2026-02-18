package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type incr struct {
	key string
}

func ParseIncr(ctx *request.Context) (*incr, error) {
	if len(ctx.Args) != 1 {
		return nil, errors.NumArgs(ctx)
	}

	return &incr{key: ctx.Args[0]}, nil
}

func (cmd *incr) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: "0"}
	}

	asStr, ok := o.Value.(string)
	if !ok {
		return nil, errors.WrongType
	}

	asInt, err := strconv.Atoi(asStr)
	if err != nil {
		return nil, errors.InvalidInteger
	}

	asInt++

	o.Value = strconv.Itoa(asInt)
	s.Set(ctx, cmd.key, o)

	return resp.Integer(asInt), nil
}
