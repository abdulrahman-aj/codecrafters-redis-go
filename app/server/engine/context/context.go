package context

import "time"

type Request struct {
	Connection   *Connection
	RequestID    int64
	RequestedAt  time.Time
	Deadline     time.Time
	Dependencies map[string]bool
	TouchedKeys  map[string]bool
}

func (ctx *Request) SetTimeout(d time.Duration) {
	deadline := ctx.RequestedAt.Add(d)
	if ctx.Deadline.IsZero() || deadline.Before(ctx.Deadline) {
		ctx.Deadline = deadline
	}
}

func (ctx *Request) DeadlineExceeded() bool {
	if ctx.Connection.InsideTx {
		return true
	}

	return !ctx.Deadline.IsZero() && time.Now().After(ctx.Deadline)
}

func (ctx *Request) HasDeadline() bool { return !ctx.Deadline.IsZero() }

type Connection struct {
	InsideTx   bool
	TxCommands []TxCommand
}

type TxCommand struct {
	Command string
	Args    []string
}
