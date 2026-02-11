package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/server"
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

func handleConnection(srv *server.Server, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		err := consumeCommand(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}

		response := srv.Do(nil)
		if _, err := conn.Write([]byte(response)); err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			break
		}
	}

}

func consumeCommand(reader *bufio.Reader) error {
	for range 3 {
		_, err := readline(reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func readline(reader *bufio.Reader) ([]byte, error) {
	var ret []byte
	for {
		bytes, isPrefix, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}
		ret = append(ret, bytes...)
		if !isPrefix {
			break
		}
	}
	return ret, nil
}
