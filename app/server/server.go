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

func New(ip, port string) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: listener,
		engine:   engine.New(),
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
		if err != nil { // TODO: should tell the server to cancel any blocking operations related to this connection
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
