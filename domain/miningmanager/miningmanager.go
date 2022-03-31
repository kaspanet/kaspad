package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensusreference"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
	"github.com/kaspanet/kaspad/util/mstime"
	"sync"
	"time"
)

// MiningManager creates block templates for mining as well as maintaining
// known transactions that have no yet been added to any block
type MiningManager interface {
	GetBlockTemplate(coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainBlock, error)
	GetBlockTemplateBuilder() miningmanagermodel.BlockTemplateBuilder
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
				// Both, virtual parents and coinbase data are equal, simply return the cached block with updated time
				newTimestamp := mstime.Now().UnixMilliseconds()
				if newTimestamp < immutableCachedTemplate.Block.Header.TimeInMilliseconds() {
					// Keep the previous time as built by internal consensus median time logic
					return immutableCachedTemplate.Block, nil
				}
				// If new time stamp is later than current, update the header
				mutableHeader := immutableCachedTemplate.Block.Header.ToMutable()
				mutableHeader.SetTimeInMilliseconds(newTimestamp)

				return &externalapi.DomainBlock{
					Header:       mutableHeader.ToImmutable(),
					Transactions: immutableCachedTemplate.Block.Transactions,
				}, nil
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
	blockTemplate, err := mm.blockTemplateBuilder.BuildBlockTemplate(coinbaseData)
	if err != nil {
		return nil, err
	}
	// Cache the built template
	mm.setImmutableCachedTemplate(blockTemplate)
	return blockTemplate.Block, nil
}

func (mm *miningManager) getImmutableCachedTemplate() *externalapi.DomainBlockTemplate {
	mm.cacheLock.Lock()
	defer mm.cacheLock.Unlock()
	if time.Since(mm.cachingTime) > time.Second {
		// No point in cache optimizations if queries are more than a second apart -- we prefer rechecking the mempool.
		// Full explanation: On the one hand this is a sub-millisecond optimization, so there is no harm in doing the full block building
		// every ~1 second. Additionally, we would like to refresh the mempool access even if virtual info was
		// unmodified for a while. All in all, caching for max 1 second is a good compromise.
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

func (mm *miningManager) GetBlockTemplateBuilder() miningmanagermodel.BlockTemplateBuilder {
	return mm.blockTemplateBuilder
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
