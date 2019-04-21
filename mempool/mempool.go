// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"container/list"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/blockdag/indexers"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/mining"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

const (
	// DefaultBlockPrioritySize is the default size in bytes for high-
	// priority / low-fee transactions.  It is used to help determine which
	// are allowed into the mempool and consequently affects their relay and
	// inclusion when generating block templates.
	DefaultBlockPrioritySize = 50000

	// orphanTTL is the maximum amount of time an orphan is allowed to
	// stay in the orphan pool before it expires and is evicted during the
	// next scan.
	orphanTTL = time.Minute * 15

	// orphanExpireScanInterval is the minimum amount of time in between
	// scans of the orphan pool to evict expired transactions.
	orphanExpireScanInterval = time.Minute * 5
)

// NewBlockMsg is the type that is used in NewBlockMsg to transfer
// data about transaction removed and added to the mempool
type NewBlockMsg struct {
	AcceptedTxs []*TxDesc
	Tx          *util.Tx
}

// Tag represents an identifier to use for tagging orphan transactions.  The
// caller may choose any scheme it desires, however it is common to use peer IDs
// so that orphans can be identified by which peer first relayed them.
type Tag uint64

// Config is a descriptor containing the memory pool configuration.
type Config struct {
	// Policy defines the various mempool configuration options related
	// to policy.
	Policy Policy

	// DAGParams identifies which chain parameters the txpool is
	// associated with.
	DAGParams *dagconfig.Params

	// BestHeight defines the function to use to access the block height of
	// the current best chain.
	BestHeight func() int32

	// MedianTimePast defines the function to use in order to access the
	// median time past calculated from the point-of-view of the current
	// chain tip within the best chain.
	MedianTimePast func() time.Time

	// CalcSequenceLockNoLock defines the function to use in order to generate
	// the current sequence lock for the given transaction using the passed
	// utxo set.
	CalcSequenceLockNoLock func(*util.Tx, blockdag.UTXOSet) (*blockdag.SequenceLock, error)

	// IsDeploymentActive returns true if the target deploymentID is
	// active, and false otherwise. The mempool uses this function to gauge
	// if transactions using new to be soft-forked rules should be allowed
	// into the mempool or not.
	IsDeploymentActive func(deploymentID uint32) (bool, error)

	// SigCache defines a signature cache to use.
	SigCache *txscript.SigCache

	// AddrIndex defines the optional address index instance to use for
	// indexing the unconfirmed transactions in the memory pool.
	// This can be nil if the address index is not enabled.
	AddrIndex *indexers.AddrIndex

	// FeeEstimatator provides a feeEstimator. If it is not nil, the mempool
	// records all new transactions it observes into the feeEstimator.
	FeeEstimator *FeeEstimator

	// DAG is the BlockDAG we want to use (mainly for UTXO checks)
	DAG *blockdag.BlockDAG
}

// Policy houses the policy (configuration parameters) which is used to
// control the mempool.
type Policy struct {
	// MaxTxVersion is the transaction version that the mempool should
	// accept.  All transactions above this version are rejected as
	// non-standard.
	MaxTxVersion int32

	// DisableRelayPriority defines whether to relay free or low-fee
	// transactions that do not have enough priority to be relayed.
	DisableRelayPriority bool

	// AcceptNonStd defines whether to accept non-standard transactions. If
	// true, non-standard transactions will be accepted into the mempool.
	// Otherwise, all non-standard transactions will be rejected.
	AcceptNonStd bool

	// FreeTxRelayLimit defines the given amount in thousands of bytes
	// per minute that transactions with no fee are rate limited to.
	FreeTxRelayLimit float64

	// MaxOrphanTxs is the maximum number of orphan transactions
	// that can be queued.
	MaxOrphanTxs int

	// MaxOrphanTxSize is the maximum size allowed for orphan transactions.
	// This helps prevent memory exhaustion attacks from sending a lot of
	// of big orphans.
	MaxOrphanTxSize int

	// MaxSigOpsPerTx is the maximum number of signature operations
	// in a single transaction we will relay or mine.  It is a fraction
	// of the max signature operations for a block.
	MaxSigOpsPerTx int

	// MinRelayTxFee defines the minimum transaction fee in BTC/kB to be
	// considered a non-zero fee.
	MinRelayTxFee util.Amount
}

// TxDesc is a descriptor containing a transaction in the mempool along with
// additional metadata.
type TxDesc struct {
	mining.TxDesc

	// StartingPriority is the priority of the transaction when it was added
	// to the pool.
	StartingPriority float64

	// depCount is not 0 for dependent transaction. Dependent transaction is
	// one that is accepted to pool, but cannot be mined in next block because it
	// depends on outputs of accepted, but still not mined transaction
	depCount int
}

// orphanTx is normal transaction that references an ancestor transaction
// that is not yet available.  It also contains additional information related
// to it such as an expiration time to help prevent caching the orphan forever.
type orphanTx struct {
	tx         *util.Tx
	tag        Tag
	expiration time.Time
}

// TxPool is used as a source of transactions that need to be mined into blocks
// and relayed to other peers.  It is safe for concurrent access from multiple
// peers.
type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated

	mtx           sync.RWMutex
	cfg           Config
	pool          map[daghash.TxID]*TxDesc
	depends       map[daghash.TxID]*TxDesc
	dependsByPrev map[wire.OutPoint]map[daghash.TxID]*TxDesc
	orphans       map[daghash.TxID]*orphanTx
	orphansByPrev map[wire.OutPoint]map[daghash.TxID]*util.Tx
	outpoints     map[wire.OutPoint]*util.Tx
	pennyTotal    float64 // exponentially decaying total for penny spends.
	lastPennyUnix int64   // unix time of last ``penny spend''

	// nextExpireScan is the time after which the orphan pool will be
	// scanned in order to evict orphans.  This is NOT a hard deadline as
	// the scan will only run when an orphan is added to the pool as opposed
	// to on an unconditional timer.
	nextExpireScan time.Time

	mpUTXOSet blockdag.UTXOSet
}

