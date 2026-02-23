package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type exec struct{}

func parseExec(command string, args []string) (*exec, error) {
	if len(args) != 0 {
		return nil, errors.NumArgs(command)
	}

	return &exec{}, nil
}

func (cmd *exec) Exec(ctx *context.Request, s *store.Store) ([]byte, error) {
	if !ctx.Connection.InsideTx {
		return nil, errors.ExecWithoutMulti
	}

	txCommands := ctx.Connection.TxCommands

	ctx.Connection.InsideTx = false
	ctx.Connection.TxCommands = nil

	var ret []any

	for _, txCommand := range txCommands {
		cmd, err := Parse(ctx, txCommand.Command, txCommand.Args)

		if err != nil { // TODO: error handling is getting a bit messy. restructure...
			if clientError, ok := err.(*errors.ClientError); ok {
				ret = append(ret, clientError.SerializeToResp())
				continue
			}
			return nil, err
		}

		res, err := cmd.Exec(ctx, s)
		if err != nil {
			if clientError, ok := err.(*errors.ClientError); ok {
				ret = append(ret, clientError.SerializeToResp())
			} else if errors.Is(err, errors.Blocked) {
				return nil, errors.New("commands should never return errors.Blocked inside a transaction")
			} else {
				return nil, err
			}
		} else {
			ret = append(ret, res)
		}
	}

	return resp.Array(ret), nil
}
