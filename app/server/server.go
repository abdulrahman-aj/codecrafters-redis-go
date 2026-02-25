package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
	"github.com/codecrafters-io/redis-starter-go/app/util"
)

type Server struct {
	listener net.Listener
	engine   *engine.Engine
	lastID   atomic.Int64
}

// can't create multiple resp.Readers for a single client
// since an earlier reader may have consumed some bytes from conn
// which will lead to non-deterministic errors in the second reader.
type client struct {
	conn   net.Conn
	reader *resp.Reader
}

func New(ip string, listeningPort int, masterAddress *types.Address) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, listeningPort))
	if err != nil {
		return nil, err
	}

	master, err := connectToMaster(listeningPort, masterAddress)
	if err != nil {
		return nil, err
	}

	s := &Server{listener: listener, engine: engine.New(master == nil)}
	if master != nil {
		// TODO: auto-repair connection to master?
		go s.handleConnection(master, &types.ConnectionCtx{ID: s.lastID.Add(1), IsMasterConn: true})
	}

	return s, nil
}

func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(
			&client{conn: conn, reader: resp.NewReader(conn)},
			&types.ConnectionCtx{ID: s.lastID.Add(1)},
		)
	}
}

func (s *Server) handleConnection(c *client, ctx *types.ConnectionCtx) {
	defer c.conn.Close()

	for {
		command, err := c.reader.ReadValue()
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from connection: ", err.Error())
			}

			break
		}

		if _, err := c.conn.Write(s.engine.Do(ctx, command)); err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			break
		}

		if ctx.IsReplicaConn {
			s.handleReplica(c, ctx)
			break
		}
	}
}

func (s *Server) handleReplica(c *client, ctx *types.ConnectionCtx) {
	util.Assert(
		ctx.IsReplicaConn && ctx.IsReplicaRegistered && ctx.ReplicationCh != nil,
		"handleReplica running without subscribing to replication log",
	)

	for req := range ctx.ReplicationCh {
		err := util.Retry(func() error {
			_, err := c.conn.Write(req)
			return err
		}, 3)

		if err != nil {
			close(ctx.ReplicationDone)
			break
		}
	}
}

func connectToMaster(listeningPort int, address *types.Address) (*client, error) {
	if address == nil {
		return nil, nil
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(address.Host, address.Port))
	if err != nil {
		return nil, err
	}

	reader := resp.NewReader(conn)

	// a bit ugly...
	handshake := [][]string{
		{"PING"}, // TODO: assert OK
		{"REPLCONF", "listening-port", strconv.Itoa(listeningPort)}, // assert OK
		{"REPLCONF", "capa", "psync2"},                              // assert OK
	}

	err = func() error {
		for _, args := range handshake {
			req := resp.Array(args)

			if _, err := conn.Write(req); err != nil {
				return err
			}

			if _, err := reader.ReadValue(); err != nil {
				return err
			}
		}

		if _, err := conn.Write(resp.Array([]string{"PSYNC", "?", "-1"})); err != nil {
			return err
		}

		if _, err := reader.ReadValue(); err != nil {
			return err
		}

		if _, err := reader.ReadRDB(); err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		conn.Close()
		return nil, err
	}

	return &client{conn: conn, reader: reader}, nil
}
