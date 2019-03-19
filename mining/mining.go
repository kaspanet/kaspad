// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"container/heap"
	"fmt"
	"sort"
	"time"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/random"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

const (
	// MinHighPriority is the minimum priority value that allows a
	// transaction to be considered high priority.
	MinHighPriority = util.SatoshiPerBitcoin * 144.0 / 250

	// blockHeaderOverhead is the max number of bytes it takes to serialize
	// a block header and max possible transaction count.
	blockHeaderOverhead = wire.MaxBlockHeaderPayload + wire.MaxVarIntPayload

	// CoinbaseFlags is added to the coinbase script of a generated block
	// and is used to monitor BIP16 support as well as blocks that are
	// generated via btcd.
	CoinbaseFlags = "/P2SH/btcd/"
)

// TxDesc is a descriptor about a transaction in a transaction source along with
// additional metadata.
type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx *util.Tx

	// Added is the time when the entry was added to the source pool.
	Added time.Time

	// Height is the block height when the entry was added to the the source
	// pool.
	Height int32

	// Fee is the total fee the transaction associated with the entry pays.
	Fee uint64

	// FeePerKB is the fee the transaction pays in Satoshi per 1000 bytes.
	FeePerKB uint64
}

// TxSource represents a source of transactions to consider for inclusion in
// new blocks.
//
// The interface contract requires that all of these methods are safe for
// concurrent access with respect to the source.
type TxSource interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(txID *daghash.TxID) bool
}

// txPrioItem houses a transaction along with extra information that allows the
// transaction to be prioritized and track dependencies on other transactions
// which have not been mined into a block yet.
type txPrioItem struct {
	tx       *util.Tx
	fee      uint64
	priority float64
	feePerKB uint64
}

// txPriorityQueueLessFunc describes a function that can be used as a compare
// function for a transaction priority queue (txPriorityQueue).
type txPriorityQueueLessFunc func(*txPriorityQueue, int, int) bool

// txPriorityQueue implements a priority queue of txPrioItem elements that
// supports an arbitrary compare function as defined by txPriorityQueueLessFunc.
type txPriorityQueue struct {
	lessFunc txPriorityQueueLessFunc
	items    []*txPrioItem
}

// Len returns the number of items in the priority queue.  It is part of the
// heap.Interface implementation.
func (pq *txPriorityQueue) Len() int {
	return len(pq.items)
}

// Less returns whether the item in the priority queue with index i should sort
// before the item with index j by deferring to the assigned less function.  It
// is part of the heap.Interface implementation.
func (pq *txPriorityQueue) Less(i, j int) bool {
	return pq.lessFunc(pq, i, j)
}

