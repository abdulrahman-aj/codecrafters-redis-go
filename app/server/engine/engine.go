package engine

import (
	"errors"
	"fmt"
	"net"
	"strconv"
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
	info       *types.Info
}

func New(port int, replicaOf string) *Engine {
	info := &types.Info{
		Port:                port,
		Role:                "master",
		MasterReplicationID: "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
	}

	if replicaOf != "" {
		info.Role = "slave"

		parts := strings.Split(replicaOf, " ")
		util.Assert(len(parts) == 2, `replica address in the format "<MASTER_HOST> <MASTER_PORT>"`)

		info.MasterIP = parts[0]
		info.MasterPort = parts[1]
	}

	return &Engine{
		inbox:     make(chan *envelope),
		store:     store.New(),
		waitQueue: newWaitQueue(),
		readyQueue: containers.NewIndexedPriorityQueue(
			(*envelope).key,
			(*envelope).before,
		),
		info: info,
	}
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

func (e *Engine) Run() {
	if e.info.MasterIP != "" {
		util.FatalOnErr(e.connectToMaster()) // TODO: handle connect to master error
	}

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
			Info:         e.info,
		},
	}

	return <-ch
}

func (e *Engine) handle(msg *envelope) {
	cmd, parseRespErr := commands.Parse(msg.ctx, msg.command, msg.args)
	if parseRespErr != nil {
		msg.responseCh <- parseRespErr
		e.wakeWaiters(msg.ctx)
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

	msg.responseCh <- res
	e.wakeWaiters(msg.ctx)
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

func (e *Engine) connectToMaster() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", e.info.MasterIP, e.info.MasterPort))
	if err != nil {
		return err
	}

	reader := resp.NewReader(conn)
	send := func(bytes []byte) error { // ignoring responses for now
		if _, err := conn.Write(bytes); err != nil {
			return err
		}

		_, err := reader.ReadValue()
		return err
	}

	if err := send(resp.Array([]string{"PING"})); err != nil {
		return err
	}

	if err := send(resp.Array([]string{"REPLCONF", "listening-port", strconv.Itoa(e.info.Port)})); err != nil {
		return err
	}

	if err := send(resp.Array([]string{"REPLCONF", "capa", "psync2"})); err != nil {
		return err
	}

	return nil
}
