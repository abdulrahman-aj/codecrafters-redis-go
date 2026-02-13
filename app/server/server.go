package server

import (
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/containers"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Server struct {
	inbox     chan *request
	storage   map[string]entry                    // TODO: need to empty this map periodically as it will cause a memory leak
	blockedBy map[string][]*request               // TODO: memory leak here. need to shrink to fit
	unblocked *containers.PriorityQueue[*request] // TODO: memory leak here. need to shrink to fit
}

type entry struct {
	value     any
	expiresAt time.Time // TODO: implement background GC in the future
}

type request struct {
	command     string
	args        []string
	responseCh  chan<- []byte
	requestedAt time.Time
}

type response struct {
	bytes       []byte
	waitingOn   string
	touchedKeys []string
}

func New() *Server {
	return &Server{
		inbox:     make(chan *request),
		storage:   map[string]entry{},
		blockedBy: map[string][]*request{},
		unblocked: containers.NewPriorityQueue(func(a, b *request) bool {
			return a.requestedAt.Before(b.requestedAt)
		}),
	}
}

func (s *Server) Run() {
	for req := range s.inbox {
		s.handle(req)

		for s.unblocked.Len() != 0 {
			req, _ := s.unblocked.Pop()
			s.handle(req)
		}
	}
}

func (s *Server) Do(c any) []byte {
	command, args, ok := parseCommand(c)
	if !ok {
		return resp.SimpleError("command should be an array of bulk strings.")
	}

	ch := make(chan []byte)

	s.inbox <- &request{
		command:     command,
		args:        args,
		responseCh:  ch,
		requestedAt: time.Now(),
	}

	return <-ch
}

func (s *Server) handle(req *request) {
	var res response

	switch req.command {
	case "ping":
		res = s.handlePing(req.command, req.args)
	case "echo":
		res = s.handleEcho(req.command, req.args)
	case "set":
		res = s.handleSet(req.command, req.args)
	case "get":
		res = s.handleGet(req.command, req.args)
	case "rpush":
		res = s.handleRpush(req.command, req.args)
	case "lpush":
		res = s.handleLpush(req.command, req.args)
	case "lrange":
		res = s.handleLrange(req.command, req.args)
	case "llen":
		res = s.handleLlen(req.command, req.args)
	case "lpop":
		res = s.handleLpop(req.command, req.args)
	case "blpop":
		res = s.handleBlpop(req.command, req.args)
	default:
		res = response{bytes: errUnknownCommand(req.command)}
	}

	if res.waitingOn != "" {
		s.blockedBy[res.waitingOn] = append(s.blockedBy[res.waitingOn], req)
	} else {
		req.responseCh <- res.bytes
		s.unblock(res.touchedKeys)
	}
}

func (s *Server) unblock(touchedKeys []string) {
	for _, key := range touchedKeys {
		for _, req := range s.blockedBy[key] {
			s.unblocked.Push(req)
		}
		delete(s.blockedBy, key)
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
