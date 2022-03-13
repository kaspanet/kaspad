package fast_pruning_ibd_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

// TestGenerateFastPruningIBDTest generates the json needed for dag-for-fast-pruning-ibd-test.json.gz
func TestGenerateFastPruningIBDTest(t *testing.T) {
	t.Skip()
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		if consensusConfig.Name != dagconfig.DevnetParams.Name {
			return
		}

		factory := consensus.NewFactory()

		// This is done to reduce the pruning depth to 6 blocks
		finalityDepth := 200
		consensusConfig.FinalityDuration = time.Duration(finalityDepth) * consensusConfig.TargetTimePerBlock
		consensusConfig.K = 0
		consensusConfig.PruningProofM = 1
		consensusConfig.MergeSetSizeLimit = 30

		tc, teardownSyncer, err := factory.NewTestConsensus(consensusConfig, "TestValidateAndInsertPruningPointSyncer")
		if err != nil {
			t.Fatalf("Error setting up tc: %+v", err)
		}
		defer teardownSyncer(false)

		numBlocks := finalityDepth
		tipHash := consensusConfig.GenesisHash
		for i := 0; i < numBlocks; i++ {
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
		}

		tip, err := tc.GetBlock(tipHash)
		if err != nil {
			t.Fatal(err)
		}

		header := tip.Header.ToMutable()

		for i := uint64(1); i < 1000; i++ {
			if i%100 == 0 {
				t.Logf("Added %d tips", i)
			}
			header.SetNonce(tip.Header.Nonce() + i)
			block := &externalapi.DomainBlock{Header: header.ToImmutable(), Transactions: tip.Transactions}
			_, err = tc.ValidateAndInsertBlock(block, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		emptyCoinbase := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		}

		pruningPoint, err := tc.PruningPoint()
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; ; i++ {
			currentPruningPoint, err := tc.PruningPoint()
			if err != nil {
				t.Fatal(err)
			}

			if !pruningPoint.Equal(currentPruningPoint) {
				t.Fatalf("Pruning point unexpectedly changed")
			}

			tips, err := tc.Tips()
			if err != nil {
				t.Fatal(err)
			}

			if len(tips) == 1 {
				break
			}

			if i%10 == 0 {
				t.Logf("Number of tips: %d", len(tips))
			}

			block, err := tc.BuildBlock(emptyCoinbase, nil)
			if err != nil {
				t.Fatal(err)
			}

			_, err = tc.ValidateAndInsertBlock(block, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		for {
			currentPruningPoint, err := tc.PruningPoint()
			if err != nil {
				t.Fatal(err)
			}

			if !pruningPoint.Equal(currentPruningPoint) {
				break
			}

			block, err := tc.BuildBlock(emptyCoinbase, nil)
			if err != nil {
				t.Fatal(err)
			}

			_, err = tc.ValidateAndInsertBlock(block, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		file, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}

		err = tc.ToJSON(file)
		if err != nil {
			t.Fatal(err)
		}
		stat, err := file.Stat()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("DAG saved at %s", path.Join(os.TempDir(), stat.Name()))
	})
}
