package request

import "time"

type Context struct {
	RequestID    int64
	Command      string
	Args         []string
	RequestedAt  time.Time
	Deadline     time.Time
	Dependencies map[string]bool
	TouchedKeys  map[string]bool
}

func (ctx *Context) SetTimeout(d time.Duration) {
	deadline := ctx.RequestedAt.Add(d)
	if ctx.Deadline.IsZero() || deadline.Before(ctx.Deadline) {
		ctx.Deadline = deadline
	}
}

func (ctx *Context) IsExpired() bool { return !ctx.Deadline.IsZero() && time.Now().After(ctx.Deadline) }

func (ctx *Context) HasDeadline() bool { return !ctx.Deadline.IsZero() }
