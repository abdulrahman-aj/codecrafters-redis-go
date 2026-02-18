package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/server/context"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	srv := server.New()
	go srv.Run()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(srv, conn)
	}
}

// TODO: move this inside the server struct
// might need to refactor the server struct a little
// e.g: extract engine component (as an actor)
func handleConnection(srv *server.Server, conn net.Conn) {
	defer conn.Close()

	reader := resp.NewReader(conn)

	ctx := &context.Connection{}

	for {
		command, err := reader.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}

		// TODO: should tell srv.Do if client canceled or closed the connection, e.g: context.Context
		if _, err := conn.Write(srv.Do(ctx, command)); err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			break
		}
	}

}
