package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/server"
)

var (
	port      = flag.Int("port", 6379, "port to listen on")
	replicaOf = flag.String("replicaof", "", `replica address in the format "<MASTER_HOST> <MASTER_PORT>"`)
)

func main() {
	fmt.Println("logs should appear here...")

	flag.Parse()

	server, err := server.New("0.0.0.0", *port, *replicaOf)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