// Ensure the TxPool type implements the mining.TxSource interface.
var _ mining.TxSource = (*TxPool)(nil)

// removeOrphan is the internal function which implements the public
// RemoveOrphan.  See the comment for RemoveOrphan for more details.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) removeOrphan(tx *util.Tx, removeRedeemers bool) {
	// Nothing to do if passed tx is not an orphan.
	txID := tx.ID()
	otx, exists := mp.orphans[*txID]
	if !exists {
		return
	}

	// Remove the reference from the previous orphan index.
	for _, txIn := range otx.tx.MsgTx().TxIn {
		orphans, exists := mp.orphansByPrev[txIn.PreviousOutPoint]
		if exists {
			delete(orphans, *txID)

			// Remove the map entry altogether if there are no
			// longer any orphans which depend on it.
			if len(orphans) == 0 {
				delete(mp.orphansByPrev, txIn.PreviousOutPoint)
			}
		}
	}

	// Remove any orphans that redeem outputs from this one if requested.
	if removeRedeemers {
		prevOut := wire.OutPoint{TxID: *txID}
		for txOutIdx := range tx.MsgTx().TxOut {
			prevOut.Index = uint32(txOutIdx)
			for _, orphan := range mp.orphansByPrev[prevOut] {
				mp.removeOrphan(orphan, true)
			}
		}
	}

	// Remove the transaction from the orphan pool.
	delete(mp.orphans, *txID)
}

// RemoveOrphan removes the passed orphan transaction from the orphan pool and
// previous orphan index.
//
// This function is safe for concurrent access.
func (mp *TxPool) RemoveOrphan(tx *util.Tx) {
	mp.mtx.Lock()
	mp.removeOrphan(tx, false)
	mp.mtx.Unlock()
}

// RemoveOrphansByTag removes all orphan transactions tagged with the provided
// identifier.
//
// This function is safe for concurrent access.
func (mp *TxPool) RemoveOrphansByTag(tag Tag) uint64 {
	var numEvicted uint64
	mp.mtx.Lock()
	for _, otx := range mp.orphans {
		if otx.tag == tag {
			mp.removeOrphan(otx.tx, true)
			numEvicted++
		}
	}
	mp.mtx.Unlock()
	return numEvicted
}

// limitNumOrphans limits the number of orphan transactions by evicting a random
// orphan if adding a new one would cause it to overflow the max allowed.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) limitNumOrphans() error {
	// Scan through the orphan pool and remove any expired orphans when it's
	// time.  This is done for efficiency so the scan only happens
	// periodically instead of on every orphan added to the pool.
	if now := time.Now(); now.After(mp.nextExpireScan) {
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
	if len(mp.orphans)+1 <= mp.cfg.Policy.MaxOrphanTxs {
		return nil
	}

	// Remove a random entry from the map.  For most compilers, Go's
	// range statement iterates starting at a random item although
	// that is not 100% guaranteed by the spec.  The iteration order
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
func (mp *TxPool) addOrphan(tx *util.Tx, tag Tag) {
	// Nothing to do if no orphans are allowed.
	if mp.cfg.Policy.MaxOrphanTxs <= 0 {
		return
	}

	// Limit the number orphan transactions to prevent memory exhaustion.
	// This will periodically remove any expired orphans and evict a random
	// orphan if space is still needed.
	mp.limitNumOrphans()

	mp.orphans[*tx.ID()] = &orphanTx{
		tx:         tx,
		tag:        tag,
		expiration: time.Now().Add(orphanTTL),
	}
	for _, txIn := range tx.MsgTx().TxIn {
		if _, exists := mp.orphansByPrev[txIn.PreviousOutPoint]; !exists {
			mp.orphansByPrev[txIn.PreviousOutPoint] =
				make(map[daghash.TxID]*util.Tx)
		}
		mp.orphansByPrev[txIn.PreviousOutPoint][*tx.ID()] = tx
	}

	log.Debugf("Stored orphan transaction %s (total: %d)", tx.ID(),
		len(mp.orphans))
}

// maybeAddOrphan potentially adds an orphan to the orphan pool.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) maybeAddOrphan(tx *util.Tx, tag Tag) error {
	// Ignore orphan transactions that are too large.  This helps avoid
	// a memory exhaustion attack based on sending a lot of really large
	// orphans.  In the case there is a valid transaction larger than this,
	// it will ultimtely be rebroadcast after the parent transactions
	// have been mined or otherwise received.
	//
	// Note that the number of orphan transactions in the orphan pool is
	// also limited, so this equates to a maximum memory used of
	// mp.cfg.Policy.MaxOrphanTxSize * mp.cfg.Policy.MaxOrphanTxs (which is ~5MB
	// using the default values at the time this comment was written).
	serializedLen := tx.MsgTx().SerializeSize()
	if serializedLen > mp.cfg.Policy.MaxOrphanTxSize {
		str := fmt.Sprintf("orphan transaction size of %d bytes is "+
			"larger than max allowed size of %d bytes",
			serializedLen, mp.cfg.Policy.MaxOrphanTxSize)
		return txRuleError(wire.RejectNonstandard, str)
	}

	// Add the orphan if the none of the above disqualified it.
	mp.addOrphan(tx, tag)

	return nil
}

