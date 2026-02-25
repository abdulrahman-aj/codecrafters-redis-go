package engine

import (
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/containers"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/commands"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
	"github.com/codecrafters-io/redis-starter-go/app/util"
)

type Engine struct {
	inbox      chan *envelope
	store      *store.Store
	waitQueue  *waitQueue                                         // Requests that are blocked on some key
	readyQueue *containers.IndexedPriorityQueue[*envelope, int64] // Requests that are potentially unblocked
	lastID     atomic.Int64                                       // Auto-increment ID for requests
	Config     *types.Config
	replicas   map[int64]replica
}

type replica struct {
	replicationCh chan<- []byte
	done          <-chan struct{}
}

func New(isMaster bool) *Engine {
	e := &Engine{
		inbox:     make(chan *envelope),
		store:     store.New(),
		waitQueue: newWaitQueue(),
		readyQueue: containers.NewIndexedPriorityQueue(
			(*envelope).key,
			(*envelope).before,
		),
		Config: &types.Config{
			IsMaster:                isMaster,
			MasterReplicationID:     "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
			MasterReplicationOffset: 0,
		},
		replicas: map[int64]replica{},
	}

	go e.run()
	return e
}

type envelope struct {
	ctx        *types.RequestCtx
	command    string
	args       []string
	responseCh chan<- []byte
}

func (e *envelope) key() int64 { return e.ctx.RequestID }

func (e *envelope) before(other *envelope) bool {
	return e.ctx.RequestedAt.Before(other.ctx.RequestedAt)
}

func (e *Engine) run() {
	nextTimeout := time.NewTimer(0)

	// TODO: cleanup this loop. extract methods...
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

func (e *Engine) Do(ctx *types.ConnectionCtx, c any) []byte {
	command, args, ok := parseCommand(c)
	if !ok {
		return resp.SimpleError("ERR Protocol error: expected array of bulk strings")
	}

	ch := make(chan []byte)

	e.inbox <- &envelope{
		responseCh: ch,
		command:    command,
		args:       args,
		ctx: &types.RequestCtx{
			Conn:         ctx,
			RequestID:    e.lastID.Add(1),
			RequestedAt:  time.Now(),
			Dependencies: map[string]bool{},
			TouchedKeys:  map[string]bool{},
			ServerCfg:    e.Config,
		},
	}

	return <-ch
}

func (e *Engine) handle(msg *envelope) {
	util.Assert(!msg.ctx.Conn.IsReplicaConn, "should not send any commands while replicating")

	cmd, parseRespErr := commands.Parse(msg.ctx, msg.command, msg.args)
	if parseRespErr != nil {
		e.respond(msg, parseRespErr)
		return
	}

	res, err := cmd.Exec(msg.ctx, e.store)
	if err != nil {
		if errors.Is(err, commands.ErrBlocked) {
			e.waitQueue.enqueue(msg)
		} else {
			util.FatalOnErr(err)
		}
		return
	}

	e.respond(msg, res)
}

func (e *Engine) respond(msg *envelope, res []byte) {
	if msg.ctx.Conn.IsReplicaConn && !msg.ctx.Conn.IsReplicaRegistered {
		e.registerReplica(msg)
	}
	e.replicate(msg)
	e.wakeWaiters(msg.ctx)
	e.updateOffset(msg)

	if !msg.ctx.Conn.IsMasterConn || msg.ctx.MustRespond {
		msg.responseCh <- res
	} else {
		msg.responseCh <- nil
	}
}

func (e *Engine) registerReplica(msg *envelope) {
	msg.ctx.Conn.IsReplicaRegistered = true

	replicationCh := make(chan []byte, 1000)
	done := make(chan struct{})

	msg.ctx.Conn.ReplicationCh = replicationCh
	msg.ctx.Conn.ReplicationDone = done

	e.replicas[msg.ctx.Conn.ID] = replica{
		replicationCh: replicationCh,
		done:          done,
	}
}

func (e *Engine) replicate(msg *envelope) {
	if len(msg.ctx.TouchedKeys) == 0 { // no state change -> nothing to replicate
		return
	}

	var deadReplicas []int64

	req := resp.Array(append([]string{msg.command}, msg.args...))
	for id, r := range e.replicas {
		select {
		case r.replicationCh <- req:
		case <-r.done:
			deadReplicas = append(deadReplicas, id)
		}
	}

	for _, id := range deadReplicas {
		delete(e.replicas, id)
	}
}

func (e *Engine) updateOffset(msg *envelope) {
	// TOOD: can optimize and get num bytes from resp.Reader.ReadValue but life is too short...
	requestBytes := resp.Array(append([]string{msg.command}, msg.args...))
	e.Config.MasterReplicationOffset += len(requestBytes)
}

func (e *Engine) wakeWaiters(ctx *types.RequestCtx) {
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
