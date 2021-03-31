package main

import (
	"time"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/stability-tests/common/rpc"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func mineLoop(syncerRPCClient, syncedRPCClient *rpc.Client) error {
	miningAddr, err := generateAddress()
	if err != nil {
		return err
	}

	syncerBlockCountBefore, err := syncerRPCClient.GetBlockCount()
	if err != nil {
		return err
	}
	syncedBlockCountBefore, err := syncedRPCClient.GetBlockCount()
	if err != nil {
		return err
	}
	log.Infof("Starting to mine")
	for i := uint64(0); i < activeConfig().NumberOfBlocks; i++ {
		log.Infof("Mining block %d...", i+1)
		err = mineBlock(syncerRPCClient.Address(), miningAddr)
		if err != nil {
			// Ignore error and instead check that the block count changed correctly.
			// TODO: Fix the race condition in kaspaminer so it won't panic (proper shutdown handler)
			log.Warnf("mineBlock returned an err: %s", err)
		}

		const timeToPropagate = 1 * time.Second
		select {
		case <-syncedRPCClient.OnBlockAdded:
		case <-time.After(timeToPropagate):
			return errors.Errorf("block %d took more than %s to propagate", i+1, timeToPropagate)
		}

		syncerResult, err := syncerRPCClient.GetBlockDAGInfo()
		if err != nil {
			return err
		}

		syncedResult, err := syncedRPCClient.GetBlockDAGInfo()
		if err != nil {
			return err
		}

		if !areTipsAreEqual(syncedResult, syncerResult) {
			return errors.Errorf("syncer node has tips %s but synced node has tips %s", syncerResult.TipHashes, syncedResult.TipHashes)
		}
	}

	log.Infof("Finished to mine")

	log.Infof("Getting syncer block count")
	syncerBlockCountAfter, err := syncerRPCClient.GetBlockCount()
	if err != nil {
		return err
	}

	log.Infof("Getting syncee block count")
	syncedBlockCountAfter, err := syncedRPCClient.GetBlockCount()
	if err != nil {
		return err
	}
	if syncerBlockCountAfter.BlockCount-syncerBlockCountBefore.BlockCount != activeConfig().NumberOfBlocks {
		return errors.Errorf("Expected to mine %d blocks, instead mined: %d", activeConfig().NumberOfBlocks, syncerBlockCountAfter.BlockCount-syncerBlockCountBefore.BlockCount)
	}
	if syncedBlockCountAfter.BlockCount-syncedBlockCountBefore.BlockCount != activeConfig().NumberOfBlocks {
		return errors.Errorf("Expected syncer to have %d new blocks, instead have: %d", activeConfig().NumberOfBlocks, syncedBlockCountAfter.BlockCount-syncedBlockCountBefore.BlockCount)
	}

	log.Infof("Finished the mine loop successfully")
	return nil
}

func generateAddress() (util.Address, error) {
	privateKey, err := secp256k1.GenerateSchnorrKeyPair()
	if err != nil {
		return nil, err
	}

	pubKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		return nil, err
	}

	pubKeySerialized, err := pubKey.Serialize()
	if err != nil {
		return nil, err
	}

	return util.NewAddressPubKeyHashFromPublicKey(pubKeySerialized[:], activeConfig().ActiveNetParams.Prefix)
}

func areTipsAreEqual(resultA, resultB *appmessage.GetBlockDAGInfoResponseMessage) bool {
	if len(resultA.TipHashes) != len(resultB.TipHashes) {
		return false
	}

	tipsASet := make(map[string]struct{})
	for _, tip := range resultA.TipHashes {
		tipsASet[tip] = struct{}{}
	}

	for _, tip := range resultB.TipHashes {
		if _, ok := tipsASet[tip]; !ok {
			return false
		}
	}

	return true
}

func mineBlock(syncerRPCAddress string, miningAddress util.Address) error {
	kaspaMinerCmd, err := common.StartCmd("MINER",
		"kaspaminer",
		common.NetworkCliArgumentFromNetParams(activeConfig().NetParams()),
		"-s", syncerRPCAddress,
		"--mine-when-not-synced",
		"--miningaddr", miningAddress.EncodeAddress(),
		"--numblocks", "1",
	)
	if err != nil {
		return err
	}
	return errors.Wrapf(kaspaMinerCmd.Wait(), "error with command '%s'", kaspaMinerCmd)
}
