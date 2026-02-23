package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type psync struct{}

func parsePsync(command string, args []string) (*psync, []byte) {
	return &psync{}, nil
}

func (cmd *psync) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	msg := fmt.Sprintf("FULLRESYNC %s %d", ctx.Info.MasterReplicationID, ctx.Info.MasterReplicationOffset)
	return resp.SimpleString(msg), nil
}
