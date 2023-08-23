package miningmanager

import (
	"sync"
	"time"

	"github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"
	"github.com/c4ei/yunseokyeol/domain/consensusreference"
	miningmanagermodel "github.com/c4ei/yunseokyeol/domain/miningmanager/model"
)

// MiningManager creates block templates for mining as well as maintaining
// known transactions that have no yet been added to any block
type MiningManager interface {
	GetBlockTemplate(coinbaseData *externalapi.DomainCoinbaseData) (block *externalapi.DomainBlock, isNearlySynced bool, err error)
	ClearBlockTemplate()
	GetBlockTemplateBuilder() miningmanagermodel.BlockTemplateBuilder
	GetTransaction(transactionID *externalapi.DomainTransactionID, includeTransactionPool bool, includeOrphanPool bool) (
		transactionPoolTransaction *externalapi.DomainTransaction,
		isOrphan bool,
		found bool)
	GetTransactionsByAddresses(includeTransactionPool bool, includeOrphanPool bool) (
		sendingInTransactionPool map[string]*externalapi.DomainTransaction,
		receivingInTransactionPool map[string]*externalapi.DomainTransaction,
		sendingInOrphanPool map[string]*externalapi.DomainTransaction,
		receivingInOrphanPool map[string]*externalapi.DomainTransaction,
		err error)
	AllTransactions(includeTransactionPool bool, includeOrphanPool bool) (
		transactionPoolTransactions []*externalapi.DomainTransaction,
		orphanPoolTransactions []*externalapi.DomainTransaction)
	TransactionCount(includeTransactionPool bool, includeOrphanPool bool) int
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
func (mm *miningManager) GetBlockTemplate(coinbaseData *externalapi.DomainCoinbaseData) (block *externalapi.DomainBlock, isNearlySynced bool, err error) {
	mm.cacheLock.Lock()
	immutableCachedTemplate := mm.getImmutableCachedTemplate()
	// We first try and use a cached template
	if immutableCachedTemplate != nil {
		mm.cacheLock.Unlock()
		if immutableCachedTemplate.CoinbaseData.Equal(coinbaseData) {
			return immutableCachedTemplate.Block, immutableCachedTemplate.IsNearlySynced, nil
		}
		// Coinbase data is new -- make the minimum changes required
		// Note we first clone the block template since it is modified by the call
		modifiedBlockTemplate, err := mm.blockTemplateBuilder.ModifyBlockTemplate(coinbaseData, immutableCachedTemplate.Clone())
		if err != nil {
			return nil, false, err
		}

		// No point in updating cache since we have no reason to believe this coinbase will be used more
		// than the previous one, and we want to maintain the original template caching time
		return modifiedBlockTemplate.Block, modifiedBlockTemplate.IsNearlySynced, nil
	}
	defer mm.cacheLock.Unlock()
	// No relevant cache, build a template
	blockTemplate, err := mm.blockTemplateBuilder.BuildBlockTemplate(coinbaseData)
	if err != nil {
		return nil, false, err
	}
	// Cache the built template
	mm.setImmutableCachedTemplate(blockTemplate)
	return blockTemplate.Block, blockTemplate.IsNearlySynced, nil
}

func (mm *miningManager) ClearBlockTemplate() {
	mm.cacheLock.Lock()
	mm.cachingTime = time.Time{}
	mm.cachedBlockTemplate = nil
	mm.cacheLock.Unlock()
}

func (mm *miningManager) getImmutableCachedTemplate() *externalapi.DomainBlockTemplate {
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
	transactionID *externalapi.DomainTransactionID,
	includeTransactionPool bool,
	includeOrphanPool bool) (
	transactionPoolTransaction *externalapi.DomainTransaction,
	isOrphan bool,
	found bool) {

	return mm.mempool.GetTransaction(transactionID, includeTransactionPool, includeOrphanPool)
}

func (mm *miningManager) AllTransactions(includeTransactionPool bool, includeOrphanPool bool) (
	transactionPoolTransactions []*externalapi.DomainTransaction,
	orphanPoolTransactions []*externalapi.DomainTransaction) {

	return mm.mempool.AllTransactions(includeTransactionPool, includeOrphanPool)
}

func (mm *miningManager) GetTransactionsByAddresses(includeTransactionPool bool, includeOrphanPool bool) (
	sendingInTransactionPool map[string]*externalapi.DomainTransaction,
	receivingInTransactionPool map[string]*externalapi.DomainTransaction,
	sendingInOrphanPool map[string]*externalapi.DomainTransaction,
	receivingInOrphanPool map[string]*externalapi.DomainTransaction,
	err error) {

	return mm.mempool.GetTransactionsByAddresses(includeTransactionPool, includeOrphanPool)
}

func (mm *miningManager) TransactionCount(includeTransactionPool bool, includeOrphanPool bool) int {
	return mm.mempool.TransactionCount(includeTransactionPool, includeOrphanPool)
}

func (mm *miningManager) RevalidateHighPriorityTransactions() (
	validTransactions []*externalapi.DomainTransaction, err error) {

	return mm.mempool.RevalidateHighPriorityTransactions()
}
