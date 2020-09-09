// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"math"
	"sort"

	"github.com/kaspanet/go-secp256k1"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

const (
	// MaxCoinbasePayloadLen is the maximum length a coinbase payload can be.
	MaxCoinbasePayloadLen = 150

	// baseSubsidy is the starting subsidy amount for mined blocks. This
	// value is halved every SubsidyHalvingInterval blocks.
	baseSubsidy = 50 * util.SompiPerKaspa

	// the following are used when calculating a transaction's mass

	// MassPerTxByte is the number of grams that any byte
	// adds to a transaction.
	MassPerTxByte = 1

	// MassPerScriptPubKeyByte is the number of grams that any
	// scriptPubKey byte adds to a transaction.
	MassPerScriptPubKeyByte = 10

	// MassPerSigOp is the number of grams that any
	// signature operation adds to a transaction.
	MassPerSigOp = 10000

	// mergeSetSizeLimit is the maximum allowed merge set size for a block.
	mergeSetSizeLimit = 1000
)

// isNullOutpoint determines whether or not a previous transaction outpoint
// is set.
func isNullOutpoint(outpoint *appmessage.Outpoint) bool {
	if outpoint.Index == math.MaxUint32 && outpoint.TxID == daghash.ZeroTxID {
		return true
	}
	return false
}

// SequenceLockActive determines if a transaction's sequence locks have been
// met, meaning that all the inputs of a given transaction have reached a
// blue score or time sufficient for their relative lock-time maturity.
func SequenceLockActive(sequenceLock *SequenceLock, blockBlueScore uint64,
	medianTimePast mstime.Time) bool {

	// If either the milliseconds, or blue score relative-lock time has not yet
	// reached, then the transaction is not yet mature according to its
	// sequence locks.
	if sequenceLock.Milliseconds >= medianTimePast.UnixMilliseconds() ||
		sequenceLock.BlockBlueScore >= int64(blockBlueScore) {
		return false
	}

	return true
}

// IsFinalizedTransaction determines whether or not a transaction is finalized.
func IsFinalizedTransaction(tx *util.Tx, blockBlueScore uint64, blockTime mstime.Time) bool {
	msgTx := tx.MsgTx()

	// Lock time of zero means the transaction is finalized.
	lockTime := msgTx.LockTime
	if lockTime == 0 {
		return true
	}

	// The lock time field of a transaction is either a block blue score at
	// which the transaction is finalized or a timestamp depending on if the
	// value is before the txscript.LockTimeThreshold. When it is under the
	// threshold it is a block blue score.
	blockTimeOrBlueScore := uint64(0)
	if lockTime < txscript.LockTimeThreshold {
		blockTimeOrBlueScore = blockBlueScore
	} else {
		blockTimeOrBlueScore = uint64(blockTime.UnixMilliseconds())
	}
	if lockTime < blockTimeOrBlueScore {
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

// CalcBlockSubsidy returns the subsidy amount a block at the provided blue score
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
//
// The subsidy is halved every SubsidyReductionInterval blocks. Mathematically
// this is: baseSubsidy / 2^(blueScore/SubsidyReductionInterval)
//
// At the target block generation rate for the main network, this is
// approximately every 4 years.
func CalcBlockSubsidy(blueScore uint64, dagParams *dagconfig.Params) uint64 {
	if dagParams.SubsidyReductionInterval == 0 {
		return baseSubsidy
	}

	// Equivalent to: baseSubsidy / 2^(blueScore/subsidyHalvingInterval)
	return baseSubsidy >> uint(blueScore/dagParams.SubsidyReductionInterval)
}

// CheckTransactionSanity performs some preliminary checks on a transaction to
// ensure it is sane. These checks are context free.
func CheckTransactionSanity(tx *util.Tx, subnetworkID *subnetworkid.SubnetworkID) error {
	err := checkTransactionInputCount(tx)
	if err != nil {
		return err
	}
	err = checkTransactionAmountRanges(tx)
	if err != nil {
		return err
	}
	err = checkDuplicateTransactionInputs(tx)
	if err != nil {
		return err
	}
	err = checkCoinbaseLength(tx)
	if err != nil {
		return err
	}
	err = checkTransactionPayloadHash(tx)
	if err != nil {
		return err
	}
	err = checkGasInBuiltInOrNativeTransactions(tx)
	if err != nil {
		return err
	}
	err = checkSubnetworkRegistryTransaction(tx)
	if err != nil {
		return err
	}
	err = checkNativeTransactionPayload(tx)
	if err != nil {
		return err
	}
	err = checkTransactionSubnetwork(tx, subnetworkID)
	if err != nil {
		return err
	}
	return nil
}

func checkTransactionInputCount(tx *util.Tx) error {
	// A non-coinbase transaction must have at least one input.
	msgTx := tx.MsgTx()
	if !tx.IsCoinBase() && len(msgTx.TxIn) == 0 {
		return ruleError(ErrNoTxInputs, "transaction has no inputs")
	}
	return nil
}

func checkTransactionAmountRanges(tx *util.Tx) error {
	// Ensure the transaction amounts are in range. Each transaction
	// output must not be negative or more than the max allowed per
	// transaction. Also, the total of all outputs must abide by the same
	// restrictions. All amounts in a transaction are in a unit value known
	// as a sompi. One kaspa is a quantity of sompi as defined by the
	// SompiPerKaspa constant.
	var totalSompi uint64
	for _, txOut := range tx.MsgTx().TxOut {
		sompi := txOut.Value
		if sompi > util.MaxSompi {
			str := fmt.Sprintf("transaction output value of %d is "+
				"higher than max allowed value of %d", sompi,
				util.MaxSompi)
			return ruleError(ErrBadTxOutValue, str)
		}

		// Binary arithmetic guarantees that any overflow is detected and reported.
		// This is impossible for Kaspa, but perhaps possible if an alt increases
		// the total money supply.
		newTotalSompi := totalSompi + sompi
		if newTotalSompi < totalSompi {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs exceeds max allowed value of %d",
				util.MaxSompi)
			return ruleError(ErrBadTxOutValue, str)
		}
		totalSompi = newTotalSompi
		if totalSompi > util.MaxSompi {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs is %d which is higher than max "+
				"allowed value of %d", totalSompi,
				util.MaxSompi)
			return ruleError(ErrBadTxOutValue, str)
		}
	}
	return nil
}

