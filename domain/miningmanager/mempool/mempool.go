// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"container/list"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"

	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"

	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

const (
	// orphanTTL is the maximum amount of time an orphan is allowed to
	// stay in the orphan pool before it expires and is evicted during the
	// next scan.
	orphanTTL = time.Minute * 15

	// orphanExpireScanInterval is the minimum amount of time in between
	// scans of the orphan pool to evict expired transactions.
	orphanExpireScanInterval = time.Minute * 5
)

// policy houses the policy (configuration parameters) which is used to
// control the mempool.
type policy struct {
	// MaxTxVersion is the transaction version that the mempool should
	// accept. All transactions above this version are rejected as
	// non-standard.
	MaxTxVersion uint16

	// AcceptNonStd defines whether to accept non-standard transactions. If
	// true, non-standard transactions will be accepted into the mempool.
	// Otherwise, all non-standard transactions will be rejected.
	AcceptNonStd bool

	// MaxOrphanTxs is the maximum number of orphan transactions
	// that can be queued.
	MaxOrphanTxs int

	// MaxOrphanTxSize is the maximum size allowed for orphan transactions.
	// This helps prevent memory exhaustion attacks from sending a lot of
	// of big orphans.
	MaxOrphanTxSize int

	// MinRelayTxFee defines the minimum transaction fee in KAS/kB to be
	// considered a non-zero fee.
	MinRelayTxFee util.Amount
}

// mempool is used as a source of transactions that need to be mined into blocks
// and relayed to other peers. It is safe for concurrent access from multiple
// peers.
type mempool struct {
	pool map[consensusexternalapi.DomainTransactionID]*txDescriptor

	chainedTransactions                  map[consensusexternalapi.DomainTransactionID]*txDescriptor
	chainedTransactionByPreviousOutpoint map[consensusexternalapi.DomainOutpoint]*txDescriptor

	orphans       map[consensusexternalapi.DomainTransactionID]*orphanTx
	orphansByPrev map[consensusexternalapi.DomainOutpoint]map[consensusexternalapi.DomainTransactionID]*consensusexternalapi.DomainTransaction

	mempoolUTXOSet *mempoolUTXOSet
	consensus      consensusexternalapi.Consensus

	orderedTransactionsByFeeRate []*consensusexternalapi.DomainTransaction

	// nextExpireScan is the time after which the orphan pool will be
	// scanned in order to evict orphans. This is NOT a hard deadline as
	// the scan will only run when an orphan is added to the pool as opposed
	// to on an unconditional timer.
	nextExpireScan mstime.Time

	mtx    sync.RWMutex
	policy policy
}

// New returns a new memory pool for validating and storing standalone
// transactions until they are mined into a block.
func New(consensus consensusexternalapi.Consensus, acceptNonStd bool) miningmanagermodel.Mempool {
	policy := policy{
		MaxTxVersion:    constants.MaxTransactionVersion,
		AcceptNonStd:    acceptNonStd,
		MaxOrphanTxs:    5,
		MaxOrphanTxSize: 100000,
		MinRelayTxFee:   1000, // 1 sompi per byte
	}
	return &mempool{
		mtx:                                  sync.RWMutex{},
		policy:                               policy,
		pool:                                 make(map[consensusexternalapi.DomainTransactionID]*txDescriptor),
		chainedTransactions:                  make(map[consensusexternalapi.DomainTransactionID]*txDescriptor),
		chainedTransactionByPreviousOutpoint: make(map[consensusexternalapi.DomainOutpoint]*txDescriptor),
		orphans:                              make(map[consensusexternalapi.DomainTransactionID]*orphanTx),
		orphansByPrev:                        make(map[consensusexternalapi.DomainOutpoint]map[consensusexternalapi.DomainTransactionID]*consensusexternalapi.DomainTransaction),
		mempoolUTXOSet:                       newMempoolUTXOSet(),
		consensus:                            consensus,
		nextExpireScan:                       mstime.Now().Add(orphanExpireScanInterval),
	}
}

func (mp *mempool) GetTransaction(
	transactionID *consensusexternalapi.DomainTransactionID) (*consensusexternalapi.DomainTransaction, bool) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	txDesc, exists := mp.fetchTxDesc(transactionID)
	if !exists {
		return nil, false
	}

	return txDesc.DomainTransaction, true
}

