package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type ping struct{}

func parsePing(command string, args []string) (*ping, []byte) {
	if len(args) != 0 {
		return nil, rediserrors.NumArgs(command)
	}

	return &ping{}, nil
}

func (cmd *ping) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	return resp.SimpleString("PONG"), nil
}
