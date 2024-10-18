package handler

import (
	"context"
	"errors"

	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"github.com/google/uuid"
)

func (s *kvStore) NewIterator(_ context.Context, req *api.NewIteratorRequest) (*api.NewIteratorResponse, error) {
	id := uuid.NewString()

	it := s.pebbleDB.NewIterator(req.GetPrefix(), req.GetStart())

	s.imLock.Lock()
	s.im[id] = it
	s.imLock.Unlock()

	return &api.NewIteratorResponse{IteratorId: id}, nil
}
func (s *kvStore) IteratorError(_ context.Context, req *api.IteratorErrorRequest) (*api.IteratorErrorResponse, error) {
	id := req.GetIteratorId()

	s.imLock.RLock()
	iter, ok := s.im[id]
	s.imLock.RUnlock()

	if !ok {
		return nil, errors.New("iterator not found")
	}

	err := iter.Error()

	return &api.IteratorErrorResponse{Error: err != nil, Msg: err.Error()}, nil
}

func (s *kvStore) IteratorKey(_ context.Context, req *api.IteratorKeyRequest) (*api.IteratorKeyResponse, error) {
	id := req.GetIteratorId()

	s.imLock.RLock()
	iter, ok := s.im[id]
	s.imLock.RUnlock()

	if !ok {
		return nil, errors.New("iterator not found")
	}

	key := iter.Key()

	return &api.IteratorKeyResponse{Key: key}, nil
}

func (s *kvStore) IteratorNext(_ context.Context, req *api.IteratorNextRequest) (*api.IteratorNextResponse, error) {
	id := req.GetIteratorId()

	s.imLock.RLock()
	iter, ok := s.im[id]
	s.imLock.RUnlock()

	if !ok {
		return nil, errors.New("iterator not found")
	}

	next := iter.Next()

	return &api.IteratorNextResponse{Next: next}, nil
}

func (s *kvStore) IteratorRelease(_ context.Context, req *api.IteratorReleaseRequest) (*api.IteratorReleaseResponse, error) {
	id := req.GetIteratorId()

	s.imLock.Lock()
	iter, ok := s.im[id]
	if ok {
		iter.Release()
		delete(s.im, id)
	}
	s.imLock.Unlock()

	return &api.IteratorReleaseResponse{}, nil
}

func (s *kvStore) IteratorValue(_ context.Context, req *api.IteratorValueRequest) (*api.IteratorValueResponse, error) {
	id := req.GetIteratorId()

	s.imLock.RLock()
	iter, ok := s.im[id]
	s.imLock.RUnlock()

	if !ok {
		return nil, errors.New("iterator not found")
	}

	value := iter.Value()

	return &api.IteratorValueResponse{Value: value}, nil
}
