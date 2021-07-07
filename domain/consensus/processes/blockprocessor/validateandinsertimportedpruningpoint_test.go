package blockprocessor_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"math"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/pkg/errors"
)

func addBlock(tc testapi.TestConsensus, parentHashes []*externalapi.DomainHash, t *testing.T) *externalapi.DomainHash {
	block, _, err := tc.BuildBlockWithParents(parentHashes, nil, nil)
	if err != nil {
		t.Fatalf("BuildBlockWithParents: %+v", err)
	}

	blockHash := consensushashing.BlockHash(block)

	_, err = tc.ValidateAndInsertBlock(block, true)
	if err != nil {
		t.Fatalf("ValidateAndInsertBlock: %+v", err)
	}

	return blockHash
}

func TestValidateAndInsertImportedPruningPoint(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		// This is done to reduce the pruning depth to 6 blocks
		finalityDepth := 3
		consensusConfig.FinalityDuration = time.Duration(finalityDepth) * consensusConfig.TargetTimePerBlock
		consensusConfig.K = 0

		synceeConfig := *consensusConfig
		synceeConfig.ShouldNotAddGenesis = true

		factory := consensus.NewFactory()

		tcSyncer, teardownSyncer, err := factory.NewTestConsensus(consensusConfig, "TestValidateAndInsertPruningPointSyncer")
		if err != nil {
			t.Fatalf("Error setting up tcSyncer: %+v", err)
		}
		defer teardownSyncer(false)

		tcSyncee, teardownSyncee, err := factory.NewTestConsensus(&synceeConfig, "TestValidateAndInsertPruningPointSyncee")
		if err != nil {
			t.Fatalf("Error setting up tcSyncee: %+v", err)
		}
		defer teardownSyncee(false)

		tipHash := consensusConfig.GenesisHash
		for i := 0; i < finalityDepth-2; i++ {
			tipHash = addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)
		}

		// Add block in the anticone of the pruning point to test such situation
		pruningPointAnticoneBlock := addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)
		tipHash = addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)
		nextPruningPoint := addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)

		tipHash = addBlock(tcSyncer, []*externalapi.DomainHash{pruningPointAnticoneBlock, nextPruningPoint}, t)

		// Add blocks until the pruning point changes
		for {
			tipHash = addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)

			pruningPoint, err := tcSyncer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(consensusConfig.GenesisHash) {
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

		pointAndItsAnticoneWithMetaData, err := tcSyncer.PruningPointAndItsAnticoneWithMetaData()
		if err != nil {
			t.Fatalf("PruningPointAndItsAnticoneWithMetaData: %+v", err)
		}

		for _, blockWithMetaData := range pointAndItsAnticoneWithMetaData {
			_, err := tcSyncee.ValidateAndInsertBlockWithMetaData(blockWithMetaData, false)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlockWithMetaData: %+v", err)
			}
		}

		syncerVirtualSelectedParent, err := tcSyncer.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("GetVirtualSelectedParent: %+v", err)
		}

		missingBlocksHashes, _, err := tcSyncer.GetHashesBetween(pruningPoint, syncerVirtualSelectedParent, math.MaxUint64)
		if err != nil {
			t.Fatalf("GetHashesBetween: %+v", err)
		}

		for _, blocksHash := range missingBlocksHashes {
			blockInfo, err := tcSyncee.GetBlockInfo(blocksHash)
			if err != nil {
				t.Fatalf("GetBlockInfo: %+v", err)
			}

			if blockInfo.Exists {
				continue
			}

			block, err := tcSyncer.GetBlock(blocksHash)
			if err != nil {
				t.Fatalf("GetBlock: %+v", err)
			}

			_, err = tcSyncee.ValidateAndInsertBlock(block, false)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		pruningPointUTXOs, err := tcSyncer.GetPruningPointUTXOs(pruningPoint, nil, 1000)
		if err != nil {
			t.Fatalf("GetPruningPointUTXOs: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(pruningPointUTXOs)
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint fails for invalid pruning point
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(tipHash)
		if !errors.Is(err, ruleerrors.ErrUnexpectedPruningPoint) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		err = tcSyncee.ClearImportedPruningPointData()
		if err != nil {
			t.Fatalf("ClearImportedPruningPointData: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(makeFakeUTXOs())
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint fails if the UTXO commitment doesn't fit the provided UTXO set.
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(pruningPoint)
		if !errors.Is(err, ruleerrors.ErrBadPruningPointUTXOSet) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		err = tcSyncee.ClearImportedPruningPointData()
		if err != nil {
			t.Fatalf("ClearImportedPruningPointData: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(pruningPointUTXOs)
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint works given the right arguments.
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(pruningPoint)
		if err != nil {
			t.Fatalf("ValidateAndInsertImportedPruningPoint: %+v", err)
		}

		err = tcSyncee.ResolveVirtual()
		if err != nil {
			t.Fatalf("ResolveVirtual: %+v", err)
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
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		// This is done to reduce the pruning depth to 6 blocks
		finalityDepth := 3
		consensusConfig.FinalityDuration = time.Duration(finalityDepth) * consensusConfig.TargetTimePerBlock
		consensusConfig.K = 0

		synceeConfig := *consensusConfig
		synceeConfig.ShouldNotAddGenesis = true

		factory := consensus.NewFactory()

		tcSyncer, teardownSyncer, err := factory.NewTestConsensus(consensusConfig, "TestValidateAndInsertPruningPointSyncer")
		if err != nil {
			t.Fatalf("Error setting up tcSyncer: %+v", err)
		}
		defer teardownSyncer(false)

		tcSyncee, teardownSyncee, err := factory.NewTestConsensus(&synceeConfig, "TestValidateAndInsertPruningPointSyncee")
		if err != nil {
			t.Fatalf("Error setting up tcSyncee: %+v", err)
		}
		defer teardownSyncee(false)

		// Mine two blocks on syncee on top of genesis
		synceeOnlyBlock := addBlock(tcSyncer, []*externalapi.DomainHash{consensusConfig.GenesisHash}, t)
		addBlock(tcSyncer, []*externalapi.DomainHash{synceeOnlyBlock}, t)

		tipHash := consensusConfig.GenesisHash
		for i := 0; i < finalityDepth-2; i++ {
			tipHash = addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)
		}

		// Add block in the anticone of the pruning point to test such situation
		pruningPointAnticoneBlock := addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)
		tipHash = addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)
		nextPruningPoint := addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)

		tipHash = addBlock(tcSyncer, []*externalapi.DomainHash{pruningPointAnticoneBlock, nextPruningPoint}, t)

		// Add blocks until the pruning point changes
		for {
			tipHash = addBlock(tcSyncer, []*externalapi.DomainHash{tipHash}, t)

			pruningPoint, err := tcSyncer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(consensusConfig.GenesisHash) {
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

		pointAndItsAnticoneWithMetaData, err := tcSyncer.PruningPointAndItsAnticoneWithMetaData()
		if err != nil {
			t.Fatalf("PruningPointAndItsAnticoneWithMetaData: %+v", err)
		}

		for _, blockWithMetaData := range pointAndItsAnticoneWithMetaData {
			_, err := tcSyncee.ValidateAndInsertBlockWithMetaData(blockWithMetaData, false)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlockWithMetaData: %+v", err)
			}
		}

		syncerVirtualSelectedParent, err := tcSyncer.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("GetVirtualSelectedParent: %+v", err)
		}

		missingBlocksHashes, _, err := tcSyncer.GetHashesBetween(pruningPoint, syncerVirtualSelectedParent, math.MaxUint64)
		if err != nil {
			t.Fatalf("GetHashesBetween: %+v", err)
		}

		for _, blocksHash := range missingBlocksHashes {
			blockInfo, err := tcSyncee.GetBlockInfo(blocksHash)
			if err != nil {
				t.Fatalf("GetBlockInfo: %+v", err)
			}

			if blockInfo.Exists {
				continue
			}

			block, err := tcSyncer.GetBlock(blocksHash)
			if err != nil {
				t.Fatalf("GetBlock: %+v", err)
			}

			_, err = tcSyncee.ValidateAndInsertBlock(block, false)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		pruningPointUTXOs, err := tcSyncer.GetPruningPointUTXOs(pruningPoint, nil, 1000)
		if err != nil {
			t.Fatalf("GetPruningPointUTXOs: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(pruningPointUTXOs)
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint fails for invalid pruning point
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(tipHash)
		if !errors.Is(err, ruleerrors.ErrUnexpectedPruningPoint) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		err = tcSyncee.ClearImportedPruningPointData()
		if err != nil {
			t.Fatalf("ClearImportedPruningPointData: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(makeFakeUTXOs())
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint fails if the UTXO commitment doesn't fit the provided UTXO set.
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(pruningPoint)
		if !errors.Is(err, ruleerrors.ErrBadPruningPointUTXOSet) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		err = tcSyncee.ClearImportedPruningPointData()
		if err != nil {
			t.Fatalf("ClearImportedPruningPointData: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(pruningPointUTXOs)
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint works given the right arguments.
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(pruningPoint)
		if err != nil {
			t.Fatalf("ValidateAndInsertImportedPruningPoint: %+v", err)
		}

		err = tcSyncee.ResolveVirtual()
		if err != nil {
			t.Fatalf("ResolveVirtual: %+v", err)
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

func makeFakeUTXOs() []*externalapi.OutpointAndUTXOEntryPair {
	return []*externalapi.OutpointAndUTXOEntryPair{
		{
			Outpoint: &externalapi.DomainOutpoint{
				TransactionID: externalapi.DomainTransactionID{},
				Index:         0,
			},
			UTXOEntry: utxo.NewUTXOEntry(
				0,
				&externalapi.ScriptPublicKey{
					Script:  nil,
					Version: 0,
				},
				false,
				0,
			),
		},
		{
			Outpoint: &externalapi.DomainOutpoint{
				TransactionID: externalapi.DomainTransactionID{},
				Index:         1,
			},
			UTXOEntry: utxo.NewUTXOEntry(
				2,
				&externalapi.ScriptPublicKey{
					Script:  nil,
					Version: 0,
				},
				true,
				3,
			),
		},
	}
}

func TestGetPruningPointUTXOs(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		// This is done to reduce the pruning depth to 8 blocks
		finalityDepth := 4
		consensusConfig.FinalityDuration = time.Duration(finalityDepth) * consensusConfig.TargetTimePerBlock
		consensusConfig.K = 0

		consensusConfig.BlockCoinbaseMaturity = 0

		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestGetPruningPointUTXOs")
		if err != nil {
			t.Fatalf("Error setting up testConsensus: %+v", err)
		}
		defer teardown(false)

		// Create a block that accepts the genesis coinbase so that we won't have script problems down the line
		emptyCoinbase := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		}
		blockAboveGenesis, err := testConsensus.BuildBlock(emptyCoinbase, nil)
		if err != nil {
			t.Fatalf("Error building block above genesis: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(blockAboveGenesis, true)
		if err != nil {
			t.Fatalf("Error validating and inserting block above genesis: %+v", err)
		}

		// Create a block whose coinbase we could spend
		scriptPublicKey, redeemScript := testutils.OpTrueScript()
		coinbaseData := &externalapi.DomainCoinbaseData{ScriptPublicKey: scriptPublicKey}
		blockWithSpendableCoinbase, err := testConsensus.BuildBlock(coinbaseData, nil)
		if err != nil {
			t.Fatalf("Error building block with spendable coinbase: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(blockWithSpendableCoinbase, true)
		if err != nil {
			t.Fatalf("Error validating and inserting block with spendable coinbase: %+v", err)
		}

		// Create a transaction that adds a lot of UTXOs to the UTXO set
		transactionToSpend := blockWithSpendableCoinbase.Transactions[0]
		signatureScript, err := txscript.PayToScriptHashSignatureScript(redeemScript, nil)
		if err != nil {
			t.Fatalf("Error creating signature script: %+v", err)
		}
		input := &externalapi.DomainTransactionInput{
			PreviousOutpoint: externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(transactionToSpend),
				Index:         0,
			},
			SignatureScript: signatureScript,
			Sequence:        constants.MaxTxInSequenceNum,
		}

		outputs := make([]*externalapi.DomainTransactionOutput, 900)
		for i := 0; i < len(outputs); i++ {
			outputs[i] = &externalapi.DomainTransactionOutput{
				ScriptPublicKey: scriptPublicKey,
				Value:           10000,
			}
		}
		spendingTransaction := &externalapi.DomainTransaction{
			Version: constants.MaxTransactionVersion,
			Inputs:  []*externalapi.DomainTransactionInput{input},
			Outputs: outputs,
			Payload: []byte{},
		}

		// Create a block with that includes the above transaction
		includingBlock, err := testConsensus.BuildBlock(emptyCoinbase, []*externalapi.DomainTransaction{spendingTransaction})
		if err != nil {
			t.Fatalf("Error building including block: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(includingBlock, true)
		if err != nil {
			t.Fatalf("Error validating and inserting including block: %+v", err)
		}

		// Add enough blocks to move the pruning point
		for {
			block, err := testConsensus.BuildBlock(emptyCoinbase, nil)
			if err != nil {
				t.Fatalf("Error building block: %+v", err)
			}
			_, err = testConsensus.ValidateAndInsertBlock(block, true)
			if err != nil {
				t.Fatalf("Error validating and inserting block: %+v", err)
			}

			pruningPoint, err := testConsensus.PruningPoint()
			if err != nil {
				t.Fatalf("Error getting the pruning point: %+v", err)
			}
			if !pruningPoint.Equal(consensusConfig.GenesisHash) {
				break
			}
		}
		pruningPoint, err := testConsensus.PruningPoint()
		if err != nil {
			t.Fatalf("Error getting the pruning point: %+v", err)
		}

		pruningRelations, err := testConsensus.BlockRelationStore().BlockRelation(
			testConsensus.DatabaseContext(), model.NewStagingArea(), pruningPoint)
		if err != nil {
			t.Fatalf("BlockRelation(): %+v", err)
		}

		if len(pruningRelations.Parents) != 1 && pruningRelations.Parents[0] != consensushashing.BlockHash(includingBlock) {
			t.Fatalf("includingBlock should be pruning point's only parent")
		}

		// Get pruning point UTXOs in a loop
		var allOutpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair
		step := 100
		var fromOutpoint *externalapi.DomainOutpoint
		for {
			outpointAndUTXOEntryPairs, err := testConsensus.GetPruningPointUTXOs(pruningPoint, fromOutpoint, step)
			if err != nil {
				t.Fatalf("Error getting pruning point UTXOs: %+v", err)
			}
			allOutpointAndUTXOEntryPairs = append(allOutpointAndUTXOEntryPairs, outpointAndUTXOEntryPairs...)
			fromOutpoint = outpointAndUTXOEntryPairs[len(outpointAndUTXOEntryPairs)-1].Outpoint

			if len(outpointAndUTXOEntryPairs) < step {
				break
			}
		}

		// Make sure the length of the UTXOs is exactly spendingTransaction.Outputs + 1 coinbase
		// output (includingBlock's coinbase)
		if len(allOutpointAndUTXOEntryPairs) != len(outputs)+1 {
			t.Fatalf("Returned an unexpected amount of UTXOs. "+
				"Want: %d, got: %d", len(outputs)+2, len(allOutpointAndUTXOEntryPairs))
		}

		// Make sure all spendingTransaction.Outputs are in the returned UTXOs
		spendingTransactionID := consensushashing.TransactionID(spendingTransaction)
		for i := range outputs {
			found := false
			for _, outpointAndUTXOEntryPair := range allOutpointAndUTXOEntryPairs {
				outpoint := outpointAndUTXOEntryPair.Outpoint
				if outpoint.TransactionID == *spendingTransactionID && outpoint.Index == uint32(i) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Outpoint %s:%d not found amongst the returned UTXOs", spendingTransactionID, i)
			}
		}
	})
}

func BenchmarkGetPruningPointUTXOs(b *testing.B) {
	consensusConfig := consensus.Config{Params: dagconfig.DevnetParams}

	// This is done to reduce the pruning depth to 200 blocks
	finalityDepth := 100
	consensusConfig.FinalityDuration = time.Duration(finalityDepth) * consensusConfig.TargetTimePerBlock
	consensusConfig.K = 0

	consensusConfig.SkipProofOfWork = true
	consensusConfig.BlockCoinbaseMaturity = 0

	factory := consensus.NewFactory()
	testConsensus, teardown, err := factory.NewTestConsensus(&consensusConfig, "TestGetPruningPointUTXOs")
	if err != nil {
		b.Fatalf("Error setting up testConsensus: %+v", err)
	}
	defer teardown(false)

	// Create a block whose coinbase we could spend
	scriptPublicKey, redeemScript := testutils.OpTrueScript()
	coinbaseData := &externalapi.DomainCoinbaseData{ScriptPublicKey: scriptPublicKey}
	blockWithSpendableCoinbase, err := testConsensus.BuildBlock(coinbaseData, nil)
	if err != nil {
		b.Fatalf("Error building block with spendable coinbase: %+v", err)
	}
	_, err = testConsensus.ValidateAndInsertBlock(blockWithSpendableCoinbase, true)
	if err != nil {
		b.Fatalf("Error validating and inserting block with spendable coinbase: %+v", err)
	}

	addBlockWithLotsOfOutputs := func(b *testing.B, transactionToSpend *externalapi.DomainTransaction) *externalapi.DomainBlock {
		// Create a transaction that adds a lot of UTXOs to the UTXO set
		signatureScript, err := txscript.PayToScriptHashSignatureScript(redeemScript, nil)
		if err != nil {
			b.Fatalf("Error creating signature script: %+v", err)
		}
		input := &externalapi.DomainTransactionInput{
			PreviousOutpoint: externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(transactionToSpend),
				Index:         0,
			},
			SignatureScript: signatureScript,
			Sequence:        constants.MaxTxInSequenceNum,
		}
		outputs := make([]*externalapi.DomainTransactionOutput, 900)
		for i := 0; i < len(outputs); i++ {
			outputs[i] = &externalapi.DomainTransactionOutput{
				ScriptPublicKey: scriptPublicKey,
				Value:           10000,
			}
		}
		transaction := &externalapi.DomainTransaction{
			Version: constants.MaxTransactionVersion,
			Inputs:  []*externalapi.DomainTransactionInput{input},
			Outputs: outputs,
			Payload: []byte{},
		}

		// Create a block that includes the above transaction
		block, err := testConsensus.BuildBlock(coinbaseData, []*externalapi.DomainTransaction{transaction})
		if err != nil {
			b.Fatalf("Error building block: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(block, true)
		if err != nil {
			b.Fatalf("Error validating and inserting block: %+v", err)
		}

		return block
	}

	// Add finalityDepth blocks, each containing lots of outputs
	tip := blockWithSpendableCoinbase
	for i := 0; i < finalityDepth; i++ {
		tip = addBlockWithLotsOfOutputs(b, tip.Transactions[0])
	}

	// Add enough blocks to move the pruning point
	for {
		block, err := testConsensus.BuildBlock(coinbaseData, nil)
		if err != nil {
			b.Fatalf("Error building block: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(block, true)
		if err != nil {
			b.Fatalf("Error validating and inserting block: %+v", err)
		}

		pruningPoint, err := testConsensus.PruningPoint()
		if err != nil {
			b.Fatalf("Error getting the pruning point: %+v", err)
		}
		if !pruningPoint.Equal(consensusConfig.GenesisHash) {
			break
		}
	}
	pruningPoint, err := testConsensus.PruningPoint()
	if err != nil {
		b.Fatalf("Error getting the pruning point: %+v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get pruning point UTXOs in a loop
		step := 100
		var fromOutpoint *externalapi.DomainOutpoint
		for {
			outpointAndUTXOEntryPairs, err := testConsensus.GetPruningPointUTXOs(pruningPoint, fromOutpoint, step)
			if err != nil {
				b.Fatalf("Error getting pruning point UTXOs: %+v", err)
			}
			fromOutpoint = outpointAndUTXOEntryPairs[len(outpointAndUTXOEntryPairs)-1].Outpoint

			if len(outpointAndUTXOEntryPairs) < step {
				break
			}
		}
	}
}