func checkDuplicateTransactionInputs(tx *util.Tx) error {
	existingTxOut := make(map[appmessage.Outpoint]struct{})
	for _, txIn := range tx.MsgTx().TxIn {
		if _, exists := existingTxOut[txIn.PreviousOutpoint]; exists {
			return ruleError(ErrDuplicateTxInputs, "transaction "+
				"contains duplicate inputs")
		}
		existingTxOut[txIn.PreviousOutpoint] = struct{}{}
	}
	return nil
}

func checkCoinbaseLength(tx *util.Tx) error {
	// Coinbase payload length must not exceed the max length.
	if tx.IsCoinBase() {
		payloadLen := len(tx.MsgTx().Payload)
		if payloadLen > MaxCoinbasePayloadLen {
			str := fmt.Sprintf("coinbase transaction payload length "+
				"of %d is out of range (max: %d)",
				payloadLen, MaxCoinbasePayloadLen)
			return ruleError(ErrBadCoinbasePayloadLen, str)
		}
	} else {
		// Previous transaction outputs referenced by the inputs to this
		// transaction must not be null.
		for _, txIn := range tx.MsgTx().TxIn {
			if isNullOutpoint(&txIn.PreviousOutpoint) {
				return ruleError(ErrBadTxInput, "transaction "+
					"input refers to previous output that "+
					"is null")
			}
		}
	}
	return nil
}

func checkTransactionPayloadHash(tx *util.Tx) error {
	msgTx := tx.MsgTx()
	if !msgTx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) {
		payloadHash := daghash.DoubleHashH(msgTx.Payload)
		if !msgTx.PayloadHash.IsEqual(&payloadHash) {
			return ruleError(ErrInvalidPayloadHash, "invalid payload hash")
		}
	} else if msgTx.PayloadHash != nil {
		return ruleError(ErrInvalidPayloadHash, "unexpected non-empty payload hash in native subnetwork")
	}
	return nil
}

func checkGasInBuiltInOrNativeTransactions(tx *util.Tx) error {
	// Transactions in native, registry and coinbase subnetworks must have Gas = 0
	msgTx := tx.MsgTx()
	if msgTx.SubnetworkID.IsBuiltInOrNative() && msgTx.Gas > 0 {
		return ruleError(ErrInvalidGas, "transaction in the native or "+
			"registry subnetworks has gas > 0 ")
	}
	return nil
}

func checkSubnetworkRegistryTransaction(tx *util.Tx) error {
	if tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDRegistry) {
		err := validateSubnetworkRegistryTransaction(tx.MsgTx())
		if err != nil {
			return err
		}
	}
	return nil
}

func checkNativeTransactionPayload(tx *util.Tx) error {
	msgTx := tx.MsgTx()
	if msgTx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && len(msgTx.Payload) > 0 {
		return ruleError(ErrInvalidPayload, "transaction in the native subnetwork includes a payload")
	}
	return nil
}

