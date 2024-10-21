package rpcdb_test

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/dbtest"
	"github.com/ethereum/go-ethereum/ethdb/rpcdb"
	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"github.com/ethereum/go-ethereum/ethdb/rpcdb/handler"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// create new grpc server, client
func newConn(t *testing.T) (*grpc.ClientConn, error) {
	ctx := context.Background()
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	server := grpc.NewServer()
	h, err := handler.NewKVStoreTest()
	if err != nil {
		return nil, err
	}

	api.RegisterKVServer(server, h)
	go func() {
		if err := server.Serve(lis); err != nil {
			assert.NoError(t, err)
		}
	}()

	conn, err := grpc.DialContext(
		ctx,
		"",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	return conn, err
}

func TestRPCDatabase(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		dbtest.TestDatabaseSuite(t, func() ethdb.KeyValueStore {
			conns := make(map[string]*grpc.ClientConn)
			for i := range 3 {
				conn, err := newConn(t)
				assert.NoError(t, err)

				conns[strconv.Itoa(i)] = conn
			}

			db := rpcdb.New(conns)

			return db
		})
	})
}