// removeOrphanDoubleSpends removes all orphans which spend outputs spent by the
// passed transaction from the orphan pool.  Removing those orphans then leads
// to removing all orphans which rely on them, recursively.  This is necessary
// when a transaction is added to the main pool because it may spend outputs
// that orphans also spend.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) removeOrphanDoubleSpends(tx *util.Tx) {
	msgTx := tx.MsgTx()
	for _, txIn := range msgTx.TxIn {
		for _, orphan := range mp.orphansByPrev[txIn.PreviousOutPoint] {
			mp.removeOrphan(orphan, true)
		}
	}
}

// isTransactionInPool returns whether or not the passed transaction already
// exists in the main pool.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *TxPool) isTransactionInPool(hash *daghash.TxID) bool {
	if _, exists := mp.pool[*hash]; exists {
		return true
	}
	return mp.isInDependPool(hash)
}

// IsTransactionInPool returns whether or not the passed transaction already
// exists in the main pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) IsTransactionInPool(hash *daghash.TxID) bool {
	// Protect concurrent access.
	mp.mtx.RLock()
	inPool := mp.isTransactionInPool(hash)
	mp.mtx.RUnlock()

	return inPool
}

// isInDependPool returns whether or not the passed transaction already
// exists in the depend pool.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *TxPool) isInDependPool(hash *daghash.TxID) bool {
	if _, exists := mp.depends[*hash]; exists {
		return true
	}

	return false
}

// IsInDependPool returns whether or not the passed transaction already
// exists in the main pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) IsInDependPool(hash *daghash.TxID) bool {
	// Protect concurrent access.
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return mp.isInDependPool(hash)
}

// isOrphanInPool returns whether or not the passed transaction already exists
// in the orphan pool.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *TxPool) isOrphanInPool(hash *daghash.TxID) bool {
	if _, exists := mp.orphans[*hash]; exists {
		return true
	}

	return false
}

// IsOrphanInPool returns whether or not the passed transaction already exists
// in the orphan pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) IsOrphanInPool(hash *daghash.TxID) bool {
	// Protect concurrent access.
	mp.mtx.RLock()
	inPool := mp.isOrphanInPool(hash)
	mp.mtx.RUnlock()

	return inPool
}

// haveTransaction returns whether or not the passed transaction already exists
// in the main pool or in the orphan pool.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *TxPool) haveTransaction(hash *daghash.TxID) bool {
	return mp.isTransactionInPool(hash) || mp.isOrphanInPool(hash)
}

// HaveTransaction returns whether or not the passed transaction already exists
// in the main pool or in the orphan pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) HaveTransaction(hash *daghash.TxID) bool {
	// Protect concurrent access.
	mp.mtx.RLock()
	haveTx := mp.haveTransaction(hash)
	mp.mtx.RUnlock()

	return haveTx
}

// removeTransaction is the internal function which implements the public
// RemoveTransaction.  See the comment for RemoveTransaction for more details.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) removeTransaction(tx *util.Tx, removeRedeemers bool, restoreInputs bool) error {
	txID := tx.ID()
	if removeRedeemers {
		// Remove any transactions which rely on this one.
		for i := uint32(0); i < uint32(len(tx.MsgTx().TxOut)); i++ {
			prevOut := wire.OutPoint{TxID: *txID, Index: i}
			if txRedeemer, exists := mp.outpoints[prevOut]; exists {
				mp.removeTransaction(txRedeemer, true, false)
			}
		}
	}

	// Remove the transaction if needed.
	if txDesc, exists := mp.fetchTransaction(txID); exists {
		// Remove unconfirmed address index entries associated with the
		// transaction if enabled.
		if mp.cfg.AddrIndex != nil {
			mp.cfg.AddrIndex.RemoveUnconfirmedTx(txID)
		}

		diff := blockdag.NewUTXODiff()
		diff.RemoveTxOuts(txDesc.Tx.MsgTx())

		// Mark the referenced outpoints as unspent by the pool.
		for _, txIn := range txDesc.Tx.MsgTx().TxIn {
			if restoreInputs {
				if prevTxDesc, exists := mp.pool[txIn.PreviousOutPoint.TxID]; exists {
					prevOut := prevTxDesc.Tx.MsgTx().TxOut[txIn.PreviousOutPoint.Index]
					entry := blockdag.NewUTXOEntry(prevOut, false, mining.UnminedHeight)
					diff.AddEntry(txIn.PreviousOutPoint, entry)
				}
				if prevTxDesc, exists := mp.depends[txIn.PreviousOutPoint.TxID]; exists {
					prevOut := prevTxDesc.Tx.MsgTx().TxOut[txIn.PreviousOutPoint.Index]
					entry := blockdag.NewUTXOEntry(prevOut, false, mining.UnminedHeight)
					diff.AddEntry(txIn.PreviousOutPoint, entry)
				}
			}
			delete(mp.outpoints, txIn.PreviousOutPoint)
		}

		if txDesc.depCount == 0 {
			delete(mp.pool, *txID)
		} else {
			delete(mp.depends, *txID)
		}

		// Process dependent transactions
		prevOut := wire.OutPoint{TxID: *txID}
		for txOutIdx := range tx.MsgTx().TxOut {
			// Skip to the next available output if there are none.
			prevOut.Index = uint32(txOutIdx)
			depends, exists := mp.dependsByPrev[prevOut]
			if !exists {
				continue
			}

			// Move independent transactions into main pool
			for _, txD := range depends {
				txD.depCount--
				if txD.depCount == 0 {
					// Transaction may be already removed by recursive calls, if removeRedeemers is true.
					// So avoid moving it into main pool
					if _, ok := mp.depends[*txD.Tx.ID()]; ok {
						delete(mp.depends, *txD.Tx.ID())
						mp.pool[*txD.Tx.ID()] = txD
					}
				}
			}
			delete(mp.dependsByPrev, prevOut)
		}

		var err error
		mp.mpUTXOSet, err = mp.mpUTXOSet.WithDiff(diff)
		if err != nil {
			return err
		}
		atomic.StoreInt64(&mp.lastUpdated, time.Now().Unix())
	}
	return nil
}

