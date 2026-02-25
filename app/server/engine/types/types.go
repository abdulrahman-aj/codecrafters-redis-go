package types

import "time"

type RequestCtx struct {
	Conn         *ConnectionCtx
	RequestID    int64
	RequestedAt  time.Time
	Deadline     time.Time
	Dependencies map[string]bool
	TouchedKeys  map[string]bool
	ServerCfg    *Config
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
	ID           int64
	InsideTx     bool
	TxCommands   []TxCommand
	IsMasterConn bool

	// Only initialized after issuing psync
	IsReplicaConn       bool
	IsReplicaRegistered bool

	// This channel will be used to subscribe to write commands
	// that the engine executes.
	// Only initialized after issuing psync
	ReplicationCh   <-chan []byte
	ReplicationDone chan<- struct{}
}

type TxCommand struct {
	Command string
	Args    []string
}

type Config struct {
	IsMaster                bool
	MasterReplicationID     string
	MasterReplicationOffset int
}

type Address struct {
	Host string
	Port string
}
