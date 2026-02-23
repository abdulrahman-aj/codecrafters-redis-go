package types

import "time"

type RequestCtx struct {
	Conn         *ConnectionCtx
	RequestID    int64
	RequestedAt  time.Time
	Deadline     time.Time
	Dependencies map[string]bool
	TouchedKeys  map[string]bool
	Info         *Info
}

func (ctx *RequestCtx) SetTimeout(d time.Duration) {
	deadline := ctx.RequestedAt.Add(d)
	if ctx.Deadline.IsZero() || deadline.Before(ctx.Deadline) {
		ctx.Deadline = deadline
	}
}

func (ctx *RequestCtx) DeadlineExceeded() bool {
	if ctx.Conn.InsideTx {
		return true
	}

	return !ctx.Deadline.IsZero() && time.Now().After(ctx.Deadline)
}

func (ctx *RequestCtx) HasDeadline() bool { return !ctx.Deadline.IsZero() }

type ConnectionCtx struct {
	InsideTx   bool
	TxCommands []TxCommand
}

type TxCommand struct {
	Command string
	Args    []string
}

type Info struct {
	Port                    int
	Role                    string
	MasterReplicationID     string
	MasterReplicationOffset int
	MasterIP                string
	MasterPort              string
}
