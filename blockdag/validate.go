// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

const (
	// MaxSigOpsPerBlock is the maximum number of signature operations
	// allowed for a block.  It is a fraction of the max block payload size.
	MaxSigOpsPerBlock = wire.MaxBlockPayload / 50

	// MaxTimeOffsetSeconds is the maximum number of seconds a block time
	// is allowed to be ahead of the current time.  This is currently 2
	// hours.
	MaxTimeOffsetSeconds = 2 * 60 * 60

	// MinCoinbaseScriptLen is the minimum length a coinbase script can be.
	MinCoinbaseScriptLen = 2

	// MaxCoinbaseScriptLen is the maximum length a coinbase script can be.
	MaxCoinbaseScriptLen = 100

	// medianTimeBlocks is the number of previous blocks which should be
	// used to calculate the median time used to validate block timestamps.
	medianTimeBlocks = 51

	// baseSubsidy is the starting subsidy amount for mined blocks.  This
	// value is halved every SubsidyHalvingInterval blocks.
	baseSubsidy = 50 * util.SatoshiPerBitcoin

	// MaxOutputsPerBlock is the maximum number of transaction outputs there
	// can be in a block of max size.
	MaxOutputsPerBlock = wire.MaxBlockPayload / wire.MinTxOutPayload

	// feeTransactionIndex is the index of the fee transaction in a block.
	feeTransactionIndex = 1
)

// isNullOutpoint determines whether or not a previous transaction output point
// is set.
func isNullOutpoint(outpoint *wire.OutPoint) bool {
	if outpoint.Index == math.MaxUint32 && outpoint.TxID == daghash.ZeroTxID {
		return true
	}
	return false
}

// IsCoinBase determines whether or not a transaction is a coinbase.  A coinbase
// is a special transaction created by miners that has no inputs.  This is
// represented in the block dag by a transaction with a single input that has
// a previous output transaction index set to the maximum value along with a
// zero hash.
func IsCoinBase(tx *util.Tx) bool {
	return tx.MsgTx().IsCoinBase()
}

// IsBlockReward determines whether or not a transaction is a block reward (a fee transaction or block reward)
func IsBlockReward(tx *util.Tx) bool {
	return tx.MsgTx().IsBlockReward()
}

// IsFeeTransaction determines whether or not a transaction is a fee transaction.  A fee
// transaction is a special transaction created by miners that distributes fees to the
// previous blocks' miners.  Each input of the fee transaction should set index to maximum
// value and reference the relevant block id, instead of previous transaction id.
func IsFeeTransaction(tx *util.Tx) bool {
	return tx.MsgTx().IsFeeTransaction()
}

// SequenceLockActive determines if a transaction's sequence locks have been
// met, meaning that all the inputs of a given transaction have reached a
// height or time sufficient for their relative lock-time maturity.
func SequenceLockActive(sequenceLock *SequenceLock, blockHeight int32,
	medianTimePast time.Time) bool {

	// If either the seconds, or height relative-lock time has not yet
	// reached, then the transaction is not yet mature according to its
	// sequence locks.
	if sequenceLock.Seconds >= medianTimePast.Unix() ||
		sequenceLock.BlockHeight >= blockHeight {
		return false
	}

	return true
}

// IsFinalizedTransaction determines whether or not a transaction is finalized.
func IsFinalizedTransaction(tx *util.Tx, blockHeight int32, blockTime time.Time) bool {
	msgTx := tx.MsgTx()

	// Lock time of zero means the transaction is finalized.
	lockTime := msgTx.LockTime
	if lockTime == 0 {
		return true
	}

	// The lock time field of a transaction is either a block height at
	// which the transaction is finalized or a timestamp depending on if the
	// value is before the txscript.LockTimeThreshold.  When it is under the
	// threshold it is a block height.
	blockTimeOrHeight := int64(0)
	if lockTime < txscript.LockTimeThreshold {
		blockTimeOrHeight = int64(blockHeight)
	} else {
		blockTimeOrHeight = blockTime.Unix()
	}
	if int64(lockTime) < blockTimeOrHeight {
		return true
	}

	// At this point, the transaction's lock time hasn't occurred yet, but
	// the transaction might still be finalized if the sequence number
	// for all transaction inputs is maxed out.
	for _, txIn := range msgTx.TxIn {
		if txIn.Sequence != math.MaxUint64 {
			return false
		}
	}
	return true
}

// CalcBlockSubsidy returns the subsidy amount a block at the provided height
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
//
// The subsidy is halved every SubsidyReductionInterval blocks.  Mathematically
// this is: baseSubsidy / 2^(height/SubsidyReductionInterval)
//
// At the target block generation rate for the main network, this is
// approximately every 4 years.
func CalcBlockSubsidy(height int32, dagParams *dagconfig.Params) uint64 {
	if dagParams.SubsidyReductionInterval == 0 {
		return baseSubsidy
	}

	// Equivalent to: baseSubsidy / 2^(height/subsidyHalvingInterval)
	return baseSubsidy >> uint(height/dagParams.SubsidyReductionInterval)
}

