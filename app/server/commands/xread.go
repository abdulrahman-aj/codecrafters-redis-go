package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/store/streams"
)

type xread struct {
	keys       []string
	ids        []string
	isBlocking bool
}

func ParseXread(ctx *request.Context) (*xread, error) {
	args := ctx.Args

	if len(args) < 3 {
		return nil, errors.NumArgs(ctx)
	}

	cmd := &xread{isBlocking: strings.EqualFold(args[0], "block")}

	if cmd.isBlocking {
		timeoutMs, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, errors.TimeoutNotInt
		}

		if timeoutMs < 0 {
			return nil, errors.TimeoutNegative
		}

		if timeoutMs != 0 {
			ctx.SetTimeout(time.Duration(timeoutMs) * time.Millisecond)
		}

		args = args[2:]
	}

	if len(args) < 3 { // STREAMS k id
		return nil, errors.NumArgs(ctx)
	}

	if !strings.EqualFold(args[0], "streams") {
		return nil, errors.SyntaxError
	}

	keysAndIDs := args[1:]
	if len(keysAndIDs)%2 != 0 {
		return nil, errors.UnbalancedXread
	}

	numKeys := len(keysAndIDs) / 2

	cmd.keys = keysAndIDs[:numKeys]
	cmd.ids = keysAndIDs[numKeys:]

	for _, key := range cmd.keys {
		ctx.Dependencies[key] = true
	}

	return cmd, nil
}

func (cmd *xread) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	if ctx.IsExpired() {
		return resp.NullArray, nil
	}

	var ret []any
	for i, key := range cmd.keys {
		id := cmd.ids[i]

		o, ok := s.Get(key)
		if !ok {
			continue
		}

		stream, ok := o.Value.(streams.Stream)
		if !ok {
			return nil, errors.WrongType
		}

		var (
			entries []streams.Entry
			err     []byte
		)

		if id == "$" {
			entries = stream.AfterTime(ctx.RequestedAt)
		} else {
			entries, err = stream.After(id)
		}

		if err != nil {
			return err, nil // TODO: revisit and do proper error handling here
		}

		if len(entries) != 0 {
			var formatted []any
			for _, e := range entries {
				formatted = append(formatted, e.Format())
			}
			ret = append(ret, []any{key, formatted})
		}
	}

	if cmd.isBlocking && len(ret) == 0 {
		return nil, errors.Blocked
	}

	return resp.Array(ret), nil
}
