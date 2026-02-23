package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type info struct {
	section string
}

func parseInfo(command string, args []string) (*info, []byte) {
	if len(args) != 1 {
		return nil, rediserrors.NumArgs(command)
	}

	return &info{section: args[0]}, nil
}

func (cmd *info) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	return resp.BulkString("# Replication\nrole:master\n"), nil
}
