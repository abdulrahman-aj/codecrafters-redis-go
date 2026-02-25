package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/types"
	"github.com/codecrafters-io/redis-starter-go/app/util"
)

var (
	listeningPort = flag.Int("port", 6379, "port to listen on")
	replicaOf     = flag.String("replicaof", "", `replica address in the format "<MASTER_HOST> <MASTER_PORT>"`)
)

func main() {
	fmt.Println("logs should appear here...")
	flag.Parse()

	var address *types.Address

	if *replicaOf != "" {
		parts := strings.Split(*replicaOf, " ")
		util.Assert(len(parts) == 2, `replica address in the format "<MASTER_HOST> <MASTER_PORT>"`)
		address = &types.Address{Host: parts[0], Port: parts[1]}
	}

	server, err := server.New("0.0.0.0", *listeningPort, address)
	util.FatalOnErr(err)
	util.FatalOnErr(server.Serve())
}
