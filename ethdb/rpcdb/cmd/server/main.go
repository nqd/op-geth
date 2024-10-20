package main

import (
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // Register pprof handlers
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/ethdb/pebble"
	api "github.com/ethereum/go-ethereum/ethdb/rpcdb/gen/go/api/v1"
	"github.com/ethereum/go-ethereum/ethdb/rpcdb/handler"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

const (
	addr           = "0.0.0.0:6789"
	maxMsgSize     = 100 * 1024 * 1024 // 100 MB
	prometheusAddr = "0.0.0.0:9801"
)

func main() {
	p, err := newPebble()
	if err != nil {
		log.Panic("Failed to create pebble database", "err", err)
	}

	h := handler.NewKVStoreWithPebble(p)

	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	api.RegisterKVServer(server, h)

	// Enable gRPC Prometheus monitoring
	grpc_prometheus.Register(server)
	startPrometheus()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Println("Starting KV server", "address: ", addr)

	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}

func newPebble() (*pebble.Database, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	directory := filepath.Join(h, "tmp-rpcdb/geth/chaindata")
	cache := 2048
	handle := 5120
	namespace := "rpcdb"
	readonly := false
	ephemeral := false

	return pebble.New(directory, cache, handle, namespace, readonly, ephemeral)
}

func startPrometheus() {
	// Start Prometheus metrics server
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("Starting metrics server at address: %s\n", prometheusAddr)
		log.Printf("\tprometheus metrics available at %s/metrics\n", prometheusAddr)
		log.Printf("\tpprof available at %s/debug/pprof/\n", prometheusAddr)
		if err := http.ListenAndServe(prometheusAddr, nil); err != nil {
			log.Panic("Failed to start Prometheus metrics server", "err", err)
		}
	}()
}
