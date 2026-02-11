package server

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Server struct {
	requests chan *request
	storage  map[string]entry
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
	command, args, ok := castCommand(commandAny)
	if !ok {
		panic("TODO: Handle errors properly")
	}

	switch command {
	case "PING":
		return resp.SimpleString("PONG")
	case "ECHO":
		if len(args) != 1 {
			panic("TODO: Handle errors properly")
		}
		return resp.BulkString(args[0])
	case "SET":
		if len(args) != 2 && len(args) != 4 {
			panic("TODO: Handle errors properly")
		}

		var expiresAt time.Time

		if len(args) == 4 {
			ttl, err := strconv.Atoi(args[3])
			if err != nil {
				panic("TODO: Handle errors properly")
			}

			switch args[2] {
			case "PX":
				expiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
			case "EX":
				expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
			default:
				panic("TODO: Handle errors properly")
			}
		}

		s.storage[args[0]] = entry{value: args[1], expiresAt: expiresAt}
		return resp.SimpleString("OK")
	case "GET":
		if len(args) != 1 {
			panic("TODO: Handle errors properly")
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
		panic("TODO: Handle errors properly")
	}
}

// Clients send commands to a Redis server as an array of bulk strings.
func castCommand(v any) (string, []string, bool) {
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

	return ret[0], ret[1:], true
}