func (mp *mempool) AllTransactions() []*consensusexternalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	transactions := make([]*consensusexternalapi.DomainTransaction, 0, len(mp.pool)+len(mp.chainedTransactions))
	for _, txDesc := range mp.pool {
		transactions = append(transactions, txDesc.DomainTransaction)
	}

	for _, txDesc := range mp.chainedTransactions {
		transactions = append(transactions, txDesc.DomainTransaction)
	}

	return transactions
}

// txDescriptor is a descriptor containing a transaction in the mempool along with
// additional metadata.
type txDescriptor struct {
	*consensusexternalapi.DomainTransaction

	// depCount is not 0 for a chained transaction. A chained transaction is
	// one that is accepted to pool, but cannot be mined in next block because it
	// depends on outputs of accepted, but still not mined transaction
	depCount int
}

// orphanTx is normal transaction that references an ancestor transaction
// that is not yet available. It also contains additional information related
// to it such as an expiration time to help prevent caching the orphan forever.
type orphanTx struct {
	tx         *consensusexternalapi.DomainTransaction
	expiration mstime.Time
}

// removeOrphan removes the passed orphan transaction from the orphan pool and
// previous orphan index.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) removeOrphan(tx *consensusexternalapi.DomainTransaction, removeRedeemers bool) {
	// Nothing to do if passed tx is not an orphan.
	txID := consensushashing.TransactionID(tx)
	otx, exists := mp.orphans[*txID]
	if !exists {
		return
	}

	// Remove the reference from the previous orphan index.
	for _, txIn := range otx.tx.Inputs {
		orphans, exists := mp.orphansByPrev[txIn.PreviousOutpoint]
		if exists {
			delete(orphans, *txID)

			// Remove the map entry altogether if there are no
			// longer any orphans which depend on it.
			if len(orphans) == 0 {
				delete(mp.orphansByPrev, txIn.PreviousOutpoint)
			}
		}
	}

	// Remove any orphans that redeem outputs from this one if requested.
	if removeRedeemers {
		prevOut := consensusexternalapi.DomainOutpoint{TransactionID: *txID}
		for txOutIdx := range tx.Outputs {
			prevOut.Index = uint32(txOutIdx)
			for _, orphan := range mp.orphansByPrev[prevOut] {
				mp.removeOrphan(orphan, true)
			}
		}
	}

	// Remove the transaction from the orphan pool.
	delete(mp.orphans, *txID)
}

// limitNumOrphans limits the number of orphan transactions by evicting a random
// orphan if adding a new one would cause it to overflow the max allowed.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) limitNumOrphans() error {
	// Scan through the orphan pool and remove any expired orphans when it's
	// time. This is done for efficiency so the scan only happens
	// periodically instead of on every orphan added to the pool.
	if now := mstime.Now(); now.After(mp.nextExpireScan) {
		origNumOrphans := len(mp.orphans)
		for _, otx := range mp.orphans {
			if now.After(otx.expiration) {
				// Remove redeemers too because the missing
				// parents are very unlikely to ever materialize
				// since the orphan has already been around more
				// than long enough for them to be delivered.
				mp.removeOrphan(otx.tx, true)
			}
		}

		// Set next expiration scan to occur after the scan interval.
		mp.nextExpireScan = now.Add(orphanExpireScanInterval)

		numOrphans := len(mp.orphans)
		if numExpired := origNumOrphans - numOrphans; numExpired > 0 {
			log.Debugf("Expired %d %s (remaining: %d)", numExpired,
				logger.PickNoun(uint64(numExpired), "orphan", "orphans"),
				numOrphans)
		}
	}

	// Nothing to do if adding another orphan will not cause the pool to
	// exceed the limit.
	if len(mp.orphans)+1 <= mp.policy.MaxOrphanTxs {
		return nil
	}

	// Remove a random entry from the map. For most compilers, Go's
	// range statement iterates starting at a random item although
	// that is not 100% guaranteed by the spec. The iteration order
	// is not important here because an adversary would have to be
	// able to pull off preimage attacks on the hashing function in
	// order to target eviction of specific entries anyways.
	for _, otx := range mp.orphans {
		// Don't remove redeemers in the case of a random eviction since
		// it is quite possible it might be needed again shortly.
		mp.removeOrphan(otx.tx, false)
		break
	}

	return nil
}