// RemoveTransaction removes the passed transaction from the mempool. When the
// removeRedeemers flag is set, any transactions that redeem outputs from the
// removed transaction will also be removed recursively from the mempool, as
// they would otherwise become orphans.
//
// This function is safe for concurrent access.
func (mp *TxPool) RemoveTransaction(tx *util.Tx, removeRedeemers bool, restoreInputs bool) error {
	// Protect concurrent access.
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	return mp.removeTransaction(tx, removeRedeemers, restoreInputs)
}

// RemoveDoubleSpends removes all transactions which spend outputs spent by the
// passed transaction from the memory pool.  Removing those transactions then
// leads to removing all transactions which rely on them, recursively.  This is
// necessary when a block is connected to the main chain because the block may
// contain transactions which were previously unknown to the memory pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) RemoveDoubleSpends(tx *util.Tx) {
	// Protect concurrent access.
	mp.mtx.Lock()
	for _, txIn := range tx.MsgTx().TxIn {
		if txRedeemer, ok := mp.outpoints[txIn.PreviousOutPoint]; ok {
			if !txRedeemer.ID().IsEqual(tx.ID()) {
				mp.removeTransaction(txRedeemer, true, false)
			}
		}
	}
	mp.mtx.Unlock()
}

// addTransaction adds the passed transaction to the memory pool.  It should
// not be called directly as it doesn't perform any validation.  This is a
// helper for maybeAcceptTransaction.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) addTransaction(tx *util.Tx, height int32, fee uint64, parentsInPool []*wire.OutPoint) *TxDesc {
	mp.cfg.DAG.RLock()
	defer mp.cfg.DAG.RUnlock()
	// Add the transaction to the pool and mark the referenced outpoints
	// as spent by the pool.
	txD := &TxDesc{
		TxDesc: mining.TxDesc{
			Tx:       tx,
			Added:    time.Now(),
			Height:   height,
			Fee:      fee,
			FeePerKB: fee * 1000 / uint64(tx.MsgTx().SerializeSize()),
		},
		StartingPriority: mining.CalcPriority(tx.MsgTx(), mp.mpUTXOSet, height),
		depCount:         len(parentsInPool),
	}

	if len(parentsInPool) == 0 {
		mp.pool[*tx.ID()] = txD
	} else {
		mp.depends[*tx.ID()] = txD
		for _, previousOutPoint := range parentsInPool {
			if _, exists := mp.dependsByPrev[*previousOutPoint]; !exists {
				mp.dependsByPrev[*previousOutPoint] = make(map[daghash.TxID]*TxDesc)
			}
			mp.dependsByPrev[*previousOutPoint][*tx.ID()] = txD
		}
	}

	for _, txIn := range tx.MsgTx().TxIn {
		mp.outpoints[txIn.PreviousOutPoint] = tx
	}
	mp.mpUTXOSet.AddTx(tx.MsgTx(), mining.UnminedHeight)
	atomic.StoreInt64(&mp.lastUpdated, time.Now().Unix())

	// Add unconfirmed address index entries associated with the transaction
	// if enabled.
	if mp.cfg.AddrIndex != nil {
		mp.cfg.AddrIndex.AddUnconfirmedTx(tx, mp.mpUTXOSet)
	}

	// Record this tx for fee estimation if enabled.
	if mp.cfg.FeeEstimator != nil {
		mp.cfg.FeeEstimator.ObserveTransaction(txD)
	}

	return txD
}

// checkPoolDoubleSpend checks whether or not the passed transaction is
// attempting to spend coins already spent by other transactions in the pool.
// Note it does not check for double spends against transactions already in the
// DAG.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *TxPool) checkPoolDoubleSpend(tx *util.Tx) error {
	for _, txIn := range tx.MsgTx().TxIn {
		if txR, exists := mp.outpoints[txIn.PreviousOutPoint]; exists {
			str := fmt.Sprintf("output %s already spent by "+
				"transaction %s in the memory pool",
				txIn.PreviousOutPoint, txR.ID())
			return txRuleError(wire.RejectDuplicate, str)
		}
	}

	return nil
}

// CheckSpend checks whether the passed outpoint is already spent by a
// transaction in the mempool. If that's the case the spending transaction will
// be returned, if not nil will be returned.
func (mp *TxPool) CheckSpend(op wire.OutPoint) *util.Tx {
	mp.mtx.RLock()
	txR := mp.outpoints[op]
	mp.mtx.RUnlock()

	return txR
}

// This function MUST be called with the mempool lock held (for reads).
func (mp *TxPool) fetchTransaction(txID *daghash.TxID) (*TxDesc, bool) {
	txDesc, exists := mp.pool[*txID]
	if !exists {
		txDesc, exists = mp.depends[*txID]
	}
	return txDesc, exists
}

// FetchTransaction returns the requested transaction from the transaction pool.
// This only fetches from the main transaction pool and does not include
// orphans.
//
// This function is safe for concurrent access.
func (mp *TxPool) FetchTransaction(txID *daghash.TxID) (*util.Tx, error) {
	// Protect concurrent access.
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	if txDesc, exists := mp.fetchTransaction(txID); exists {
		return txDesc.Tx, nil
	}

	return nil, fmt.Errorf("transaction is not in the pool")
}

