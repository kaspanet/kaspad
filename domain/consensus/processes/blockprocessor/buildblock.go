package blockprocessor

import (
	"sort"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/util/mstime"
)

func (bp *blockProcessor) buildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	coinbase, err := bp.newBlockCoinbaseTransaction(coinbaseData)
	if err != nil {
		return nil, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	header, err := bp.buildHeader(transactionsWithCoinbase)
	if err != nil {
		return nil, err
	}
	headerHash := hashserialization.HeaderHash(header)

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactionsWithCoinbase,
		Hash:         headerHash,
	}, nil
}

func (bp *blockProcessor) newBlockCoinbaseTransaction(
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransaction, error) {

	return bp.coinbaseManager.ExpectedCoinbaseTransaction(model.VirtualBlockHash, coinbaseData)
}

func (bp *blockProcessor) buildHeader(transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlockHeader, error) {
	virtualData, err := bp.consensusStateManager.VirtualData()
	if err != nil {
		return nil, err
	}
	parentHashes, err := bp.newBlockParentHashes(virtualData)
	if err != nil {
		return nil, err
	}
	timeInMilliseconds, err := bp.newBlockTime(virtualData)
	if err != nil {
		return nil, err
	}
	bits, err := bp.newBlockDifficulty(virtualData)
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
		Version:              constants.BlockVersion,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       *hashMerkleRoot,
		AcceptedIDMerkleRoot: *acceptedIDMerkleRoot,
		UTXOCommitment:       *utxoCommitment,
		TimeInMilliseconds:   timeInMilliseconds,
		Bits:                 bits,
	}, nil
}

func (bp *blockProcessor) newBlockParentHashes(virtualData *model.VirtualData) ([]*externalapi.DomainHash, error) {
	return virtualData.ParentHashes, nil
}

func (bp *blockProcessor) newBlockTime(virtualData *model.VirtualData) (int64, error) {
	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks. Thus, choose the maximum between the
	// current time and one second after the past median time. The current
	// timestamp is truncated to a millisecond boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// millisecond.
	newTimestamp := mstime.Now().UnixMilliseconds() + 1
	minTimestamp, err := bp.pastMedianTimeManager.PastMedianTime(virtualData.SelectedParent)
	if err != nil {
		return 0, err
	}
	if newTimestamp < minTimestamp {
		newTimestamp = minTimestamp
	}
	return newTimestamp, nil
}

func (bp *blockProcessor) newBlockDifficulty(virtualData *model.VirtualData) (uint32, error) {
	return bp.difficultyManager.RequiredDifficulty(virtualData.SelectedParent)
}

func (bp *blockProcessor) newBlockHashMerkleRoot(transactions []*externalapi.DomainTransaction) *externalapi.DomainHash {
	return merkle.CalculateHashMerkleRoot(transactions)
}

func (bp *blockProcessor) newBlockAcceptedIDMerkleRoot() (*externalapi.DomainHash, error) {
	newBlockAcceptanceData, err := bp.acceptanceDataStore.Get(bp.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	var acceptedTransactions []*externalapi.DomainTransaction
	for _, blockAcceptanceData := range newBlockAcceptanceData {
		for _, transactionAcceptance := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptance.IsAccepted {
				continue
			}
			acceptedTransactions = append(acceptedTransactions, transactionAcceptance.Transaction)
		}
	}
	sort.Slice(acceptedTransactions, func(i, j int) bool {
		acceptedTransactionIID := hashserialization.TransactionID(acceptedTransactions[i])
		acceptedTransactionJID := hashserialization.TransactionID(acceptedTransactions[j])
		return transactionid.Less(acceptedTransactionIID, acceptedTransactionJID)
	})

	return merkle.CalculateIDMerkleRoot(acceptedTransactions), nil
}

func (bp *blockProcessor) newBlockUTXOCommitment() (*externalapi.DomainHash, error) {
	newBlockMultiset, err := bp.multisetStore.Get(bp.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	newBlockUTXOCommitment := newBlockMultiset.Hash()
	return newBlockUTXOCommitment, nil
}
