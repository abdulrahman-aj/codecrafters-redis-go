package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/streams"
)

type xadd struct {
	key     string
	entryID string
	kvs     []string
}

func ParseXadd(ctx *request.Context) (*xadd, error) {
	if len(ctx.Args) < 4 {
		return nil, errors.NumArgs(ctx)
	}

	var (
		key     = ctx.Args[0]
		entryID = ctx.Args[1]
		kvs     = ctx.Args[2:]
	)

	if len(kvs)%2 != 0 {
		return nil, errors.NumArgs(ctx)
	}

	return &xadd{key: key, entryID: entryID, kvs: kvs}, nil
}

func (cmd *xadd) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: streams.Stream{}}
	}

	stream, ok := o.Value.(streams.Stream)
	if !ok {
		return nil, errors.WrongType
	}

	var fields []streams.Field
	for i := 0; i < len(cmd.kvs); i += 2 {
		fields = append(fields, streams.Field{
			Key:   cmd.kvs[i],
			Value: cmd.kvs[i+1],
		})
	}

	id, err := stream.Append(cmd.entryID, fields, ctx.RequestedAt)
	if err != nil {
		return err, nil // TODO: revisit this and do proper error mapping
	}

	o.Value = stream
	s.Set(ctx, cmd.key, o)

	return resp.BulkString(id), nil
}