// CheckTransactionSanity performs some preliminary checks on a transaction to
// ensure it is sane.  These checks are context free.
func CheckTransactionSanity(tx *util.Tx, subnetworkID *subnetworkid.SubnetworkID, isFeeTransaction bool) error {
	// A transaction must have at least one input.
	msgTx := tx.MsgTx()
	if len(msgTx.TxIn) == 0 {
		return ruleError(ErrNoTxInputs, "transaction has no inputs")
	}

	// A transaction must not exceed the maximum allowed block payload when
	// serialized.
	serializedTxSize := msgTx.SerializeSize()
	if serializedTxSize > wire.MaxBlockPayload {
		str := fmt.Sprintf("serialized transaction is too big - got "+
			"%d, max %d", serializedTxSize, wire.MaxBlockPayload)
		return ruleError(ErrTxTooBig, str)
	}

	// Ensure the transaction amounts are in range.  Each transaction
	// output must not be negative or more than the max allowed per
	// transaction.  Also, the total of all outputs must abide by the same
	// restrictions.  All amounts in a transaction are in a unit value known
	// as a satoshi.  One bitcoin is a quantity of satoshi as defined by the
	// SatoshiPerBitcoin constant.
	var totalSatoshi uint64
	for _, txOut := range msgTx.TxOut {
		satoshi := txOut.Value
		if satoshi > util.MaxSatoshi {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v", satoshi,
				util.MaxSatoshi)
			return ruleError(ErrBadTxOutValue, str)
		}

		// Binary arithmetic guarantees that any overflow is detected and reported.
		// This is impossible for Bitcoin, but perhaps possible if an alt increases
		// the total money supply.
		newTotalSatoshi := totalSatoshi + satoshi
		if newTotalSatoshi < totalSatoshi {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs exceeds max allowed value of %v",
				util.MaxSatoshi)
			return ruleError(ErrBadTxOutValue, str)
		}
		totalSatoshi = newTotalSatoshi
		if totalSatoshi > util.MaxSatoshi {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs is %v which is higher than max "+
				"allowed value of %v", totalSatoshi,
				util.MaxSatoshi)
			return ruleError(ErrBadTxOutValue, str)
		}
	}

	// Check for duplicate transaction inputs.
	existingTxOut := make(map[wire.OutPoint]struct{})
	for _, txIn := range msgTx.TxIn {
		if _, exists := existingTxOut[txIn.PreviousOutPoint]; exists {
			return ruleError(ErrDuplicateTxInputs, "transaction "+
				"contains duplicate inputs")
		}
		existingTxOut[txIn.PreviousOutPoint] = struct{}{}
	}

	// Coinbase script length must be between min and max length.
	if IsCoinBase(tx) {
		slen := len(msgTx.TxIn[0].SignatureScript)
		if slen < MinCoinbaseScriptLen || slen > MaxCoinbaseScriptLen {
			str := fmt.Sprintf("coinbase transaction script length "+
				"of %d is out of range (min: %d, max: %d)",
				slen, MinCoinbaseScriptLen, MaxCoinbaseScriptLen)
			return ruleError(ErrBadCoinbaseScriptLen, str)
		}
	} else {
		// Previous transaction outputs referenced by the inputs to this
		// transaction must not be null.
		for _, txIn := range msgTx.TxIn {
			if isNullOutpoint(&txIn.PreviousOutPoint) {
				return ruleError(ErrBadTxInput, "transaction "+
					"input refers to previous output that "+
					"is null")
			}
		}
	}

	// Transactions in native and subnetwork registry subnetworks must have Gas = 0
	if (msgTx.SubnetworkID == wire.SubnetworkIDNative ||
		msgTx.SubnetworkID == wire.SubnetworkIDRegistry) &&
		msgTx.Gas > 0 {

		return ruleError(ErrInvalidGas, "transaction in the native or "+
			"registry subnetworks has gas > 0 ")
	}

	if msgTx.SubnetworkID == wire.SubnetworkIDNative &&
		len(msgTx.Payload) > 0 {

		return ruleError(ErrInvalidPayload,
			"transaction in the native subnetwork includes a payload")
	}

	if msgTx.SubnetworkID == wire.SubnetworkIDRegistry &&
		len(msgTx.Payload) != 8 {

		return ruleError(ErrInvalidPayload,
			"transaction in the subnetwork registry include a payload "+
				"with length != 8 bytes")
	}

	// If we are a partial node, only transactions on the Registry subnetwork
	// or our own subnetwork may have a payload
	isLocalNodeFull := subnetworkID.IsEqual(&wire.SubnetworkIDSupportsAll)
	shouldTxBeFull := msgTx.SubnetworkID.IsEqual(&wire.SubnetworkIDRegistry) ||
		msgTx.SubnetworkID.IsEqual(subnetworkID)
	if !isLocalNodeFull && !shouldTxBeFull && len(msgTx.Payload) > 0 {
		return ruleError(ErrInvalidPayload,
			"transaction that was expected to be partial has a payload "+
				"with length > 0")
	}

	return nil
}

// checkProofOfWork ensures the block header bits which indicate the target
// difficulty is in min/max range and that the block hash is less than the
// target difficulty as claimed.
//
// The flags modify the behavior of this function as follows:
//  - BFNoPoWCheck: The check to ensure the block hash is less than the target
//    difficulty is not performed.
func (dag *BlockDAG) checkProofOfWork(header *wire.BlockHeader, flags BehaviorFlags) error {
	// The target difficulty must be larger than zero.
	target := CompactToBig(header.Bits)
	if target.Sign() <= 0 {
		str := fmt.Sprintf("block target difficulty of %064x is too low",
			target)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(dag.dagParams.PowLimit) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, dag.dagParams.PowLimit)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if flags&BFNoPoWCheck != BFNoPoWCheck {
		// The block hash must be less than the claimed target.
		hash := header.BlockHash()
		hashNum := daghash.HashToBig(&hash)
		if hashNum.Cmp(target) > 0 {
			str := fmt.Sprintf("block hash of %064x is higher than "+
				"expected max of %064x", hashNum, target)
			return ruleError(ErrHighHash, str)
		}
	}

	return nil
}

