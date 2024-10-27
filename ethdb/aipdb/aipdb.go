package aipdb

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
	"github.com/golang/groupcache/consistenthash"
)

type Database struct {
	peers *consistenthash.Map
	pdb   map[string]*pebble.Database
}

var _ ethdb.KeyValueStore = (*Database)(nil)

// Close implements ethdb.KeyValueStore.
func (d *Database) Close() error {
	for _, pdb := range d.pdb {
		if err := pdb.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Compact implements ethdb.KeyValueStore.
func (d *Database) Compact(start []byte, limit []byte) error {
	for _, pdb := range d.pdb {
		if err := pdb.Compact(start, limit); err != nil {
			return err
		}
	}

	return nil
}

// Delete implements ethdb.KeyValueStore.
func (d *Database) Delete(key []byte) error {
	peer := d.peers.Get(string(key))

	pdb := d.pdb[peer]

	return pdb.Delete(key)
}

// Get implements ethdb.KeyValueStore.
func (d *Database) Get(key []byte) ([]byte, error) {
	peer := d.peers.Get(string(key))

	pdb := d.pdb[peer]

	return pdb.Get(key)
}

// Has implements ethdb.KeyValueStore.
func (d *Database) Has(key []byte) (bool, error) {
	peer := d.peers.Get(string(key))

	pdb := d.pdb[peer]

	return pdb.Has(key)
}

// NewBatch implements ethdb.KeyValueStore.
func (d *Database) NewBatch() ethdb.Batch {
	panic("unimplemented")
}

// NewBatchWithSize implements ethdb.KeyValueStore.
func (d *Database) NewBatchWithSize(size int) ethdb.Batch {
	panic("unimplemented")
}

// TODO
type batch struct {
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
