package main

import (
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	fmt.Println("logs should appear hear...")

	server, err := server.New("0.0.0.0", "6379")
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
