package handler

import (
	"context"
	"sync"

	cockroachdbpebble "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
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

func NewKVStoreWithPebble(pebbleDB *pebble.Database) *kvStore {
	return &kvStore{
		pebbleDB: pebbleDB,
		im:       make(map[string]ethdb.Iterator, 100),
	}
}

func NewKVStoreTest() (*kvStore, error) {
	p, err := cockroachdbpebble.Open("", &cockroachdbpebble.Options{
		FS: vfs.NewMem(),
	})

	if err != nil {
		return nil, err
	}
	return &kvStore{
		pebbleDB: pebble.NewRaw(p),
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

func (s *kvStore) Has(_ context.Context, req *api.HasRequest) (*api.HasResponse, error) {
	found, err := s.pebbleDB.Has(req.GetKey())
	if err != nil {
		return nil, err
	}

	return &api.HasResponse{Found: found}, nil
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

func (s *kvStore) Compact(_ context.Context, req *api.CompactRequest) (*api.CompactResponse, error) {
	start := req.GetStart()
	limit := req.GetLimit()
	if err := s.pebbleDB.Compact(start, limit); err != nil {
		return nil, err
	}

	return &api.CompactResponse{}, nil
}

func (s *kvStore) Delete(_ context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	if err := s.pebbleDB.Delete(req.GetKey()); err != nil {
		return nil, err
	}

	return &api.DeleteResponse{}, nil
}

func (s *kvStore) Put(_ context.Context, req *api.PutRequest) (*api.PutResponse, error) {
	if err := s.pebbleDB.Put(req.GetKey(), req.GetValue()); err != nil {
		return nil, err
	}

	return &api.PutResponse{}, nil
}

func (s *kvStore) Reset(_ context.Context, req *api.ResetRequest) (*api.ResetResponse, error) {
	// todo: implement

	return &api.ResetResponse{}, nil
}

func (s *kvStore) Close(_ context.Context, req *api.CloseRequest) (*api.CloseResponse, error) {
	if err := s.pebbleDB.Close(); err != nil {
		return nil, err
	}

	return &api.CloseResponse{}, nil
}