// Swap swaps the items at the passed indices in the priority queue.  It is
// part of the heap.Interface implementation.
func (pq *txPriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push pushes the passed item onto the priority queue.  It is part of the
// heap.Interface implementation.
func (pq *txPriorityQueue) Push(x interface{}) {
	pq.items = append(pq.items, x.(*txPrioItem))
}

// Pop removes the highest priority item (according to Less) from the priority
// queue and returns it.  It is part of the heap.Interface implementation.
func (pq *txPriorityQueue) Pop() interface{} {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil
	pq.items = pq.items[0 : n-1]
	return item
}

// SetLessFunc sets the compare function for the priority queue to the provided
// function.  It also invokes heap.Init on the priority queue using the new
// function so it can immediately be used with heap.Push/Pop.
func (pq *txPriorityQueue) SetLessFunc(lessFunc txPriorityQueueLessFunc) {
	pq.lessFunc = lessFunc
	heap.Init(pq)
}

// txPQByPriority sorts a txPriorityQueue by transaction priority and then fees
// per kilobyte.
func txPQByPriority(pq *txPriorityQueue, i, j int) bool {
	// Using > here so that pop gives the highest priority item as opposed
	// to the lowest.  Sort by priority first, then fee.
	if pq.items[i].priority == pq.items[j].priority {
		return pq.items[i].feePerKB > pq.items[j].feePerKB
	}
	return pq.items[i].priority > pq.items[j].priority

}

// txPQByFee sorts a txPriorityQueue by fees per kilobyte and then transaction
// priority.
func txPQByFee(pq *txPriorityQueue, i, j int) bool {
	// Using > here so that pop gives the highest fee item as opposed
	// to the lowest.  Sort by fee first, then priority.
	if pq.items[i].feePerKB == pq.items[j].feePerKB {
		return pq.items[i].priority > pq.items[j].priority
	}
	return pq.items[i].feePerKB > pq.items[j].feePerKB
}

// newTxPriorityQueue returns a new transaction priority queue that reserves the
// passed amount of space for the elements.  The new priority queue uses either
// the txPQByPriority or the txPQByFee compare function depending on the
// sortByFee parameter and is already initialized for use with heap.Push/Pop.
// The priority queue can grow larger than the reserved space, but extra copies
// of the underlying array can be avoided by reserving a sane value.
func newTxPriorityQueue(reserve int, sortByFee bool) *txPriorityQueue {
	pq := &txPriorityQueue{
		items: make([]*txPrioItem, 0, reserve),
	}
	if sortByFee {
		pq.SetLessFunc(txPQByFee)
	} else {
		pq.SetLessFunc(txPQByPriority)
	}
	return pq
}

// BlockTemplate houses a block that has yet to be solved along with additional
// details about the fees and the number of signature operations for each
// transaction in the block.
type BlockTemplate struct {
	// Block is a block that is ready to be solved by miners.  Thus, it is
	// completely valid with the exception of satisfying the proof-of-work
	// requirement.
	Block *wire.MsgBlock

	// Fees contains the amount of fees each transaction in the generated
	// template pays in base units.  Since the first transaction is the
	// coinbase, the first entry (offset 0) will contain the negative of the
	// sum of the fees of all other transactions.
	Fees []uint64

	// SigOpCounts contains the number of signature operations each
	// transaction in the generated template performs.
	SigOpCounts []int64

	// Height is the height at which the block template connects to the main
	// chain.
	Height int32

	// ValidPayAddress indicates whether or not the template coinbase pays
	// to an address or is redeemable by anyone.  See the documentation on
	// NewBlockTemplate for details on which this can be useful to generate
	// templates without a coinbase payment address.
	ValidPayAddress bool
}

// StandardCoinbaseScript returns a standard script suitable for use as the
// signature script of the coinbase transaction of a new block.  In particular,
// it starts with the block height that is required by version 2 blocks and adds
// the extra nonce as well as additional coinbase flags.
func StandardCoinbaseScript(nextBlockHeight int32, extraNonce uint64) ([]byte, error) {
	return txscript.NewScriptBuilder().AddInt64(int64(nextBlockHeight)).
		AddInt64(int64(extraNonce)).AddData([]byte(CoinbaseFlags)).
		Script()
}

// CreateCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// based on the passed block height to the provided address.  When the address
// is nil, the coinbase transaction will instead be redeemable by anyone.
//
// See the comment for NewBlockTemplate for more information about why the nil
// address handling is useful.
func CreateCoinbaseTx(params *dagconfig.Params, coinbaseScript []byte, nextBlockHeight int32, addr util.Address) (*util.Tx, error) {
	// Create the script to pay to the provided payment address if one was
	// specified.  Otherwise create a script that allows the coinbase to be
	// redeemable by anyone.
	var pkScript []byte
	if addr != nil {
		var err error
		pkScript, err = txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		scriptBuilder := txscript.NewScriptBuilder()
		pkScript, err = scriptBuilder.AddOp(txscript.OpTrue).Script()
		if err != nil {
			return nil, err
		}
	}

	txIn := &wire.TxIn{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOutPoint: *wire.NewOutPoint(&daghash.TxID{},
			wire.MaxPrevOutIndex),
		SignatureScript: coinbaseScript,
		Sequence:        wire.MaxTxInSequenceNum,
	}
	txOut := &wire.TxOut{
		Value:    blockdag.CalcBlockSubsidy(nextBlockHeight, params),
		PkScript: pkScript,
	}
	return util.NewTx(wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})), nil
}

