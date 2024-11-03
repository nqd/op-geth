package redis

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/valkey-io/valkey-go"
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

func newBatch(db *Database, size int) *batch {
	return &batch{
		db:     db,
		writes: make([]keyvalue, 0, size),
		size:   0,
	}
}

// Delete implements ethdb.Batch.
func (b *batch) Delete(key []byte) error {
	b.db.batchDelete.Add(1)

	b.writes = append(b.writes, keyvalue{string(key), nil, true})
	b.size += len(key)

	return nil
}

// Put implements ethdb.Batch.
func (b *batch) Put(key []byte, value []byte) error {
	b.db.batchPut.Add(1)

	b.writes = append(b.writes, keyvalue{string(key), common.CopyBytes(value), false})
	b.size += len(key) + len(value)

	return nil
}

// Replay implements ethdb.Batch.
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	b.db.batchReplay.Add(1)

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
	b.db.batchReset.Add(1)

	b.writes = b.writes[:0]
	b.size = 0
}

// ValueSize implements ethdb.Batch.
func (b *batch) ValueSize() int {
	b.db.batchValueSize.Add(1)

	return b.size
}

// Write implements ethdb.Batch.
func (b *batch) Write() error {
	b.db.batchWrite.Add(1)

	ctx := context.Background()

	cmds := make(valkey.Commands, 0, len(b.writes)*2)

	for _, kv := range b.writes {
		k := b.db.conHash.Get(kv.key)

		if kv.delete {
			cmds = append(cmds, b.db.client.B().Del().Key(kv.key).Build())
			cmds = append(cmds, b.db.client.B().Zrem().Key(k).Member(kv.key).Build())
		} else {
			cmds = append(cmds, b.db.client.B().Set().Key(kv.key).Value(string(kv.value)).Build())
			cmds = append(cmds, b.db.client.B().Zadd().Key(k).ScoreMember().ScoreMember(0, kv.key).Build())
		}
	}
	for _, res := range b.db.client.DoMulti(ctx, cmds...) {
		if res.Error() != nil {
			return res.Error()
		}
	}

	return nil
}

var _ ethdb.Batch = (*batch)(nil)
