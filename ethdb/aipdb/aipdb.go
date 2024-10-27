package aipdb

import (
	"slices"

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

type iterKey struct {
	it  ethdb.Iterator
	key []byte
}
type iterator struct {
	d *Database
	// iterMap         map[string]ethdb.Iterator // peer -> iterator
	iterKeys []*iterKey
	// currIter ethdb.Iterator // current iterator
}

var _ ethdb.Iterator = (*iterator)(nil)

// NewIterator implements ethdb.KeyValueStore.
func (d *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	i := &iterator{
		d: d,
		// iterMap:         make(map[string]ethdb.Iterator, len(d.pebbleMap)),
		// lookaheadKeyMap: make(map[ethdb.Iterator][]byte, len(d.pebbleMap)),
		iterKeys: make([]*iterKey, 0, len(d.pebbleMap)),
	}

	for peer, _ := range d.pebbleMap {
		it := d.pebbleMap[peer].NewIterator(prefix, start)
		if it.Next() {
			i.iterKeys = append(i.iterKeys, &iterKey{it: it, key: it.Key()})
		}
	}

	slices.SortStableFunc(i.iterKeys, func(a, b *iterKey) int {
		return slices.Compare(a.key, b.key)
	})

	return i
}

// Error implements ethdb.Iterator.
func (i *iterator) Error() error {
	if len(i.iterKeys) == 0 {
		return nil
	}

	return i.iterKeys[0].it.Error()
}

// Next implements ethdb.Iterator.
func (i *iterator) Next() bool {
	currIter := i.iterKeys[0].it

	if !currIter.Next() {
		if len(i.iterKeys) == 1 {
			return false
		}

		// remove the iterator that has reached the end
		currIter.Release()

		// the next iterator will be the first one, with the key preloaded
		i.iterKeys = i.iterKeys[1:]

		return true
	}

	// update the key
	i.iterKeys[0].key = currIter.Key()
	// sort the keys again
	slices.SortStableFunc(i.iterKeys, func(a, b *iterKey) int {
		return slices.Compare(a.key, b.key)
	})

	return true
}

// Key implements ethdb.Iterator.
func (i *iterator) Key() []byte {
	return i.iterKeys[0].key
}

// Value implements ethdb.Iterator.
func (i *iterator) Value() []byte {
	return i.iterKeys[0].it.Value()
}

// Release implements ethdb.Iterator.
func (i *iterator) Release() {
	for _, iterKey := range i.iterKeys {
		iterKey.it.Release()
	}
}

// Stat implements ethdb.KeyValueStore.
func (d *Database) Stat() (string, error) {
	return "", nil
}
