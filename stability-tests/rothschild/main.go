package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util/profiling"
	"os"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/os/signal"
	"github.com/kaspanet/kaspad/util/panics"

	"github.com/kaspanet/automation/stability-tests/common"
)

var shutdown int32 = 0

func main() {
	interrupt := signal.InterruptListener()
	err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %+v", err)
		os.Exit(1)
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())

	defer panics.HandlePanic(log, "main", nil)

	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	addresses, err := loadAddresses()
	if err != nil {
		panic(err)
	}

	client, err := rpcclient.NewRPCClient(activeConfig().RPCServer)
	if err != nil {
		panic(err)
	}

	client.SetTimeout(5 * time.Minute)

	utxosChangedNotificationChan := make(chan *appmessage.UTXOsChangedNotificationMessage, 100)
	err = client.RegisterForUTXOsChangedNotifications([]string{addresses.myAddress.EncodeAddress()},
		func(notification *appmessage.UTXOsChangedNotificationMessage) {
			utxosChangedNotificationChan <- notification
		})
	if err != nil {
		panic(err)
	}

	spendLoopDoneChan := spendLoop(client, addresses, utxosChangedNotificationChan)

	<-interrupt

	atomic.AddInt32(&shutdown, 1)

	<-spendLoopDoneChan
}