// addOrphan adds an orphan transaction to the orphan pool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) addOrphan(tx *consensusexternalapi.DomainTransaction) {
	// Nothing to do if no orphans are allowed.
	if mp.policy.MaxOrphanTxs <= 0 {
		return
	}

	// Limit the number orphan transactions to prevent memory exhaustion.
	// This will periodically remove any expired orphans and evict a random
	// orphan if space is still needed.
	mp.limitNumOrphans()
	txID := consensushashing.TransactionID(tx)
	mp.orphans[*txID] = &orphanTx{
		tx:         tx,
		expiration: mstime.Now().Add(orphanTTL),
	}
	for _, txIn := range tx.Inputs {
		if _, exists := mp.orphansByPrev[txIn.PreviousOutpoint]; !exists {
			mp.orphansByPrev[txIn.PreviousOutpoint] =
				make(map[consensusexternalapi.DomainTransactionID]*consensusexternalapi.DomainTransaction)
		}
		mp.orphansByPrev[txIn.PreviousOutpoint][*txID] = tx
	}

	log.Debugf("Stored orphan transaction %s (total: %d)", consensushashing.TransactionID(tx),
		len(mp.orphans))
}

// maybeAddOrphan potentially adds an orphan to the orphan pool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) maybeAddOrphan(tx *consensusexternalapi.DomainTransaction) error {
	// Ignore orphan transactions that are too large. This helps avoid
	// a memory exhaustion attack based on sending a lot of really large
	// orphans. In the case there is a valid transaction larger than this,
	// it will ultimtely be rebroadcast after the parent transactions
	// have been mined or otherwise received.
	//
	// Note that the number of orphan transactions in the orphan pool is
	// also limited, so this equates to a maximum memory used of
	// mp.policy.MaxOrphanTxSize * mp.policy.MaxOrphanTxs (which is ~5MB
	// using the default values at the time this comment was written).
	serializedLen := estimatedsize.TransactionEstimatedSerializedSize(tx)
	if serializedLen > uint64(mp.policy.MaxOrphanTxSize) {
		str := fmt.Sprintf("orphan transaction size of %d bytes is "+
			"larger than max allowed size of %d bytes",
			serializedLen, mp.policy.MaxOrphanTxSize)
		return txRuleError(RejectNonstandard, str)
	}

	// Add the orphan if the none of the above disqualified it.
	mp.addOrphan(tx)

	return nil
}

// removeOrphanDoubleSpends removes all orphans which spend outputs spent by the
// passed transaction from the orphan pool. Removing those orphans then leads
// to removing all orphans which rely on them, recursively. This is necessary
// when a transaction is added to the main pool because it may spend outputs
// that orphans also spend.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) removeOrphanDoubleSpends(tx *consensusexternalapi.DomainTransaction) {
	for _, txIn := range tx.Inputs {
		for _, orphan := range mp.orphansByPrev[txIn.PreviousOutpoint] {
			mp.removeOrphan(orphan, true)
		}
	}
}

// isTransactionInPool returns whether or not the passed transaction already
// exists in the main pool.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *mempool) isTransactionInPool(txID *consensusexternalapi.DomainTransactionID) bool {
	if _, exists := mp.pool[*txID]; exists {
		return true
	}
	return mp.isInDependPool(txID)
}

// isInDependPool returns whether or not the passed transaction already
// exists in the depend pool.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *mempool) isInDependPool(hash *consensusexternalapi.DomainTransactionID) bool {
	if _, exists := mp.chainedTransactions[*hash]; exists {
		return true
	}

	return false
}

// isOrphanInPool returns whether or not the passed transaction already exists
// in the orphan pool.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *mempool) isOrphanInPool(txID *consensusexternalapi.DomainTransactionID) bool {
	if _, exists := mp.orphans[*txID]; exists {
		return true
	}

	return false
}

// haveTransaction returns whether or not the passed transaction already exists
// in the main pool or in the orphan pool.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *mempool) haveTransaction(txID *consensusexternalapi.DomainTransactionID) bool {
	return mp.isTransactionInPool(txID) || mp.isOrphanInPool(txID)
}

// removeBlockTransactionsFromPool removes the transactions that are found in the block
// from the mempool, and move their chained mempool transactions (if any) to the main pool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) removeBlockTransactionsFromPool(txs []*consensusexternalapi.DomainTransaction) error {
	for _, tx := range txs[transactionhelper.CoinbaseTransactionIndex+1:] {
		txID := consensushashing.TransactionID(tx)

		// We use the mempool transaction, because it has populated fee and mass
		mempoolTx, exists := mp.fetchTxDesc(txID)
		if !exists {
			continue
		}

		err := mp.cleanTransactionFromSets(mempoolTx.DomainTransaction)
		if err != nil {
			return err
		}

		mp.updateBlockTransactionChainedTransactions(mempoolTx.DomainTransaction)
	}
	return nil
}

