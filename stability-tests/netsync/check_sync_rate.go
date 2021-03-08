package main

import (
	"time"

	"github.com/kaspanet/kaspad/stability-tests/common/rpc"

	"github.com/pkg/errors"
)

func checkSyncRate(syncerClient, syncedClient *rpc.Client) error {
	log.Info("Checking the sync rate")
	syncerBlockCountResponse, err := syncerClient.GetBlockCount()
	if err != nil {
		return err
	}

	syncerHeadersCount := syncerBlockCountResponse.HeaderCount
	syncerBlockCount := syncerBlockCountResponse.BlockCount
	log.Infof("SYNCER block count: %d headers and %d blocks", syncerHeadersCount, syncerBlockCount)
	// We give 5 seconds for IBD to start and then 100 milliseconds for each block.
	expectedTime := time.Now().Add(5*time.Second + time.Duration(syncerHeadersCount)*100*time.Millisecond)
	start := time.Now()
	const tickDuration = 10 * time.Second
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()
	for range ticker.C {
		log.Info("Getting SYNCED block count")
		syncedBlockCountResponse, err := syncedClient.GetBlockCount()
		if err != nil {
			return err
		}
		log.Infof("SYNCED block count: %d headers and %d blocks", syncedBlockCountResponse.HeaderCount,
			syncedBlockCountResponse.BlockCount)
		if syncedBlockCountResponse.BlockCount >= syncerBlockCount &&
			syncedBlockCountResponse.HeaderCount >= syncerHeadersCount {
			break
		}
		if time.Now().After(expectedTime) {
			return errors.Errorf("SYNCED is not synced in the expected rate")
		}
	}
	log.Infof("IBD took approximately %s", time.Since(start))
	return nil
}
