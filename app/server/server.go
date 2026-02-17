package server

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/containers"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/commands"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/request"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type Server struct {
	inbox chan *envelope
	store *store.Store

	// Contains requests that are blocked on some key
	waitQueue *waitQueue

	// Contains requests that are potentially unblocked
	readyQueue *containers.IndexedPriorityQueue[*envelope, int64] // TODO: memory leak here. need to shrink to fit

	// auto-increment ID for requests
	lastID atomic.Int64
}

type envelope struct {
	ctx        *request.Context
	responseCh chan<- []byte
}

func (e *envelope) key() int64 { return e.ctx.RequestID }

func (e *envelope) before(other *envelope) bool {
	return e.ctx.RequestedAt.Before(other.ctx.RequestedAt)
}

func New() *Server {
	return &Server{
		inbox:     make(chan *envelope),
		store:     store.New(),
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
		ctx: &request.Context{
			RequestID:    s.lastID.Add(1),
			Command:      command,
			Args:         args,
			RequestedAt:  time.Now(),
			Dependencies: map[string]bool{},
			TouchedKeys:  map[string]bool{},
		},
	}

	return <-ch
}

func (s *Server) handle(msg *envelope) {
	var (
		cmd commands.Command
		err error
	)

	switch msg.ctx.Command {
	case "ping":
		cmd, err = commands.ParsePing(msg.ctx)
	case "echo":
		cmd, err = commands.ParseEcho(msg.ctx)
	case "set":
		cmd, err = commands.ParseSet(msg.ctx)
	case "get":
		cmd, err = commands.ParseGet(msg.ctx)
	case "rpush":
		cmd, err = commands.ParseRpush(msg.ctx)
	case "lpush":
		cmd, err = commands.ParseLpush(msg.ctx)
	case "lrange":
		cmd, err = commands.ParseLrange(msg.ctx)
	case "llen":
		cmd, err = commands.ParseLlen(msg.ctx)
	case "lpop":
		cmd, err = commands.ParseLpop(msg.ctx)
	case "blpop":
		cmd, err = commands.ParseBlpop(msg.ctx)
	case "type":
		cmd, err = commands.ParseType(msg.ctx)
	case "xadd":
		cmd, err = commands.ParseXadd(msg.ctx)
	case "xrange":
		cmd, err = commands.ParseXrange(msg.ctx)
	case "xread":
		cmd, err = commands.ParseXread(msg.ctx)
	default:
		err = errors.UknownCommand(msg.ctx)
	}

	// TODO: consider adding --verbose logging
	if err != nil {
		s.handleErr(msg, err)
		return
	}

	res, err := cmd.Exec(msg.ctx, s.store)
	if err != nil {
		s.handleErr(msg, err)
		return
	}

	msg.responseCh <- res
	s.wakeWaiters(msg.ctx)
}

func (s *Server) handleErr(msg *envelope, err error) {
	if clientError, ok := err.(*errors.ClientError); ok {
		msg.responseCh <- clientError.SerializeToResp()
		s.wakeWaiters(msg.ctx) // TODO: unneeded?
	} else if errors.Is(err, errors.Blocked) {
		s.waitQueue.enqueue(msg)
	} else {
		panic(clientError)
	}
}

func (s *Server) wakeWaiters(ctx *request.Context) {
	for key := range ctx.TouchedKeys {
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
