package server

import (
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Server struct {
	requests chan *request
	storage  map[string]entry // TODO: need to empty this map periodically as it will cause a memory leak
}

type entry struct {
	value     any
	expiresAt time.Time // TODO: implement background GC in the future
}

type request struct {
	command    any
	responseCh chan<- []byte
}

func New() *Server {
	return &Server{
		requests: make(chan *request),
		storage:  map[string]entry{},
	}
}

func (s *Server) Run() {
	for req := range s.requests {
		req.responseCh <- s.handle(req.command)
	}
}

func (s *Server) Do(command any) []byte {
	ch := make(chan []byte)
	req := &request{
		command:    command,
		responseCh: ch,
	}

	s.requests <- req
	return <-ch
}

func (s *Server) handle(commandAny any) []byte {
	command, args, ok := parseCommand(commandAny)
	if !ok {
		return resp.SimpleError("command should be an array of bulk strings.")
	}

	switch command {
	case "ping":
		return s.handlePing(command, args)
	case "echo":
		return s.handleEcho(command, args)
	case "set":
		return s.handleSet(command, args)
	case "get":
		return s.handleGet(command, args)
	case "rpush":
		return s.handleRpush(command, args)
	case "lpush":
		return s.handleLpush(command, args)
	case "lrange":
		return s.handleLrange(command, args)
	case "llen":
		return s.handleLlen(command, args)
	case "lpop":
		return s.handleLpop(command, args)
	default:
		return errUnknownCommand(command)
	}
}

func parseCommand(v any) (string, []string, bool) {
	args, ok := v.([]any)
	if !ok {
		return "", nil, false
	}

	var ret []string
	for _, arg := range args {
		argStr, ok := arg.(string)
		if !ok {
			return "", nil, false
		}
		ret = append(ret, argStr)
	}

	if len(ret) == 0 {
		return "", nil, false
	}

	return strings.ToLower(ret[0]), ret[1:], true
}
