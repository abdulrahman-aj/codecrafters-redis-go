package server

import (
	"fmt"
	"io"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
)

type Server struct {
	listener net.Listener
	engine   *engine.Engine
}

func New(ip string, port int, replicaOf string) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: listener,
		engine:   engine.New(replicaOf),
	}, nil
}

func (s *Server) Serve() error {
	go s.engine.Run()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := resp.NewReader(conn)

	ctx := &context.Connection{}

	for {
		command, err := reader.ReadValue()
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Error reading from connection: ", err.Error())
			return
		}

		if _, err := conn.Write(s.engine.Do(ctx, command)); err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			break
		}
	}
}
