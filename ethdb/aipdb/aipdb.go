package aipdb

import (
	"slices"

	cockroachpebble "github.com/cockroachdb/pebble"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
	"github.com/golang/groupcache/consistenthash"
)

type Database struct {
	peers     *consistenthash.Map
	pebbleMap map[string]*pebble.Database // peer -> pebble database
}

var _ ethdb.KeyValueStore = (*Database)(nil)

func New(pebbleMap map[string]*pebble.Database) *Database {
	d := &Database{
		pebbleMap: pebbleMap,
		// todo: may provide an option to use custom hash function
		peers: consistenthash.New(len(pebbleMap), nil),
	}

	for key, _ := range pebbleMap {
		d.peers.Add(key)
	}

	return d
}

func (d *Database) getPebbleDB(key []byte) *pebble.Database {
	peer := d.peers.Get(string(key))

	return d.pebbleMap[peer]
}

// Close implements ethdb.KeyValueStore.
func (d *Database) Close() error {
	for _, pdb := range d.pebbleMap {
		if err := pdb.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Compact implements ethdb.KeyValueStore.
func (d *Database) Compact(start []byte, limit []byte) error {
	for _, pdb := range d.pebbleMap {
		if err := pdb.Compact(start, limit); err != nil {
			return err
		}
	}

	return nil
}

// Delete implements ethdb.KeyValueStore.
func (d *Database) Delete(key []byte) error {
	pdb := d.getPebbleDB(key)

	return pdb.Delete(key)
}

// Get implements ethdb.KeyValueStore.
func (d *Database) Get(key []byte) ([]byte, error) {
	pdb := d.getPebbleDB(key)

	return pdb.Get(key)
}

// Has implements ethdb.KeyValueStore.
func (d *Database) Has(key []byte) (bool, error) {
	pdb := d.getPebbleDB(key)

	return pdb.Has(key)
}

// Put implements ethdb.KeyValueStore.
func (d *Database) Put(key []byte, value []byte) error {
	pdb := d.getPebbleDB(key)

	return pdb.Put(key, value)
}

// NewBatch implements ethdb.KeyValueStore.
func (d *Database) NewBatch() ethdb.Batch {
	b := &batch{
		d:        d,
		batchMap: make(map[string]ethdb.Batch, len(d.pebbleMap)),
	}

	for peer, _ := range d.pebbleMap {
		b.batchMap[peer] = d.pebbleMap[peer].NewBatch()
	}

	return b
}

// NewBatchWithSize implements ethdb.KeyValueStore.
func (d *Database) NewBatchWithSize(size int) ethdb.Batch {
	b := &batch{
		d:        d,
		batchMap: make(map[string]ethdb.Batch, len(d.pebbleMap)),
	}

	for peer, _ := range d.pebbleMap {
		b.batchMap[peer] = d.pebbleMap[peer].NewBatchWithSize(size)
	}

	return b
}

type batch struct {
	d        *Database
	batchMap map[string]ethdb.Batch // peer -> batch
}

var _ ethdb.Batch = (*batch)(nil)

func (b *batch) getPebbleBatch(key []byte) ethdb.Batch {
	peer := b.d.peers.Get(string(key))

	return b.batchMap[peer]
}

// Delete implements ethdb.Batch.
func (b *batch) Delete(key []byte) error {
	pb := b.getPebbleBatch(key)

	return pb.Delete(key)
}

// Put implements ethdb.Batch.
func (b *batch) Put(key []byte, value []byte) error {
	pb := b.getPebbleBatch(key)

	return pb.Put(key, value)
}

// Replay implements ethdb.Batch.
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	for _, pb := range b.batchMap {
		if err := pb.Replay(w); err != nil {
			return err
		}
	}

	return nil
}

// Reset implements ethdb.Batch.
func (b *batch) Reset() {
	for _, pb := range b.batchMap {
		pb.Reset()
	}
}

// ValueSize implements ethdb.Batch.
func (b *batch) ValueSize() int {
	s := 0
	for _, pb := range b.batchMap {
		s += pb.ValueSize()
	}

	return s
}

// Write implements ethdb.Batch.
func (b *batch) Write() error {
	// TODO: consider using goroutine to write in parallel
	for _, pb := range b.batchMap {
		if err := pb.Write(); err != nil {
			return err
		}
	}
	return nil
}

type pebbleIterator struct {
	iter *cockroachpebble.Iterator
	key  []byte
}
type iterator struct {
	d           *Database
	pebbleIters []*pebbleIterator
	moved       bool
	released    bool
}

var _ ethdb.Iterator = (*iterator)(nil)

// copied from pebble/pebble.go
func upperBound(prefix []byte) (limit []byte) {
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c == 0xff {
			continue
		}
		limit = make([]byte, i+1)
		copy(limit, prefix)
		limit[i] = c + 1
		break
	}
	return limit
}

// NewIterator implements ethdb.KeyValueStore.
func (d *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	i := &iterator{
		d:           d,
		pebbleIters: make([]*pebbleIterator, 0, len(d.pebbleMap)),
		moved:       true,
		released:    false,
	}

	for _, pdb := range d.pebbleMap {
		iter, _ := pdb.GetDB().NewIter(&cockroachpebble.IterOptions{
			LowerBound: append(prefix, start...),
			UpperBound: upperBound(prefix),
		})
		iter.First()

		if iter.Valid() {
			i.pebbleIters = append(i.pebbleIters, &pebbleIterator{
				iter: iter,
				key:  iter.Key(),
			})
		}
	}

	slices.SortStableFunc(i.pebbleIters, func(a, b *pebbleIterator) int {
		return slices.Compare(a.key, b.key)
	})

	return i
}

// Error implements ethdb.Iterator.
func (i *iterator) Error() error {
	if len(i.pebbleIters) == 0 {
		return nil
	}

	return i.pebbleIters[0].iter.Error()
}

// Next implements ethdb.Iterator.
func (i *iterator) Next() bool {
	if len(i.pebbleIters) == 0 {
		return false
	}

	if i.moved {
		i.moved = false

		return true
	}

	currIter := i.pebbleIters[0].iter

	if !currIter.Next() {
		currIter.Close()

		// the next iterator will be the first one, with the key preloaded
		i.pebbleIters = i.pebbleIters[1:]

		return len(i.pebbleIters) != 0
	}

	// update the key
	i.pebbleIters[0].key = currIter.Key()
	// sort the keys again
	slices.SortStableFunc(i.pebbleIters, func(a, b *pebbleIterator) int {
		return slices.Compare(a.key, b.key)
	})

	return true
}

// Key implements ethdb.Iterator.
func (i *iterator) Key() []byte {
	return i.pebbleIters[0].key
}

// Value implements ethdb.Iterator.
func (i *iterator) Value() []byte {
	return i.pebbleIters[0].iter.Value()
}

// Release implements ethdb.Iterator.
func (i *iterator) Release() {
	if i.released {
		return
	}

	for _, iterKey := range i.pebbleIters {
		iterKey.iter.Close()
	}
}

// Stat implements ethdb.KeyValueStore.
func (d *Database) Stat() (string, error) {
	return "", nil
}