// maybeAcceptTransaction is the internal function which implements the public
// MaybeAcceptTransaction.  See the comment for MaybeAcceptTransaction for
// more details.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) maybeAcceptTransaction(tx *util.Tx, isNew, rateLimit, rejectDupOrphans bool) ([]*daghash.TxID, *TxDesc, error) {
	mp.cfg.DAG.RLock()
	defer mp.cfg.DAG.RUnlock()
	txID := tx.ID()

	// Don't accept the transaction if it already exists in the pool.  This
	// applies to orphan transactions as well when the reject duplicate
	// orphans flag is set.  This check is intended to be a quick check to
	// weed out duplicates.
	if mp.isTransactionInPool(txID) || (rejectDupOrphans &&
		mp.isOrphanInPool(txID)) {

		str := fmt.Sprintf("already have transaction %s", txID)
		return nil, nil, txRuleError(wire.RejectDuplicate, str)
	}

	// Don't accept the transaction if it's from an incompatible subnetwork.
	subnetworkID := mp.cfg.DAG.SubnetworkID()
	if !tx.MsgTx().IsSubnetworkCompatible(subnetworkID) {
		str := fmt.Sprintf("tx %s belongs to an invalid subnetwork %s, DAG subnetwork %s", tx.ID(),
			tx.MsgTx().SubnetworkID, subnetworkID)
		return nil, nil, txRuleError(wire.RejectInvalid, str)
	}

	// Perform preliminary sanity checks on the transaction.  This makes
	// use of blockDAG which contains the invariant rules for what
	// transactions are allowed into blocks.
	err := blockdag.CheckTransactionSanity(tx, subnetworkID, false)
	if err != nil {
		if cerr, ok := err.(blockdag.RuleError); ok {
			return nil, nil, dagRuleError(cerr)
		}
		return nil, nil, err
	}

	// Check that transaction does not overuse GAS
	msgTx := tx.MsgTx()
	if !msgTx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) {
		gasLimit, err := mp.cfg.DAG.SubnetworkStore.GasLimit(&msgTx.SubnetworkID)
		if err != nil {
			return nil, nil, err
		}
		if msgTx.Gas > gasLimit {
			str := fmt.Sprintf("transaction wants more gas %d, than allowed %d",
				msgTx.Gas, gasLimit)
			return nil, nil, dagRuleError(blockdag.RuleError{
				ErrorCode:   blockdag.ErrInvalidGas,
				Description: str})
		}
	}

	// A standalone transaction must not be a block reward transaction.
	if blockdag.IsBlockReward(tx) {
		str := fmt.Sprintf("transaction %s is an individual block reward transaction",
			txID)
		return nil, nil, txRuleError(wire.RejectInvalid, str)
	}

	// Get the current height of the main chain.  A standalone transaction
	// will be mined into the next block at best, so its height is at least
	// one more than the current height.
	bestHeight := mp.cfg.BestHeight()
	nextBlockHeight := bestHeight + 1

	medianTimePast := mp.cfg.MedianTimePast()

	// Don't allow non-standard transactions if the network parameters
	// forbid their acceptance.
	if !mp.cfg.Policy.AcceptNonStd {
		err = checkTransactionStandard(tx, nextBlockHeight,
			medianTimePast, &mp.cfg.Policy)
		if err != nil {
			// Attempt to extract a reject code from the error so
			// it can be retained.  When not possible, fall back to
			// a non standard error.
			rejectCode, found := extractRejectCode(err)
			if !found {
				rejectCode = wire.RejectNonstandard
			}
			str := fmt.Sprintf("transaction %s is not standard: %s",
				txID, err)
			return nil, nil, txRuleError(rejectCode, str)
		}
	}

	// The transaction may not use any of the same outputs as other
	// transactions already in the pool as that would ultimately result in a
	// double spend.  This check is intended to be quick and therefore only
	// detects double spends within the transaction pool itself.  The
	// transaction could still be double spending coins from the main chain
	// at this point.  There is a more in-depth check that happens later
	// after fetching the referenced transaction inputs from the main chain
	// which examines the actual spend data and prevents double spends.
	err = mp.checkPoolDoubleSpend(tx)
	if err != nil {
		return nil, nil, err
	}

	// Don't allow the transaction if it exists in the DAG and is
	// not already fully spent.
	prevOut := wire.OutPoint{TxID: *txID}
	for txOutIdx := range tx.MsgTx().TxOut {
		prevOut.Index = uint32(txOutIdx)
		_, ok := mp.mpUTXOSet.Get(prevOut)
		if ok {
			return nil, nil, txRuleError(wire.RejectDuplicate,
				"transaction already exists")
		}
	}

	// Transaction is an orphan if any of the referenced transaction outputs
	// don't exist or are already spent.  Adding orphans to the orphan pool
	// is not handled by this function, and the caller should use
	// maybeAddOrphan if this behavior is desired.
	var missingParents []*daghash.TxID
	var parentsInPool []*wire.OutPoint
	for _, txIn := range tx.MsgTx().TxIn {
		if _, ok := mp.mpUTXOSet.Get(txIn.PreviousOutPoint); !ok {
			// Must make a copy of the hash here since the iterator
			// is replaced and taking its address directly would
			// result in all of the entries pointing to the same
			// memory location and thus all be the final hash.
			hashCopy := txIn.PreviousOutPoint.TxID
			missingParents = append(missingParents, &hashCopy)
		}
		if mp.isTransactionInPool(&txIn.PreviousOutPoint.TxID) {
			parentsInPool = append(parentsInPool, &txIn.PreviousOutPoint)
		}
	}
	if len(missingParents) > 0 {
		return missingParents, nil, nil
	}

	// Don't allow the transaction into the mempool unless its sequence
	// lock is active, meaning that it'll be allowed into the next block
	// with respect to its defined relative lock times.
	sequenceLock, err := mp.cfg.CalcSequenceLockNoLock(tx, mp.mpUTXOSet)
	if err != nil {
		if cerr, ok := err.(blockdag.RuleError); ok {
			return nil, nil, dagRuleError(cerr)
		}
		return nil, nil, err
	}
	if !blockdag.SequenceLockActive(sequenceLock, nextBlockHeight,
		medianTimePast) {
		return nil, nil, txRuleError(wire.RejectNonstandard,
			"transaction's sequence locks on inputs not met")
	}

	// Perform several checks on the transaction inputs using the invariant
	// rules in blockchain for what transactions are allowed into blocks.
	// Also returns the fees associated with the transaction which will be
	// used later.
	txFee, err := blockdag.CheckTransactionInputsAndCalulateFee(tx, nextBlockHeight,
		mp.mpUTXOSet, mp.cfg.DAGParams, false)
	if err != nil {
		if cerr, ok := err.(blockdag.RuleError); ok {
			return nil, nil, dagRuleError(cerr)
		}
		return nil, nil, err
	}

	// Don't allow transactions with non-standard inputs if the network
	// parameters forbid their acceptance.
	if !mp.cfg.Policy.AcceptNonStd {
		err := checkInputsStandard(tx, mp.mpUTXOSet)
		if err != nil {
			// Attempt to extract a reject code from the error so
			// it can be retained.  When not possible, fall back to
			// a non standard error.
			rejectCode, found := extractRejectCode(err)
			if !found {
				rejectCode = wire.RejectNonstandard
			}
			str := fmt.Sprintf("transaction %s has a non-standard "+
				"input: %s", txID, err)
			return nil, nil, txRuleError(rejectCode, str)
		}
	}

	// NOTE: if you modify this code to accept non-standard transactions,
	// you should add code here to check that the transaction does a
	// reasonable number of ECDSA signature verifications.

	// Don't allow transactions with an excessive number of signature
	// operations which would result in making it impossible to mine.  Since
	// the coinbase address itself can contain signature operations, the
	// maximum allowed signature operations per transaction is less than
	// the maximum allowed signature operations per block.
	sigOpCount, err := blockdag.CountP2SHSigOps(tx, false, mp.mpUTXOSet)
	if err != nil {
		if cerr, ok := err.(blockdag.RuleError); ok {
			return nil, nil, dagRuleError(cerr)
		}
		return nil, nil, err
	}
	if sigOpCount > mp.cfg.Policy.MaxSigOpsPerTx {
		str := fmt.Sprintf("transaction %s sigop count is too high: %d > %d",
			txID, sigOpCount, mp.cfg.Policy.MaxSigOpsPerTx)
		return nil, nil, txRuleError(wire.RejectNonstandard, str)
	}

	// Don't allow transactions with fees too low to get into a mined block.
	//
	// Most miners allow a free transaction area in blocks they mine to go
	// alongside the area used for high-priority transactions as well as
	// transactions with fees.  A transaction size of up to 1000 bytes is
	// considered safe to go into this section.  Further, the minimum fee
	// calculated below on its own would encourage several small
	// transactions to avoid fees rather than one single larger transaction
	// which is more desirable.  Therefore, as long as the size of the
	// transaction does not exceeed 1000 less than the reserved space for
	// high-priority transactions, don't require a fee for it.
	serializedSize := int64(tx.MsgTx().SerializeSize())
	minFee := uint64(calcMinRequiredTxRelayFee(serializedSize,
		mp.cfg.Policy.MinRelayTxFee))
	if serializedSize >= (DefaultBlockPrioritySize-1000) && txFee < minFee {
		str := fmt.Sprintf("transaction %s has %d fees which is under "+
			"the required amount of %d", txID, txFee,
			minFee)
		return nil, nil, txRuleError(wire.RejectInsufficientFee, str)
	}

	// Require that free transactions have sufficient priority to be mined
	// in the next block.  Transactions which are being added back to the
	// memory pool from blocks that have been disconnected during a reorg
	// are exempted.
	if isNew && !mp.cfg.Policy.DisableRelayPriority && txFee < minFee {
		currentPriority := mining.CalcPriority(tx.MsgTx(), mp.mpUTXOSet,
			nextBlockHeight)
		if currentPriority <= mining.MinHighPriority {
			str := fmt.Sprintf("transaction %s has insufficient "+
				"priority (%g <= %g)", txID,
				currentPriority, mining.MinHighPriority)
			return nil, nil, txRuleError(wire.RejectInsufficientFee, str)
		}
	}

	// Free-to-relay transactions are rate limited here to prevent
	// penny-flooding with tiny transactions as a form of attack.
	if rateLimit && txFee < minFee {
		nowUnix := time.Now().Unix()
		// Decay passed data with an exponentially decaying ~10 minute
		// window - matches bitcoind handling.
		mp.pennyTotal *= math.Pow(1.0-1.0/600.0,
			float64(nowUnix-mp.lastPennyUnix))
		mp.lastPennyUnix = nowUnix

		// Are we still over the limit?
		if mp.pennyTotal >= mp.cfg.Policy.FreeTxRelayLimit*10*1000 {
			str := fmt.Sprintf("transaction %s has been rejected "+
				"by the rate limiter due to low fees", txID)
			return nil, nil, txRuleError(wire.RejectInsufficientFee, str)
		}
		oldTotal := mp.pennyTotal

		mp.pennyTotal += float64(serializedSize)
		log.Tracef("rate limit: curTotal %d, nextTotal: %d, "+
			"limit %d", oldTotal, mp.pennyTotal,
			mp.cfg.Policy.FreeTxRelayLimit*10*1000)
	}

	// Verify crypto signatures for each input and reject the transaction if
	// any don't verify.
	err = blockdag.ValidateTransactionScripts(tx, mp.mpUTXOSet,
		txscript.StandardVerifyFlags, mp.cfg.SigCache)
	if err != nil {
		if cerr, ok := err.(blockdag.RuleError); ok {
			return nil, nil, dagRuleError(cerr)
		}
		return nil, nil, err
	}

	// Add to transaction pool.
	txD := mp.addTransaction(tx, bestHeight, txFee, parentsInPool)

	log.Debugf("Accepted transaction %s (pool size: %d)", txID,
		len(mp.pool))

	return nil, txD, nil
}

