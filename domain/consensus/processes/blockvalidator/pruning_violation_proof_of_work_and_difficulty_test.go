package blockvalidator_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/pow"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/domain/consensus/utils/mining"
	"github.com/kaspanet/kaspad/util/difficulty"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

// TestPOW tests the validation of the block's POW.
func TestPOW(t *testing.T) {
	// We set the flag "skip pow" to be false (second argument in the function) for not skipping the check of POW and validate its correctness.
	testutils.ForAllNets(t, false, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestPOW")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Builds and checks block with invalid POW.
		invalidBlockWrongPOW, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		invalidBlockWrongPOW = solveBlockWithWrongPOW(invalidBlockWrongPOW)
		_, err = tc.ValidateAndInsertBlock(invalidBlockWrongPOW)
		if !errors.Is(err, ruleerrors.ErrInvalidPoW) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrInvalidPoW, err)
		}

		abovePowMaxBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		abovePowMaxTarget := big.NewInt(0).Add(big.NewInt(1), params.PowMax)
		abovePowMaxBlock.Header = blockheader.NewImmutableBlockHeader(
			abovePowMaxBlock.Header.Version(),
			abovePowMaxBlock.Header.ParentHashes(),
			abovePowMaxBlock.Header.HashMerkleRoot(),
			abovePowMaxBlock.Header.AcceptedIDMerkleRoot(),
			abovePowMaxBlock.Header.UTXOCommitment(),
			abovePowMaxBlock.Header.TimeInMilliseconds(),
			difficulty.BigToCompact(abovePowMaxTarget),
			abovePowMaxBlock.Header.Nonce(),
		)

		_, err = tc.ValidateAndInsertBlock(abovePowMaxBlock)
		if !errors.Is(err, ruleerrors.ErrTargetTooHigh) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		negativeTargetBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		negativeTargetBlock.Header = blockheader.NewImmutableBlockHeader(
			negativeTargetBlock.Header.Version(),
			negativeTargetBlock.Header.ParentHashes(),
			negativeTargetBlock.Header.HashMerkleRoot(),
			negativeTargetBlock.Header.AcceptedIDMerkleRoot(),
			negativeTargetBlock.Header.UTXOCommitment(),
			negativeTargetBlock.Header.TimeInMilliseconds(),
			0x00800000,
			negativeTargetBlock.Header.Nonce(),
		)

		_, err = tc.ValidateAndInsertBlock(negativeTargetBlock)
		if !errors.Is(err, ruleerrors.ErrNegativeTarget) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		// test on a valid block.
		validBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		random := rand.New(rand.NewSource(0))
		mining.SolveBlock(validBlock, random)
		_, err = tc.ValidateAndInsertBlock(validBlock)
		if err != nil {
			t.Fatal(err)
		}
	})
}

// solveBlockWithWrongPOW increments the given block's nonce until it gets wrong POW (for test!).
func solveBlockWithWrongPOW(block *externalapi.DomainBlock) *externalapi.DomainBlock {
	targetDifficulty := difficulty.CompactToBig(block.Header.Bits())
	headerForMining := block.Header.ToMutable()
	initialNonce := uint64(0)
	for i := initialNonce; i < math.MaxUint64; i++ {
		headerForMining.SetNonce(i)
		if !pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
			block.Header = headerForMining.ToImmutable()
			return block
		}
	}

	panic("Failed to solve block! cannot find a invalid POW for the test")
}

func TestCheckParentHeadersExist(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(params, false, "TestCheckParentHeadersExist")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		orphanBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		parentHash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{}) // Non existing parent hash
		orphanBlock.Header = blockheader.NewImmutableBlockHeader(
			orphanBlock.Header.Version(),
			[]*externalapi.DomainHash{
				parentHash,
			},
			orphanBlock.Header.HashMerkleRoot(),
			orphanBlock.Header.AcceptedIDMerkleRoot(),
			orphanBlock.Header.UTXOCommitment(),
			orphanBlock.Header.TimeInMilliseconds(),
			orphanBlock.Header.Bits(),
			orphanBlock.Header.Nonce(),
		)

		_, err = tc.ValidateAndInsertBlock(orphanBlock)
		errMissingParents := &ruleerrors.ErrMissingParents{}
		if !errors.As(err, errMissingParents) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		if !externalapi.HashesEqual(errMissingParents.MissingParentHashes, []*externalapi.DomainHash{parentHash}) {
			t.Fatalf("unexpected missing parents %s", errMissingParents.MissingParentHashes)
		}

		invalidBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		invalidBlock.Transactions[0].Version = constants.MaxTransactionVersion + 1 // This should invalidate the block
		invalidBlock.Header = blockheader.NewImmutableBlockHeader(
			invalidBlock.Header.Version(),
			invalidBlock.Header.ParentHashes(),
			merkle.CalculateHashMerkleRoot(invalidBlock.Transactions),
			orphanBlock.Header.AcceptedIDMerkleRoot(),
			orphanBlock.Header.UTXOCommitment(),
			orphanBlock.Header.TimeInMilliseconds(),
			orphanBlock.Header.Bits(),
			orphanBlock.Header.Nonce(),
		)

		_, err = tc.ValidateAndInsertBlock(invalidBlock)
		if !errors.Is(err, ruleerrors.ErrTransactionVersionIsUnknown) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		invalidBlockHash := consensushashing.BlockHash(invalidBlock)

		invalidBlockChild, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		invalidBlockChild.Header = blockheader.NewImmutableBlockHeader(
			invalidBlockChild.Header.Version(),
			[]*externalapi.DomainHash{invalidBlockHash},
			invalidBlockChild.Header.HashMerkleRoot(),
			invalidBlockChild.Header.AcceptedIDMerkleRoot(),
			invalidBlockChild.Header.UTXOCommitment(),
			invalidBlockChild.Header.TimeInMilliseconds(),
			invalidBlockChild.Header.Bits(),
			invalidBlockChild.Header.Nonce(),
		)

		_, err = tc.ValidateAndInsertBlock(invalidBlockChild)
		if !errors.Is(err, ruleerrors.ErrInvalidAncestorBlock) {
			t.Fatalf("Unexpected error: %+v", err)
		}
	})
}

