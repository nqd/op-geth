package redis

import "github.com/ethereum/go-ethereum/ethdb"

type iterator struct {
	index int
	kvs   []keyvalue
	err   error
}

// Error implements ethdb.Iterator.
func (i *iterator) Error() error {
	return i.err
}

// Key implements ethdb.Iterator.
func (i *iterator) Key() []byte {
	if i.index < 0 || i.index >= len(i.kvs) {
		return nil
	}
	return []byte(i.kvs[i.index].key)
}

// Next implements ethdb.Iterator.
func (i *iterator) Next() bool {
	if i.index >= len(i.kvs) {
		return false
	}
	i.index += 1

	return i.index < len(i.kvs)
}

// Release implements ethdb.Iterator.
func (i *iterator) Release() {
	i.index = -1
	i.kvs = nil
	i.err = nil
}

// Value implements ethdb.Iterator.
func (i *iterator) Value() []byte {
	if i.index < 0 || i.index >= len(i.kvs) {
		return nil
	}
	return i.kvs[i.index].value
}

var _ ethdb.Iterator = (*iterator)(nil)
