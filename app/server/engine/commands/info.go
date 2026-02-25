package commands

import (
	"bytes"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
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

func (cmd *info) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	kvs := []struct{ k, v string }{
		{"role", "master"},
		{"master_replid", ctx.ServerCfg.MasterReplicationID},
		{"master_repl_offset", strconv.Itoa(ctx.ServerCfg.MasterReplicationOffset)},
	}

	if !ctx.ServerCfg.IsMaster {
		kvs[0].v = "slave"
	}

	var buf bytes.Buffer
	buf.WriteString("# Replication\n")
	for _, kv := range kvs {
		buf.WriteString(kv.k)
		buf.WriteString(":")
		buf.WriteString(kv.v)
		buf.WriteString("\n")
	}

	return resp.BulkString(buf.String()), nil
}
