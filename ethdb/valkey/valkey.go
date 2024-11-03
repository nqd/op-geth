package redis

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/golang/groupcache/consistenthash"
	"github.com/redis/go-redis/v9"
	"github.com/valkey-io/valkey-go"
	"golang.org/x/sync/errgroup"
)

type Database struct {
	client valkey.Client

	deleteCountMeter metrics.Meter
	getCountMeter    metrics.Meter
	hasCountMeter    metrics.Meter
	putCountMeter    metrics.Meter

	batchCountMeter          metrics.Meter
	batchDeleteCountMeter    metrics.Meter
	batchPutCountMeter       metrics.Meter
	batchReplayCountMeter    metrics.Meter
	batchResetCountMeter     metrics.Meter
	batchWriteCountMeter     metrics.Meter
	batchValueSizeCountMeter metrics.Meter

	batchWithSizeMeter metrics.Meter
	iteratorCountMeter metrics.Meter
	iteratorSetupMeter metrics.Meter

	deleteCount atomic.Int64 // Total number of delete operations
	getCount    atomic.Int64 // Total number of get operations
	hasCount    atomic.Int64 // Total number of has operations
	putCount    atomic.Int64 // Total number of put operations
	batchCount  atomic.Int64 // Total number of new batch operations

	batchDelete    atomic.Int64 // Total number of new batch.delete operations
	batchPut       atomic.Int64 // Total number of new batch.put operations
	batchReplay    atomic.Int64 // Total number of new batch.replay operations
	batchReset     atomic.Int64 // Total number of new batch.reset operations
	batchWrite     atomic.Int64 // Total number of new batch.write operations
	batchValueSize atomic.Int64 // Total number of new batch.valueSize operations

	batchWithSizeCount atomic.Int64 // Total number of new batch with size operations
	iteratorCount      atomic.Int64 // Total number of new iterator operations

	conHash      *consistenthash.Map
	conHashPeers []string
}

const (
	metricsGatheringInterval = 3 * time.Second
)

var errValkeyNotFound = errors.New("not found")

func New(client valkey.Client, namespace string) *Database {
	db := &Database{
		client: client,
	}

	db.deleteCountMeter = metrics.GetOrRegisterMeter(namespace+"deletecount", nil)
	db.getCountMeter = metrics.GetOrRegisterMeter(namespace+"getcount", nil)
	db.hasCountMeter = metrics.GetOrRegisterMeter(namespace+"hascount", nil)
	db.putCountMeter = metrics.GetOrRegisterMeter(namespace+"putcount", nil)

	db.batchCountMeter = metrics.GetOrRegisterMeter(namespace+"batchcount", nil)
	db.batchDeleteCountMeter = metrics.GetOrRegisterMeter(namespace+"batchdeletecount", nil)
	db.batchPutCountMeter = metrics.GetOrRegisterMeter(namespace+"batchputcount", nil)
	db.batchReplayCountMeter = metrics.GetOrRegisterMeter(namespace+"batchreplaycount", nil)
	db.batchResetCountMeter = metrics.GetOrRegisterMeter(namespace+"batchresetcount", nil)
	db.batchWriteCountMeter = metrics.GetOrRegisterMeter(namespace+"batchwritecount", nil)
	db.batchValueSizeCountMeter = metrics.GetOrRegisterMeter(namespace+"batchvaluesizecount", nil)

	db.batchWithSizeMeter = metrics.GetOrRegisterMeter(namespace+"batchwithsizecount", nil)
	db.iteratorCountMeter = metrics.GetOrRegisterMeter(namespace+"iteratorcount", nil)
	db.iteratorSetupMeter = metrics.GetOrRegisterMeter(namespace+"iteratorsetup", nil)

	go db.meter(metricsGatheringInterval)

	chPeerCount := 16
	conHashPeers := make([]string, 0, chPeerCount)
	conHash := consistenthash.New(chPeerCount, nil)
	for i := range chPeerCount {
		peer := fmt.Sprintf("peer-%d", i)
		conHashPeers = append(conHashPeers, peer)
		conHash.Add(peer)
	}

	db.conHash = conHash
	db.conHashPeers = conHashPeers

	return db
}

var _ ethdb.KeyValueStore = (*Database)(nil)

// Close implements ethdb.KeyValueStore.
func (d *Database) Close() error {
	d.client.Close()

	return nil
}

// Compact implements ethdb.KeyValueStore.
// Does not support compaction.
func (d *Database) Compact(start []byte, limit []byte) error {
	return nil
}

// Delete implements ethdb.KeyValueStore.
func (d *Database) Delete(key []byte) error {
	keyStr := string(key)

	d.deleteCount.Add(1)

	ctx := context.Background()
	delCmd := d.client.Do(
		ctx,
		d.client.B().Del().Key(keyStr).Build(),
	)

	// todo: use lua to ensure atomicity
	// ignore error for now
	ordSetKey := d.conHash.Get(keyStr)
	d.client.Do(
		ctx,
		d.client.B().Zrem().Key(ordSetKey).Member(keyStr).Build(),
	)

	return delCmd.Error()
}