// removeTransactionAndItsChainedTransactions removes a transaction and all of its chained transaction from the mempool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) removeTransactionAndItsChainedTransactions(tx *consensusexternalapi.DomainTransaction) error {
	txID := consensushashing.TransactionID(tx)
	// Remove any transactions which rely on this one.
	for i := uint32(0); i < uint32(len(tx.Outputs)); i++ {
		prevOut := consensusexternalapi.DomainOutpoint{TransactionID: *txID, Index: i}
		if txRedeemer, exists := mp.mempoolUTXOSet.poolTransactionBySpendingOutpoint(prevOut); exists {
			err := mp.removeTransactionAndItsChainedTransactions(txRedeemer)
			if err != nil {
				return err
			}
		}
	}

	if _, exists := mp.chainedTransactions[*txID]; exists {
		mp.removeChainTransaction(tx)
	}

	err := mp.cleanTransactionFromSets(tx)
	if err != nil {
		return err
	}

	return nil
}

// cleanTransactionFromSets removes the transaction from all mempool related transaction sets.
// It assumes that any chained transaction is already cleaned from the mempool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) cleanTransactionFromSets(tx *consensusexternalapi.DomainTransaction) error {
	err := mp.mempoolUTXOSet.removeTx(tx)
	if err != nil {
		return err
	}

	txID := consensushashing.TransactionID(tx)
	delete(mp.pool, *txID)
	delete(mp.chainedTransactions, *txID)

	return mp.removeTransactionFromOrderedTransactionsByFeeRate(tx)
}

// updateBlockTransactionChainedTransactions processes the dependencies of a
// transaction that was included in a block and was just now removed from the mempool.
//
// This function MUST be called with the mempool lock held (for writes).

func (mp *mempool) updateBlockTransactionChainedTransactions(tx *consensusexternalapi.DomainTransaction) {
	prevOut := consensusexternalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(tx)}
	for txOutIdx := range tx.Outputs {
		// Skip to the next available output if there are none.
		prevOut.Index = uint32(txOutIdx)
		txDesc, exists := mp.chainedTransactionByPreviousOutpoint[prevOut]
		if !exists {
			continue
		}

		txDesc.depCount--
		// If the transaction is not chained anymore, move it into the main pool
		if txDesc.depCount == 0 {
			// Transaction may be already removed by recursive calls, if removeRedeemers is true.
			// So avoid moving it into main pool
			txDescID := consensushashing.TransactionID(txDesc.DomainTransaction)
			if _, ok := mp.chainedTransactions[*txDescID]; ok {
				delete(mp.chainedTransactions, *txDescID)
				mp.pool[*txDescID] = txDesc
			}
		}
		delete(mp.chainedTransactionByPreviousOutpoint, prevOut)
	}
}

// removeChainTransaction removes a chain transaction and all of its relation as a result of double spend.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) removeChainTransaction(tx *consensusexternalapi.DomainTransaction) {
	delete(mp.chainedTransactions, *consensushashing.TransactionID(tx))
	for _, txIn := range tx.Inputs {
		delete(mp.chainedTransactionByPreviousOutpoint, txIn.PreviousOutpoint)
	}
}

