package main

import (
	"log"
	"net"

	"github.com/ethereum/go-ethereum/ethdb/pebble"
	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"github.com/ethereum/go-ethereum/ethdb/rpcdb/handler"
	"google.golang.org/grpc"
)

func main() {
	p, err := newPebble()
	if err != nil {
		log.Panic("Failed to create pebble database", "err", err)
	}

	h := handler.NewKVStoreWithPebble(p)

	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024), // 100 MB
		grpc.MaxSendMsgSize(100*1024*1024), // 100 MB
	)

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

func newPebble() (*pebble.Database, error) {
	directory := "~/tmp-rpcdb/geth/chaindata"
	cache := 2048
	handle := 5120
	namespace := "rpcdb"
	readonly := false
	ephemeral := false

	return pebble.New(directory, cache, handle, namespace, readonly, ephemeral)
}
