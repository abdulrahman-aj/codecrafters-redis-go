package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/streams"
)

type xadd struct {
	key     string
	entryID string
	kvs     []string
}

func parseXadd(command string, args []string) (*xadd, []byte) {
	if len(args) < 4 {
		return nil, rediserrors.NumArgs(command)
	}

	var (
		key     = args[0]
		entryID = args[1]
		kvs     = args[2:]
	)

	if len(kvs)%2 != 0 {
		return nil, rediserrors.NumArgs(command)
	}

	return &xadd{key: key, entryID: entryID, kvs: kvs}, nil
}

func (cmd *xadd) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: streams.Stream{}}
	}

	stream, ok := o.Value.(streams.Stream)
	if !ok {
		return rediserrors.WrongType, nil
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
		switch {
		case errors.Is(err, streams.ErrInvalidID):
			return rediserrors.InvalidStreamID, nil
		case errors.Is(err, streams.ErrNotIncreasing):
			return rediserrors.XaddEqualOrSmaller, nil
		case errors.Is(err, streams.ErrZeroID):
			return rediserrors.XaddZeroID, nil
		default:
			return nil, err
		}
	}

	o.Value = stream
	s.Set(ctx, cmd.key, o)

	return resp.BulkString(id), nil
}