func checkTransactionSubnetwork(tx *util.Tx, subnetworkID *subnetworkid.SubnetworkID) error {
	// If we are a partial node, only transactions on built in subnetworks
	// or our own subnetwork may have a payload
	msgTx := tx.MsgTx()
	isLocalNodeFull := subnetworkID == nil
	shouldTxBeFull := msgTx.SubnetworkID.IsBuiltIn() ||
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
func (dag *BlockDAG) checkProofOfWork(header *appmessage.BlockHeader, flags BehaviorFlags) error {
	// The target difficulty must be larger than zero.
	target := util.CompactToBig(header.Bits)
	if target.Sign() <= 0 {
		str := fmt.Sprintf("block target difficulty of %064x is too low",
			target)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(dag.Params.PowMax) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, dag.Params.PowMax)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if flags&BFNoPoWCheck != BFNoPoWCheck {
		// The block hash must be less than the claimed target.
		hash := header.BlockHash()
		hashNum := daghash.HashToBig(hash)
		if hashNum.Cmp(target) > 0 {
			str := fmt.Sprintf("block hash of %064x is higher than "+
				"expected max of %064x", hashNum, target)
			return ruleError(ErrHighHash, str)
		}
	}

	return nil
}

// ValidateTxMass makes sure that the given transaction's mass does not exceed
// the maximum allowed limit. Currently, it is equivalent to the block mass limit.
// See CalcTxMass for further details.
func ValidateTxMass(tx *util.Tx, utxoSet UTXOSet) error {
	txMass, err := CalcTxMassFromUTXOSet(tx, utxoSet)
	if err != nil {
		return err
	}
	if txMass > appmessage.MaxMassAcceptedByBlock {
		str := fmt.Sprintf("tx %s has mass %d, which is above the "+
			"allowed limit of %d", tx.ID(), txMass, appmessage.MaxMassAcceptedByBlock)
		return ruleError(ErrTxMassTooHigh, str)
	}
	return nil
}

func calcTxMassFromInputsWithUTXOEntries(
	tx *util.Tx, inputsWithUTXOEntries []*txInputAndUTXOEntry) uint64 {

	if tx.IsCoinBase() {
		return calcCoinbaseTxMass(tx)
	}

	previousScriptPubKeys := make([][]byte, 0, len(tx.MsgTx().TxIn))

	for _, inputWithUTXOEntry := range inputsWithUTXOEntries {
		utxoEntry := inputWithUTXOEntry.utxoEntry

		previousScriptPubKeys = append(previousScriptPubKeys, utxoEntry.ScriptPubKey())
	}
	return CalcTxMass(tx, previousScriptPubKeys)
}

// CalcTxMassFromUTXOSet calculates the transaction mass based on the
// UTXO set in its past.
//
// See CalcTxMass for more details.
func CalcTxMassFromUTXOSet(tx *util.Tx, utxoSet UTXOSet) (uint64, error) {
	if tx.IsCoinBase() {
		return calcCoinbaseTxMass(tx), nil
	}
	previousScriptPubKeys := make([][]byte, len(tx.MsgTx().TxIn))
	for txInIndex, txIn := range tx.MsgTx().TxIn {
		entry, ok := utxoSet.Get(txIn.PreviousOutpoint)
		if !ok {
			str := fmt.Sprintf("output %s referenced from "+
				"transaction %s input %d either does not exist or "+
				"has already been spent", txIn.PreviousOutpoint,
				tx.ID(), txInIndex)
			return 0, ruleError(ErrMissingTxOut, str)
		}
		previousScriptPubKeys[txInIndex] = entry.ScriptPubKey()
	}
	return CalcTxMass(tx, previousScriptPubKeys), nil
}

func calcCoinbaseTxMass(tx *util.Tx) uint64 {
	return CalcTxMass(tx, nil)
}

// CalcTxMass sums up and returns the "mass" of a transaction. This number
// is an approximation of how many resources (CPU, RAM, etc.) it would take
// to process the transaction.
// The following properties are considered in the calculation:
// * The transaction length in bytes
// * The length of all output scripts in bytes
// * The count of all input sigOps
func CalcTxMass(tx *util.Tx, previousScriptPubKeys [][]byte) uint64 {
	txSize := tx.MsgTx().SerializeSize()

	if tx.IsCoinBase() {
		return uint64(txSize * MassPerTxByte)
	}

	scriptPubKeySize := 0
	for _, txOut := range tx.MsgTx().TxOut {
		scriptPubKeySize += len(txOut.ScriptPubKey)
	}

	sigOpsCount := 0
	for txInIndex, txIn := range tx.MsgTx().TxIn {
		// Count the precise number of signature operations in the
		// referenced public key script.
		sigScript := txIn.SignatureScript
		isP2SH := txscript.IsPayToScriptHash(previousScriptPubKeys[txInIndex])
		sigOpsCount += txscript.GetPreciseSigOpCount(sigScript, previousScriptPubKeys[txInIndex], isP2SH)
	}

	return uint64(txSize*MassPerTxByte +
		scriptPubKeySize*MassPerScriptPubKeyByte +
		sigOpsCount*MassPerSigOp)
}

// checkBlockHeaderSanity performs some preliminary checks on a block header to
// ensure it is sane before continuing with processing. These checks are
// context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkProofOfWork.
func (dag *BlockDAG) checkBlockHeaderSanity(block *util.Block, flags BehaviorFlags) error {
	// Ensure the proof of work bits in the block header is in min/max range
	// and the block hash is less than the target value described by the
	// bits.
	header := &block.MsgBlock().Header
	err := dag.checkProofOfWork(header, flags)
	if err != nil {
		return err
	}

	if len(header.ParentHashes) == 0 {
		if !header.BlockHash().IsEqual(dag.Params.GenesisHash) {
			return ruleError(ErrNoParents, "block has no parents")
		}
	} else {
		err = checkBlockParentsOrder(header)
		if err != nil {
			return err
		}
	}

	return nil
}

//checkBlockParentsOrder ensures that the block's parents are ordered by hash
func checkBlockParentsOrder(header *appmessage.BlockHeader) error {
	sortedHashes := make([]*daghash.Hash, header.NumParentBlocks())
	for i, hash := range header.ParentHashes {
		sortedHashes[i] = hash
	}
	sort.Slice(sortedHashes, func(i, j int) bool {
		return daghash.Less(sortedHashes[i], sortedHashes[j])
	})
	if !daghash.AreEqual(header.ParentHashes, sortedHashes) {
		return ruleError(ErrWrongParentsOrder, "block parents are not ordered by hash")
	}
	return nil
}

// checkBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing. These checks are context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkBlockHeaderSanity.
func (dag *BlockDAG) checkBlockSanity(block *util.Block, flags BehaviorFlags) error {
	err := dag.checkBlockHeaderSanity(block, flags)
	if err != nil {
		return err
	}
	err = dag.checkBlockContainsAtLeastOneTransaction(block)
	if err != nil {
		return err
	}
	err = dag.checkBlockContainsLessThanMaxBlockMassTransactions(block)
	if err != nil {
		return err
	}
	err = dag.checkFirstBlockTransactionIsCoinbase(block)
	if err != nil {
		return err
	}
	err = dag.checkBlockContainsOnlyOneCoinbase(block)
	if err != nil {
		return err
	}
	err = dag.checkBlockTransactionOrder(block)
	if err != nil {
		return err
	}
	err = dag.checkNoNonNativeTransactions(block)
	if err != nil {
		return err
	}
	err = dag.checkBlockTransactionSanity(block)
	if err != nil {
		return err
	}
	err = dag.checkBlockHashMerkleRoot(block)
	if err != nil {
		return err
	}

	// The following check will be fairly quick since the transaction IDs
	// are already cached due to building the merkle tree above.
	err = dag.checkBlockDuplicateTransactions(block)
	if err != nil {
		return err
	}

	err = dag.checkBlockDoubleSpends(block)
	if err != nil {
		return err
	}
	return nil
}

func (dag *BlockDAG) checkBlockContainsAtLeastOneTransaction(block *util.Block) error {
	transactions := block.Transactions()
	numTx := len(transactions)
	if numTx == 0 {
		return ruleError(ErrNoTransactions, "block does not contain "+
			"any transactions")
	}
	return nil
}

func (dag *BlockDAG) checkBlockContainsLessThanMaxBlockMassTransactions(block *util.Block) error {
	// A block must not have more transactions than the max block mass or
	// else it is certainly over the block mass limit.
	transactions := block.Transactions()
	numTx := len(transactions)
	if numTx > appmessage.MaxMassAcceptedByBlock {
		str := fmt.Sprintf("block contains too many transactions - "+
			"got %d, max %d", numTx, appmessage.MaxMassAcceptedByBlock)
		return ruleError(ErrBlockMassTooHigh, str)
	}
	return nil
}

func (dag *BlockDAG) checkFirstBlockTransactionIsCoinbase(block *util.Block) error {
	transactions := block.Transactions()
	if !transactions[util.CoinbaseTransactionIndex].IsCoinBase() {
		return ruleError(ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}
	return nil
}

func (dag *BlockDAG) checkBlockContainsOnlyOneCoinbase(block *util.Block) error {
	transactions := block.Transactions()
	for i, tx := range transactions[util.CoinbaseTransactionIndex+1:] {
		if tx.IsCoinBase() {
			str := fmt.Sprintf("block contains second coinbase at "+
				"index %d", i+2)
			return ruleError(ErrMultipleCoinbases, str)
		}
	}
	return nil
}

func (dag *BlockDAG) checkBlockTransactionOrder(block *util.Block) error {
	transactions := block.Transactions()
	for i, tx := range transactions[util.CoinbaseTransactionIndex+1:] {
		if i != 0 && subnetworkid.Less(&tx.MsgTx().SubnetworkID, &transactions[i].MsgTx().SubnetworkID) {
			return ruleError(ErrTransactionsNotSorted, "transactions must be sorted by subnetwork")
		}
	}
	return nil
}

func (dag *BlockDAG) checkNoNonNativeTransactions(block *util.Block) error {
	// Disallow non-native/coinbase subnetworks in networks that don't allow them
	if !dag.Params.EnableNonNativeSubnetworks {
		transactions := block.Transactions()
		for _, tx := range transactions {
			if !(tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) ||
				tx.MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDCoinbase)) {
				return ruleError(ErrInvalidSubnetwork, "non-native/coinbase subnetworks are not allowed")
			}
		}
	}
	return nil
}

