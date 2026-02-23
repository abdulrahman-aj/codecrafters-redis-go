package engine

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/containers"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/commands"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/errors"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
)

type Engine struct {
	inbox      chan *envelope
	store      *store.Store
	waitQueue  *waitQueue                                         // Requests that are blocked on some key
	readyQueue *containers.IndexedPriorityQueue[*envelope, int64] // Requests that are potentially unblocked
	lastID     atomic.Int64                                       // Auto-increment ID for requests
}

func New() *Engine {
	return &Engine{
		inbox:     make(chan *envelope),
		store:     store.New(),
		waitQueue: newWaitQueue(),
		readyQueue: containers.NewIndexedPriorityQueue(
			(*envelope).key,
			(*envelope).before,
		),
	}
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

func (e *Engine) Run() {
	nextTimeout := time.NewTimer(0)

	for {
		select {
		case msg := <-e.inbox:
			e.handle(msg)
		case <-nextTimeout.C:
			for _, msg := range e.waitQueue.dequeueExpired() {
				e.readyQueue.Enqueue(msg)
			}
		}

		for e.readyQueue.Len() > 0 {
			req, _ := e.readyQueue.Dequeue()
			e.handle(req)
		}

		if t, ok := e.waitQueue.timeUntilNext(); ok {
			nextTimeout.Reset(t)
		}
	}
}

func (e *Engine) Do(connectionCtx *context.Connection, c any) []byte {
	command, args, ok := parseCommand(c)
	if !ok {
		return resp.SimpleError("ERR Protocol error: expected array of bulk strings")
	}

	ch := make(chan []byte)

	e.inbox <- &envelope{
		responseCh: ch,
		command:    command,
		args:       args,
		ctx: &context.Request{
			Connection:   connectionCtx,
			RequestID:    e.lastID.Add(1),
			RequestedAt:  time.Now(),
			Dependencies: map[string]bool{},
			TouchedKeys:  map[string]bool{},
		},
	}

	return <-ch
}

func (e *Engine) handle(msg *envelope) {
	cmd, err := commands.Parse(msg.ctx, msg.command, msg.args)
	if err != nil {
		e.handleErr(msg, err)
		return
	}

	res, err := cmd.Exec(msg.ctx, e.store)
	if err != nil {
		e.handleErr(msg, err)
		return
	}

	msg.responseCh <- res
	e.wakeWaiters(msg.ctx)
}

func (e *Engine) handleErr(msg *envelope, err error) {
	if clientError, ok := err.(*errors.ClientError); ok {
		msg.responseCh <- clientError.SerializeToResp()
		e.wakeWaiters(msg.ctx) // TODO: unneeded?
	} else if errors.Is(err, errors.Blocked) {
		e.waitQueue.enqueue(msg)
	} else {
		panic(err)
	}
}

func (e *Engine) wakeWaiters(ctx *context.Request) {
	for key := range ctx.TouchedKeys {
		for _, msg := range e.waitQueue.dequeueWaiters(key) {
			e.readyQueue.Enqueue(msg)
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