// Get implements ethdb.KeyValueStore.
func (d *Database) Get(key []byte) ([]byte, error) {
	d.getCount.Add(1)

	return d.get(key)
}

func (d *Database) get(key []byte) ([]byte, error) {
	ctx := context.Background()
	getCmd := d.client.Do(
		ctx,
		d.client.B().Get().Key(string(key)).Build(),
	)

	if getCmd.Error() == redis.Nil {
		return nil, errValkeyNotFound
	}

	if getCmd.Error() != nil {
		return nil, getCmd.Error()
	}

	return getCmd.AsBytes()
}

// Has implements ethdb.KeyValueStore.
func (d *Database) Has(key []byte) (bool, error) {
	d.hasCount.Add(1)

	_, err := d.get(key)
	if err != nil {
		if err == errValkeyNotFound {
			return false, nil
		}

		return false, err
	}
	return true, nil
}

// NewBatch implements ethdb.KeyValueStore.
func (d *Database) NewBatch() ethdb.Batch {
	d.batchCount.Add(1)

	return newBatch(d, 0)
}

// NewBatchWithSize implements ethdb.KeyValueStore.
func (d *Database) NewBatchWithSize(size int) ethdb.Batch {
	d.batchWithSizeCount.Add(1)

	return newBatch(d, size)
}

// upperBound returns the upper bound for the given prefix
func upperBound(prefix []byte) (limit []byte) {
	ub := make([]byte, len(prefix), len(prefix)+1)
	copy(ub, prefix)
	return append(ub, 0xff)
}

// NewIterator implements ethdb.KeyValueStore.
func (d *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	t1 := time.Now()
	defer func() {
		du := time.Since(t1)

		log.Info("NewIterator setup", "durationInSec", du.Seconds())

		d.iteratorSetupMeter.Mark(int64(du))
	}()

	d.iteratorCount.Add(1)

	ctx := context.Background()
	size := 128
	lowerBoundBytes := (append(prefix, start...))
	upperBoundBytes := (upperBound(prefix))
	lowerBound := string(append([]byte("["), lowerBoundBytes...))
	upperBound := string(append([]byte("["), upperBoundBytes...))

	// get all kv at one then sort the result
	// though this is not efficient, but this is the only way to implement sorted iterator in redis
	// for each master function run in parallel
	iter := iterator{
		db:    d,
		index: -1,
		keys:  make([]string, 0, size),
	}

	var keysLock sync.Mutex

	errg, ctx := errgroup.WithContext(ctx)

	for _, peer := range d.conHashPeers {
		errg.Go(func() error {
			ssCmd := d.client.Do(
				ctx,
				d.client.B().Zrangebylex().Key(peer).Min(lowerBound).Max(upperBound).Limit(0, 0).Build(),
			)
			if err := ssCmd.Error(); err != nil {
				return err
			}

			keysLock.Lock()
			iter.keys = append(iter.keys, ssCmd.Val()...)
			keysLock.Unlock()

			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		iter.err = err
		return &iter
	}

	// sort by the key
	slices.SortFunc(iter.keys, func(a, b string) int {
		return strings.Compare(a, b)
	})

	return &iter
}

// Put implements ethdb.KeyValueStore.
func (d *Database) Put(key []byte, value []byte) error {
	d.putCount.Add(1)

	ctx := context.Background()
	setCmd := d.client.Do(
		ctx,
		d.client.B().Set().Key(string(key)).Value(string(value)).Build(),
	)

	// todo: use lua to ensure atomicity
	// also ignore error for now
	ordSetKey := d.conHash.Get(string(key))
	d.client.Do(
		ctx,
		d.client.B().Zadd().Key(ordSetKey).ScoreMember().ScoreMember(0, string(key)).Build(),
	)
	// d.client.ZAdd(ctx, ordSetKey, redis.Z{Member: string(key)})

	return setCmd.Error()
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
		d.batchDeleteCountMeter.Mark(d.batchDelete.Swap(0))
		d.batchPutCountMeter.Mark(d.batchPut.Swap(0))
		d.batchReplayCountMeter.Mark(d.batchReplay.Swap(0))
		d.batchResetCountMeter.Mark(d.batchReset.Swap(0))
		d.batchWriteCountMeter.Mark(d.batchWrite.Swap(0))
		d.batchValueSizeCountMeter.Mark(d.batchValueSize.Swap(0))

		d.batchWithSizeMeter.Mark(d.batchWithSizeCount.Swap(0))
		d.iteratorCountMeter.Mark(d.iteratorCount.Swap(0))
	}
}

func (d *Database) Reset() error {
	flushallCmd := d.client.Do(
		context.Background(),
		d.client.B().Flushall().Build(),
	)

	return flushallCmd.Error()
}
