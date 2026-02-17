package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type Command interface {
	Exec(ctx *request.Context, s *store.Store) ([]byte, error)
}