// removeDoubleSpends removes all transactions which spend outputs spent by the
// passed transaction from the memory pool. Removing those transactions then
// leads to removing all transactions which rely on them, recursively. This is
// necessary when a block is connected to the DAG because the block may
// contain transactions which were previously unknown to the memory pool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) removeDoubleSpends(tx *consensusexternalapi.DomainTransaction) error {
	txID := consensushashing.TransactionID(tx)
	for _, txIn := range tx.Inputs {
		if txRedeemer, ok := mp.mempoolUTXOSet.poolTransactionBySpendingOutpoint(txIn.PreviousOutpoint); ok {
			if !consensushashing.TransactionID(txRedeemer).Equal(txID) {
				err := mp.removeTransactionAndItsChainedTransactions(txRedeemer)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// addTransaction adds the passed transaction to the memory pool. It should
// not be called directly as it doesn't perform any validation. This is a
// helper for maybeAcceptTransaction.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) addTransaction(tx *consensusexternalapi.DomainTransaction, parentsInPool []consensusexternalapi.DomainOutpoint) (*txDescriptor, error) {
	// Add the transaction to the pool and mark the referenced outpoints
	// as spent by the pool.
	txDescriptor := &txDescriptor{
		DomainTransaction: tx,
		depCount:          len(parentsInPool),
	}
	txID := *consensushashing.TransactionID(tx)

	if len(parentsInPool) == 0 {
		mp.pool[txID] = txDescriptor
	} else {
		mp.chainedTransactions[txID] = txDescriptor
		for _, previousOutpoint := range parentsInPool {
			mp.chainedTransactionByPreviousOutpoint[previousOutpoint] = txDescriptor
		}
	}

	err := mp.mempoolUTXOSet.addTx(tx)
	if err != nil {
		return nil, err
	}

	err = mp.addTransactionToOrderedTransactionsByFeeRate(tx)
	if err != nil {
		return nil, err
	}

	return txDescriptor, nil
}

func (mp *mempool) findTxIndexInOrderedTransactionsByFeeRate(tx *consensusexternalapi.DomainTransaction) (int, error) {
	if tx.Fee == 0 || tx.Mass == 0 {
		return 0, errors.Errorf("findTxIndexInOrderedTransactionsByFeeRate expects a transaction with " +
			"populated fee and as")
	}
	txID := consensushashing.TransactionID(tx)
	txFeeRate := float64(tx.Fee) / float64(tx.Mass)

	return sort.Search(len(mp.orderedTransactionsByFeeRate), func(i int) bool {
		elementFeeRate := float64(mp.orderedTransactionsByFeeRate[i].Fee) / float64(mp.orderedTransactionsByFeeRate[i].Mass)
		return elementFeeRate > txFeeRate ||
			(elementFeeRate == txFeeRate &&
				consensusexternalapi.LessOrEqual(
					(*consensusexternalapi.DomainHash)(txID),
					(*consensusexternalapi.DomainHash)(consensushashing.TransactionID(mp.orderedTransactionsByFeeRate[i])),
				))
	}), nil
}

func (mp *mempool) addTransactionToOrderedTransactionsByFeeRate(tx *consensusexternalapi.DomainTransaction) error {
	index, err := mp.findTxIndexInOrderedTransactionsByFeeRate(tx)
	if err != nil {
		return err
	}

	mp.orderedTransactionsByFeeRate = append(mp.orderedTransactionsByFeeRate[:index],
		append([]*consensusexternalapi.DomainTransaction{tx}, mp.orderedTransactionsByFeeRate[index:]...)...)

	return nil
}

func (mp *mempool) removeTransactionFromOrderedTransactionsByFeeRate(tx *consensusexternalapi.DomainTransaction) error {
	index, err := mp.findTxIndexInOrderedTransactionsByFeeRate(tx)
	if err != nil {
		return err
	}

	txID := consensushashing.TransactionID(tx)
	if !consensushashing.TransactionID(mp.orderedTransactionsByFeeRate[index]).Equal(txID) {
		return errors.Errorf("Couldn't find %s in mp.orderedTransactionsByFeeRate", txID)
	}

	mp.orderedTransactionsByFeeRate = append(mp.orderedTransactionsByFeeRate[:index], mp.orderedTransactionsByFeeRate[index+1:]...)
	return nil
}

func (mp *mempool) enforceTransactionLimit() error {
	const limit = 1_000_000
	if len(mp.pool)+len(mp.chainedTransactions) > limit {
		// mp.orderedTransactionsByFeeRate[0] is the least profitable transaction
		txToRemove := mp.orderedTransactionsByFeeRate[0]
		log.Debugf("Mempool size is over the limit of %d transactions. Removing %s",
			limit,
			consensushashing.TransactionID(txToRemove),
		)
		return mp.removeTransactionAndItsChainedTransactions(txToRemove)
	}
	return nil
}

// checkPoolDoubleSpend checks whether or not the passed transaction is
// attempting to spend coins already spent by other transactions in the pool.
// Note it does not check for double spends against transactions already in the
// DAG.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *mempool) checkPoolDoubleSpend(tx *consensusexternalapi.DomainTransaction) error {
	for _, txIn := range tx.Inputs {
		if txR, exists := mp.mempoolUTXOSet.poolTransactionBySpendingOutpoint(txIn.PreviousOutpoint); exists {
			str := fmt.Sprintf("output %s already spent by "+
				"transaction %s in the memory pool",
				txIn.PreviousOutpoint, consensushashing.TransactionID(txR))
			return txRuleError(RejectDuplicate, str)
		}
	}

	return nil
}

// This function MUST be called with the mempool lock held (for reads).
// This only fetches from the main transaction pool and does not include
// orphans.
// returns false in the second return parameter if transaction was not found
func (mp *mempool) fetchTxDesc(txID *consensusexternalapi.DomainTransactionID) (*txDescriptor, bool) {
	txDesc, exists := mp.pool[*txID]
	if !exists {
		txDesc, exists = mp.chainedTransactions[*txID]
	}
	return txDesc, exists
}

// maybeAcceptTransaction is the main workhorse for handling insertion of new
// free-standing transactions into a memory pool. It includes functionality
// such as rejecting duplicate transactions, ensuring transactions follow all
// rules, detecting orphan transactions, and insertion into the memory pool.
//
// If the transaction is an orphan (missing parent transactions), the
// transaction is NOT added to the orphan pool, but each unknown referenced
// parent is returned. Use ProcessTransaction instead if new orphans should
// be added to the orphan pool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) maybeAcceptTransaction(tx *consensusexternalapi.DomainTransaction, rejectDupOrphans bool) (
	[]*consensusexternalapi.DomainOutpoint, *txDescriptor, error) {

	txID := consensushashing.TransactionID(tx)

	// Don't accept the transaction if it already exists in the pool. This
	// applies to orphan transactions as well when the reject duplicate
	// orphans flag is set. This check is intended to be a quick check to
	// weed out duplicates.
	if mp.isTransactionInPool(txID) || (rejectDupOrphans &&
		mp.isOrphanInPool(txID)) {

		str := fmt.Sprintf("already have transaction %s", txID)
		return nil, nil, txRuleError(RejectDuplicate, str)
	}

	// Don't allow non-standard transactions if the network parameters
	// forbid their acceptance.
	if !mp.policy.AcceptNonStd {
		err := checkTransactionStandard(tx, &mp.policy)
		if err != nil {
			// Attempt to extract a reject code from the error so
			// it can be retained. When not possible, fall back to
			// a non standard error.
			rejectCode, found := extractRejectCode(err)
			if !found {
				rejectCode = RejectNonstandard
			}
			str := fmt.Sprintf("transaction %s is not standard: %s",
				txID, err)
			return nil, nil, txRuleError(rejectCode, str)
		}
	}

	// The transaction may not use any of the same outputs as other
	// transactions already in the pool as that would ultimately result in a
	// double spend. This check is intended to be quick and therefore only
	// detects double spends within the transaction pool itself. The
	// transaction could still be double spending coins from the DAG
	// at this point. There is a more in-depth check that happens later
	// after fetching the referenced transaction inputs from the DAG
	// which examines the actual spend data and prevents double spends.
	err := mp.checkPoolDoubleSpend(tx)
	if err != nil {
		return nil, nil, err
	}

	// Don't allow the transaction if it exists in the DAG and is
	// not already fully spent.
	if mp.mempoolUTXOSet.checkExists(tx) {
		return nil, nil, txRuleError(RejectDuplicate, "transaction already exists")
	}

	// Transaction is an orphan if any of the referenced transaction outputs
	// don't exist or are already spent. Adding orphans to the orphan pool
	// is not handled by this function, and the caller should use
	// maybeAddOrphan if this behavior is desired.
	parentsInPool := mp.mempoolUTXOSet.populateUTXOEntries(tx)

	// This will populate the missing UTXOEntries.
	err = mp.consensus.ValidateTransactionAndPopulateWithConsensusData(tx)
	missingOutpoints := ruleerrors.ErrMissingTxOut{}
	if err != nil {
		if errors.As(err, &missingOutpoints) {
			return missingOutpoints.MissingOutpoints, nil, nil
		}
		if errors.As(err, &ruleerrors.RuleError{}) {
			return nil, nil, newRuleError(err)
		}
		return nil, nil, err
	}

	// Don't allow transactions with non-standard inputs if the network
	// parameters forbid their acceptance.
	if !mp.policy.AcceptNonStd {
		err := checkInputsStandard(tx)
		if err != nil {
			// Attempt to extract a reject code from the error so
			// it can be retained. When not possible, fall back to
			// a non standard error.
			rejectCode, found := extractRejectCode(err)
			if !found {
				rejectCode = RejectNonstandard
			}
			str := fmt.Sprintf("transaction %s has a non-standard "+
				"input: %s", txID, err)
			return nil, nil, txRuleError(rejectCode, str)
		}
	}

	// Don't allow transactions with fees too low to get into a mined block
	serializedSize := int64(estimatedsize.TransactionEstimatedSerializedSize(tx))
	minFee := uint64(calcMinRequiredTxRelayFee(serializedSize,
		mp.policy.MinRelayTxFee))
	if tx.Fee < minFee {
		str := fmt.Sprintf("transaction %s has %d fees which is under "+
			"the required amount of %d", txID, tx.Fee,
			minFee)
		return nil, nil, txRuleError(RejectInsufficientFee, str)
	}
	// Add to transaction pool.
	txDesc, err := mp.addTransaction(tx, parentsInPool)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("Accepted transaction %s (pool size: %d)", txID,
		len(mp.pool))

	err = mp.enforceTransactionLimit()
	if err != nil {
		return nil, nil, err
	}

	return nil, txDesc, nil
}

