package server

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"

	"github.com/c4ei/yunseokyeol/util/txmass"

	"github.com/c4ei/yunseokyeol/util/profiling"

	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/daemon/pb"
	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/keys"
	"github.com/c4ei/yunseokyeol/domain/dagconfig"
	"github.com/c4ei/yunseokyeol/infrastructure/network/rpcclient"
	"github.com/c4ei/yunseokyeol/infrastructure/os/signal"
	"github.com/c4ei/yunseokyeol/util/panics"
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
	addressSet          walletAddressSet
	txMassCalculator    *txmass.Calculator
	usedOutpoints       map[externalapi.DomainOutpoint]time.Time

	isLogFinalProgressLineShown bool
	maxUsedAddressesForLog      uint32
	maxProcessedAddressesForLog uint32
}

// MaxDaemonSendMsgSize is the max send message size used for the daemon server.
// Currently, set to 100MB
const MaxDaemonSendMsgSize = 100_000_000

// Start starts the c4exwalletd server
func Start(params *dagconfig.Params, listen, rpcServer string, keysFilePath string, profile string, timeout uint32) error {
	initLog(defaultLogFile, defaultErrLogFile)

	defer panics.HandlePanic(log, "MAIN", nil)
	interrupt := signal.InterruptListener()

	if profile != "" {
		profiling.Start(profile, log)
	}

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return (errors.Wrapf(err, "Error listening to TCP on %s", listen))
	}
	log.Infof("Listening to TCP on %s", listen)

	log.Infof("Connecting to a node at %s...", rpcServer)
	rpcClient, err := connectToRPC(params, rpcServer, timeout)
	if err != nil {
		return (errors.Wrapf(err, "Error connecting to RPC server %s", rpcServer))
	}

	log.Infof("Connected, reading keys file %s...", keysFilePath)
	keysFile, err := keys.ReadKeysFile(params, keysFilePath)
	if err != nil {
		return (errors.Wrapf(err, "Error reading keys file %s", keysFilePath))
	}

	err = keysFile.TryLock()
	if err != nil {
		return err
	}

	serverInstance := &server{
		rpcClient:                   rpcClient,
		params:                      params,
		utxosSortedByAmount:         []*walletUTXO{},
		nextSyncStartIndex:          0,
		keysFile:                    keysFile,
		shutdown:                    make(chan struct{}),
		addressSet:                  make(walletAddressSet),
		txMassCalculator:            txmass.NewCalculator(params.MassPerTxByte, params.MassPerScriptPubKeyByte, params.MassPerSigOp),
		usedOutpoints:               map[externalapi.DomainOutpoint]time.Time{},
		isLogFinalProgressLineShown: false,
		maxUsedAddressesForLog:      0,
		maxProcessedAddressesForLog: 0,
	}

	log.Infof("Read, syncing the wallet...")
	spawn("serverInstance.sync", func() {
		err := serverInstance.sync()
		if err != nil {
			printErrorAndExit(errors.Wrap(err, "error syncing the wallet"))
		}
	})

	grpcServer := grpc.NewServer(grpc.MaxSendMsgSize(MaxDaemonSendMsgSize))
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
