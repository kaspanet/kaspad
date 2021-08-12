package mine

import (
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/mining"
	"github.com/kaspanet/kaspad/stability-tests/common/rpc"
	"github.com/pkg/errors"
)

// FromFile mines all blocks as described by `jsonFile`
func FromFile(jsonFile string, consensusConfig *consensus.Config, rpcClient *rpc.Client, dataDir string) error {
	log.Infof("Mining blocks from JSON file %s from data directory %s", jsonFile, dataDir)
	blockChan, err := readBlocks(jsonFile)
	if err != nil {
		return err
	}

	return mineBlocks(consensusConfig, rpcClient, blockChan, dataDir)
}

func mineBlocks(consensusConfig *consensus.Config, rpcClient *rpc.Client, blockChan <-chan JSONBlock, dataDir string) error {
	mdb, err := newMiningDB(dataDir)
	if err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "db")
	factory := consensus.NewFactory()
	factory.SetTestDataDir(dbPath)
	testConsensus, tearDownFunc, err := factory.NewTestConsensus(consensusConfig, "minejson")
	if err != nil {
		return err
	}
	defer tearDownFunc(true)

	info, err := testConsensus.GetSyncInfo()
	if err != nil {
		return err
	}

	log.Infof("Starting with data directory with %d headers and %d blocks", info.HeaderCount, info.BlockCount)

	err = mdb.putID("0", consensusConfig.GenesisHash)
	if err != nil {
		return err
	}

	totalBlocksSubmitted := 0
	lastLogTime := time.Now()
	rpcWaitInInterval := 0 * time.Second
	for blockData := range blockChan {
		if hash := mdb.hashByID(blockData.ID); hash != nil {
			_, err := rpcClient.GetBlock(hash.String(), false)
			if err == nil {
				continue
			}

			if !strings.Contains(err.Error(), "not found") {
				return err
			}
		}

		block, err := mineOrFetchBlock(blockData, mdb, testConsensus)
		if err != nil {
			return err
		}

		beforeSubmitBlockTime := time.Now()
		rejectReason, err := rpcClient.SubmitBlock(block)
		if err != nil {
			return errors.Wrap(err, "error in SubmitBlock")
		}
		if rejectReason != appmessage.RejectReasonNone {
			return errors.Errorf("block rejected in SubmitBlock")
		}
		rpcWaitInInterval += time.Since(beforeSubmitBlockTime)

		totalBlocksSubmitted++
		const logInterval = 1000
		if totalBlocksSubmitted%logInterval == 0 {
			intervalDuration := time.Since(lastLogTime)
			blocksPerSecond := logInterval / intervalDuration.Seconds()
			log.Infof("It took %s to submit %d blocks (%f blocks/sec) while %s of it it waited for RPC response"+
				" (total blocks sent %d)", intervalDuration, logInterval, blocksPerSecond, rpcWaitInInterval,
				totalBlocksSubmitted)
			rpcWaitInInterval = 0
			lastLogTime = time.Now()
		}

		blockHash := consensushashing.BlockHash(block)
		log.Tracef("Submitted block %s with hash %s", blockData.ID, blockHash)
	}
	return nil
}

func mineOrFetchBlock(blockData JSONBlock, mdb *miningDB, testConsensus testapi.TestConsensus) (*externalapi.DomainBlock, error) {
	hash := mdb.hashByID(blockData.ID)
	if mdb.hashByID(blockData.ID) != nil {
		return testConsensus.GetBlock(hash)
	}

	parentHashes := make([]*externalapi.DomainHash, len(blockData.Parents))
	for i, parentID := range blockData.Parents {
		parentHashes[i] = mdb.hashByID(parentID)
	}
	block, _, err := testConsensus.BuildBlockWithParents(parentHashes,
		&externalapi.DomainCoinbaseData{ScriptPublicKey: &externalapi.ScriptPublicKey{}}, []*externalapi.DomainTransaction{})
	if err != nil {
		return nil, errors.Wrap(err, "error in BuildBlockWithParents")
	}

	if !testConsensus.DAGParams().SkipProofOfWork {
		SolveBlock(block)
	}

	_, err = testConsensus.ValidateAndInsertBlock(block, true)
	if err != nil {
		return nil, errors.Wrap(err, "error in ValidateAndInsertBlock")
	}

	blockHash := consensushashing.BlockHash(block)
	err = mdb.putID(blockData.ID, blockHash)
	if err != nil {
		return nil, err
	}

	err = mdb.updateLastMinedBlock(blockData.ID)
	if err != nil {
		return nil, err
	}

	return block, nil
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// SolveBlock increments the given block's nonce until it matches the difficulty requirements in its bits field
func SolveBlock(block *externalapi.DomainBlock) {
	mining.SolveBlock(block, random)
}
