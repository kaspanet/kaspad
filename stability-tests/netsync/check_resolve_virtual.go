package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/stability-tests/common/rpc"
	"github.com/pkg/errors"
	"time"
)

func checkResolveVirtual(syncerClient, syncedClient *rpc.Client) error {
	syncedBlockCountResponse, err := syncerClient.GetBlockCount()
	if err != nil {
		return err
	}

	rejectReason, err := mineOnTips(syncerClient)
	if err != nil {
		panic(err)
	}
	if rejectReason != appmessage.RejectReasonNone {
		panic(fmt.Sprintf("mined block rejected: %s", rejectReason))
	}

	expectedDuration := time.Duration(syncedBlockCountResponse.BlockCount) * 100 * time.Millisecond
	start := time.Now()
	select {
	case <-time.After(expectedDuration):
		return errors.Errorf("it took more than %s to resolve the virtual", expectedDuration)
	case <-syncedClient.OnBlockAdded:
	}

	log.Infof("It took %s to resolve the virtual", time.Since(start))
	return nil
}
