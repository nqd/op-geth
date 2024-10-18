package handler

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"github.com/google/uuid"
)

type kvStore struct {
	api.UnimplementedKVServer

	pebbleDB *pebble.Database

	imLock sync.RWMutex
	im     map[string]ethdb.Iterator
}

func NewKVStore() (*kvStore, error) {
	// TODO: read these values from config
	directory := "data"
	cache := 1000000
	handle := 100
	namespace := "rpcdb"
	readonly := false
	ephemeral := false

	pebbleDB, err := pebble.New(directory, cache, handle, namespace, readonly, ephemeral)
	if err != nil {
		return nil, err
	}

	return &kvStore{
		pebbleDB: pebbleDB,
		im:       make(map[string]ethdb.Iterator, 100),
	}, nil
}

func (s *kvStore) Get(_ context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	res, err := s.pebbleDB.Get(req.GetKey())
	if err != nil {
		return nil, err
	}

	return &api.GetResponse{Value: res}, nil
}

func (s *kvStore) Batch(_ context.Context, req *api.BatchRequest) (*api.BatchResponse, error) {
	batch := s.pebbleDB.NewBatch()
	for _, op := range req.GetBatches() {
		if op.Delete {
			err := batch.Delete(op.GetKey())
			if err != nil {
				return nil, err
			}
			continue
		}

		err := batch.Put(op.GetKey(), op.GetValue())
		if err != nil {
			return nil, err
		}
	}

	if err := batch.Write(); err != nil {
		return nil, err
	}

	return &api.BatchResponse{}, nil
}

func (s *kvStore) NewIterator(_ context.Context, req *api.NewIteratorRequest) (*api.NewIteratorResponse, error) {
	id := uuid.NewString()

	it := s.pebbleDB.NewIterator(req.GetPrefix(), req.GetStart())

	s.imLock.Lock()
	s.im[id] = it
	s.imLock.Unlock()

	return &api.NewIteratorResponse{IteratorId: id}, nil
}
