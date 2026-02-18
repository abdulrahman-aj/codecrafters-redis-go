package server

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/containers"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/commands"
	"github.com/codecrafters-io/redis-starter-go/app/server/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/errors"
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
	ctx        *context.Request
	command    string
	args       []string
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

func (s *Server) Do(connectionCtx *context.Connection, c any) []byte {
	command, args, ok := parseCommand(c)
	if !ok {
		return resp.SimpleError("command should be an array of bulk strings.") // TODO: check the proper redis error returned here
	}

	ch := make(chan []byte)

	s.inbox <- &envelope{
		responseCh: ch,
		command:    command,
		args:       args,
		ctx: &context.Request{
			Connection:   connectionCtx,
			RequestID:    s.lastID.Add(1),
			RequestedAt:  time.Now(),
			Dependencies: map[string]bool{},
			TouchedKeys:  map[string]bool{},
		},
	}

	return <-ch
}

// TODO: consider adding --verbose logging
func (s *Server) handle(msg *envelope) {
	cmd, err := commands.Parse(msg.ctx, msg.command, msg.args)
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
		panic(err)
	}
}

func (s *Server) wakeWaiters(ctx *context.Request) {
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
