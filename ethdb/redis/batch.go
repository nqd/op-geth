package redis

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/redis/go-redis/v9"
)

type keyvalue struct {
	key    string
	value  []byte
	delete bool
}

type batch struct {
	db     *Database
	writes []keyvalue
	size   int
}

// Delete implements ethdb.Batch.
func (b *batch) Delete(key []byte) error {
	b.writes = append(b.writes, keyvalue{string(key), nil, true})
	b.size += len(key)

	return nil
}

// Put implements ethdb.Batch.
func (b *batch) Put(key []byte, value []byte) error {
	b.writes = append(b.writes, keyvalue{string(key), common.CopyBytes(value), false})
	b.size += len(key) + len(value)

	return nil
}

// Replay implements ethdb.Batch.
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	for _, kv := range b.writes {
		if kv.delete {
			if err := w.Delete([]byte(kv.key)); err != nil {
				return err
			}

			continue
		}
		if err := w.Put([]byte(kv.key), kv.value); err != nil {
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
	ctx := context.Background()

	_, err := b.db.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		var err error
		for _, kv := range b.writes {
			if kv.delete {
				err = pipe.Del(ctx, kv.key).Err()
			} else {
				err = pipe.Set(ctx, kv.key, kv.value, 0).Err()
			}
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

var _ ethdb.Batch = (*batch)(nil)
