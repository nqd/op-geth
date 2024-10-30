package redis

import "github.com/ethereum/go-ethereum/ethdb"

type iterator struct {
	db    *Database
	index int
	keys  []string
	err   error
}

// Error implements ethdb.Iterator.
func (i *iterator) Error() error {
	return i.err
}

// Key implements ethdb.Iterator.
func (i *iterator) Key() []byte {
	if i.index < 0 || i.index >= len(i.keys) {
		return nil
	}
	return []byte(i.keys[i.index])
}

// Next implements ethdb.Iterator.
func (i *iterator) Next() bool {
	if i.index >= len(i.keys) {
		return false
	}
	i.index += 1

	return i.index < len(i.keys)
}

// Release implements ethdb.Iterator.
func (i *iterator) Release() {
	i.index = -1
	i.keys = nil
	i.err = nil
}

// Value implements ethdb.Iterator.
func (i *iterator) Value() []byte {
	if i.index < 0 || i.index >= len(i.keys) {
		return nil
	}

	i.db.get([]byte(i.keys[i.index]))
}

var _ ethdb.Iterator = (*iterator)(nil)