// CountSigOps returns the number of signature operations for all transaction
// input and output scripts in the provided transaction.  This uses the
// quicker, but imprecise, signature operation counting mechanism from
// txscript.
func CountSigOps(tx *util.Tx) int {
	msgTx := tx.MsgTx()

	// Accumulate the number of signature operations in all transaction
	// inputs.
	totalSigOps := 0
	for _, txIn := range msgTx.TxIn {
		numSigOps := txscript.GetSigOpCount(txIn.SignatureScript)
		totalSigOps += numSigOps
	}

	// Accumulate the number of signature operations in all transaction
	// outputs.
	for _, txOut := range msgTx.TxOut {
		numSigOps := txscript.GetSigOpCount(txOut.PkScript)
		totalSigOps += numSigOps
	}

	return totalSigOps
}

// CountP2SHSigOps returns the number of signature operations for all input
// transactions which are of the pay-to-script-hash type.  This uses the
// precise, signature operation counting mechanism from the script engine which
// requires access to the input transaction scripts.
func CountP2SHSigOps(tx *util.Tx, isBlockReward bool, utxoSet UTXOSet) (int, error) {
	// Block reward transactions have no interesting inputs.
	if isBlockReward {
		return 0, nil
	}

	// Accumulate the number of signature operations in all transaction
	// inputs.
	msgTx := tx.MsgTx()
	totalSigOps := 0
	for txInIndex, txIn := range msgTx.TxIn {
		// Ensure the referenced input transaction is available.
		entry, ok := utxoSet.Get(txIn.PreviousOutPoint)
		if !ok {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOutPoint,
				tx.ID(), txInIndex)
			return 0, ruleError(ErrMissingTxOut, str)
		}

		// We're only interested in pay-to-script-hash types, so skip
		// this input if it's not one.
		pkScript := entry.PkScript()
		if !txscript.IsPayToScriptHash(pkScript) {
			continue
		}

		// Count the precise number of signature operations in the
		// referenced public key script.
		sigScript := txIn.SignatureScript
		numSigOps := txscript.GetPreciseSigOpCount(sigScript, pkScript,
			true)

		// We could potentially overflow the accumulator so check for
		// overflow.
		lastSigOps := totalSigOps
		totalSigOps += numSigOps
		if totalSigOps < lastSigOps {
			str := fmt.Sprintf("the public key script from output "+
				"%v contains too many signature operations - "+
				"overflow", txIn.PreviousOutPoint)
			return 0, ruleError(ErrTooManySigOps, str)
		}
	}

	return totalSigOps, nil
}

// checkBlockHeaderSanity performs some preliminary checks on a block header to
// ensure it is sane before continuing with processing.  These checks are
// context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkProofOfWork.
func (dag *BlockDAG) checkBlockHeaderSanity(header *wire.BlockHeader, flags BehaviorFlags) error {
	// Ensure the proof of work bits in the block header is in min/max range
	// and the block hash is less than the target value described by the
	// bits.
	err := dag.checkProofOfWork(header, flags)
	if err != nil {
		return err
	}

	if len(header.ParentHashes) == 0 {
		if header.BlockHash() != *dag.dagParams.GenesisHash {
			return ruleError(ErrNoParents, "block has no parents")
		}
	} else {
		err = checkBlockParentsOrder(header)
		if err != nil {
			return err
		}
	}

	// A block timestamp must not have a greater precision than one second.
	// This check is necessary because Go time.Time values support
	// nanosecond precision whereas the consensus rules only apply to
	// seconds and it's much nicer to deal with standard Go time values
	// instead of converting to seconds everywhere.
	if !header.Timestamp.Equal(time.Unix(header.Timestamp.Unix(), 0)) {
		str := fmt.Sprintf("block timestamp of %v has a higher "+
			"precision than one second", header.Timestamp)
		return ruleError(ErrInvalidTime, str)
	}

	// Ensure the block time is not too far in the future.
	maxTimestamp := dag.timeSource.AdjustedTime().Add(time.Second *
		MaxTimeOffsetSeconds)
	if header.Timestamp.After(maxTimestamp) {
		str := fmt.Sprintf("block timestamp of %v is too far in the "+
			"future", header.Timestamp)
		return ruleError(ErrTimeTooNew, str)
	}

	return nil
}

//checkBlockParentsOrder ensures that the block's parents are ordered by hash
func checkBlockParentsOrder(header *wire.BlockHeader) error {
	sortedHashes := make([]daghash.Hash, 0, header.NumParentBlocks())
	for _, hash := range header.ParentHashes {
		sortedHashes = append(sortedHashes, hash)
	}
	sort.Slice(sortedHashes, func(i, j int) bool {
		return daghash.Less(&sortedHashes[i], &sortedHashes[j])
	})
	if !daghash.AreEqual(header.ParentHashes, sortedHashes) {
		return ruleError(ErrWrongParentsOrder, "block parents are not ordered by hash")
	}
	return nil
}

// checkBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing.  These checks are context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkBlockHeaderSanity.
func (dag *BlockDAG) checkBlockSanity(block *util.Block, flags BehaviorFlags) error {

	msgBlock := block.MsgBlock()
	header := &msgBlock.Header
	err := dag.checkBlockHeaderSanity(header, flags)
	if err != nil {
		return err
	}

	// A block must have at least one transaction.
	numTx := len(msgBlock.Transactions)
	if numTx == 0 {
		return ruleError(ErrNoTransactions, "block does not contain "+
			"any transactions")
	}

	// A block must not have more transactions than the max block payload or
	// else it is certainly over the block size limit.
	if numTx > wire.MaxBlockPayload {
		str := fmt.Sprintf("block contains too many transactions - "+
			"got %d, max %d", numTx, wire.MaxBlockPayload)
		return ruleError(ErrBlockTooBig, str)
	}

	// A block must not exceed the maximum allowed block payload when
	// serialized.
	serializedSize := msgBlock.SerializeSize()
	if serializedSize > wire.MaxBlockPayload {
		str := fmt.Sprintf("serialized block is too big - got %d, "+
			"max %d", serializedSize, wire.MaxBlockPayload)
		return ruleError(ErrBlockTooBig, str)
	}

	// The first transaction in a block must be a coinbase.
	transactions := block.Transactions()
	if !IsCoinBase(transactions[0]) {
		return ruleError(ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}

	isGenesis := block.MsgBlock().Header.IsGenesis()
	if !isGenesis && !IsFeeTransaction(transactions[1]) {
		return ruleError(ErrSecondTxNotFeeTransaction, "second transaction in "+
			"block is not a fee transaction")
	}

	txOffset := 2
	if isGenesis {
		txOffset = 1
	}

	// A block must not have more than one coinbase. And transactions must be
	// ordered by subnetwork
	for i, tx := range transactions[txOffset:] {
		if IsCoinBase(tx) {
			str := fmt.Sprintf("block contains second coinbase at "+
				"index %d", i+2)
			return ruleError(ErrMultipleCoinbases, str)
		}
		if IsFeeTransaction(tx) {
			str := fmt.Sprintf("block contains second fee transaction at "+
				"index %d", i+2)
			return ruleError(ErrMultipleFeeTransactions, str)
		}
		if subnetworkid.Less(&tx.MsgTx().SubnetworkID, &transactions[i].MsgTx().SubnetworkID) {
			return ruleError(ErrTransactionsNotSorted, "transactions must be sorted by subnetwork")
		}
	}

	// Do some preliminary checks on each transaction to ensure they are
	// sane before continuing.
	for i, tx := range transactions {
		isFeeTransaction := i == feeTransactionIndex
		err := CheckTransactionSanity(tx, dag.subnetworkID, isFeeTransaction)
		if err != nil {
			return err
		}
	}

	// Build merkle tree and ensure the calculated merkle root matches the
	// entry in the block header.  This also has the effect of caching all
	// of the transaction hashes in the block to speed up future hash
	// checks.  Bitcoind builds the tree here and checks the merkle root
	// after the following checks, but there is no reason not to check the
	// merkle root matches here.
	hashMerkleTree := BuildHashMerkleTreeStore(block.Transactions())
	calculatedHashMerkleRoot := hashMerkleTree.Root()
	if !header.HashMerkleRoot.IsEqual(calculatedHashMerkleRoot) {
		str := fmt.Sprintf("block hash merkle root is invalid - block "+
			"header indicates %v, but calculated value is %v",
			header.HashMerkleRoot, calculatedHashMerkleRoot)
		return ruleError(ErrBadMerkleRoot, str)
	}

	idMerkleTree := BuildIDMerkleTreeStore(block.Transactions())
	calculatedIDMerkleRoot := idMerkleTree.Root()
	if !header.IDMerkleRoot.IsEqual(calculatedIDMerkleRoot) {
		str := fmt.Sprintf("block ID merkle root is invalid - block "+
			"header indicates %v, but calculated value is %v",
			header.IDMerkleRoot, calculatedIDMerkleRoot)
		return ruleError(ErrBadMerkleRoot, str)
	}

	// Check for duplicate transactions.  This check will be fairly quick
	// since the transaction IDs are already cached due to building the
	// merkle tree above.
	existingTxIDs := make(map[daghash.TxID]struct{})
	for _, tx := range transactions {
		id := tx.ID()
		if _, exists := existingTxIDs[*id]; exists {
			str := fmt.Sprintf("block contains duplicate "+
				"transaction %v", id)
			return ruleError(ErrDuplicateTx, str)
		}
		existingTxIDs[*id] = struct{}{}
	}

	// The number of signature operations must be less than the maximum
	// allowed per block.
	totalSigOps := 0
	for _, tx := range transactions {
		// We could potentially overflow the accumulator so check for
		// overflow.
		lastSigOps := totalSigOps
		totalSigOps += CountSigOps(tx)
		if totalSigOps < lastSigOps || totalSigOps > MaxSigOpsPerBlock {
			str := fmt.Sprintf("block contains too many signature "+
				"operations - got %v, max %v", totalSigOps,
				MaxSigOpsPerBlock)
			return ruleError(ErrTooManySigOps, str)
		}
	}

	// Amount of gas consumed per sub-network shouldn't be more than the subnetwork's limit
	gasUsageInAllSubnetworks := map[subnetworkid.SubnetworkID]uint64{}
	for _, tx := range transactions {
		msgTx := tx.MsgTx()
		// In DAGCoin and Registry sub-networks all txs must have Gas = 0, and that is validated in checkTransactionSanity
		// Therefore - no need to check them here.
		if msgTx.SubnetworkID != wire.SubnetworkIDNative && msgTx.SubnetworkID != wire.SubnetworkIDRegistry {
			gasUsageInSubnetwork := gasUsageInAllSubnetworks[msgTx.SubnetworkID]
			gasUsageInSubnetwork += msgTx.Gas
			if gasUsageInSubnetwork < gasUsageInAllSubnetworks[msgTx.SubnetworkID] { // protect from overflows
				str := fmt.Sprintf("Block gas usage in subnetwork with ID %s has overflown", msgTx.SubnetworkID)
				return ruleError(ErrInvalidGas, str)
			}
			gasUsageInAllSubnetworks[msgTx.SubnetworkID] = gasUsageInSubnetwork

			gasLimit, err := dag.SubnetworkStore.GasLimit(&msgTx.SubnetworkID)
			if err != nil {
				return err
			}
			if gasUsageInSubnetwork > gasLimit {
				str := fmt.Sprintf("Block wastes too much gas in subnetwork with ID %s", msgTx.SubnetworkID)
				return ruleError(ErrInvalidGas, str)
			}
		}
	}

	return nil
}

// CheckBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing.  These checks are context free.
func (dag *BlockDAG) CheckBlockSanity(block *util.Block, powLimit *big.Int,
	timeSource MedianTimeSource) error {

	return dag.checkBlockSanity(block, BFNone)
}

// ExtractCoinbaseHeight attempts to extract the height of the block from the
// scriptSig of a coinbase transaction.
func ExtractCoinbaseHeight(coinbaseTx *util.Tx) (int32, error) {
	sigScript := coinbaseTx.MsgTx().TxIn[0].SignatureScript
	if len(sigScript) < 1 {
		str := "the coinbase signature script" +
			"must start with the " +
			"length of the serialized block height"
		str = fmt.Sprintf(str)
		return 0, ruleError(ErrMissingCoinbaseHeight, str)
	}

	// Detect the case when the block height is a small integer encoded with
	// a single byte.
	opcode := int(sigScript[0])
	if opcode == txscript.Op0 {
		return 0, nil
	}
	if opcode >= txscript.Op1 && opcode <= txscript.Op16 {
		return int32(opcode - (txscript.Op1 - 1)), nil
	}

	// Otherwise, the opcode is the length of the following bytes which
	// encode in the block height.
	serializedLen := int(sigScript[0])
	if len(sigScript[1:]) < serializedLen {
		str := "the coinbase signature script " +
			"must start with the " +
			"serialized block height"
		str = fmt.Sprintf(str, serializedLen)
		return 0, ruleError(ErrMissingCoinbaseHeight, str)
	}

	serializedHeightBytes := make([]byte, 8)
	copy(serializedHeightBytes, sigScript[1:serializedLen+1])
	serializedHeight := binary.LittleEndian.Uint64(serializedHeightBytes)

	return int32(serializedHeight), nil
}

// checkSerializedHeight checks if the signature script in the passed
// transaction starts with the serialized block height of wantHeight.
func checkSerializedHeight(coinbaseTx *util.Tx, wantHeight int32) error {
	serializedHeight, err := ExtractCoinbaseHeight(coinbaseTx)
	if err != nil {
		return err
	}

	if serializedHeight != wantHeight {
		str := fmt.Sprintf("the coinbase signature script serialized "+
			"block height is %d when %d was expected",
			serializedHeight, wantHeight)
		return ruleError(ErrBadCoinbaseHeight, str)
	}
	return nil
}

// checkBlockHeaderContext performs several validation checks on the block header
// which depend on its position within the block dag.
//
// The flags modify the behavior of this function as follows:
//  - BFFastAdd: All checks except those involving comparing the header against
//    the checkpoints are not performed.
//
// This function MUST be called with the dag state lock held (for writes).
func (dag *BlockDAG) checkBlockHeaderContext(header *wire.BlockHeader, bluestParent *blockNode, blockHeight int32, flags BehaviorFlags) error {
	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		// Ensure the difficulty specified in the block header matches
		// the calculated difficulty based on the previous block and
		// difficulty retarget rules.
		expectedDifficulty, err := dag.calcNextRequiredDifficulty(bluestParent,
			header.Timestamp)
		if err != nil {
			return err
		}
		blockDifficulty := header.Bits
		if blockDifficulty != expectedDifficulty {
			str := "block difficulty of %d is not the expected value of %d"
			str = fmt.Sprintf(str, blockDifficulty, expectedDifficulty)
			return ruleError(ErrUnexpectedDifficulty, str)
		}

		if !header.IsGenesis() {
			// Ensure the timestamp for the block header is not before the
			// median time of the last several blocks (medianTimeBlocks).
			medianTime := bluestParent.CalcPastMedianTime()
			if header.Timestamp.Before(medianTime) {
				str := "block timestamp of %s is not after expected %s"
				str = fmt.Sprintf(str, header.Timestamp.String(), medianTime.String())
				return ruleError(ErrTimeTooOld, str)
			}
		}
	}

	// Ensure dag matches up to predetermined checkpoints.
	blockHash := header.BlockHash()
	if !dag.verifyCheckpoint(blockHeight, &blockHash) {
		str := fmt.Sprintf("block at height %d does not match "+
			"checkpoint hash", blockHeight)
		return ruleError(ErrBadCheckpoint, str)
	}

	// Find the previous checkpoint and prevent blocks which fork the main
	// dag before it.  This prevents storage of new, otherwise valid,
	// blocks which build off of old blocks that are likely at a much easier
	// difficulty and therefore could be used to waste cache and disk space.
	checkpointNode, err := dag.findPreviousCheckpoint()
	if err != nil {
		return err
	}
	if checkpointNode != nil && blockHeight < checkpointNode.height {
		str := fmt.Sprintf("block at height %d forks the main dag "+
			"before the previous checkpoint at height %d",
			blockHeight, checkpointNode.height)
		return ruleError(ErrForkTooOld, str)
	}

	return nil
}