// processOrphans determines if there are any orphans which depend on the passed
// transaction hash (it is possible that they are no longer orphans) and
// potentially accepts them to the memory pool. It repeats the process for the
// newly accepted transactions (to detect further orphans which may no longer be
// orphans) until there are no more.
//
// It returns a slice of transactions added to the mempool. A nil slice means
// no transactions were moved from the orphan pool to the mempool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *mempool) processOrphans(acceptedTx *consensusexternalapi.DomainTransaction) []*txDescriptor {
	var acceptedTxns []*txDescriptor

	// Start with processing at least the passed transaction.
	processList := list.New()
	processList.PushBack(acceptedTx)
	for processList.Len() > 0 {
		// Pop the transaction to process from the front of the list.
		firstElement := processList.Remove(processList.Front())
		processItem := firstElement.(*consensusexternalapi.DomainTransaction)

		prevOut := consensusexternalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(processItem)}
		for txOutIdx := range processItem.Outputs {
			// Look up all orphans that redeem the output that is
			// now available. This will typically only be one, but
			// it could be multiple if the orphan pool contains
			// double spends. While it may seem odd that the orphan
			// pool would allow this since there can only possibly
			// ultimately be a single redeemer, it's important to
			// track it this way to prevent malicious actors from
			// being able to purposely constructing orphans that
			// would otherwise make outputs unspendable.
			//
			// Skip to the next available output if there are none.
			prevOut.Index = uint32(txOutIdx)
			orphans, exists := mp.orphansByPrev[prevOut]
			if !exists {
				continue
			}

			// Potentially accept an orphan into the tx pool.
			for _, tx := range orphans {
				missing, txD, err := mp.maybeAcceptTransaction(
					tx, false)
				if err != nil {
					// The orphan is now invalid, so there
					// is no way any other orphans which
					// redeem any of its outputs can be
					// accepted. Remove them.
					mp.removeOrphan(tx, true)
					break
				}

				// Transaction is still an orphan. Try the next
				// orphan which redeems this output.
				if len(missing) > 0 {
					continue
				}

				// Transaction was accepted into the main pool.
				//
				// Add it to the list of accepted transactions
				// that are no longer orphans, remove it from
				// the orphan pool, and add it to the list of
				// transactions to process so any orphans that
				// depend on it are handled too.
				acceptedTxns = append(acceptedTxns, txD)
				mp.removeOrphan(tx, false)
				processList.PushBack(tx)

				// Only one transaction for this outpoint can be
				// accepted, so the rest are now double spends
				// and are removed later.
				break
			}
		}
	}

	// Recursively remove any orphans that also redeem any outputs redeemed
	// by the accepted transactions since those are now definitive double
	// spends.
	mp.removeOrphanDoubleSpends(acceptedTx)
	for _, txDescriptor := range acceptedTxns {
		mp.removeOrphanDoubleSpends(txDescriptor.DomainTransaction)
	}

	return acceptedTxns
}