// MaybeAcceptTransaction is the main workhorse for handling insertion of new
// free-standing transactions into a memory pool.  It includes functionality
// such as rejecting duplicate transactions, ensuring transactions follow all
// rules, detecting orphan transactions, and insertion into the memory pool.
//
// If the transaction is an orphan (missing parent transactions), the
// transaction is NOT added to the orphan pool, but each unknown referenced
// parent is returned.  Use ProcessTransaction instead if new orphans should
// be added to the orphan pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) MaybeAcceptTransaction(tx *util.Tx, isNew, rateLimit bool) ([]*daghash.TxID, *TxDesc, error) {
	// Protect concurrent access.
	mp.mtx.Lock()
	hashes, txD, err := mp.maybeAcceptTransaction(tx, isNew, rateLimit, true)
	mp.mtx.Unlock()

	return hashes, txD, err
}

// processOrphans is the internal function which implements the public
// ProcessOrphans.  See the comment for ProcessOrphans for more details.
//
// This function MUST be called with the mempool lock held (for writes).
func (mp *TxPool) processOrphans(acceptedTx *util.Tx) []*TxDesc {
	var acceptedTxns []*TxDesc

	// Start with processing at least the passed transaction.
	processList := list.New()
	processList.PushBack(acceptedTx)
	for processList.Len() > 0 {
		// Pop the transaction to process from the front of the list.
		firstElement := processList.Remove(processList.Front())
		processItem := firstElement.(*util.Tx)

		prevOut := wire.OutPoint{TxID: *processItem.ID()}
		for txOutIdx := range processItem.MsgTx().TxOut {
			// Look up all orphans that redeem the output that is
			// now available.  This will typically only be one, but
			// it could be multiple if the orphan pool contains
			// double spends.  While it may seem odd that the orphan
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
					tx, true, true, false)
				if err != nil {
					// The orphan is now invalid, so there
					// is no way any other orphans which
					// redeem any of its outputs can be
					// accepted.  Remove them.
					mp.removeOrphan(tx, true)
					break
				}

				// Transaction is still an orphan.  Try the next
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
	for _, txD := range acceptedTxns {
		mp.removeOrphanDoubleSpends(txD.Tx)
	}

	return acceptedTxns
}

// ProcessOrphans determines if there are any orphans which depend on the passed
// transaction hash (it is possible that they are no longer orphans) and
// potentially accepts them to the memory pool.  It repeats the process for the
// newly accepted transactions (to detect further orphans which may no longer be
// orphans) until there are no more.
//
// It returns a slice of transactions added to the mempool.  A nil slice means
// no transactions were moved from the orphan pool to the mempool.
//
// This function is safe for concurrent access.
func (mp *TxPool) ProcessOrphans(acceptedTx *util.Tx) []*TxDesc {
	mp.mtx.Lock()
	acceptedTxns := mp.processOrphans(acceptedTx)
	mp.mtx.Unlock()

	return acceptedTxns
}

// ProcessTransaction is the main workhorse for handling insertion of new
// free-standing transactions into the memory pool.  It includes functionality
// such as rejecting duplicate transactions, ensuring transactions follow all
// rules, orphan transaction handling, and insertion into the memory pool.
//
// It returns a slice of transactions added to the mempool.  When the
// error is nil, the list will include the passed transaction itself along
// with any additional orphan transaactions that were added as a result of
// the passed one being accepted.
//
// This function is safe for concurrent access.
func (mp *TxPool) ProcessTransaction(tx *util.Tx, allowOrphan, rateLimit bool, tag Tag) ([]*TxDesc, error) {
	log.Tracef("Processing transaction %s", tx.ID())

	// Protect concurrent access.
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	// Potentially accept the transaction to the memory pool.
	missingParents, txD, err := mp.maybeAcceptTransaction(tx, true, rateLimit,
		true)
	if err != nil {
		return nil, err
	}

	if len(missingParents) == 0 {
		// Accept any orphan transactions that depend on this
		// transaction (they may no longer be orphans if all inputs
		// are now available) and repeat for those accepted
		// transactions until there are no more.
		newTxs := mp.processOrphans(tx)
		acceptedTxs := make([]*TxDesc, len(newTxs)+1)

		// Add the parent transaction first so remote nodes
		// do not add orphans.
		acceptedTxs[0] = txD
		copy(acceptedTxs[1:], newTxs)

		return acceptedTxs, nil
	}

	// The transaction is an orphan (has inputs missing).  Reject
	// it if the flag to allow orphans is not set.
	if !allowOrphan {
		// Only use the first missing parent transaction in
		// the error message.
		//
		// NOTE: RejectDuplicate is really not an accurate
		// reject code here, but it matches the reference
		// implementation and there isn't a better choice due
		// to the limited number of reject codes.  Missing
		// inputs is assumed to mean they are already spent
		// which is not really always the case.
		str := fmt.Sprintf("orphan transaction %s references "+
			"outputs of unknown or fully-spent "+
			"transaction %s", tx.ID(), missingParents[0])
		return nil, txRuleError(wire.RejectDuplicate, str)
	}

	// Potentially add the orphan transaction to the orphan pool.
	err = mp.maybeAddOrphan(tx, tag)
	return nil, err
}

// Count returns the number of transactions in the main pool.  It does not
// include the orphan pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) Count() int {
	mp.mtx.RLock()
	count := len(mp.pool)
	mp.mtx.RUnlock()

	return count
}