// validateParents validates that no parent is an ancestor of another parent
func validateParents(blockHeader *wire.BlockHeader, parents blockSet) error {
	minHeight := int32(math.MaxInt32)
	queue := NewHeap()
	visited := newSet()
	for _, parent := range parents {
		if parent.height < minHeight {
			minHeight = parent.height
		}
		for _, grandParent := range parent.parents {
			if !visited.contains(grandParent) {
				queue.Push(grandParent)
				visited.add(grandParent)
			}
		}
	}
	for queue.Len() > 0 {
		current := queue.pop()
		if parents.contains(current) {
			return fmt.Errorf("Block %s is both a parent of %s and an"+
				" ancestor of another parent",
				current.hash,
				blockHeader.BlockHash())
		}
		if current.height > minHeight {
			for _, parent := range current.parents {
				if !visited.contains(parent) {
					queue.Push(parent)
					visited.add(parent)
				}
			}
		}
	}
	return nil
}

// checkBlockContext peforms several validation checks on the block which depend
// on its position within the block DAG.
//
// The flags modify the behavior of this function as follows:
//  - BFFastAdd: The transaction are not checked to see if they are finalized
//    and the somewhat expensive BIP0034 validation is not performed.
//
// The flags are also passed to checkBlockHeaderContext.  See its documentation
// for how the flags modify its behavior.
//
// This function MUST be called with the dag state lock held (for writes).
func (dag *BlockDAG) checkBlockContext(block *util.Block, parents blockSet, bluestParent *blockNode, flags BehaviorFlags) error {
	err := validateParents(&block.MsgBlock().Header, parents)
	if err != nil {
		return err
	}

	// Perform all block header related validation checks.
	header := &block.MsgBlock().Header
	err = dag.checkBlockHeaderContext(header, bluestParent, block.Height(), flags)
	if err != nil {
		return err
	}

	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		blockTime := header.Timestamp
		if !block.IsGenesis() {
			blockTime = bluestParent.CalcPastMedianTime()
		}

		// Ensure all transactions in the block are finalized.
		for _, tx := range block.Transactions() {
			if !IsFinalizedTransaction(tx, block.Height(),
				blockTime) {

				str := fmt.Sprintf("block contains unfinalized "+
					"transaction %v", tx.ID())
				return ruleError(ErrUnfinalizedTx, str)
			}
		}

		// Ensure coinbase starts with serialized block heights

		coinbaseTx := block.Transactions()[0]
		err := checkSerializedHeight(coinbaseTx, block.Height())
		if err != nil {
			return err
		}

	}

	return nil
}

