package server

import "github.com/codecrafters-io/redis-starter-go/app/resp"

type request struct {
	command    any
	responseCh chan<- []byte
}

type Server struct {
	requests chan *request
}

func New() *Server {
	return &Server{requests: make(chan *request)}
}

func (s *Server) Run() {
	for req := range s.requests {
		req.responseCh <- handle(req.command)
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

func handle(argsAny any) []byte {
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
		return []byte("+PONG\r\n")
	case "ECHO":
		if len(args) != 1 {
			panic("TODO: Handle errors properly")
		}
		arg, ok := args[0].(string)
		if !ok {
			panic("TODO: Handle errors properly")
		}
		return resp.SerializeBulkString(arg)
	default:
		panic("TODO: Handle errors properly")
	}
}