func (dag *BlockDAG) checkBlockTransactionSanity(block *util.Block) error {
	transactions := block.Transactions()
	for _, tx := range transactions {
		err := CheckTransactionSanity(tx, dag.subnetworkID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dag *BlockDAG) checkBlockHashMerkleRoot(block *util.Block) error {
	// Build merkle tree and ensure the calculated merkle root matches the
	// entry in the block header. This also has the effect of caching all
	// of the transaction hashes in the block to speed up future hash
	// checks.
	hashMerkleTree := BuildHashMerkleTreeStore(block.Transactions())
	calculatedHashMerkleRoot := hashMerkleTree.Root()
	if !block.MsgBlock().Header.HashMerkleRoot.IsEqual(calculatedHashMerkleRoot) {
		str := fmt.Sprintf("block hash merkle root is invalid - block "+
			"header indicates %s, but calculated value is %s",
			block.MsgBlock().Header.HashMerkleRoot, calculatedHashMerkleRoot)
		return ruleError(ErrBadMerkleRoot, str)
	}
	return nil
}

func (dag *BlockDAG) checkBlockDuplicateTransactions(block *util.Block) error {
	existingTxIDs := make(map[daghash.TxID]struct{})
	transactions := block.Transactions()
	for _, tx := range transactions {
		id := tx.ID()
		if _, exists := existingTxIDs[*id]; exists {
			str := fmt.Sprintf("block contains duplicate "+
				"transaction %s", id)
			return ruleError(ErrDuplicateTx, str)
		}
		existingTxIDs[*id] = struct{}{}
	}
	return nil
}

func (dag *BlockDAG) checkBlockDoubleSpends(block *util.Block) error {
	usedOutpoints := make(map[appmessage.Outpoint]*daghash.TxID)
	transactions := block.Transactions()
	for _, tx := range transactions {
		for _, txIn := range tx.MsgTx().TxIn {
			if spendingTxID, exists := usedOutpoints[txIn.PreviousOutpoint]; exists {
				str := fmt.Sprintf("transaction %s spends "+
					"outpoint %s that was already spent by "+
					"transaction %s in this block", tx.ID(), txIn.PreviousOutpoint, spendingTxID)
				return ruleError(ErrDoubleSpendInSameBlock, str)
			}
			usedOutpoints[txIn.PreviousOutpoint] = tx.ID()
		}
	}
	return nil
}

// checkBlockHeaderContext performs several validation checks on the block header
// which depend on its position within the block dag.
//
// The flags modify the behavior of this function as follows:
//  - BFFastAdd: No checks are performed.
//
// This function MUST be called with the dag state lock held (for writes).
func (dag *BlockDAG) checkBlockHeaderContext(header *appmessage.BlockHeader, bluestParent *blockNode, fastAdd bool) error {
	if !fastAdd {
		if err := dag.validateDifficulty(header, bluestParent); err != nil {
			return err
		}

		if err := validateMedianTime(header, bluestParent); err != nil {
			return err
		}
	}
	return nil
}

func validateMedianTime(header *appmessage.BlockHeader, bluestParent *blockNode) error {
	if !header.IsGenesis() {
		// Ensure the timestamp for the block header is not before the
		// median time of the last several blocks (medianTimeBlocks).
		medianTime := bluestParent.PastMedianTime()
		if header.Timestamp.Before(medianTime) {
			str := fmt.Sprintf("block timestamp of %s is not after expected %s", header.Timestamp, medianTime)
			return ruleError(ErrTimeTooOld, str)
		}
	}

	return nil
}

func (dag *BlockDAG) validateDifficulty(header *appmessage.BlockHeader, bluestParent *blockNode) error {
	// Ensure the difficulty specified in the block header matches
	// the calculated difficulty based on the previous block and
	// difficulty retarget rules.
	expectedDifficulty := dag.requiredDifficulty(bluestParent,
		header.Timestamp)
	blockDifficulty := header.Bits
	if blockDifficulty != expectedDifficulty {
		str := fmt.Sprintf("block difficulty of %d is not the expected value of %d", blockDifficulty, expectedDifficulty)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	return nil
}

// validateParents validates that no parent is an ancestor of another parent, and no parent is finalized
func (dag *BlockDAG) validateParents(blockHeader *appmessage.BlockHeader, parents blockSet) error {
	if len(parents) > appmessage.MaxNumParentBlocks {
		return ruleError(ErrTooManyParents,
			fmt.Sprintf("block %s points to %d parents > MaxNumParentBlocks: %d",
				blockHeader.BlockHash(), len(parents), appmessage.MaxNumParentBlocks))
	}

	for parentA := range parents {
		for parentB := range parents {
			if parentA == parentB {
				continue
			}

			isAncestorOf, err := dag.isInPast(parentA, parentB)
			if err != nil {
				return err
			}
			if isAncestorOf {
				return ruleError(ErrInvalidParentsRelation, fmt.Sprintf("block %s is both a parent of %s and an"+
					" ancestor of another parent %s",
					parentA.hash,
					blockHeader.BlockHash(),
					parentB.hash,
				))
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
// The flags are also passed to checkBlockHeaderContext. See its documentation
// for how the flags modify its behavior.
//
// This function MUST be called with the dag state lock held (for writes).
func (dag *BlockDAG) checkBlockContext(block *util.Block, flags BehaviorFlags) error {
	parents, err := lookupParentNodes(block, dag)
	if err != nil {
		return dag.handleLookupParentNodesError(block, err)
	}

	bluestParent := parents.bluest()
	fastAdd := flags&BFFastAdd == BFFastAdd

	err = dag.validateParents(&block.MsgBlock().Header, parents)
	if err != nil {
		return err
	}

	// Perform all block header related validation checks.
	header := &block.MsgBlock().Header
	if err = dag.checkBlockHeaderContext(header, bluestParent, fastAdd); err != nil {
		return err
	}

	return nil
}

func (node *blockNode) checkDAGRelations() error {
	err := node.checkMergeSizeLimit()
	if err != nil {
		return err
	}

	err = node.checkBoundedMergeDepth()
	if err != nil {
		return err
	}

	return nil
}

func (dag *BlockDAG) handleLookupParentNodesError(block *util.Block, err error) error {
	var ruleErr RuleError
	if ok := errors.As(err, &ruleErr); ok && ruleErr.ErrorCode == ErrInvalidAncestorBlock {
		err := dag.addNodeToIndexWithInvalidAncestor(block)
		if err != nil {
			return err
		}
	}
	return err
}

func (dag *BlockDAG) checkBlockTransactionsFinalized(block *util.Block, node *blockNode, flags BehaviorFlags) error {
	fastAdd := flags&BFFastAdd == BFFastAdd || dag.index.BlockNodeStatus(node).KnownValid()
	if fastAdd {
		return nil
	}

	blockTime := block.MsgBlock().Header.Timestamp
	if !block.IsGenesis() {
		blockTime = node.selectedParent.PastMedianTime()
	}

	// Ensure all transactions in the block are finalized.
	for _, tx := range block.Transactions() {
		if !IsFinalizedTransaction(tx, node.blueScore, blockTime) {
			str := fmt.Sprintf("block contains unfinalized "+
				"transaction %s", tx.ID())
			return ruleError(ErrUnfinalizedTx, str)
		}
	}

	return nil
}

func (dag *BlockDAG) checkBlockHasNoChainedTransactions(block *util.Block, node *blockNode, flags BehaviorFlags) error {
	fastAdd := flags&BFFastAdd == BFFastAdd || dag.index.BlockNodeStatus(node).KnownValid()
	if fastAdd {
		return nil
	}

	transactions := block.Transactions()
	transactionsSet := make(map[daghash.TxID]struct{}, len(transactions))
	for _, transaction := range transactions {
		transactionsSet[*transaction.ID()] = struct{}{}
	}

	for _, transaction := range transactions {
		for i, transactionInput := range transaction.MsgTx().TxIn {
			if _, ok := transactionsSet[transactionInput.PreviousOutpoint.TxID]; ok {
				str := fmt.Sprintf("block contains chained transactions: Input %d of transaction %s spend"+
					"an output of transaction %s", i, transaction.ID(), transactionInput.PreviousOutpoint.TxID)
				return ruleError(ErrChainedTransactions, str)
			}
		}
	}

	return nil
}

// ensureNoDuplicateTx ensures blocks do not contain duplicate transactions which
// 'overwrite' older transactions that are not fully spent. This prevents an
// attack where a coinbase and all of its dependent transactions could be
// duplicated to effectively revert the overwritten transactions to a single
// confirmation thereby making them vulnerable to a double spend.
//
// For more details, see http://r6.ca/blog/20120206T005236Z.html.
//
// This function MUST be called with the dag state lock held (for reads).
func ensureNoDuplicateTx(utxoSet UTXOSet, transactions []*util.Tx) error {
	// Fetch utxos for all of the transaction ouputs in this block.
	// Typically, there will not be any utxos for any of the outputs.
	fetchSet := make(map[appmessage.Outpoint]struct{})
	for _, tx := range transactions {
		prevOut := appmessage.Outpoint{TxID: *tx.ID()}
		for txOutIdx := range tx.MsgTx().TxOut {
			prevOut.Index = uint32(txOutIdx)
			fetchSet[prevOut] = struct{}{}
		}
	}

	// Duplicate transactions are only allowed if the previous transaction
	// is fully spent.
	for outpoint := range fetchSet {
		if _, ok := utxoSet.Get(outpoint); ok {
			str := fmt.Sprintf("tried to overwrite transaction %s "+
				"that is not fully spent", outpoint.TxID)
			return ruleError(ErrOverwriteTx, str)
		}
	}

	return nil
}

func checkTxIsNotDuplicate(tx *util.Tx, utxoSet UTXOSet) error {
	fetchSet := make(map[appmessage.Outpoint]struct{})

	// Fetch utxos for all of the ouputs in this transaction.
	// Typically, there will not be any utxos for any of the outputs.
	prevOut := appmessage.Outpoint{TxID: *tx.ID()}
	for txOutIdx := range tx.MsgTx().TxOut {
		prevOut.Index = uint32(txOutIdx)
		fetchSet[prevOut] = struct{}{}
	}

	// Duplicate transactions are only allowed if the previous transaction
	// is fully spent.
	for outpoint := range fetchSet {
		if _, ok := utxoSet.Get(outpoint); ok {
			str := fmt.Sprintf("tried to overwrite transaction %s "+
				"that is not fully spent", outpoint.TxID)
			return ruleError(ErrOverwriteTx, str)
		}
	}

	return nil
}

// CheckTransactionInputsAndCalulateFee performs a series of checks on the inputs to a
// transaction to ensure they are valid. An example of some of the checks
// include verifying all inputs exist, ensuring the block reward seasoning
// requirements are met, detecting double spends, validating all values and fees
// are in the legal range and the total output amount doesn't exceed the input
// amount. As it checks the inputs, it also calculates the total fees for the
// transaction and returns that value.
//
// NOTE: The transaction MUST have already been sanity checked with the
// CheckTransactionSanity function prior to calling this function.
func CheckTransactionInputsAndCalulateFee(
	tx *util.Tx, txBlueScore uint64, utxoSet UTXOSet, dagParams *dagconfig.Params, fastAdd bool) (
	txFeeInSompi uint64, err error) {

	// Coinbase transactions have no standard inputs to validate.
	if tx.IsCoinBase() {
		return 0, nil
	}

	txID := tx.ID()
	var totalSompiIn uint64
	for txInIndex, txIn := range tx.MsgTx().TxIn {
		entry, err := checkReferencedOutputsAreAvailable(tx, utxoSet, txIn, txInIndex)
		if err != nil {
			return 0, err
		}

		if !fastAdd {
			if err = validateCoinbaseMaturity(dagParams, entry, txBlueScore, txIn); err != nil {
				return 0, err
			}
		}

		totalSompiIn, err = checkEntryAmounts(entry, totalSompiIn)
		if err != nil {
			return 0, err
		}
	}

	totalSompiOut, err := checkOutputsAmounts(tx, totalSompiIn, txID)
	if err != nil {
		return 0, err
	}

	txFeeInSompi = totalSompiIn - totalSompiOut
	return txFeeInSompi, nil
}

func checkOutputsAmounts(tx *util.Tx, totalSompiIn uint64, txID *daghash.TxID) (totalSompiOut uint64, err error) {
	// Calculate the total output amount for this transaction. It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	for _, txOut := range tx.MsgTx().TxOut {
		totalSompiOut += txOut.Value
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalSompiIn < totalSompiOut {
		str := fmt.Sprintf("total value of all transaction inputs for "+
			"transaction %s is %d which is less than the amount "+
			"spent of %d", txID, totalSompiIn, totalSompiOut)
		return 0, ruleError(ErrSpendTooHigh, str)
	}
	return totalSompiOut, nil
}

func checkEntryAmounts(entry *UTXOEntry, totalSompiInBefore uint64) (totalSompiInAfter uint64, err error) {
	// The total of all outputs must not be more than the max
	// allowed per transaction. Also, we could potentially overflow
	// the accumulator so check for overflow.
	lastSompiIn := totalSompiInBefore
	originTxSompi := entry.Amount()
	totalSompiInAfter = totalSompiInBefore + originTxSompi
	if totalSompiInBefore < lastSompiIn ||
		totalSompiInBefore > util.MaxSompi {
		str := fmt.Sprintf("total value of all transaction "+
			"inputs is %d which is higher than max "+
			"allowed value of %d", totalSompiInBefore,
			util.MaxSompi)
		return 0, ruleError(ErrBadTxOutValue, str)
	}
	return totalSompiInAfter, nil
}

func checkReferencedOutputsAreAvailable(
	tx *util.Tx, utxoSet UTXOSet, txIn *appmessage.TxIn, txInIndex int) (*UTXOEntry, error) {

	entry, ok := utxoSet.Get(txIn.PreviousOutpoint)
	if !ok {
		str := fmt.Sprintf("output %s referenced from "+
			"transaction %s input %d either does not exist or "+
			"has already been spent", txIn.PreviousOutpoint,
			tx.ID(), txInIndex)
		return nil, ruleError(ErrMissingTxOut, str)
	}
	return entry, nil
}

func validateCoinbaseMaturity(dagParams *dagconfig.Params, entry *UTXOEntry, txBlueScore uint64, txIn *appmessage.TxIn) error {
	// Ensure the transaction is not spending coins which have not
	// yet reached the required coinbase maturity.
	if entry.IsCoinbase() {
		originBlueScore := entry.BlockBlueScore()
		blueScoreSincePrev := txBlueScore - originBlueScore
		if blueScoreSincePrev < dagParams.BlockCoinbaseMaturity {
			str := fmt.Sprintf("tried to spend coinbase "+
				"transaction output %s from blue score %d "+
				"to blue score %d before required maturity "+
				"of %d", txIn.PreviousOutpoint,
				originBlueScore, txBlueScore,
				dagParams.BlockCoinbaseMaturity)
			return ruleError(ErrImmatureSpend, str)
		}
	}
	return nil
}

func (dag *BlockDAG) checkConnectBlockToPastUTXO(
	node *blockNode, pastUTXO UTXOSet, transactions []*util.Tx) (err error) {

	selectedParentMedianTime := node.selectedParentMedianTime()

	totalFee := uint64(0)

	for _, tx := range transactions {
		txFee, _, err :=
			dag.checkConnectTransactionToPastUTXO(node, tx, pastUTXO, 0, selectedParentMedianTime)

		if err != nil {
			return err
		}

		totalFee, err = dag.checkTotalFee(totalFee, txFee)
		if err != nil {
			return err
		}
	}

	return nil
}

type txInputAndUTXOEntry struct {
	txIn      *appmessage.TxIn
	utxoEntry *UTXOEntry
}

func (dag *BlockDAG) checkConnectTransactionToPastUTXO(
	node *blockNode, tx *util.Tx, pastUTXO UTXOSet, accumulatedMassBefore uint64, selectedParentMedianTime mstime.Time) (
	txFee uint64, accumulatedMassAfter uint64, err error) {

	err = checkTxIsNotDuplicate(tx, pastUTXO)
	if err != nil {
		return 0, 0, err
	}

	inputsWithUTXOEntries, err := dag.getReferencedUTXOEntries(tx, pastUTXO)
	if err != nil {
		return 0, 0, err
	}

	accumulatedMassAfter, err = dag.checkTxMass(tx, inputsWithUTXOEntries, accumulatedMassBefore)
	if err != nil {
		return 0, 0, err
	}

	err = dag.checkTxCoinbaseMaturity(node, inputsWithUTXOEntries)
	if err != nil {
		return 0, 0, nil
	}

	totalSompiIn, err := dag.checkTxInputAmounts(inputsWithUTXOEntries)
	if err != nil {
		return 0, 0, nil
	}

	totalSompiOut, err := dag.checkTxOutputAmounts(tx, totalSompiIn)
	if err != nil {
		return 0, 0, nil
	}

	txFee = totalSompiIn - totalSompiOut

	err = dag.checkTxSequenceLock(node, tx, inputsWithUTXOEntries, selectedParentMedianTime)
	if err != nil {
		return 0, 0, nil
	}

	err = ValidateTransactionScripts(tx, pastUTXO, txscript.ScriptNoFlags, dag.sigCache)
	if err != nil {
		return 0, 0, err
	}

	return txFee, accumulatedMassAfter, nil
}

func (dag *BlockDAG) checkTxSequenceLock(node *blockNode, tx *util.Tx,
	inputsWithUTXOEntries []*txInputAndUTXOEntry, medianTime mstime.Time) error {

	// A transaction can only be included within a block
	// once the sequence locks of *all* its inputs are
	// active.
	sequenceLock, err := dag.calcTxSequenceLockFromInputsWithUTXOEntries(node, tx, inputsWithUTXOEntries)
	if err != nil {
		return err
	}
	if !SequenceLockActive(sequenceLock, node.blueScore, medianTime) {
		str := fmt.Sprintf("block contains " +
			"transaction whose input sequence " +
			"locks are not met")
		return ruleError(ErrUnfinalizedTx, str)
	}

	return nil
}

func (dag *BlockDAG) checkTxOutputAmounts(tx *util.Tx, totalSompiIn uint64) (uint64, error) {
	totalSompiOut := uint64(0)
	// Calculate the total output amount for this transaction. It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	for _, txOut := range tx.MsgTx().TxOut {
		totalSompiOut += txOut.Value
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalSompiIn < totalSompiOut {
		str := fmt.Sprintf("total value of all transaction inputs for "+
			"transaction %s is %d which is less than the amount "+
			"spent of %d", tx.ID(), totalSompiIn, totalSompiOut)
		return 0, ruleError(ErrSpendTooHigh, str)
	}
	return totalSompiOut, nil
}

func (dag *BlockDAG) checkTxInputAmounts(
	inputsWithUTXOEntries []*txInputAndUTXOEntry) (totalSompiIn uint64, err error) {

	totalSompiIn = 0

	for _, txInAndReferencedUTXOEntry := range inputsWithUTXOEntries {
		utxoEntry := txInAndReferencedUTXOEntry.utxoEntry

		// Ensure the transaction amounts are in range. Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction. All amounts in
		// a transaction are in a unit value known as a sompi. One
		// kaspa is a quantity of sompi as defined by the
		// SompiPerKaspa constant.
		originTxSompi := utxoEntry.Amount()
		if originTxSompi > util.MaxSompi {
			str := fmt.Sprintf("transaction output value of %s is "+
				"higher than max allowed value of %d",
				util.Amount(originTxSompi),
				util.MaxSompi)
			return 0, ruleError(ErrBadTxOutValue, str)
		}

		// The total of all outputs must not be more than the max
		// allowed per transaction. Also, we could potentially overflow
		// the accumulator so check for overflow.
		totalSompiInAfter := totalSompiIn + originTxSompi
		if totalSompiInAfter < totalSompiIn || totalSompiInAfter > util.MaxSompi {
			str := fmt.Sprintf("total value of all transaction "+
				"inputs is %d which is higher than max "+
				"allowed value of %d", totalSompiInAfter, util.MaxSompi)
			return 0, ruleError(ErrBadTxOutValue, str)
		}
		totalSompiIn = totalSompiInAfter
	}

	return totalSompiIn, nil
}

func (dag *BlockDAG) checkTxCoinbaseMaturity(
	node *blockNode, inputsWithUTXOEntries []*txInputAndUTXOEntry) error {
	txBlueScore := node.blueScore
	for _, txInAndReferencedUTXOEntry := range inputsWithUTXOEntries {
		txIn := txInAndReferencedUTXOEntry.txIn
		utxoEntry := txInAndReferencedUTXOEntry.utxoEntry

		if utxoEntry.IsCoinbase() {
			originBlueScore := utxoEntry.BlockBlueScore()
			blueScoreSincePrev := txBlueScore - originBlueScore
			if blueScoreSincePrev < dag.Params.BlockCoinbaseMaturity {
				str := fmt.Sprintf("tried to spend coinbase "+
					"transaction output %s from blue score %d "+
					"to blue score %d before required maturity "+
					"of %d", txIn.PreviousOutpoint,
					originBlueScore, txBlueScore,
					dag.Params.BlockCoinbaseMaturity)

				return ruleError(ErrImmatureSpend, str)
			}
		}
	}

	return nil
}

func (dag *BlockDAG) checkTxMass(tx *util.Tx, inputsWithUTXOEntries []*txInputAndUTXOEntry,
	accumulatedMassBefore uint64) (accumulatedMassAfter uint64, err error) {

	txMass := calcTxMassFromInputsWithUTXOEntries(tx, inputsWithUTXOEntries)

	accumulatedMassAfter = accumulatedMassBefore + txMass

	// We could potentially overflow the accumulator so check for
	// overflow as well.
	if accumulatedMassAfter < txMass || accumulatedMassAfter > appmessage.MaxMassAcceptedByBlock {
		str := fmt.Sprintf("block accepts transactions with accumulated mass higher then allowed limit of %d",
			appmessage.MaxMassAcceptedByBlock)
		return 0, ruleError(ErrBlockMassTooHigh, str)
	}

	return accumulatedMassAfter, nil
}

func (dag *BlockDAG) getReferencedUTXOEntries(tx *util.Tx, utxoSet UTXOSet) (
	[]*txInputAndUTXOEntry, error) {

	txIns := tx.MsgTx().TxIn
	inputsWithUTXOEntries := make([]*txInputAndUTXOEntry, 0, len(txIns))

	for txInIndex, txIn := range txIns {
		utxoEntry, ok := utxoSet.Get(txIn.PreviousOutpoint)
		if !ok {
			str := fmt.Sprintf("output %s referenced from "+
				"transaction %s input %d either does not exist or "+
				"has already been spent", txIn.PreviousOutpoint,
				tx.ID(), txInIndex)
			return nil, ruleError(ErrMissingTxOut, str)
		}

		inputsWithUTXOEntries = append(inputsWithUTXOEntries, &txInputAndUTXOEntry{
			txIn:      txIn,
			utxoEntry: utxoEntry,
		})
	}

	return inputsWithUTXOEntries, nil
}

func (dag *BlockDAG) checkTotalFee(totalFees uint64, txFee uint64) (uint64, error) {
	// Sum the total fees and ensure we don't overflow the
	// accumulator.
	lastTotalFees := totalFees
	totalFees += txFee
	if totalFees < lastTotalFees || totalFees > util.MaxSompi {
		str := fmt.Sprintf("total fees are higher then max allowed value of %d", util.MaxSompi)
		return 0, ruleError(ErrBadFees, str)
	}
	return totalFees, nil
}

func (node *blockNode) validateUTXOCommitment(multiset *secp256k1.MultiSet) error {
	calculatedMultisetHash := daghash.Hash(*multiset.Finalize())
	if !calculatedMultisetHash.IsEqual(node.utxoCommitment) {
		str := fmt.Sprintf("block %s UTXO commitment is invalid - block "+
			"header indicates %s, but calculated value is %s", node.hash,
			node.utxoCommitment, calculatedMultisetHash)
		return ruleError(ErrBadUTXOCommitment, str)
	}

	return nil
}

// CheckConnectBlockTemplate fully validates that connecting the passed block to
// the DAG does not violate any consensus rules, aside from the proof of
// work requirement.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) CheckConnectBlockTemplate(block *util.Block) error {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()
	return dag.CheckConnectBlockTemplateNoLock(block)
}

// CheckConnectBlockTemplateNoLock fully validates that connecting the passed block to
// the DAG does not violate any consensus rules, aside from the proof of
// work requirement. The block must connect to the current tip of the main dag.
func (dag *BlockDAG) CheckConnectBlockTemplateNoLock(block *util.Block) error {
	// Skip the proof of work check as this is just a block template.
	flags := BFNoPoWCheck

	header := block.MsgBlock().Header

	err := dag.checkBlockSanity(block, flags)
	if err != nil {
		return err
	}

	_, isDelayed := dag.shouldBlockBeDelayed(block)
	if isDelayed {
		return errors.Errorf("Block timestamp is too far in the future")
	}

	err = dag.checkBlockContext(block, flags)
	if err != nil {
		return err
	}

	templateParents := newBlockSet()
	for _, parentHash := range header.ParentHashes {
		parent, ok := dag.index.LookupNode(parentHash)
		if !ok {
			return errors.Errorf("Couldn't find parent of block template with hash `%s`", parentHash)
		}
		templateParents.add(parent)
	}

	templateNode, _ := dag.newBlockNode(&header, templateParents)

	err = dag.checkConnectBlockToPastUTXO(templateNode, dag.UTXOSet(), block.Transactions())

	return err
}

func (dag *BlockDAG) checkDuplicateBlock(blockHash *daghash.Hash, flags BehaviorFlags) error {
	wasBlockStored := flags&BFWasStored == BFWasStored
	if dag.IsInDAG(blockHash) && !wasBlockStored {
		str := fmt.Sprintf("already have block %s", blockHash)
		return ruleError(ErrDuplicateBlock, str)
	}

	// The block must not already exist as an orphan.
	if _, exists := dag.orphans[*blockHash]; exists {
		str := fmt.Sprintf("already have block (orphan) %s", blockHash)
		return ruleError(ErrDuplicateBlock, str)
	}

	if dag.isKnownDelayedBlock(blockHash) {
		str := fmt.Sprintf("already have block (delayed) %s", blockHash)
		return ruleError(ErrDuplicateBlock, str)
	}

	return nil
}
