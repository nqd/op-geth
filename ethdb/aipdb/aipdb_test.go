package aipdb

import (
	"strconv"
	"testing"

	cockroachpebble "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/dbtest"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
)

func TestRPCDatabase(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		dbtest.TestDatabaseSuite(t, func() ethdb.KeyValueStore {
			pebbleMap := make(map[string]*pebble.Database)

			for i := range 10 {
				db, err := cockroachpebble.Open("", &cockroachpebble.Options{
					FS: vfs.NewMem(),
				})
				if err != nil {
					t.Fatal(err)
				}

				pebbleMap[strconv.Itoa(i)] = pebble.NewRaw(db)
			}

			return New(pebbleMap)
		})
	})
}
