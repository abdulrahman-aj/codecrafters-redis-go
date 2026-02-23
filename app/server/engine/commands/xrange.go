package commands

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/streams"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type xrange struct {
	key   string
	start string
	end   string
}

func parseXrange(command string, args []string) (*xrange, []byte) {
	if len(args) != 3 {
		return nil, rediserrors.NumArgs(command)
	}

	return &xrange{key: args[0], start: args[1], end: args[2]}, nil
}

func (cmd *xrange) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	o, ok := s.Get(cmd.key)
	if !ok {
		o = store.Object{Value: streams.Stream{}}
	}

	stream, ok := o.Value.(streams.Stream)
	if !ok {
		return rediserrors.WrongType, nil
	}

	entries, err := stream.Between(cmd.start, cmd.end)
	if err != nil {
		if errors.Is(err, streams.ErrInvalidID) {
			return rediserrors.InvalidStreamID, nil
		}
		return nil, err
	}

	ret := []any{}
	for _, e := range entries {
		ret = append(ret, e.Format())
	}

	return resp.Array(ret), nil
}
