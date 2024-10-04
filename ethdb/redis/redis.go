package redis

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/redis/go-redis/v9"
)

type Database struct {
	client *redis.ClusterClient

	deleteCountMeter   metrics.Meter
	getCountMeter      metrics.Meter
	hasCountMeter      metrics.Meter
	putCountMeter      metrics.Meter
	batchCountMeter    metrics.Meter
	batchWithSizeMeter metrics.Meter
	iteratorCountMeter metrics.Meter

	deleteCount        atomic.Int64 // Total number of delete operations
	getCount           atomic.Int64 // Total number of get operations
	hasCount           atomic.Int64 // Total number of has operations
	putCount           atomic.Int64 // Total number of put operations
	batchCount         atomic.Int64 // Total number of new batch operations
	batchWithSizeCount atomic.Int64 // Total number of new batch with size operations
	iteratorCount      atomic.Int64 // Total number of new iterator operations
}

const (
	metricsGatheringInterval = 3 * time.Second
)

var errRemoteRedisNotFound = errors.New("not found")

func New(client *redis.ClusterClient, namespace string) *Database {
	db := &Database{
		client: client,
	}

	db.deleteCountMeter = metrics.GetOrRegisterMeter(namespace+"deletecount", nil)
	db.getCountMeter = metrics.GetOrRegisterMeter(namespace+"getcount", nil)
	db.hasCountMeter = metrics.GetOrRegisterMeter(namespace+"hascount", nil)
	db.putCountMeter = metrics.GetOrRegisterMeter(namespace+"putcount", nil)
	db.batchCountMeter = metrics.GetOrRegisterMeter(namespace+"batchcount", nil)
	db.batchWithSizeMeter = metrics.GetOrRegisterMeter(namespace+"batchwithsizecount", nil)
	db.iteratorCountMeter = metrics.GetOrRegisterMeter(namespace+"iteratorcount", nil)

	go db.meter(metricsGatheringInterval)

	return db
}

var _ ethdb.KeyValueStore = (*Database)(nil)

// Close implements ethdb.KeyValueStore.
func (d *Database) Close() error {
	return d.client.Close()
}

// Compact implements ethdb.KeyValueStore.
// Does not support compaction.
func (d *Database) Compact(start []byte, limit []byte) error {
	return nil
}

// Delete implements ethdb.KeyValueStore.
func (d *Database) Delete(key []byte) error {
	d.deleteCount.Add(1)

	ctx := context.Background()
	delCmd := d.client.Del(ctx, string(key))

	return delCmd.Err()
}

// Get implements ethdb.KeyValueStore.
func (d *Database) Get(key []byte) ([]byte, error) {
	d.getCount.Add(1)

	return d.get(key)
}

func (d *Database) get(key []byte) ([]byte, error) {
	ctx := context.Background()
	getCmd := d.client.Get(ctx, string(key))

	if getCmd.Err() == redis.Nil {
		return nil, errRemoteRedisNotFound
	}

	if getCmd.Err() != nil {
		return nil, getCmd.Err()
	}

	return getCmd.Bytes()
}

// Has implements ethdb.KeyValueStore.
func (d *Database) Has(key []byte) (bool, error) {
	d.hasCount.Add(1)

	_, err := d.get(key)
	if err != nil {
		if err == errRemoteRedisNotFound {
			return false, nil
		}

		return false, err
	}
	return true, nil
}

// NewBatch implements ethdb.KeyValueStore.
func (d *Database) NewBatch() ethdb.Batch {
	d.batchCount.Add(1)

	return &batch{
		db:     d,
		writes: make([]keyvalue, 0),
	}
}

// NewBatchWithSize implements ethdb.KeyValueStore.
func (d *Database) NewBatchWithSize(size int) ethdb.Batch {
	d.batchWithSizeCount.Add(1)

	return &batch{
		db:     d,
		writes: make([]keyvalue, 0, size),
	}
}

// NewIterator implements ethdb.KeyValueStore.
func (d *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	d.iteratorCount.Add(1)

	ctx := context.Background()
	size := 100
	pr := string(prefix)
	st := string(append(prefix, start...))

	// get all kv at one then sort the result
	// though this is not efficient, but this is the only way to implement sorted iterator in redis
	// for each master function run in parallel
	iter := iterator{
		index: -1,
		kvs:   make([]keyvalue, 0, size),
	}
	var kvsLock sync.Mutex
	err := d.client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
		rit := client.Scan(ctx, 0, pr+"*", 0).Iterator()

		for rit.Next(ctx) {
			k := rit.Val()
			if k >= st {
				kvsLock.Lock()
				iter.kvs = append(iter.kvs, keyvalue{
					key:   k,
					value: nil,
				})
				kvsLock.Unlock()
			}
		}
		return rit.Err()
	})

	if err != nil {
		iter.err = err
	}

	// sort by the key
	slices.SortFunc(iter.kvs, func(a, b keyvalue) int {
		return strings.Compare(a.key, b.key)
	})

	// query values via pipeline
	cmds, err := d.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i := range iter.kvs {
			cmd := pipe.Get(ctx, iter.kvs[i].key)
			if cmd.Err() != nil {
				return cmd.Err()
			}
		}
		return nil
	})
	if err != nil {
		iter.err = err
	}

	for i, cmd := range cmds {
		if cmd.Err() != nil {
			iter.err = cmd.Err()
			break
		}
		val, err := cmd.(*redis.StringCmd).Bytes()
		if err != nil {
			iter.err = err
			break
		}

		iter.kvs[i].value = val
	}

	return &iter
}

// Put implements ethdb.KeyValueStore.
func (d *Database) Put(key []byte, value []byte) error {
	d.putCount.Add(1)

	ctx := context.Background()
	setCmd := d.client.Set(ctx, string(key), value, 0)

	return setCmd.Err()
}

// Stat implements ethdb.KeyValueStore.
func (d *Database) Stat() (string, error) {
	return "", nil
}

func (d *Database) meter(refresh time.Duration) {
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	for range ticker.C {
		d.deleteCountMeter.Mark(d.deleteCount.Swap(0))
		d.getCountMeter.Mark(d.getCount.Swap(0))
		d.hasCountMeter.Mark(d.hasCount.Swap(0))
		d.putCountMeter.Mark(d.putCount.Swap(0))
		d.batchCountMeter.Mark(d.batchCount.Swap(0))
		d.batchWithSizeMeter.Mark(d.batchWithSizeCount.Swap(0))
		d.iteratorCountMeter.Mark(d.iteratorCount.Swap(0))
	}
}

func (d *Database) Reset() error {
	return d.client.ForEachMaster(context.Background(), func(ctx context.Context, client *redis.Client) error {
		return client.FlushAll(context.Background()).Err()
	})
}
