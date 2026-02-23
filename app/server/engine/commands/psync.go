package commands

import (
	"encoding/base64"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
	"github.com/codecrafters-io/redis-starter-go/app/util"
)

var emptyRDB = func() []byte {
	rdbBase64 := `UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog==`
	res, err := base64.StdEncoding.DecodeString(rdbBase64)
	util.FatalOnErr(err)
	return res
}()

type psync struct{}

func parsePsync(command string, args []string) (*psync, []byte) {
	return &psync{}, nil
}

func (cmd *psync) Exec(ctx *types.RequestCtx, s *store.Store) ([]byte, error) {
	msg := resp.SimpleString(fmt.Sprintf("FULLRESYNC %s %d", ctx.Info.MasterReplicationID, ctx.Info.MasterReplicationOffset))
	msg = fmt.Appendf(msg, "$%d\r\n%s", len(emptyRDB), emptyRDB)
	return msg, nil
}