// MinimumMedianTime returns the minimum allowed timestamp for a block building
// on the end of the provided best chain.  In particular, it is one second after
// the median timestamp of the last several blocks per the chain consensus
// rules.
func MinimumMedianTime(dagMedianTime time.Time) time.Time {
	return dagMedianTime.Add(time.Second)
}

// medianAdjustedTime returns the current time adjusted to ensure it is at least
// one second after the median timestamp of the last several blocks per the
// chain consensus rules.
func medianAdjustedTime(dagMedianTime time.Time, timeSource blockdag.MedianTimeSource) time.Time {
	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks.  Thus, choose the maximum between the
	// current time and one second after the past median time.  The current
	// timestamp is truncated to a second boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// second.
	newTimestamp := timeSource.AdjustedTime()
	minTimestamp := MinimumMedianTime(dagMedianTime)
	if newTimestamp.Before(minTimestamp) {
		newTimestamp = minTimestamp
	}

	return newTimestamp
}

// BlkTmplGenerator provides a type that can be used to generate block templates
// based on a given mining policy and source of transactions to choose from.
// It also houses additional state required in order to ensure the templates
// are built on top of the current best chain and adhere to the consensus rules.
type BlkTmplGenerator struct {
	policy     *Policy
	dagParams  *dagconfig.Params
	txSource   TxSource
	dag        *blockdag.BlockDAG
	timeSource blockdag.MedianTimeSource
	sigCache   *txscript.SigCache
}

// NewBlkTmplGenerator returns a new block template generator for the given
// policy using transactions from the provided transaction source.
//
// The additional state-related fields are required in order to ensure the
// templates are built on top of the current best chain and adhere to the
// consensus rules.
func NewBlkTmplGenerator(policy *Policy, params *dagconfig.Params,
	txSource TxSource, dag *blockdag.BlockDAG,
	timeSource blockdag.MedianTimeSource,
	sigCache *txscript.SigCache) *BlkTmplGenerator {

	return &BlkTmplGenerator{
		policy:     policy,
		dagParams:  params,
		txSource:   txSource,
		dag:        dag,
		timeSource: timeSource,
		sigCache:   sigCache,
	}
}