// ensureNoDuplicateTx ensures blocks do not contain duplicate transactions which
// 'overwrite' older transactions that are not fully spent.  This prevents an
// attack where a coinbase and all of its dependent transactions could be
// duplicated to effectively revert the overwritten transactions to a single
// confirmation thereby making them vulnerable to a double spend.
//
// For more details, see
// https://github.com/bitcoin/bips/blob/master/bip-0030.mediawiki and
// http://r6.ca/blog/20120206T005236Z.html.
//
// This function MUST be called with the dag state lock held (for reads).
func ensureNoDuplicateTx(block *blockNode, utxoSet UTXOSet,
	transactions []*util.Tx) error {
	// Fetch utxos for all of the transaction ouputs in this block.
	// Typically, there will not be any utxos for any of the outputs.
	fetchSet := make(map[wire.OutPoint]struct{})
	for _, tx := range transactions {
		prevOut := wire.OutPoint{TxID: *tx.ID()}
		for txOutIdx := range tx.MsgTx().TxOut {
			prevOut.Index = uint32(txOutIdx)
			fetchSet[prevOut] = struct{}{}
		}
	}

	// Duplicate transactions are only allowed if the previous transaction
	// is fully spent.
	for outpoint := range fetchSet {
		utxo, ok := utxoSet.Get(outpoint)
		if ok {
			str := fmt.Sprintf("tried to overwrite transaction %v "+
				"at block height %d that is not fully spent",
				outpoint.TxID, utxo.BlockHeight())
			return ruleError(ErrOverwriteTx, str)
		}
	}

	return nil
}

