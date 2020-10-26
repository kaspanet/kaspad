package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/util/mstime"
)

const blockVersion = 1

func (bp *blockProcessor) buildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	header, err := bp.buildHeader(transactions)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactions,
		Hash:         nil,
	}, nil
}

func (bp *blockProcessor) buildHeader(transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlockHeader, error) {
	parentHashes := bp.newBlockParentHashes()
	timeInMilliseconds, err := bp.newBlockTime()
	if err != nil {
		return nil, err
	}
	bits, err := bp.newBlockDifficulty()
	if err != nil {
		return nil, err
	}
	hashMerkleRoot := bp.newBlockHashMerkleRoot(transactions)
	acceptedIDMerkleRoot, err := bp.newBlockAcceptedIDMerkleRoot()
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := bp.newBlockUTXOCommitment()
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlockHeader{
		Version:              blockVersion,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       hashMerkleRoot,
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       utxoCommitment,
		TimeInMilliseconds:   timeInMilliseconds,
		Bits:                 bits,
	}, nil
}

func (bp *blockProcessor) newBlockParentHashes() []*externalapi.DomainHash {
	return bp.consensusStateManager.VirtualParentHashes()
}

func (bp *blockProcessor) newBlockTime() (int64, error) {
	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks. Thus, choose the maximum between the
	// current time and one second after the past median time. The current
	// timestamp is truncated to a millisecond boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// millisecond.
	newTimestamp := mstime.Now().UnixMilliseconds() + 1
	minTimestamp, err := bp.pastMedianTimeManager.PastMedianTime(bp.consensusStateManager.VirtualSelectedParent())
	if err != nil {
		return 0, err
	}
	if newTimestamp < minTimestamp {
		newTimestamp = minTimestamp
	}
	return newTimestamp, nil
}

func (bp *blockProcessor) newBlockDifficulty() (uint32, error) {
	return bp.difficultyManager.RequiredDifficulty(bp.consensusStateManager.VirtualSelectedParent())
}

func (bp *blockProcessor) newBlockHashMerkleRoot(transactions []*externalapi.DomainTransaction) *externalapi.DomainHash {
	return merkle.CalculateHashMerkleRoot(transactions)
}

func (bp *blockProcessor) newBlockAcceptedIDMerkleRoot() (*externalapi.DomainHash, error) {
	newBlockAcceptanceData, err := bp.acceptanceDataStore.Get(bp.databaseContext, model.VirtualHash)
	if err != nil {
		return nil, err
	}
	newBlockAcceptedIDMerkleRoot := calculateAcceptedIDMerkleRoot(newBlockAcceptanceData)
	return newBlockAcceptedIDMerkleRoot.Hash(), nil
}

func (bp *blockProcessor) newBlockUTXOCommitment() (*externalapi.DomainHash, error) {
	newBlockMultiset, err := bp.multisetStore.Get(bp.databaseContext, model.VirtualHash)
	if err != nil {
		return nil, err
	}
	newBlockUTXOCommitment := newBlockMultiset.Hash()
	return newBlockUTXOCommitment, nil
}
