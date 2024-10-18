package handler

import api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"

type kvStore struct {
	api.UnimplementedKVServer
}

func NewKVStore() *kvStore {
	return &kvStore{}
}
