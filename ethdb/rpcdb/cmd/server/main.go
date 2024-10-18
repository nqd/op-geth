package main

import (
	"log"
	"net"

	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"github.com/ethereum/go-ethereum/ethdb/rpcdb/handler"
	"google.golang.org/grpc"
)

func main() {
	h, err := handler.NewKVStoreTest()
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()

	api.RegisterKVServer(server, h)

	addr := "0.0.0.0:6789"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Println("Starting KV server", "address: ", addr)

	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}
