package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensusreference"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
	"sync"
	"time"
)

// MiningManager creates block templates for mining as well as maintaining
// known transactions that have no yet been added to any block
type MiningManager interface {
	GetBlockTemplate(coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainBlock, error)
	GetTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool)
	AllTransactions() []*externalapi.DomainTransaction
	TransactionCount() int
	HandleNewBlockTransactions(txs []*externalapi.DomainTransaction) ([]*externalapi.DomainTransaction, error)
	ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool, allowOrphan bool) (
		acceptedTransactions []*externalapi.DomainTransaction, err error)
	RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error)
}

type miningManager struct {
	consensusReference   consensusreference.ConsensusReference
	mempool              miningmanagermodel.Mempool
	blockTemplateBuilder miningmanagermodel.BlockTemplateBuilder
	cachedBlockTemplate  *externalapi.DomainBlockTemplate
	cachingTime          time.Time
	cacheLock            *sync.Mutex
}

// GetBlockTemplate obtains a block template for a miner to consume
func (mm *miningManager) GetBlockTemplate(coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainBlock, error) {
	immutableCachedTemplate := mm.getImmutableCachedTemplate()
	// We first try and use a cached template
	if immutableCachedTemplate != nil {
		virtualInfo, err := mm.consensusReference.Consensus().GetVirtualInfo()
		if err != nil {
			return nil, err
		}
		if externalapi.HashesEqual(virtualInfo.ParentHashes, immutableCachedTemplate.Block.Header.DirectParents()) {
			if immutableCachedTemplate.CoinbaseData.Equal(coinbaseData) {
				// Both, virtual parents and coinbase data are equal, simply return the cached block
				return immutableCachedTemplate.Block, nil
			}

			// Virtual parents are equal, but coinbase data is new -- make the minimum changes required
			// Note we first clone the block template since it is modified by the call
			modifiedBlockTemplate, err := mm.blockTemplateBuilder.ModifyBlockTemplate(coinbaseData, immutableCachedTemplate.Clone())
			if err != nil {
				return nil, err
			}

			// No point in updating cache since we have no reason to believe this coinbase will be used more
			// than the previous one, and we want to maintain the original template caching time
			return modifiedBlockTemplate.Block, nil
		}
	}
	// No relevant cache, build a template
	blockTemplate, err := mm.blockTemplateBuilder.GetBlockTemplate(coinbaseData)
	if err != nil {
		return nil, err
	}
	// Cache the built template
	mm.setImmutableCachedTemplate(blockTemplate)
	return blockTemplate.Block, err
}

func (mm *miningManager) getImmutableCachedTemplate() *externalapi.DomainBlockTemplate {
	mm.cacheLock.Lock()
	defer mm.cacheLock.Unlock()
	if time.Now().Sub(mm.cachingTime) > time.Second {
		// No point in cache optimizations if queries are more than a second apart -- we prefer rechecking the mempool
		mm.cachedBlockTemplate = nil
	}
	return mm.cachedBlockTemplate
}

func (mm *miningManager) setImmutableCachedTemplate(blockTemplate *externalapi.DomainBlockTemplate) {
	mm.cacheLock.Lock()
	defer mm.cacheLock.Unlock()
	mm.cachingTime = time.Now()
	mm.cachedBlockTemplate = blockTemplate
}

// HandleNewBlockTransactions handles the transactions for a new block that was just added to the DAG
func (mm *miningManager) HandleNewBlockTransactions(txs []*externalapi.DomainTransaction) ([]*externalapi.DomainTransaction, error) {
	return mm.mempool.HandleNewBlockTransactions(txs)
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the set of known transactions that have not yet been
// added to any block
func (mm *miningManager) ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction,
	isHighPriority bool, allowOrphan bool) (acceptedTransactions []*externalapi.DomainTransaction, err error) {

	return mm.mempool.ValidateAndInsertTransaction(transaction, isHighPriority, allowOrphan)
}

func (mm *miningManager) GetTransaction(
	transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool) {

	return mm.mempool.GetTransaction(transactionID)
}

func (mm *miningManager) AllTransactions() []*externalapi.DomainTransaction {
	return mm.mempool.AllTransactions()
}

func (mm *miningManager) TransactionCount() int {
	return mm.mempool.TransactionCount()
}

func (mm *miningManager) RevalidateHighPriorityTransactions() (
	validTransactions []*externalapi.DomainTransaction, err error) {

	return mm.mempool.RevalidateHighPriorityTransactions()
}
