package rpcdb

import (
	"context"
	"errors"
	"io"
	"slices"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"github.com/ethereum/go-ethereum/log"
	"github.com/golang/groupcache/consistenthash"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Database struct {
	mu      sync.RWMutex
	peers   *consistenthash.Map
	clients map[string]api.KVClient // peer -> client
}

func New(conns map[string]*grpc.ClientConn) *Database {
	d := &Database{
		clients: make(map[string]api.KVClient),
	}

	// use default hash function crc32.ChecksumIEEE
	// todo: may provide an option to use custom hash function
	d.peers = consistenthash.New(len(conns), nil)

	for key, conn := range conns {
		kvClient := api.NewKVClient(conn)
		d.clients[key] = kvClient
		d.peers.Add(key)
	}

	return d
}

var _ ethdb.KeyValueStore = (*Database)(nil)

var ctx = context.Background()

// Close implements ethdb.KeyValueStore.
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, kvClient := range d.clients {
		_, err := kvClient.Close(ctx, &api.CloseRequest{})
		if err != nil {
			return err
		}
	}

	return nil
}

// Compact implements ethdb.KeyValueStore.
func (d *Database) Compact(start []byte, limit []byte) error {
	errg, ctx := errgroup.WithContext(ctx)

	for _, kvClient := range d.clients {
		errg.Go(func() error {
			_, err := kvClient.Compact(ctx, &api.CompactRequest{Start: start, Limit: limit})
			return err
		})
	}

	return errg.Wait()
}

// Delete implements ethdb.KeyValueStore.
func (d *Database) Delete(key []byte) error {
	peer := d.peers.Get(string(key))
	kvClient := d.clients[peer]

	_, err := kvClient.Delete(ctx, &api.DeleteRequest{Key: key})
	return err
}

// Get implements ethdb.KeyValueStore.
func (d *Database) Get(key []byte) ([]byte, error) {
	peer := d.peers.Get(string(key))
	kvClient := d.clients[peer]

	res, err := kvClient.Get(ctx, &api.GetRequest{Key: key})
	if err != nil {
		return nil, err
	}

	return res.GetValue(), nil
}

// Has implements ethdb.KeyValueStore.
func (d *Database) Has(key []byte) (bool, error) {
	peer := d.peers.Get(string(key))
	kvClient := d.clients[peer]

	res, err := kvClient.Has(ctx, &api.HasRequest{Key: key})
	if err != nil {
		return false, err
	}

	return res.GetFound(), nil
}

// Put implements ethdb.KeyValueStore.
func (d *Database) Put(key []byte, value []byte) error {
	peer := d.peers.Get(string(key))
	kvClient := d.clients[peer]

	_, err := kvClient.Put(ctx, &api.PutRequest{Key: key, Value: value})

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
	peerWrites := make(map[string][]*api.Batch)

	for _, kv := range b.writes {
		peer := b.d.peers.Get(string(kv.GetKey()))
		peerWrites[peer] = append(peerWrites[peer], kv)
	}

	errg, ctx := errgroup.WithContext(ctx)

	for peer, writes := range peerWrites {
		kvClient := b.d.clients[peer]
		errg.Go(func() error {
			_, err := kvClient.Batch(ctx, &api.BatchRequest{Batches: writes})
			return err
		})
	}

	return errg.Wait()
}

// NewIterator implements ethdb.KeyValueStore.
// func (d *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
// 	i := iterator{
// 		index: -1,
// 		kvs:   make([]*keyvalue, 0),
// 	}

// 	errg, ctx := errgroup.WithContext(ctx)

// 	for _, kvClient := range d.clients {
// 		errg.Go(func() error {
// 			res, err := kvClient.NewIterator(ctx, &api.NewIteratorRequest{Prefix: prefix, Start: start})
// 			if err != nil {
// 				return err
// 			}
// 			iterID := res.GetIteratorId()
// 			for {
// 				nextRes, err := kvClient.IteratorNext(ctx, &api.IteratorNextRequest{IteratorId: iterID})
// 				if err != nil {
// 					return err
// 				}
// 				if !nextRes.GetNext() {
// 					// close the iterator
// 					// todo: may need to handle the error
// 					_, _ = kvClient.IteratorRelease(ctx, &api.IteratorReleaseRequest{IteratorId: iterID})

// 					break
// 				}
// 				// todo: could parallelize this part
// 				keyRes, err := kvClient.IteratorKey(ctx, &api.IteratorKeyRequest{IteratorId: iterID})
// 				if err != nil {
// 					return err
// 				}
// 				valRes, err := kvClient.IteratorValue(ctx, &api.IteratorValueRequest{IteratorId: iterID})
// 				if err != nil {
// 					return err
// 				}
// 				i.mu.Lock()
// 				i.kvs = append(i.kvs, &keyvalue{key: keyRes.GetKey(), value: valRes.GetValue()})
// 				i.mu.Unlock()
// 			}

// 			return nil
// 		})
// 	}

// 	err := errg.Wait()
// 	if err != nil {
// 		i.err = err
// 		return nil
// 	}

// 	// sort by the key
// 	slices.SortFunc(i.kvs, func(a, b *keyvalue) int {
// 		return slices.Compare(a.key, b.key)
// 	})

// 	return &i
// }

func (d *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	i := iterator{
		index: -1,
		kvs:   make([]*keyvalue, 0, 500),
	}
	var lock sync.Mutex

	errg, ctx := errgroup.WithContext(ctx)

	for _, kvClient := range d.clients {
		errg.Go(func() error {
			stream, err := kvClient.NewIteratorStream(ctx, &api.NewIteratorStreamRequest{Prefix: prefix, Start: start})
			if err != nil {
				return err
			}

			for {
				res, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return err
				}
				lock.Lock()
				i.kvs = append(i.kvs, &keyvalue{key: res.GetKey(), value: res.GetValue()})
				lock.Unlock()
			}

			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		i.err = err
		return nil
	}

	// sort by the key
	slices.SortFunc(i.kvs, func(a, b *keyvalue) int {
		return slices.Compare(a.key, b.key)
	})

	i.size = len(i.kvs)

	return &i
}

type keyvalue struct {
	key   []byte
	value []byte
}

type iterator struct {
	mu    sync.RWMutex
	index int
	kvs   []*keyvalue
	err   error
	size  int
}

var _ ethdb.Iterator = (*iterator)(nil)

// Error implements ethdb.Iterator.
func (i *iterator) Error() error {
	return i.err
}

// Key implements ethdb.Iterator.
func (i *iterator) Key() []byte {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if i.index < 0 || i.index >= len(i.kvs) {
		return nil
	}
	return i.kvs[i.index].key
}

// Next implements ethdb.Iterator.
func (i *iterator) Next() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if i.index >= len(i.kvs) {
		return false
	}
	i.index += 1

	return i.index < len(i.kvs)
}

// Release implements ethdb.Iterator.
func (i *iterator) Release() {
	// for debug
	log.Info("=== release iterator ===", "size", i.size, "index", i.index)

	i.mu.Lock()
	defer i.mu.Unlock()

	i.index = -1
	i.kvs = i.kvs[:0]
	i.err = nil
}

// Value implements ethdb.Iterator.
func (i *iterator) Value() []byte {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if i.index < 0 || i.index >= len(i.kvs) {
		return nil
	}
	return i.kvs[i.index].value
}
