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

		go s.handleConnection(conn, &types.ConnectionCtx{ID: s.lastID.Add(1)})
	}
}

func (s *Server) handleConnection(conn net.Conn, ctx *types.ConnectionCtx) {
	defer conn.Close()
	reader := resp.NewReader(conn)

	for {
		command, err := reader.ReadValue()
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from connection: ", err.Error())
			}

			break
		}

		if _, err := conn.Write(s.engine.Do(ctx, command)); err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			break
		}

		if ctx.IsReplicaConn {
			s.handleReplica(conn, ctx)
			break
		}
	}
}

func (s *Server) handleReplica(conn net.Conn, ctx *types.ConnectionCtx) {
	util.Assert(
		ctx.IsReplicaConn && ctx.IsReplicaRegistered && ctx.ReplicationCh != nil,
		"handleReplica running without subscribing to replication log",
	)

	for req := range ctx.ReplicationCh {
		err := util.Retry(func() error {
			_, err := conn.Write(req)
			return err
		}, 3)

		if err != nil {
			close(ctx.ReplicationDone)
			break
		}
	}
}

func connectToMaster(listeningPort int, address *types.Address) (net.Conn, error) {
	if address == nil {
		return nil, nil
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(address.Host, address.Port))
	if err != nil {
		return nil, err
	}

	reader := resp.NewReader(conn)

	// TODO: check responses
	handshake := [][]string{
		{"PING"}, // OK
		{"REPLCONF", "listening-port", strconv.Itoa(listeningPort)}, // OK
		{"REPLCONF", "capa", "psync2"},                              // OK
		{"PSYNC", "?", "-1"},                                        // FULLRESYNC
	}

	for _, args := range handshake {
		req := resp.Array(args)

		if _, err := conn.Write(req); err != nil {
			conn.Close()
			return nil, err
		}

		if _, err := reader.ReadValue(); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return conn, nil
}