// DepCount returns the number of dependent transactions in the main pool.  It does not
// include the orphan pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) DepCount() int {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return len(mp.depends)
}

// TxIDs returns a slice of IDs for all of the transactions in the memory
// pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) TxIDs() []*daghash.TxID {
	mp.mtx.RLock()
	ids := make([]*daghash.TxID, len(mp.pool))
	i := 0
	for txID := range mp.pool {
		idCopy := txID
		ids[i] = &idCopy
		i++
	}
	mp.mtx.RUnlock()

	return ids
}

// TxDescs returns a slice of descriptors for all the transactions in the pool.
// The descriptors are to be treated as read only.
//
// This function is safe for concurrent access.
func (mp *TxPool) TxDescs() []*TxDesc {
	mp.mtx.RLock()
	descs := make([]*TxDesc, len(mp.pool))
	i := 0
	for _, desc := range mp.pool {
		descs[i] = desc
		i++
	}
	mp.mtx.RUnlock()

	return descs
}

// MiningDescs returns a slice of mining descriptors for all the transactions
// in the pool.
//
// This is part of the mining.TxSource interface implementation and is safe for
// concurrent access as required by the interface contract.
func (mp *TxPool) MiningDescs() []*mining.TxDesc {
	mp.mtx.RLock()
	descs := make([]*mining.TxDesc, len(mp.pool))
	i := 0
	for _, desc := range mp.pool {
		descs[i] = &desc.TxDesc
		i++
	}
	mp.mtx.RUnlock()

	return descs
}

