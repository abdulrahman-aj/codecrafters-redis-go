package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type set struct {
	key       string
	value     string
	expiresAt time.Time
}

func parseSet(command string, args []string) (*set, []byte) {
	if len(args) != 2 && len(args) != 4 {
		return nil, rediserrors.NumArgs(command)
	}

	var expiresAt time.Time

	if len(args) == 4 {
		ttl, err := strconv.Atoi(args[3])
		if err != nil {
			return nil, rediserrors.InvalidInteger
		}

		switch strings.ToLower(args[2]) {
		case "px":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
		case "ex":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
		default:
			return nil, rediserrors.SyntaxError
		}
	}

	return &set{key: args[0], value: args[1], expiresAt: expiresAt}, nil
}

func (cmd *set) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	o := store.Object{Value: cmd.value, ExpiresAt: cmd.expiresAt}
	s.Set(ctx, cmd.key, o)
	return resp.SimpleString("OK"), nil
}
