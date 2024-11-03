package redis_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/dbtest"
	ethValkeyDB "github.com/ethereum/go-ethereum/ethdb/valkey"
	"github.com/stretchr/testify/assert"
	"github.com/valkey-io/valkey-go"
)

func TestRedisDatabase(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		dbtest.TestDatabaseSuite(t, func() ethdb.KeyValueStore {
			client, err := valkey.NewClient(valkey.ClientOption{
				InitAddress: []string{"localhost:6379"},
			})

			assert.NoError(t, err)

			db := ethValkeyDB.New(client, "test")

			assert.NoError(t, db.Reset())

			return db
		})
	})
}
