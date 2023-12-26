package server

import (
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/version"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/util/txmass"

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

	rpcClient           *rpcclient.RPCClient // RPC client for ongoing user requests
	backgroundRPCClient *rpcclient.RPCClient // RPC client dedicated for address and UTXO background fetching
	params              *dagconfig.Params

	lock                            sync.RWMutex
	utxosSortedByAmount             []*walletUTXO
	nextSyncStartIndex              uint32
	keysFile                        *keys.File
	shutdown                        chan struct{}
	forceSyncChan                   chan struct{}
	startTimeOfLastCompletedRefresh time.Time
	addressSet                      walletAddressSet
	txMassCalculator                *txmass.Calculator
	usedOutpoints                   map[externalapi.DomainOutpoint]time.Time
	firstSyncDone                   atomic.Bool

	isLogFinalProgressLineShown bool
	maxUsedAddressesForLog      uint32
	maxProcessedAddressesForLog uint32
}

// MaxDaemonSendMsgSize is the max send message size used for the daemon server.
// Currently, set to 100MB
const MaxDaemonSendMsgSize = 100_000_000

// Start starts the kaspawalletd server
func Start(params *dagconfig.Params, listen, rpcServer string, keysFilePath string, profile string, timeout uint32) error {
	initLog(defaultLogFile, defaultErrLogFile)

	defer panics.HandlePanic(log, "MAIN", nil)
	interrupt := signal.InterruptListener()

	if profile != "" {
		profiling.Start(profile, log)
	}

	log.Infof("Version %s", version.Version())
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
	backgroundRPCClient, err := connectToRPC(params, rpcServer, timeout)
	if err != nil {
		return (errors.Wrapf(err, "Error making a second connection to RPC server %s", rpcServer))
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
		backgroundRPCClient:         backgroundRPCClient,
		params:                      params,
		utxosSortedByAmount:         []*walletUTXO{},
		nextSyncStartIndex:          0,
		keysFile:                    keysFile,
		shutdown:                    make(chan struct{}),
		forceSyncChan:               make(chan struct{}),
		addressSet:                  make(walletAddressSet),
		txMassCalculator:            txmass.NewCalculator(params.MassPerTxByte, params.MassPerScriptPubKeyByte, params.MassPerSigOp),
		usedOutpoints:               map[externalapi.DomainOutpoint]time.Time{},
		isLogFinalProgressLineShown: false,
		maxUsedAddressesForLog:      0,
		maxProcessedAddressesForLog: 0,
	}

	log.Infof("Read, syncing the wallet...")
	spawn("serverInstance.syncLoop", func() {
		err := serverInstance.syncLoop()
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
