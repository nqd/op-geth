package rpcdb

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
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

var ctx = context.Background()

// Close implements ethdb.KeyValueStore.
func (d *Database) Close() error {
	_, err := d.kvClient.Close(ctx, &api.CloseRequest{})

	return err
}

// Compact implements ethdb.KeyValueStore.
func (d *Database) Compact(start []byte, limit []byte) error {
	_, err := d.kvClient.Compact(ctx, &api.CompactRequest{Start: start, Limit: limit})

	return err
}

// Delete implements ethdb.KeyValueStore.
func (d *Database) Delete(key []byte) error {
	_, err := d.kvClient.Delete(ctx, &api.DeleteRequest{Key: key})

	return err
}

// Get implements ethdb.KeyValueStore.
func (d *Database) Get(key []byte) ([]byte, error) {
	res, err := d.kvClient.Get(ctx, &api.GetRequest{Key: key})
	if err != nil {
		return nil, err
	}

	return res.GetValue(), nil
}

// Has implements ethdb.KeyValueStore.
func (d *Database) Has(key []byte) (bool, error) {
	res, err := d.kvClient.Has(ctx, &api.HasRequest{Key: key})
	if err != nil {
		return false, err
	}

	return res.GetFound(), nil
}

// Put implements ethdb.KeyValueStore.
func (d *Database) Put(key []byte, value []byte) error {
	_, err := d.kvClient.Put(ctx, &api.PutRequest{Key: key, Value: value})

	return err
}

// Stat implements ethdb.KeyValueStore.
func (d *Database) Stat() (string, error) {
	return "", nil
}

// NewBatch implements ethdb.KeyValueStore.
func (d *Database) NewBatch() ethdb.Batch {
	return &batch{
		d:      d,
		writes: make([]*api.Batch, 0),
	}
}

// NewBatchWithSize implements ethdb.KeyValueStore.
func (d *Database) NewBatchWithSize(size int) ethdb.Batch {
	return &batch{
		d:      d,
		writes: make([]*api.Batch, 0, size),
	}
}

type batch struct {
	d      *Database
	writes []*api.Batch
	size   int
}

var _ ethdb.Batch = (*batch)(nil)

// Delete implements ethdb.Batch.
func (b *batch) Delete(key []byte) error {
	b.writes = append(b.writes, &api.Batch{Key: key, Value: nil, Delete: true})
	b.size += len(key)

	return nil
}

// Put implements ethdb.Batch.
func (b *batch) Put(key []byte, value []byte) error {
	b.writes = append(b.writes, &api.Batch{Key: key, Value: common.CopyBytes(value), Delete: false})
	b.size += len(key) + len(value)

	return nil
}

// Replay implements ethdb.Batch.
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	for _, kv := range b.writes {
		if kv.GetDelete() {
			if err := w.Delete(kv.GetKey()); err != nil {
				return err
			}

			continue
		}
		if err := w.Put(kv.GetKey(), kv.GetValue()); err != nil {
			return err
		}
	}

	return nil
}

// Reset implements ethdb.Batch.
func (b *batch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}

// ValueSize implements ethdb.Batch.
func (b *batch) ValueSize() int {
	return b.size
}

// Write implements ethdb.Batch.
func (b *batch) Write() error {
	_, err := b.d.kvClient.Batch(ctx, &api.BatchRequest{Batches: b.writes})

	return err
}

// NewIterator implements ethdb.KeyValueStore.
func (d *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	res, err := d.kvClient.NewIterator(ctx, &api.NewIteratorRequest{Prefix: prefix, Start: start})
	if err != nil {
		return &iterator{
			d:   d,
			err: err,
		}
	}

	return &iterator{
		d:      d,
		iterID: res.GetIteratorId(),
	}
}

type iterator struct {
	d      *Database
	iterID string
	err    error
}

var _ ethdb.Iterator = (*iterator)(nil)

// Error implements ethdb.Iterator.
func (i *iterator) Error() error {
	// TODO: need to call server?
	return i.err
}

// Key implements ethdb.Iterator.
func (i *iterator) Key() []byte {
	res, err := i.d.kvClient.IteratorKey(ctx, &api.IteratorKeyRequest{IteratorId: i.iterID})
	if err != nil {
		i.err = err
		return nil
	}

	return res.GetKey()
}

// Next implements ethdb.Iterator.
func (i *iterator) Next() bool {
	res, err := i.d.kvClient.IteratorNext(ctx, &api.IteratorNextRequest{IteratorId: i.iterID})
	if err != nil {
		i.err = err
		return false
	}

	return res.GetNext()
}

// Release implements ethdb.Iterator.
func (i *iterator) Release() {
	// ignore error
	_, _ = i.d.kvClient.IteratorRelease(ctx, &api.IteratorReleaseRequest{IteratorId: i.iterID})
}

// Value implements ethdb.Iterator.
func (i *iterator) Value() []byte {
	res, err := i.d.kvClient.IteratorValue(ctx, &api.IteratorValueRequest{IteratorId: i.iterID})
	if err != nil {
		i.err = err
		return nil
	}

	return res.GetValue()
}
