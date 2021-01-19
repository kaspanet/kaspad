package blockprocessor_test

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxoserialization"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func addBlock(tcSyncer, tcSyncee testapi.TestConsensus, parentHashes []*externalapi.DomainHash, t *testing.T) *externalapi.DomainHash {
	block, _, err := tcSyncer.BuildBlockWithParents(parentHashes, nil, nil)
	if err != nil {
		t.Fatalf("BuildBlockWithParents: %+v", err)
	}

	_, err = tcSyncer.ValidateAndInsertBlock(block)
	if err != nil {
		t.Fatalf("ValidateAndInsertBlock: %+v", err)
	}

	_, err = tcSyncee.ValidateAndInsertBlock(&externalapi.DomainBlock{
		Header:       block.Header,
		Transactions: nil,
	})
	if err != nil {
		t.Fatalf("ValidateAndInsertBlock: %+v", err)
	}

	return consensushashing.BlockHash(block)
}

func TestValidateAndInsertPruningPoint(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// This is done to reduce the pruning depth to 6 blocks
		finalityDepth := 3
		params.FinalityDuration = time.Duration(finalityDepth) * params.TargetTimePerBlock
		params.K = 0

		factory := consensus.NewFactory()

		tcSyncer, teardownSyncer, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertPruningPointSyncer")
		if err != nil {
			t.Fatalf("Error setting up tcSyncer: %+v", err)
		}
		defer teardownSyncer(false)

		tcSyncee, teardownSyncee, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertPruningPointSyncee")
		if err != nil {
			t.Fatalf("Error setting up tcSyncee: %+v", err)
		}
		defer teardownSyncee(false)

		tipHash := params.GenesisHash
		for i := 0; i < finalityDepth-2; i++ {
			tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		}

		// Add block in the anticone of the pruning point to test such situation
		pruningPointAnticoneBlock := addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		nextPruningPoint := addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)

		tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{pruningPointAnticoneBlock, nextPruningPoint}, t)

		// Add blocks until the pruning point changes
		for {
			tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)

			pruningPoint, err := tcSyncer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(params.GenesisHash) {
				break
			}
		}

		pruningPoint, err := tcSyncer.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !pruningPoint.Equal(nextPruningPoint) {
			t.Fatalf("Unexpected pruning point %s", pruningPoint)
		}

		pruningPointUTXOSet, err := tcSyncer.GetPruningPointUTXOSet(pruningPoint)
		if err != nil {
			t.Fatalf("GetPruningPointUTXOSet: %+v", err)
		}

		tip, err := tcSyncer.GetBlock(tipHash)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		// Check that ValidateAndInsertPruningPoint fails for invalid pruning point
		err = tcSyncee.ValidateAndInsertPruningPoint(tip, pruningPointUTXOSet)
		if !errors.Is(err, ruleerrors.ErrUnexpectedPruningPoint) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		pruningPointBlock, err := tcSyncer.GetBlock(pruningPoint)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		invalidPruningPointBlock := pruningPointBlock.Clone()
		invalidPruningPointBlock.Transactions[0].Version += 1

		// Check that ValidateAndInsertPruningPoint fails for invalid block
		err = tcSyncee.ValidateAndInsertPruningPoint(invalidPruningPointBlock, pruningPointUTXOSet)
		if !errors.Is(err, ruleerrors.ErrBadMerkleRoot) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		serializedFakeUTXOSet, err := makeSerializedFakeUTXOSet()
		if err != nil {
			t.Fatalf("makeSerializedFakeUTXOSet: %+v", err)
		}

		// Check that ValidateAndInsertPruningPoint fails if the UTXO commitment doesn't fit the provided UTXO set.
		err = tcSyncee.ValidateAndInsertPruningPoint(pruningPointBlock, serializedFakeUTXOSet)
		if !errors.Is(err, ruleerrors.ErrBadPruningPointUTXOSet) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		// Check that ValidateAndInsertPruningPoint works given the right arguments.
		err = tcSyncee.ValidateAndInsertPruningPoint(pruningPointBlock, pruningPointUTXOSet)
		if err != nil {
			t.Fatalf("ValidateAndInsertPruningPoint: %+v", err)
		}

		virtualSelectedParent, err := tcSyncer.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("GetVirtualSelectedParent: %+v", err)
		}

		missingBlockBodyHashes, err := tcSyncee.GetMissingBlockBodyHashes(virtualSelectedParent)
		if err != nil {
			t.Fatalf("GetMissingBlockBodyHashes: %+v", err)
		}

		for _, missingHash := range missingBlockBodyHashes {
			block, err := tcSyncer.GetBlock(missingHash)
			if err != nil {
				t.Fatalf("GetBlock: %+v", err)
			}

			_, err = tcSyncee.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		synceeTips, err := tcSyncee.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		syncerTips, err := tcSyncer.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		if !externalapi.HashesEqual(synceeTips, syncerTips) {
			t.Fatalf("Syncee's tips are %s while syncer's are %s", synceeTips, syncerTips)
		}

		synceePruningPoint, err := tcSyncee.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !synceePruningPoint.Equal(pruningPoint) {
			t.Fatalf("The syncee pruning point has not changed as exepcted")
		}
	})
}

