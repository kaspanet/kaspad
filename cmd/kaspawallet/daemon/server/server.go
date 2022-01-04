package server

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/kaspanet/kaspad/util/profiling"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/infrastructure/os/signal"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedKaspawalletdServer

	rpcClient *rpcclient.RPCClient
	params    *dagconfig.Params

	lock                sync.RWMutex
	utxosSortedByAmount []*walletUTXO
	nextSyncStartIndex  uint32
	keysFile            *keys.File
	shutdown            chan struct{}
}

// Start starts the kaspawalletd server
func Start(params *dagconfig.Params, listen, rpcServer string, keysFilePath string, profile string) error {
	initLog(defaultLogFile, defaultErrLogFile)

	defer panics.HandlePanic(log, "MAIN", nil)
	interrupt := signal.InterruptListener()

	if profile != "" {
		profiling.Start(profile, log)
	}

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return (errors.Wrapf(err, "Error listening to tcp at %s", listen))
	}
	log.Infof("Listening on %s", listen)

	rpcClient, err := connectToRPC(params, rpcServer)
	if err != nil {
		return (errors.Wrapf(err, "Error connecting to RPC server %s", rpcServer))
	}

	keysFile, err := keys.ReadKeysFile(params, keysFilePath)
	if err != nil {
		return (errors.Wrapf(err, "Error connecting to RPC server %s", rpcServer))
	}

	serverInstance := &server{
		rpcClient:           rpcClient,
		params:              params,
		utxosSortedByAmount: []*walletUTXO{},
		nextSyncStartIndex:  0,
		keysFile:            keysFile,
		shutdown:            make(chan struct{}),
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

	return nil
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
	os.Exit(1)
}
