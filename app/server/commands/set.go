package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type set struct {
	key       string
	value     string
	expiresAt time.Time
}

func ParseSet(ctx *request.Context) (*set, error) {
	if len(ctx.Args) != 2 && len(ctx.Args) != 4 {
		return nil, errors.NumArgs(ctx)
	}

	var expiresAt time.Time

	if len(ctx.Args) == 4 {
		ttl, err := strconv.Atoi(ctx.Args[3])
		if err != nil {
			return nil, errors.InvalidInteger
		}

		switch strings.ToLower(ctx.Args[2]) {
		case "px":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
		case "ex":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
		default:
			return nil, errors.SyntaxError
		}
	}

	return &set{key: ctx.Args[0], value: ctx.Args[1], expiresAt: expiresAt}, nil
}

func (cmd *set) Exec(ctx *request.Context, s *store.Store) ([]byte, error) {
	o := store.Object{Value: cmd.value, ExpiresAt: cmd.expiresAt}
	s.Set(ctx, cmd.key, o)
	return resp.SimpleString("OK"), nil
}