// RawMempoolVerbose returns all of the entries in the mempool as a fully
// populated btcjson result.
//
// This function is safe for concurrent access.
func (mp *TxPool) RawMempoolVerbose() map[string]*btcjson.GetRawMempoolVerboseResult {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	result := make(map[string]*btcjson.GetRawMempoolVerboseResult,
		len(mp.pool))
	bestHeight := mp.cfg.BestHeight()

	for _, desc := range mp.pool {
		// Calculate the current priority based on the inputs to
		// the transaction.  Use zero if one or more of the
		// input transactions can't be found for some reason.
		tx := desc.Tx
		currentPriority := mining.CalcPriority(tx.MsgTx(), mp.mpUTXOSet,
			bestHeight+1)

		mpd := &btcjson.GetRawMempoolVerboseResult{
			Size:             int32(tx.MsgTx().SerializeSize()),
			Fee:              util.Amount(desc.Fee).ToBTC(),
			Time:             desc.Added.Unix(),
			Height:           int64(desc.Height),
			StartingPriority: desc.StartingPriority,
			CurrentPriority:  currentPriority,
			Depends:          make([]string, 0),
		}
		for _, txIn := range tx.MsgTx().TxIn {
			hash := &txIn.PreviousOutPoint.TxID
			if mp.haveTransaction(hash) {
				mpd.Depends = append(mpd.Depends,
					hash.String())
			}
		}

		result[tx.ID().String()] = mpd
	}

	return result
}

// LastUpdated returns the last time a transaction was added to or removed from
// the main pool.  It does not include the orphan pool.
//
// This function is safe for concurrent access.
func (mp *TxPool) LastUpdated() time.Time {
	return time.Unix(atomic.LoadInt64(&mp.lastUpdated), 0)
}

// HandleNewBlock removes all the transactions in the new block
// from the mempool and the orphan pool, and it also removes
// from the mempool transactions that double spend a
// transaction that is already in the DAG
func (mp *TxPool) HandleNewBlock(block *util.Block, txChan chan NewBlockMsg) error {

	oldUTXOSet := mp.mpUTXOSet

	// Remove all of the transactions (except the coinbase) in the
	// connected block from the transaction pool.  Secondly, remove any
	// transactions which are now double spends as a result of these
	// new transactions.  Finally, remove any transaction that is
	// no longer an orphan. Transactions which depend on a confirmed
	// transaction are NOT removed recursively because they are still
	// valid.
	for _, tx := range block.Transactions()[1:] {
		err := mp.RemoveTransaction(tx, false, false)
		if err != nil {
			mp.mpUTXOSet = oldUTXOSet
			return err
		}
		mp.RemoveDoubleSpends(tx)
		mp.RemoveOrphan(tx)
		acceptedTxs := mp.ProcessOrphans(tx)
		txChan <- NewBlockMsg{
			AcceptedTxs: acceptedTxs,
			Tx:          tx,
		}
	}
	return nil
}

// New returns a new memory pool for validating and storing standalone
// transactions until they are mined into a block.
func New(cfg *Config) *TxPool {
	virtualUTXO := cfg.DAG.UTXOSet()
	mpUTXO := blockdag.NewDiffUTXOSet(virtualUTXO, blockdag.NewUTXODiff())
	return &TxPool{
		cfg:            *cfg,
		pool:           make(map[daghash.TxID]*TxDesc),
		depends:        make(map[daghash.TxID]*TxDesc),
		dependsByPrev:  make(map[wire.OutPoint]map[daghash.TxID]*TxDesc),
		orphans:        make(map[daghash.TxID]*orphanTx),
		orphansByPrev:  make(map[wire.OutPoint]map[daghash.TxID]*util.Tx),
		nextExpireScan: time.Now().Add(orphanExpireScanInterval),
		outpoints:      make(map[wire.OutPoint]*util.Tx),
		mpUTXOSet:      mpUTXO,
	}
}
