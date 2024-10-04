package redis_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/dbtest"
	redisDB "github.com/ethereum/go-ethereum/ethdb/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisDatabase(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		dbtest.TestDatabaseSuite(t, func() ethdb.KeyValueStore {
			client := redis.NewClusterClient(&redis.ClusterOptions{
				// Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002", "localhost:7003", "localhost:7004", "localhost:7005"},
				Addrs:        []string{"localhost:7000"},
				MaxRedirects: 16,
			})

			db := redisDB.New(client, "test")

			assert.NoError(t, db.Reset())

			return db
		})
	})
}
