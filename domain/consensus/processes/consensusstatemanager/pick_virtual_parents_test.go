package consensusstatemanager_test

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"io/ioutil"
	"os/user"
	"path"
	"runtime/pprof"
	"testing"
	"time"
)
var log, _ = logger.Get(logger.SubsystemTags.CMGR)

func TestPickVirtualParents(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	const chainSize = 97

	params := dagconfig.DevnetParams
	params.SkipProofOfWork = true

	factory := consensus.NewFactory()

	var chains [10][]*externalapi.DomainBlock
	// Build three chains over the genesis
	for chainIndex := range chains {
		func() {
			tipHash := params.GenesisHash
			builder, teardown, err := factory.NewTestConsensus(&params, false, fmt.Sprintf("TestPickVirtualParents: %d", chainIndex))
			if err != nil {
				t.Fatalf("Error setting up consensus: %+v", err)
			}
			defer teardown(false)
			for blockIndex := 0; blockIndex < chainSize; blockIndex++ {
				scriptPubKey, _ := testutils.OpTrueScript()
				extraData := []byte{byte(chainIndex)}
				block, _, err := builder.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, &externalapi.DomainCoinbaseData{scriptPubKey, extraData}, nil)
				if err != nil {
					t.Fatalf("Could not build block: %s", err)
				}
				_, err = builder.ValidateAndInsertBlock(block)
				if err != nil {
					t.Fatalf("Could not build block: %s", err)
				}
				chains[chainIndex] = append(chains[chainIndex], block)
				tipHash = consensushashing.BlockHash(block)
			}
			fmt.Printf("Finished Building chain: %d\n", chainIndex)
		}()
	}


	testConsensus, teardown, err := factory.NewTestConsensus(&params, false, "TestPickVirtualParents")
	if err != nil {
		t.Fatalf("Error setting up consensus: %+v", err)
	}
	defer teardown(false)

	var maxTime time.Duration
	var maxString string
	var profName string
	maxProf := make([]byte, 0, 1024)
	// Build three chains over the genesis
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	for chainIndex, chain := range chains {
		accumulatedValidationTime := time.Duration(0)
		for blockIndex, block := range chain {
			if chainIndex == 9 && blockIndex > 90 {
				logger.InitLog(path.Join(usr.HomeDir, "TestPickVirtualParents.log"), path.Join(usr.HomeDir, "TestPickVirtualParents_err.log"))
				logger.SetLogLevels("debug")
			}
			log.Debugf("Starting chain:#%d, block: #%d", chainIndex, blockIndex)
			blockHash := consensushashing.BlockHash(block)
			buf.Reset()
			err = pprof.StartCPUProfile(buf)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			_, err := testConsensus.ValidateAndInsertBlock(block)
			validationTime := time.Since(start)
			pprof.StopCPUProfile()
			if err != nil {
				t.Fatalf("Failed to validate block %s: %s", blockHash, err)
			}
			if validationTime > maxTime {
				maxTime = validationTime
				maxString = fmt.Sprintf("Chain: %d, Block: %d", chainIndex, blockIndex)
				profName = fmt.Sprintf("TestPickVirtualParents-chain-%d-block-%d.pprof", chainIndex, blockIndex)
				maxProf = append(maxProf[:0], buf.Bytes()...)
			}

			accumulatedValidationTime += validationTime
			log.Debugf("Validated block #%d in chain #%d, took %s\n", blockIndex, chainIndex, validationTime)

		}

		averageValidationTime := accumulatedValidationTime / chainSize
		fmt.Printf("Average validation time for chain #%d: %s\n", chainIndex, averageValidationTime)
	}

	err = ioutil.WriteFile(path.Join(usr.HomeDir, profName), maxProf, 0644)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s, took: %s\n", maxString, maxTime)
}
