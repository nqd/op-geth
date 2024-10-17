package rpcdb

import (
	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"google.golang.org/grpc"
)

type Database struct {
	kvClient api.KVClient
}

func New(conn *grpc.ClientConn) *Database {
	kvClient := api.NewKVClient(conn)

	return &Database{
		kvClient: kvClient,
	}
}