// ProcessTransaction is the main workhorse for handling insertion of new
// free-standing transactions into the memory pool. It includes functionality
// such as rejecting duplicate transactions, ensuring transactions follow all
// rules, orphan transaction handling, and insertion into the memory pool.
//
// It returns a slice of transactions added to the mempool. When the
// error is nil, the list will include the passed transaction itself along
// with any additional orphan transaactions that were added as a result of
// the passed one being accepted.
//
// This function is safe for concurrent access.
func (mp *mempool) ValidateAndInsertTransaction(tx *consensusexternalapi.DomainTransaction, allowOrphan bool) error {
	log.Tracef("Processing transaction %s", consensushashing.TransactionID(tx))

	// Protect concurrent access.
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	// Potentially accept the transaction to the memory pool.
	missingParents, txD, err := mp.maybeAcceptTransaction(tx, true)
	if err != nil {
		return err
	}

	if len(missingParents) == 0 {
		// Accept any orphan transactions that depend on this
		// transaction (they may no longer be orphans if all inputs
		// are now available) and repeat for those accepted
		// transactions until there are no more.
		newTxs := mp.processOrphans(tx)
		acceptedTxs := make([]*txDescriptor, len(newTxs)+1)

		// Add the parent transaction first so remote nodes
		// do not add orphans.
		acceptedTxs[0] = txD
		copy(acceptedTxs[1:], newTxs)

		return nil
	}

	// The transaction is an orphan (has inputs missing). Reject
	// it if the flag to allow orphans is not set.
	if !allowOrphan {
		// Only use the first missing parent transaction in
		// the error message.
		//
		// NOTE: RejectDuplicate is really not an accurate
		// reject code here, but it matches the reference
		// implementation and there isn't a better choice due
		// to the limited number of reject codes. Missing
		// inputs is assumed to mean they are already spent
		// which is not really always the case.
		str := fmt.Sprintf("orphan transaction %s references "+
			"outputs of unknown or fully-spent "+
			"transaction %s", consensushashing.TransactionID(tx), missingParents[0])
		return txRuleError(RejectDuplicate, str)
	}

	// Potentially add the orphan transaction to the orphan pool.
	return mp.maybeAddOrphan(tx)
}

