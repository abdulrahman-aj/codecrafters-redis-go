package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type replConf struct{}

func parseReplConf(command string, args []string) (*replConf, []byte) {
	return &replConf{}, nil
}

func (cmd *replConf) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	return resp.SimpleString("OK"), nil
}
