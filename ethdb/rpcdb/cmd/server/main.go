package main

import (
	"net"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	ethdbpebble "github.com/ethereum/go-ethereum/ethdb/pebble"
	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"github.com/ethereum/go-ethereum/ethdb/rpcdb/handler"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
)

func main() {
	p, err := pebble.Open("", &pebble.Options{
		FS: vfs.NewMem(),
	})

	if err != nil {
		panic(err)
	}

	defer p.Close()

	h := handler.NewKVStoreWithPebble(ethdbpebble.NewRaw(p))

	server := grpc.NewServer()

	api.RegisterKVServer(server, h)

	listener, err := net.Listen("tcp", "0.0.0.0:6789")
	if err != nil {
		panic(err)
	}

	log.Info("Starting server")

	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}
