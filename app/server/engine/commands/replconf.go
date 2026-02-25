package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
)

type replConf struct {
	key   string
	value string
}

func parseReplConf(command string, args []string) (*replConf, []byte) {
	if len(args) != 2 {
		return nil, rediserrors.NumArgs(command)
	}

	return &replConf{key: args[0], value: args[1]}, nil
}

func (cmd *replConf) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	if strings.EqualFold(cmd.key, "getack") {
		ctx.MustRespond = true
		return resp.Array([]string{"REPLCONF", "ACK", "0"}), nil
	}

	return resp.SimpleString("OK"), nil
}
