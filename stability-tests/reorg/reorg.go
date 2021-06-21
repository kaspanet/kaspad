package main

import (
	"compress/gzip"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func testReorg(cfg *configFlags) {
	consensusConfig := consensus.Config{Params: dagconfig.DevnetParams}
	consensusConfig.SkipProofOfWork = true
	consensusConfig.DisableDifficultyAdjustment = true

	factory := consensus.NewFactory()
	tc, teardown, err := factory.NewTestConsensus(&consensusConfig, "ReorgHonest")
	if err != nil {
		panic(err)
	}
	defer teardown(false)

	f, err := os.Open(cfg.DAGFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	defer gzipReader.Close()

	_, err = tc.MineJSON(gzipReader, testapi.MineJSONBlockTypeUTXOValidBlock)
	if err != nil {
		panic(err)
	}

	tcAttacker, teardownAttacker, err := factory.NewTestConsensus(&consensusConfig, "ReorgAttacker")
	if err != nil {
		panic(err)
	}
	defer teardownAttacker(false)

	virtualSelectedParent, err := tc.GetVirtualSelectedParent()
	if err != nil {
		panic(err)
	}

	stagingArea := model.NewStagingArea()
	virtualSelectedParentGHOSTDAGData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), stagingArea, virtualSelectedParent)
	if err != nil {
		panic(err)
	}

	log.Infof("Selected tip blue score %d", virtualSelectedParentGHOSTDAGData.BlueScore())

	sideChain := make([]*externalapi.DomainBlock, 0)

	for i := uint64(0); ; i++ {
		tips, err := tcAttacker.Tips()
		if err != nil {
			panic(err)
		}

		block, _, err := tcAttacker.BuildBlockWithParents(tips, nil, nil)
		if err != nil {
			panic(err)
		}

		// We change the nonce of the first block so its hash won't be similar to any of the
		// honest DAG blocks. As a result the rest of the side chain should have unique hashes
		// as well.
		if i == 0 {
			mutableHeader := block.Header.ToMutable()
			mutableHeader.SetNonce(uint64(rand.NewSource(84147).Int63()))
			block.Header = mutableHeader.ToImmutable()
		}

		_, err = tcAttacker.ValidateAndInsertBlock(block, true)
		if err != nil {
			panic(err)
		}

		sideChain = append(sideChain, block)

		if i%100 == 0 {
			log.Infof("Attacker side chain mined %d blocks", i)
		}

		blockHash := consensushashing.BlockHash(block)
		ghostdagData, err := tcAttacker.GHOSTDAGDataStore().Get(tcAttacker.DatabaseContext(), stagingArea, blockHash)
		if err != nil {
			panic(err)
		}

		if virtualSelectedParentGHOSTDAGData.BlueWork().Cmp(ghostdagData.BlueWork()) == -1 {
			break
		}
	}

	sideChainTipHash := consensushashing.BlockHash(sideChain[len(sideChain)-1])
	sideChainTipGHOSTDAGData, err := tcAttacker.GHOSTDAGDataStore().Get(tcAttacker.DatabaseContext(), stagingArea, sideChainTipHash)
	if err != nil {
		panic(err)
	}

	log.Infof("Side chain tip (%s) blue score %d", sideChainTipHash, sideChainTipGHOSTDAGData.BlueScore())

	doneChan := make(chan struct{})
	spawn("add-sidechain-to-honest", func() {
		for i, block := range sideChain {
			if i%100 == 0 {
				log.Infof("Validated %d blocks from the attacker chain", i)
			}
			_, err := tc.ValidateAndInsertBlock(block, true)
			if err != nil {
				panic(err)
			}
		}

		doneChan <- struct{}{}
	})

	const timeout = 12 * time.Hour
	select {
	case <-doneChan:
	case <-time.After(timeout):
		fail("Adding the side chain took more than %s", timeout)
	}

	sideChainTipGHOSTDAGData, err = tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), stagingArea, sideChainTipHash)
	if err != nil {
		panic(err)
	}

	log.Infof("Side chain tip (%s) blue score %d", sideChainTipHash, sideChainTipGHOSTDAGData.BlueScore())

	newVirtualSelectedParent, err := tc.GetVirtualSelectedParent()
	if err != nil {
		panic(err)
	}

	if !newVirtualSelectedParent.Equal(sideChainTipHash) {
		fail("No reorg happened")
	}
}

func fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, msg)
	log.Criticalf(msg)
	backendLog.Close()
	os.Exit(1)
}
