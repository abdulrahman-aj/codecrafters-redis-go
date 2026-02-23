package commands

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/rediserrors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/util"
)

type exec struct{}

func parseExec(command string, args []string) (*exec, []byte) {
	if len(args) != 0 {
		return nil, rediserrors.NumArgs(command)
	}

	return &exec{}, nil
}

func (cmd *exec) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	if !ctx.Connection.InsideTx {
		return rediserrors.ExecWithoutMulti, nil
	}

	txCommands := ctx.Connection.TxCommands

	ctx.Connection.InsideTx = false
	ctx.Connection.TxCommands = nil

	var ret []any

	for _, txCommand := range txCommands {
		cmd, respErr := Parse(ctx, txCommand.Command, txCommand.Args)
		if respErr != nil {
			ret = append(ret, respErr)
			continue
		}

		res, err := cmd.Exec(ctx, s)
		if err != nil {
			util.Assert(
				!errors.Is(err, ErrBlocked),
				"commands should never return ErrBlocked inside a transaction",
			)
			return nil, err
		}

		ret = append(ret, res)
	}

	return resp.Array(ret), nil
}