// NewBlockTemplate returns a new block template that is ready to be solved
// using the transactions from the passed transaction source pool and a coinbase
// that either pays to the passed address if it is not nil, or a coinbase that
// is redeemable by anyone if the passed address is nil.  The nil address
// functionality is useful since there are cases such as the getblocktemplate
// RPC where external mining software is responsible for creating their own
// coinbase which will replace the one generated for the block template.  Thus
// the need to have configured address can be avoided.
//
// The transactions selected and included are prioritized according to several
// factors.  First, each transaction has a priority calculated based on its
// value, age of inputs, and size.  Transactions which consist of larger
// amounts, older inputs, and small sizes have the highest priority.  Second, a
// fee per kilobyte is calculated for each transaction.  Transactions with a
// higher fee per kilobyte are preferred.  Finally, the block generation related
// policy settings are all taken into account.
//
// Transactions which only spend outputs from other transactions already in the
// block chain are immediately added to a priority queue which either
// prioritizes based on the priority (then fee per kilobyte) or the fee per
// kilobyte (then priority) depending on whether or not the BlockPrioritySize
// policy setting allots space for high-priority transactions.  Transactions
// which spend outputs from other transactions in the source pool are added to a
// dependency map so they can be added to the priority queue once the
// transactions they depend on have been included.
//
// Once the high-priority area (if configured) has been filled with
// transactions, or the priority falls below what is considered high-priority,
// the priority queue is updated to prioritize by fees per kilobyte (then
// priority).
//
// When the fees per kilobyte drop below the TxMinFreeFee policy setting, the
// transaction will be skipped unless the BlockMinSize policy setting is
// nonzero, in which case the block will be filled with the low-fee/free
// transactions until the block size reaches that minimum size.
//
// Any transactions which would cause the block to exceed the BlockMaxSize
// policy setting, exceed the maximum allowed signature operations per block, or
// otherwise cause the block to be invalid are skipped.
//
// Given the above, a block generated by this function is of the following form:
//
//   -----------------------------------  --  --
//  |      Coinbase Transaction         |   |   |
//  |-----------------------------------|   |   |
//  |                                   |   |   | ----- policy.BlockPrioritySize
//  |   High-priority Transactions      |   |   |
//  |                                   |   |   |
//  |-----------------------------------|   | --
//  |                                   |   |
//  |                                   |   |
//  |                                   |   |--- policy.BlockMaxSize
//  |  Transactions prioritized by fee  |   |
//  |  until <= policy.TxMinFreeFee     |   |
//  |                                   |   |
//  |                                   |   |
//  |                                   |   |
//  |-----------------------------------|   |
//  |  Low-fee/Non high-priority (free) |   |
//  |  transactions (while block size   |   |
//  |  <= policy.BlockMinSize)          |   |
//   -----------------------------------  --
func (g *BlkTmplGenerator) NewBlockTemplate(payToAddress util.Address) (*BlockTemplate, error) {
	// Extend the most recently known best block.
	nextBlockHeight := g.dag.Height() + 1

	// Create a standard coinbase transaction paying to the provided
	// address.  NOTE: The coinbase value will be updated to include the
	// fees from the selected transactions later after they have actually
	// been selected.  It is created here to detect any errors early
	// before potentially doing a lot of work below.  The extra nonce helps
	// ensure the transaction is not a duplicate transaction (paying the
	// same value to the same public key address would otherwise be an
	// identical transaction for block version 1).
	extraNonce, err := random.Uint64()
	if err != nil {
		return nil, err
	}
	coinbaseScript, err := StandardCoinbaseScript(nextBlockHeight, extraNonce)
	if err != nil {
		return nil, err
	}
	coinbaseTx, err := CreateCoinbaseTx(g.dagParams, coinbaseScript,
		nextBlockHeight, payToAddress)
	if err != nil {
		return nil, err
	}
	numCoinbaseSigOps := int64(blockdag.CountSigOps(coinbaseTx))

	msgFeeTransaction, err := g.dag.NextBlockFeeTransaction()
	if err != nil {
		return nil, err
	}
	feeTransaction := util.NewTx(msgFeeTransaction)
	feeTxSigOps := int64(blockdag.CountSigOps(feeTransaction))

	// Get the current source transactions and create a priority queue to
	// hold the transactions which are ready for inclusion into a block
	// along with some priority related and fee metadata.  Reserve the same
	// number of items that are available for the priority queue.  Also,
	// choose the initial sort order for the priority queue based on whether
	// or not there is an area allocated for high-priority transactions.
	sourceTxns := g.txSource.MiningDescs()
	sortedByFee := g.policy.BlockPrioritySize == 0
	priorityQueue := newTxPriorityQueue(len(sourceTxns), sortedByFee)

	// Create a slice to hold the transactions to be included in the
	// generated block with reserved space.  Also create a utxo view to
	// house all of the input transactions so multiple lookups can be
	// avoided.
	blockTxns := make([]*util.Tx, 0, len(sourceTxns)+2)
	blockTxns = append(blockTxns, coinbaseTx, feeTransaction)

	// The starting block size is the size of the block header plus the max
	// possible transaction count size, plus the size of the coinbase
	// transaction.
	blockSize := blockHeaderOverhead + uint32(coinbaseTx.MsgTx().SerializeSize())
	blockSigOps := numCoinbaseSigOps + feeTxSigOps
	totalFees := uint64(0)

	// Create slices to hold the fees and number of signature operations
	// for each of the selected transactions and add an entry for the
	// coinbase.  This allows the code below to simply append details about
	// a transaction as it is selected for inclusion in the final block.
	// However, since the total fees aren't known yet, use a dummy value for
	// the coinbase fee which will be updated later.
	txFees := make([]uint64, 0, len(sourceTxns))
	txSigOpCounts := make([]int64, 0, len(sourceTxns))
	txFees = append(txFees, 0) // Updated once known
	txSigOpCounts = append(txSigOpCounts, numCoinbaseSigOps, feeTxSigOps)

	log.Debugf("Considering %d transactions for inclusion to new block",
		len(sourceTxns))

	for _, txDesc := range sourceTxns {
		// A block can't have more than one coinbase or contain
		// non-finalized transactions.
		tx := txDesc.Tx
		if blockdag.IsBlockReward(tx) {
			log.Tracef("Skipping block reward tx %s", tx.ID())
			continue
		}
		if !blockdag.IsFinalizedTransaction(tx, nextBlockHeight,
			g.timeSource.AdjustedTime()) {

			log.Tracef("Skipping non-finalized tx %s", tx.ID())
			continue
		}

		// Calculate the final transaction priority using the input
		// value age sum as well as the adjusted transaction size.  The
		// formula is: sum(inputValue * inputAge) / adjustedTxSize
		prioItem := &txPrioItem{tx: tx}
		prioItem.priority = CalcPriority(tx.MsgTx(), g.dag.UTXOSet(),
			nextBlockHeight)

		// Calculate the fee in Satoshi/kB.
		prioItem.feePerKB = txDesc.FeePerKB
		prioItem.fee = txDesc.Fee

		heap.Push(priorityQueue, prioItem)
	}

	// Create map of GAS usage per subnetwork
	gasUsageMap := make(map[subnetworkid.SubnetworkID]uint64)

	// Choose which transactions make it into the block.
	for priorityQueue.Len() > 0 {
		// Grab the highest priority (or highest fee per kilobyte
		// depending on the sort order) transaction.
		prioItem := heap.Pop(priorityQueue).(*txPrioItem)
		tx := prioItem.tx

		if !tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && !tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDRegistry) {
			subnetworkID := tx.MsgTx().SubnetworkID
			gasUsage, ok := gasUsageMap[subnetworkID]
			if !ok {
				gasUsage = 0
			}
			gasLimit, err := g.dag.SubnetworkStore.GasLimit(&subnetworkID)
			if err != nil {
				log.Errorf("Cannot get GAS limit for subnetwork %s", subnetworkID)
				continue
			}
			txGas := tx.MsgTx().Gas
			if gasLimit-gasUsage < txGas {
				log.Tracef("Transaction %s (GAS=%d) ignored because gas overusage (GASUsage=%d) in subnetwork %s (GASLimit=%d)",
					tx.MsgTx().TxID(), txGas, gasUsage, subnetworkID, gasLimit)
				continue
			}
			gasUsageMap[subnetworkID] = gasUsage + txGas
		}

		// Enforce maximum block size.  Also check for overflow.
		txSize := uint32(tx.MsgTx().SerializeSize())
		blockPlusTxSize := blockSize + txSize
		if blockPlusTxSize < blockSize ||
			blockPlusTxSize >= g.policy.BlockMaxSize {

			log.Tracef("Skipping tx %s because it would exceed "+
				"the max block size", tx.ID())
			continue
		}

		// Enforce maximum signature operations per block.  Also check
		// for overflow.
		numSigOps := int64(blockdag.CountSigOps(tx))
		if blockSigOps+numSigOps < blockSigOps ||
			blockSigOps+numSigOps > blockdag.MaxSigOpsPerBlock {
			log.Tracef("Skipping tx %s because it would exceed "+
				"the maximum sigops per block", tx.ID())
			continue
		}
		numP2SHSigOps, err := blockdag.CountP2SHSigOps(tx, false,
			g.dag.UTXOSet())
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"GetSigOpCost: %s", tx.ID(), err)
			continue
		}
		numSigOps += int64(numP2SHSigOps)
		if blockSigOps+numSigOps < blockSigOps ||
			blockSigOps+numSigOps > blockdag.MaxSigOpsPerBlock {
			log.Tracef("Skipping tx %s because it would "+
				"exceed the maximum sigops per block", tx.ID())
			continue
		}

		// Skip free transactions once the block is larger than the
		// minimum block size.
		if sortedByFee &&
			prioItem.feePerKB < uint64(g.policy.TxMinFreeFee) &&
			blockPlusTxSize >= g.policy.BlockMinSize {

			log.Tracef("Skipping tx %s with feePerKB %.2f "+
				"< TxMinFreeFee %d and block size %d >= "+
				"minBlockSize %d", tx.ID(), prioItem.feePerKB,
				g.policy.TxMinFreeFee, blockPlusTxSize,
				g.policy.BlockMinSize)
			continue
		}

		// Prioritize by fee per kilobyte once the block is larger than
		// the priority size or there are no more high-priority
		// transactions.
		if !sortedByFee && (blockPlusTxSize >= g.policy.BlockPrioritySize ||
			prioItem.priority <= MinHighPriority) {

			log.Tracef("Switching to sort by fees per kilobyte "+
				"blockSize %d >= BlockPrioritySize %d || "+
				"priority %.2f <= minHighPriority %.2f",
				blockPlusTxSize, g.policy.BlockPrioritySize,
				prioItem.priority, MinHighPriority)

			sortedByFee = true
			priorityQueue.SetLessFunc(txPQByFee)

			// Put the transaction back into the priority queue and
			// skip it so it is re-priortized by fees if it won't
			// fit into the high-priority section or the priority
			// is too low.  Otherwise this transaction will be the
			// final one in the high-priority section, so just fall
			// though to the code below so it is added now.
			if blockPlusTxSize > g.policy.BlockPrioritySize ||
				prioItem.priority < MinHighPriority {

				heap.Push(priorityQueue, prioItem)
				continue
			}
		}

		// Ensure the transaction inputs pass all of the necessary
		// preconditions before allowing it to be added to the block.
		_, err = blockdag.CheckTransactionInputsAndCalulateFee(tx, nextBlockHeight,
			g.dag.UTXOSet(), g.dagParams, false)
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"CheckTransactionInputs: %s", tx.ID(), err)
			continue
		}
		err = blockdag.ValidateTransactionScripts(tx, g.dag.UTXOSet(),
			txscript.StandardVerifyFlags, g.sigCache)
		if err != nil {
			log.Tracef("Skipping tx %s due to error in "+
				"ValidateTransactionScripts: %s", tx.ID(), err)
			continue
		}

		// Add the transaction to the block, increment counters, and
		// save the fees and signature operation counts to the block
		// template.
		blockTxns = append(blockTxns, tx)
		blockSize += txSize
		blockSigOps += int64(numSigOps)
		totalFees += prioItem.fee
		txFees = append(txFees, prioItem.fee)
		txSigOpCounts = append(txSigOpCounts, numSigOps)

		log.Tracef("Adding tx %s (priority %.2f, feePerKB %.2f)",
			prioItem.tx.ID(), prioItem.priority, prioItem.feePerKB)
	}

	// Now that the actual transactions have been selected, update the
	// block size for the real transaction count and coinbase value with
	// the total fees accordingly.
	blockSize -= wire.MaxVarIntPayload -
		uint32(wire.VarIntSerializeSize(uint64(len(blockTxns))))
	coinbaseTx.MsgTx().TxOut[0].Value += totalFees
	txFees[0] = -totalFees

	// Calculate the required difficulty for the block.  The timestamp
	// is potentially adjusted to ensure it comes after the median time of
	// the last several blocks per the chain consensus rules.
	ts := medianAdjustedTime(g.dag.CalcPastMedianTime(), g.timeSource)
	reqDifficulty, err := g.dag.CalcNextRequiredDifficulty(ts)
	if err != nil {
		return nil, err
	}

	// Calculate the next expected block version based on the state of the
	// rule change deployments.
	nextBlockVersion, err := g.dag.CalcNextBlockVersion()
	if err != nil {
		return nil, err
	}

	// Sort transactions by subnetwork ID before building Merkle tree
	sort.Slice(blockTxns, func(i, j int) bool {
		return subnetworkid.Less(&blockTxns[i].MsgTx().SubnetworkID, &blockTxns[j].MsgTx().SubnetworkID)
	})

	// Create a new block ready to be solved.
	hashMerkleTree := blockdag.BuildHashMerkleTreeStore(blockTxns)
	idMerkleTree := blockdag.BuildIDMerkleTreeStore(blockTxns)
	var msgBlock wire.MsgBlock
	msgBlock.Header = wire.BlockHeader{
		Version:        nextBlockVersion,
		ParentHashes:   g.dag.TipHashes(),
		HashMerkleRoot: *hashMerkleTree.Root(),
		IDMerkleRoot:   *idMerkleTree.Root(),
		Timestamp:      ts,
		Bits:           reqDifficulty,
	}
	for _, tx := range blockTxns {
		if err := msgBlock.AddTransaction(tx.MsgTx()); err != nil {
			return nil, err
		}
	}

	// Finally, perform a full check on the created block against the chain
	// consensus rules to ensure it properly connects to the current best
	// chain with no issues.
	block := util.NewBlock(&msgBlock)
	block.SetHeight(nextBlockHeight)
	if err := g.dag.CheckConnectBlockTemplate(block); err != nil {
		return nil, err
	}

	log.Debugf("Created new block template (%d transactions, %d in fees, "+
		"%d signature operations, %d bytes, target difficulty %064x)",
		len(msgBlock.Transactions), totalFees, blockSigOps, blockSize,
		blockdag.CompactToBig(msgBlock.Header.Bits))

	return &BlockTemplate{
		Block:           &msgBlock,
		Fees:            txFees,
		SigOpCounts:     txSigOpCounts,
		Height:          nextBlockHeight,
		ValidPayAddress: payToAddress != nil,
	}, nil
}