func TestCheckPruningPointViolation(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// This is done to reduce the pruning depth to 6 blocks
		params.FinalityDuration = 2 * params.TargetTimePerBlock
		params.K = 0

		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(params, false, "TestCheckPruningPointViolation")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Add blocks until the pruning point changes
		tipHash := params.GenesisHash
		for {
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			pruningPoint, err := tc.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(params.GenesisHash) {
				break
			}
		}

		_, _, err = tc.AddUTXOInvalidBlock([]*externalapi.DomainHash{params.GenesisHash})
		if !errors.Is(err, ruleerrors.ErrPruningPointViolation) {
			t.Fatalf("Unexpected error: %+v", err)
		}
	})
}

// TestValidateDifficulty verifies that in case of a block with an unexpected difficulty,
// an appropriate error message (ErrUnexpectedDifficulty) will be returned on the
// function ValidatePruningPointViolationAndProofOfWorkAndDifficulty. The required difficulty is
// "calculated" by the function (dm *mocDifficultyManager) UpdateDAADataAndReturnDifficultyBits ,
// where mocDifficultyManager is special implementation of the type DifficultyManager for this test (defined below).
func TestValidateDifficulty(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		factory := consensus.NewFactory()
		mocDifficulty := &mocDifficultyManager{}
		factory.SetTestDifficultyManager(func(_ model.DBReader, _ model.GHOSTDAGManager, _ model.GHOSTDAGDataStore,
			_ model.BlockHeaderStore, daaBlocksStore model.DAABlocksStore, _ model.DAGTopologyManager,
			_ model.DAGTraversalManager, _ *big.Int, _ int, _ bool, _ time.Duration,
			_ *externalapi.DomainHash) model.DifficultyManager {

			mocDifficulty.daaBlocksStore = daaBlocksStore
			return mocDifficulty
		})
		genesisDifficulty := params.GenesisBlock.Header.Bits()
		mocDifficulty.testDifficulty = genesisDifficulty
		mocDifficulty.testGenesisBits = genesisDifficulty
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestValidateDifficulty")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		emptyCoinbase := externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		}
		block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, &emptyCoinbase, nil)
		if err != nil {
			t.Fatalf("TestValidateDifficulty: Failed build block with parents: %v.", err)
		}
		blockHash := consensushashing.BlockHash(block)
		tc.BlockStore().Stage(blockHash, block)
		tc.BlockHeaderStore().Stage(blockHash, block.Header)
		wrongTestDifficulty := mocDifficulty.testDifficulty + uint32(5)
		mocDifficulty.testDifficulty = wrongTestDifficulty

		err = tc.BlockValidator().ValidatePruningPointViolationAndProofOfWorkAndDifficulty(blockHash)
		if err == nil || !errors.Is(err, ruleerrors.ErrUnexpectedDifficulty) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrUnexpectedDifficulty, err)
		}
	})
}

type mocDifficultyManager struct {
	testDifficulty  uint32
	testGenesisBits uint32
	daaBlocksStore  model.DAABlocksStore
}

// UpdateDAADataAndReturnDifficultyBits returns the difficulty required for the test
func (dm *mocDifficultyManager) UpdateDAADataAndReturnDifficultyBits(blockHash *externalapi.DomainHash) (uint32, error) {
	// Populate daaBlocksStore with fake values
	dm.daaBlocksStore.StageDAAScore(blockHash, 0)
	dm.daaBlocksStore.StageBlockDAAAddedBlocks(blockHash, nil)

	return dm.testDifficulty, nil
}
