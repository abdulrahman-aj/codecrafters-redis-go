package server

import "github.com/codecrafters-io/redis-starter-go/app/resp"

type request struct {
	command    any
	responseCh chan<- []byte
}

type Server struct {
	requests chan *request
	storage  map[string]string
}

func New() *Server {
	return &Server{
		requests: make(chan *request),
		storage:  map[string]string{},
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

func (s *Server) handle(argsAny any) []byte {
	args, ok := argsAny.([]any)
	if !ok || len(args) == 0 {
		panic("TODO: Handle errors properly")
	}

	command, ok := args[0].(string)
	if !ok {
		panic("TODO: Handle errors properly")
	}

	args = args[1:]
	switch command {
	case "PING":
		return resp.SimpleString("PONG")
	case "ECHO":
		if len(args) != 1 {
			panic("TODO: Handle errors properly")
		}
		arg, ok := args[0].(string)
		if !ok {
			panic("TODO: Handle errors properly")
		}
		return resp.BulkString(arg)
	case "SET":
		if len(args) != 2 {
			panic("TODO: Handle errors properly")
		}

		key, ok := args[0].(string)
		if !ok {
			panic("TODO: Handle errors properly")
		}
		value, ok := args[1].(string)
		if !ok {
			panic("TODO: Handle errors properly")
		}

		s.storage[key] = value
		return resp.SimpleString("OK")
	case "GET":
		if len(args) != 1 {
			panic("TODO: Handle errors properly")
		}
		key, ok := args[0].(string)
		if !ok {
			panic("TODO: Handle errors properly")
		}
		val, ok := s.storage[key]
		if !ok {
			return resp.Null
		}
		return resp.BulkString(val)
	default:
		panic("TODO: Handle errors properly")
	}
}
