package rpcdb

import (
	"context"

	"github.com/ethereum/go-ethereum/ethdb"
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

var _ ethdb.KeyValueStore = (*Database)(nil)

// Close implements ethdb.KeyValueStore.
func (d *Database) Close() error {
	panic("unimplemented")
}

// Compact implements ethdb.KeyValueStore.
func (d *Database) Compact(start []byte, limit []byte) error {
	panic("unimplemented")
}

// Delete implements ethdb.KeyValueStore.
func (d *Database) Delete(key []byte) error {
	panic("unimplemented")
}

// Get implements ethdb.KeyValueStore.
func (d *Database) Get(key []byte) ([]byte, error) {
	ctx := context.Background()

	res, err := d.kvClient.Get(ctx, &api.GetRequest{Key: key})
	if err != nil {
		return nil, err
	}

	return res.GetValue(), nil
}

// Has implements ethdb.KeyValueStore.
func (d *Database) Has(key []byte) (bool, error) {
	panic("unimplemented")
}

// NewBatch implements ethdb.KeyValueStore.
func (d *Database) NewBatch() ethdb.Batch {
	panic("unimplemented")
}

// NewBatchWithSize implements ethdb.KeyValueStore.
func (d *Database) NewBatchWithSize(size int) ethdb.Batch {
	panic("unimplemented")
}

// NewIterator implements ethdb.KeyValueStore.
func (d *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	panic("unimplemented")
}

// Put implements ethdb.KeyValueStore.
func (d *Database) Put(key []byte, value []byte) error {
	panic("unimplemented")
}

// Stat implements ethdb.KeyValueStore.
func (d *Database) Stat() (string, error) {
	panic("unimplemented")
}
