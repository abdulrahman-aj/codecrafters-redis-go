package server

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Server struct {
	requests chan *request
	storage  map[string]entry // TODO: need to empty this map periodically as it will cause a memory leak
}

type entry struct {
	value     string
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
		if len(args) != 0 {
			return errNumArgs(command)
		}
		return resp.SimpleString("PONG")
	case "echo":
		if len(args) != 1 {
			return errNumArgs(command)
		}
		return resp.BulkString(args[0])
	case "set":
		if len(args) != 2 && len(args) != 4 {
			return errNumArgs(command)
		}

		var expiresAt time.Time

		if len(args) == 4 {
			ttl, err := strconv.Atoi(args[3])
			if err != nil {
				return errInvalidInteger
			}

			switch args[2] {
			case "PX":
				expiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
			case "EX":
				expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
			default:
				return errSyntaxError
			}
		}

		s.storage[args[0]] = entry{value: args[1], expiresAt: expiresAt}
		return resp.SimpleString("OK")
	case "get":
		if len(args) != 1 {
			return errNumArgs(command)
		}
		entry, ok := s.storage[args[0]]
		if !ok {
			return resp.NullBulkString
		}
		if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
			delete(s.storage, args[0])
			return resp.NullBulkString
		}
		return resp.BulkString(entry.value)
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

func errNumArgs(command string) []byte {
	msg := fmt.Sprintf("ERR wrong number of arguments for '%s' command", command)
	return resp.SimpleError(msg)
}

func errUnknownCommand(command string) []byte {
	msg := fmt.Sprintf("ERR unknown command '%s'", command)
	return resp.SimpleError(msg)
}

var (
	errSyntaxError    = resp.SimpleError("ERR syntax error")
	errInvalidInteger = resp.SimpleError("ERR value is not an integer or out of range")
)
