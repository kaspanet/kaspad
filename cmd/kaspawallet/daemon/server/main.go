package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/infrastructure/os/signal"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedKaspawalletdServer

	rpcClient *rpcclient.RPCClient

	lock               sync.RWMutex
	utxos              map[externalapi.DomainOutpoint]*walletUTXO
	nextSyncStartIndex uint32
	keysFile           *keys.Data
	cfg                *configFlags
	shutdown           chan struct{}
}

func main() {
	defer panics.HandlePanic(log, "MAIN", nil)
	interrupt := signal.InterruptListener()

	cfg, err := parseConfig()
	if err != nil {
		printErrorAndExit(errors.Wrap(err, "Error parsing command-line arguments"))
	}
	defer backendLog.Close()

	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		printErrorAndExit(errors.Wrapf(err, "Error listening to tcp at %s", cfg.Listen))
	}

	rpcClient, err := connectToRPC(cfg.NetParams(), cfg.RPCServer)
	if err != nil {
		printErrorAndExit(errors.Wrapf(err, "Error connecting to RPC server %s", cfg.RPCServer))
	}

	keysFile, err := keys.ReadKeysFile(cfg.NetParams(), cfg.KeysFile)
	if err != nil {
		return
	}

	serverInstance := &server{
		rpcClient:          rpcClient,
		utxos:              make(map[externalapi.DomainOutpoint]*walletUTXO),
		nextSyncStartIndex: 0,
		keysFile:           keysFile,
		cfg:                cfg,
		shutdown:           make(chan struct{}),
	}

	spawn("serverInstance.sync", func() {
		err := serverInstance.sync()
		if err != nil {
			printErrorAndExit(errors.Wrap(err, "error syncing the wallet"))
		}
	})

	grpcServer := grpc.NewServer()
	pb.RegisterKaspawalletdServer(grpcServer, serverInstance)

	spawn("grpcServer.Serve", func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			printErrorAndExit(errors.Wrap(err, "Error serving gRPC"))
		}
	})

	select {
	case <-serverInstance.shutdown:
	case <-interrupt:
		const stopTimeout = 2 * time.Second

		stopChan := make(chan interface{})
		spawn("gRPCServer.Stop", func() {
			grpcServer.GracefulStop()
			close(stopChan)
		})

		select {
		case <-stopChan:
		case <-time.After(stopTimeout):
			log.Warnf("Could not gracefully stop: timed out after %s", stopTimeout)
			grpcServer.Stop()
		}
	}
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