// CheckTransactionInputs performs a series of checks on the inputs to a
// transaction to ensure they are valid.  An example of some of the checks
// include verifying all inputs exist, ensuring the block reward seasoning
// requirements are met, detecting double spends, validating all values and fees
// are in the legal range and the total output amount doesn't exceed the input
// amount, and verifying the signatures to prove the spender was the owner of
// the bitcoins and therefore allowed to spend them.  As it checks the inputs,
// it also calculates the total fees for the transaction and returns that value.
//
// NOTE: The transaction MUST have already been sanity checked with the
// CheckTransactionSanity function prior to calling this function.
func CheckTransactionInputs(tx *util.Tx, txHeight int32, utxoSet UTXOSet, dagParams *dagconfig.Params) (uint64, error) {
	// Block reward transactions have no inputs.
	if IsBlockReward(tx) {
		return 0, nil
	}

	txID := tx.ID()
	var totalSatoshiIn uint64
	for txInIndex, txIn := range tx.MsgTx().TxIn {
		// Ensure the referenced input transaction is available.
		entry, ok := utxoSet.Get(txIn.PreviousOutPoint)
		if !ok {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOutPoint,
				tx.ID(), txInIndex)
			return 0, ruleError(ErrMissingTxOut, str)
		}

		// Ensure the transaction is not spending coins which have not
		// yet reached the required block reward maturity.
		if entry.IsBlockReward() {
			originHeight := entry.BlockHeight()
			blocksSincePrev := txHeight - originHeight
			blockRewardMaturity := int32(dagParams.BlockRewardMaturity)
			if blocksSincePrev < blockRewardMaturity {
				str := fmt.Sprintf("tried to spend block reward "+
					"transaction output %v from height %v "+
					"at height %v before required maturity "+
					"of %v blocks", txIn.PreviousOutPoint,
					originHeight, txHeight,
					blockRewardMaturity)
				return 0, ruleError(ErrImmatureSpend, str)
			}
		}

		// Ensure the transaction amounts are in range.  Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction.  All amounts in
		// a transaction are in a unit value known as a satoshi.  One
		// bitcoin is a quantity of satoshi as defined by the
		// SatoshiPerBitcoin constant.
		originTxSatoshi := entry.Amount()
		if originTxSatoshi > util.MaxSatoshi {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v",
				util.Amount(originTxSatoshi),
				util.MaxSatoshi)
			return 0, ruleError(ErrBadTxOutValue, str)
		}

		// The total of all outputs must not be more than the max
		// allowed per transaction.  Also, we could potentially overflow
		// the accumulator so check for overflow.
		lastSatoshiIn := totalSatoshiIn
		totalSatoshiIn += originTxSatoshi
		if totalSatoshiIn < lastSatoshiIn ||
			totalSatoshiIn > util.MaxSatoshi {
			str := fmt.Sprintf("total value of all transaction "+
				"inputs is %v which is higher than max "+
				"allowed value of %v", totalSatoshiIn,
				util.MaxSatoshi)
			return 0, ruleError(ErrBadTxOutValue, str)
		}
	}

	// Calculate the total output amount for this transaction.  It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	var totalSatoshiOut uint64
	for _, txOut := range tx.MsgTx().TxOut {
		totalSatoshiOut += txOut.Value
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalSatoshiIn < totalSatoshiOut {
		str := fmt.Sprintf("total value of all transaction inputs for "+
			"transaction %v is %v which is less than the amount "+
			"spent of %v", txID, totalSatoshiIn, totalSatoshiOut)
		return 0, ruleError(ErrSpendTooHigh, str)
	}

	// NOTE: bitcoind checks if the transaction fees are < 0 here, but that
	// is an impossible condition because of the check above that ensures
	// the inputs are >= the outputs.
	txFeeInSatoshi := totalSatoshiIn - totalSatoshiOut
	return txFeeInSatoshi, nil
}

