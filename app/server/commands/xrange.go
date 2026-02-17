package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/streams"
)

type xrange struct {
	key   string
	start string
	end   string
}

func ParseXrange(ctx *request.Context) (*xrange, error) {
	if len(ctx.Args) != 3 {
		return nil, errors.NumArgs(ctx)
	}

	return &xrange{key: ctx.Args[0], start: ctx.Args[1], end: ctx.Args[2]}, nil
}

func (cmd *xrange) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: streams.Stream{}}
	}

	stream, ok := o.Value.(streams.Stream)
	if !ok {
		return nil, errors.WrongType
	}

	entries, err := stream.Between(cmd.start, cmd.end)
	if err != nil {
		return err, nil // TODO: revisit and do proper error handling here
	}

	ret := []any{}
	for _, e := range entries {
		ret = append(ret, e.Format())
	}

	return resp.Array(ret), nil
}
