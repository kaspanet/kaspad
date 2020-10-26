package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
)

func (bp *blockProcessor) buildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	header, err := bp.buildHeader()
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactions,
		Hash:         nil,
	}, nil
}

func (bp *blockProcessor) buildHeader() (*externalapi.DomainBlockHeader, error) {
	parentHashes := bp.newBlockParentHashes()
	timeInMilliseconds, err := bp.newBlockTime()
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlockHeader{
		Version:              0,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       nil,
		AcceptedIDMerkleRoot: nil,
		UTXOCommitment:       nil,
		TimeInMilliseconds:   timeInMilliseconds,
		Bits:                 0,
		Nonce:                0,
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