// Count returns the number of transactions in the main pool. It does not
// include the orphan pool.
//
// This function is safe for concurrent access.
func (mp *mempool) Count() int {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	count := len(mp.pool)

	return count
}

// ChainedCount returns the number of chained transactions in the mempool. It does not
// include the orphan pool.
//
// This function is safe for concurrent access.
func (mp *mempool) ChainedCount() int {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return len(mp.chainedTransactions)
}

// BlockCandidateTransactions returns a slice of all the candidate transactions for the next block
// This is safe for concurrent use
func (mp *mempool) BlockCandidateTransactions() []*consensusexternalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	descs := make([]*consensusexternalapi.DomainTransaction, len(mp.pool))
	i := 0
	for _, desc := range mp.pool {
		descs[i] = desc.DomainTransaction
		i++
	}

	return descs
}

// HandleNewBlockTransactions removes all the transactions in the new block
// from the mempool and the orphan pool, and it also removes
// from the mempool transactions that double spend a
// transaction that is already in the DAG
func (mp *mempool) HandleNewBlockTransactions(txs []*consensusexternalapi.DomainTransaction) ([]*consensusexternalapi.DomainTransaction, error) {
	// Protect concurrent access.
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	// Remove all of the transactions (except the coinbase) in the
	// connected block from the transaction pool. Secondly, remove any
	// transactions which are now double spends as a result of these
	// new transactions. Finally, remove any transaction that is
	// no longer an orphan. Transactions which depend on a confirmed
	// transaction are NOT removed recursively because they are still
	// valid.
	err := mp.removeBlockTransactionsFromPool(txs)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed removing txs from pool")
	}
	acceptedTxs := make([]*consensusexternalapi.DomainTransaction, 0)
	for _, tx := range txs[transactionhelper.CoinbaseTransactionIndex+1:] {
		err := mp.removeDoubleSpends(tx)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed removing tx from mempool: %s", consensushashing.TransactionID(tx))
		}
		mp.removeOrphan(tx, false)
		acceptedOrphans := mp.processOrphans(tx)
		for _, acceptedOrphan := range acceptedOrphans {
			acceptedTxs = append(acceptedTxs, acceptedOrphan.DomainTransaction)
		}
	}

	return acceptedTxs, nil
}

func (mp *mempool) RemoveTransactions(txs []*consensusexternalapi.DomainTransaction) {
	// Protect concurrent access.
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	for _, tx := range txs {
		err := mp.removeDoubleSpends(tx)
		if err != nil {
			log.Infof("Failed removing tx from mempool: %s, '%s'", consensushashing.TransactionID(tx), err)
		}
		mp.removeOrphan(tx, true)
	}
}