// TestValidateAndInsertPruningPointWithSideBlocks makes sure that when a node applies a UTXO-Set downloaded during
// IBD, while it already has a non-empty UTXO-Set originating from blocks mined on top of genesis - the resulting
// UTXO set is correct
func TestValidateAndInsertPruningPointWithSideBlocks(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// This is done to reduce the pruning depth to 6 blocks
		finalityDepth := 3
		params.FinalityDuration = time.Duration(finalityDepth) * params.TargetTimePerBlock
		params.K = 0

		factory := consensus.NewFactory()

		tcSyncer, teardownSyncer, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertPruningPointSyncer")
		if err != nil {
			t.Fatalf("Error setting up tcSyncer: %+v", err)
		}
		defer teardownSyncer(false)

		tcSyncee, teardownSyncee, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertPruningPointSyncee")
		if err != nil {
			t.Fatalf("Error setting up tcSyncee: %+v", err)
		}
		defer teardownSyncee(false)

		// Mine 2 block in the syncee on top of genesis
		side, _, err := tcSyncee.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, &externalapi.DomainCoinbaseData{ScriptPublicKey: &externalapi.ScriptPublicKey{}, ExtraData: []byte{1, 2}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = tcSyncee.AddBlock([]*externalapi.DomainHash{side}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		tipHash := params.GenesisHash
		for i := 0; i < finalityDepth-2; i++ {
			tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		}

		tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		nextPruningPoint := addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)

		tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{nextPruningPoint}, t)

		// Add blocks until the pruning point changes
		for {
			tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)

			pruningPoint, err := tcSyncer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(params.GenesisHash) {
				break
			}
		}

		pruningPoint, err := tcSyncer.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !pruningPoint.Equal(nextPruningPoint) {
			t.Fatalf("Unexpected pruning point %s", pruningPoint)
		}

		pruningPointUTXOSet, err := tcSyncer.GetPruningPointUTXOSet(pruningPoint)
		if err != nil {
			t.Fatalf("GetPruningPointUTXOSet: %+v", err)
		}

		// Check that ValidateAndInsertPruningPoint works.
		pruningPointBlock, err := tcSyncer.GetBlock(pruningPoint)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}
		err = tcSyncee.ValidateAndInsertPruningPoint(pruningPointBlock, pruningPointUTXOSet)
		if err != nil {
			t.Fatalf("ValidateAndInsertPruningPoint: %+v", err)
		}

		// Insert the rest of the blocks atop pruning point
		virtualSelectedParent, err := tcSyncer.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("GetVirtualSelectedParent: %+v", err)
		}

		missingBlockBodyHashes, err := tcSyncee.GetMissingBlockBodyHashes(virtualSelectedParent)
		if err != nil {
			t.Fatalf("GetMissingBlockBodyHashes: %+v", err)
		}

		for _, missingHash := range missingBlockBodyHashes {
			block, err := tcSyncer.GetBlock(missingHash)
			if err != nil {
				t.Fatalf("GetBlock: %+v", err)
			}

			_, err = tcSyncee.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		// Verify that syncee and syncer tips are equal
		synceeTips, err := tcSyncee.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		syncerTips, err := tcSyncer.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		if !externalapi.HashesEqual(synceeTips, syncerTips) {
			t.Fatalf("Syncee's tips are %s while syncer's are %s", synceeTips, syncerTips)
		}

		// Verify that syncee and syncer pruning points are equal
		synceePruningPoint, err := tcSyncee.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !synceePruningPoint.Equal(pruningPoint) {
			t.Fatalf("The syncee pruning point has not changed as exepcted")
		}

		pruningPointOld := pruningPoint

		// Add blocks until the pruning point moves, and verify it moved to the same point on both syncer and syncee
		for {
			block, _, err := tcSyncer.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("BuildBlockWithParents: %+v", err)
			}

			_, err = tcSyncer.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}

			_, err = tcSyncee.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}

			tipHash = consensushashing.BlockHash(block)

			pruningPoint, err = tcSyncer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(pruningPointOld) {
				break
			}
		}

		synceePruningPoint, err = tcSyncee.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !synceePruningPoint.Equal(pruningPoint) {
			t.Fatalf("The syncee pruning point(%s) is not equal to syncer pruning point (%s) after it moved. "+
				"pruning point before move: %s", synceePruningPoint, pruningPoint, pruningPointOld)
		}
	})
}

type fakeUTXOSetIterator struct {
	nextCalled bool
}

func (f *fakeUTXOSetIterator) Next() bool {
	if f.nextCalled {
		return false
	}
	f.nextCalled = true
	return true
}

func (f *fakeUTXOSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	return &externalapi.DomainOutpoint{
			TransactionID: *externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{0x01}),
			Index:         0,
		}, utxo.NewUTXOEntry(1000, &externalapi.ScriptPublicKey{
			Script:  []byte{1, 2, 3},
			Version: 0,
		}, false, 2000), nil
}

func makeSerializedFakeUTXOSet() ([]byte, error) {
	serializedUtxo, err := utxoserialization.ReadOnlyUTXOSetToProtoUTXOSet(&fakeUTXOSetIterator{})
	if err != nil {
		return nil, err
	}

	return proto.Marshal(serializedUtxo)
}
