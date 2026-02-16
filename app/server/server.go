package server

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/containers"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Server struct {
	inbox chan *envelope
	store *store

	// Contains requests that are blocked on some key
	waitQueue *waitQueue

	// Contains requests that are potentially unblocked
	readyQueue *containers.IndexedPriorityQueue[*envelope, int64] // TODO: memory leak here. need to shrink to fit

	// auto-increment ID for requests
	lastID atomic.Int64
}

type envelope struct {
	req        *request
	responseCh chan<- []byte
}

func (e *envelope) key() int64 { return e.req.id }

func (e *envelope) before(other *envelope) bool {
	return e.req.requestedAt.Before(other.req.requestedAt)
}

type request struct {
	id          int64
	command     string
	args        []string
	requestedAt time.Time
	deadline    time.Time
	dependency  string
	touchedKeys []string
}

func (r *request) isExpired() bool {
	return !r.deadline.IsZero() && time.Now().After(r.deadline)
}

func New() *Server {
	return &Server{
		inbox:     make(chan *envelope),
		store:     newStore(),
		waitQueue: newWaitQueue(),
		readyQueue: containers.NewIndexedPriorityQueue(
			(*envelope).key,
			(*envelope).before,
		),
	}
}

func (s *Server) Run() {
	nextTimeout := time.NewTimer(0)

	for {
		select {
		case msg := <-s.inbox:
			s.handle(msg)
		case <-nextTimeout.C:
			for _, msg := range s.waitQueue.dequeueExpired() {
				s.readyQueue.Enqueue(msg)
			}
		}

		for s.readyQueue.Len() > 0 {
			req, _ := s.readyQueue.Dequeue()
			s.handle(req)
		}

		if t, ok := s.waitQueue.timeUntilNext(); ok {
			nextTimeout.Reset(t)
		}
	}
}

func (s *Server) Do(c any) []byte {
	command, args, ok := parseCommand(c)
	if !ok {
		return resp.SimpleError("command should be an array of bulk strings.")
	}

	ch := make(chan []byte)

	s.inbox <- &envelope{
		responseCh: ch,
		req: &request{
			id:          s.lastID.Add(1),
			command:     command,
			args:        args,
			requestedAt: time.Now(),
		},
	}

	return <-ch
}

func (s *Server) handle(msg *envelope) {
	var (
		res  []byte
		done = true
	)

	switch msg.req.command {
	case "ping":
		res = s.handlePing(msg.req)
	case "echo":
		res = s.handleEcho(msg.req)
	case "set":
		res = s.handleSet(msg.req)
	case "get":
		res = s.handleGet(msg.req)
	case "rpush":
		res = s.handleRpush(msg.req)
	case "lpush":
		res = s.handleLpush(msg.req)
	case "lrange":
		res = s.handleLrange(msg.req)
	case "llen":
		res = s.handleLlen(msg.req)
	case "lpop":
		res = s.handleLpop(msg.req)
	case "blpop":
		res, done = s.handleBlpop(msg.req)
	case "type":
		res = s.handleType(msg.req)
	case "xadd":
		res = s.handleXadd(msg.req)
	case "xrange":
		res = s.handleXrange(msg.req)
	case "xread":
		res = s.handleXread(msg.req)
	default:
		res = errUnknownCommand(msg.req.command)
	}

	if !done {
		s.waitQueue.enqueue(msg)
	} else {
		msg.responseCh <- res
		s.wakeWaiters(msg.req.touchedKeys)
	}
}

func (s *Server) wakeWaiters(touchedKeys []string) {
	for _, key := range touchedKeys {
		for _, msg := range s.waitQueue.dequeueWaiters(key) {
			s.readyQueue.Enqueue(msg)
		}
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