// checkConnectToPastUTXO performs several checks to confirm connecting the passed
// block to the DAG represented by the passed view does not violate any rules.
//
// An example of some of the checks performed are ensuring connecting the block
// would not cause any duplicate transaction hashes for old transactions that
// aren't already fully spent, double spends, exceeding the maximum allowed
// signature operations per block, invalid values in relation to the expected
// block subsidy, or fail transaction script validation.
//
// This function MUST be called with the dag state lock held (for writes).
func (dag *BlockDAG) checkConnectToPastUTXO(block *blockNode, pastUTXO UTXOSet,
	transactions []*util.Tx) error {

	err := ensureNoDuplicateTx(block, pastUTXO, transactions)
	if err != nil {
		return err
	}

	// The number of signature operations must be less than the maximum
	// allowed per block.  Note that the preliminary sanity checks on a
	// block also include a check similar to this one, but this check
	// expands the count to include a precise count of pay-to-script-hash
	// signature operations in each of the input transaction public key
	// scripts.
	totalSigOps := 0
	for i, tx := range transactions {
		numsigOps := CountSigOps(tx)
		// Since the first transaction has already been verified to be a
		// coinbase transaction, and the second transaction has already
		// been verified to be a fee transaction, use i < 2 as an
		// optimization for the flag to countP2SHSigOps for whether or
		// not the transaction is a block reward transaction rather than
		// having to do a full coinbase and fee transaction check again.
		numP2SHSigOps, err := CountP2SHSigOps(tx, i < 2, pastUTXO)
		if err != nil {
			return err
		}
		numsigOps += numP2SHSigOps

		// Check for overflow or going over the limits.  We have to do
		// this on every loop iteration to avoid overflow.
		lastSigops := totalSigOps
		totalSigOps += numsigOps
		if totalSigOps < lastSigops || totalSigOps > MaxSigOpsPerBlock {
			str := fmt.Sprintf("block contains too many "+
				"signature operations - got %v, max %v",
				totalSigOps, MaxSigOpsPerBlock)
			return ruleError(ErrTooManySigOps, str)
		}
	}

	// Perform several checks on the inputs for each transaction.  Also
	// accumulate the total fees.  This could technically be combined with
	// the loop above instead of running another loop over the transactions,
	// but by separating it we can avoid running the more expensive (though
	// still relatively cheap as compared to running the scripts) checks
	// against all the inputs when the signature operations are out of
	// bounds.
	var totalFees uint64
	for _, tx := range transactions {
		txFee, err := CheckTransactionInputs(tx, block.height, pastUTXO,
			dag.dagParams)
		if err != nil {
			return err
		}

		// Sum the total fees and ensure we don't overflow the
		// accumulator.
		lastTotalFees := totalFees
		totalFees += txFee
		if totalFees < lastTotalFees {
			return ruleError(ErrBadFees, "total fees for block "+
				"overflows accumulator")
		}
	}

	// The total output values of the coinbase transaction must not exceed
	// the expected subsidy value plus total transaction fees gained from
	// mining the block.  It is safe to ignore overflow and out of range
	// errors here because those error conditions would have already been
	// caught by checkTransactionSanity.
	var totalSatoshiOut uint64
	for _, txOut := range transactions[0].MsgTx().TxOut {
		totalSatoshiOut += txOut.Value
	}
	expectedSatoshiOut := CalcBlockSubsidy(block.height, dag.dagParams)
	if totalSatoshiOut > expectedSatoshiOut {
		str := fmt.Sprintf("coinbase transaction for block pays %v "+
			"which is more than expected value of %v",
			totalSatoshiOut, expectedSatoshiOut)
		return ruleError(ErrBadCoinbaseValue, str)
	}

	// Don't run scripts if this node is before the latest known good
	// checkpoint since the validity is verified via the checkpoints (all
	// transactions are included in the merkle root hash and any changes
	// will therefore be detected by the next checkpoint).  This is a huge
	// optimization because running the scripts is the most time consuming
	// portion of block handling.
	checkpoint := dag.LatestCheckpoint()
	runScripts := true
	if checkpoint != nil && block.height <= checkpoint.Height {
		runScripts = false
	}

	scriptFlags := txscript.ScriptNoFlags

	// We obtain the MTP of the *previous* block (unless it's genesis block)
	// in order to determine if transactions in the current block are final.
	medianTime := block.Header().Timestamp
	if !block.isGenesis() {
		medianTime = block.selectedParent.CalcPastMedianTime()
	}

	// We also enforce the relative sequence number based
	// lock-times within the inputs of all transactions in this
	// candidate block.
	for _, tx := range transactions {
		// A transaction can only be included within a block
		// once the sequence locks of *all* its inputs are
		// active.
		sequenceLock, err := dag.calcSequenceLock(block, pastUTXO, tx, false)
		if err != nil {
			return err
		}
		if !SequenceLockActive(sequenceLock, block.height,
			medianTime) {
			str := fmt.Sprintf("block contains " +
				"transaction whose input sequence " +
				"locks are not met")
			return ruleError(ErrUnfinalizedTx, str)
		}
	}

	// Now that the inexpensive checks are done and have passed, verify the
	// transactions are actually allowed to spend the coins by running the
	// expensive ECDSA signature check scripts.  Doing this last helps
	// prevent CPU exhaustion attacks.
	if runScripts {
		err := checkBlockScripts(block, pastUTXO, transactions, scriptFlags, dag.sigCache)
		if err != nil {
			return err
		}
	}

	return nil
}

// countSpentOutputs returns the number of utxos the passed block spends.
func countSpentOutputs(block *util.Block) int {
	// Exclude the block reward transactions since they can't spend anything.
	var numSpent int
	for _, tx := range block.Transactions()[1:] {
		if !IsFeeTransaction(tx) {
			numSpent += len(tx.MsgTx().TxIn)
		}
	}
	return numSpent
}

// CheckConnectBlockTemplate fully validates that connecting the passed block to
// the DAG does not violate any consensus rules, aside from the proof of
// work requirement. The block must connect to the current tip of the main dag.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) CheckConnectBlockTemplate(block *util.Block) error {
	dag.dagLock.Lock()
	defer dag.dagLock.Unlock()

	// Skip the proof of work check as this is just a block template.
	flags := BFNoPoWCheck

	// This only checks whether the block can be connected to the tip of the
	// current dag.
	tips := dag.virtual.tips()
	header := block.MsgBlock().Header
	parentHashes := header.ParentHashes
	if !tips.hashesEqual(parentHashes) {
		str := fmt.Sprintf("parent blocks must be the currents tips %v, "+
			"instead got %v", tips, parentHashes)
		return ruleError(ErrParentBlockNotCurrentTips, str)
	}

	err := dag.checkBlockSanity(block, flags)
	if err != nil {
		return err
	}

	parents, err := lookupParentNodes(block, dag)
	if err != nil {
		return err
	}

	err = dag.checkBlockContext(block, parents, dag.selectedTip(), flags)
	if err != nil {
		return err
	}

	return dag.checkConnectToPastUTXO(newBlockNode(&header, dag.virtual.tips(), dag.dagParams.K),
		dag.UTXOSet(), block.Transactions())
}