// UpdateBlockTime updates the timestamp in the header of the passed block to
// the current time while taking into account the median time of the last
// several blocks to ensure the new time is after that time per the chain
// consensus rules.  Finally, it will update the target difficulty if needed
// based on the new time for the test networks since their target difficulty can
// change based upon time.
func (g *BlkTmplGenerator) UpdateBlockTime(msgBlock *wire.MsgBlock) error {
	// The new timestamp is potentially adjusted to ensure it comes after
	// the median time of the last several blocks per the chain consensus
	// rules.
	dagMedianTime := g.dag.CalcPastMedianTime()
	newTime := medianAdjustedTime(dagMedianTime, g.timeSource)
	msgBlock.Header.Timestamp = newTime

	// Recalculate the difficulty if running on a network that requires it.
	if g.dagParams.ReduceMinDifficulty {
		difficulty, err := g.dag.CalcNextRequiredDifficulty(newTime)
		if err != nil {
			return err
		}
		msgBlock.Header.Bits = difficulty
	}

	return nil
}

// UpdateExtraNonce updates the extra nonce in the coinbase script of the passed
// block by regenerating the coinbase script with the passed value and block
// height.  It also recalculates and updates the new merkle root that results
// from changing the coinbase script.
func (g *BlkTmplGenerator) UpdateExtraNonce(msgBlock *wire.MsgBlock, blockHeight int32, extraNonce uint64) error {
	coinbaseScript, err := StandardCoinbaseScript(blockHeight, extraNonce)
	if err != nil {
		return err
	}
	if len(coinbaseScript) > blockdag.MaxCoinbaseScriptLen {
		return fmt.Errorf("coinbase transaction script length "+
			"of %d is out of range (min: %d, max: %d)",
			len(coinbaseScript), blockdag.MinCoinbaseScriptLen,
			blockdag.MaxCoinbaseScriptLen)
	}
	msgBlock.Transactions[0].TxIn[0].SignatureScript = coinbaseScript

	// TODO(davec): A util.Block should use saved in the state to avoid
	// recalculating all of the other transaction hashes.
	// block.Transactions[0].InvalidateCache()

	// Recalculate the merkle roots with the updated extra nonce.
	block := util.NewBlock(msgBlock)
	hashMerkleTree := blockdag.BuildHashMerkleTreeStore(block.Transactions())
	msgBlock.Header.HashMerkleRoot = *hashMerkleTree.Root()
	idMerkleTree := blockdag.BuildIDMerkleTreeStore(block.Transactions())
	msgBlock.Header.IDMerkleRoot = *idMerkleTree.Root()

	return nil
}

// DAGHeight returns the DAG's height
func (g *BlkTmplGenerator) DAGHeight() int32 {
	return g.dag.Height()
}

// TipHashes returns the hashes of the DAG's tips
func (g *BlkTmplGenerator) TipHashes() []daghash.Hash {
	return g.dag.TipHashes()
}

// TxSource returns the associated transaction source.
//
// This function is safe for concurrent access.
func (g *BlkTmplGenerator) TxSource() TxSource {
	return g.txSource
}
