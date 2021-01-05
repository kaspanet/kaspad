// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"

	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
)

const (
	// maxStandardP2SHSigOps is the maximum number of signature operations
	// that are considered standard in a pay-to-script-hash script.
	maxStandardP2SHSigOps = 15

	// maxStandardSigScriptSize is the maximum size allowed for a
	// transaction input signature script to be considered standard. This
	// value allows for a 15-of-15 CHECKMULTISIG pay-to-script-hash with
	// compressed keys.
	//
	// The form of the overall script is: OP_0 <15 signatures> OP_PUSHDATA2
	// <2 bytes len> [OP_15 <15 pubkeys> OP_15 OP_CHECKMULTISIG]
	//
	// For the p2sh script portion, each of the 15 compressed pubkeys are
	// 33 bytes (plus one for the OP_DATA_33 opcode), and the thus it totals
	// to (15*34)+3 = 513 bytes. Next, each of the 15 signatures is a max
	// of 73 bytes (plus one for the OP_DATA_73 opcode). Also, there is one
	// extra byte for the initial extra OP_0 push and 3 bytes for the
	// OP_PUSHDATA2 needed to specify the 513 bytes for the script push.
	// That brings the total to 1+(15*74)+3+513 = 1627. This value also
	// adds a few extra bytes to provide a little buffer.
	// (1 + 15*74 + 3) + (15*34 + 3) + 23 = 1650
	maxStandardSigScriptSize = 1650

	// MaxStandardTxSize is the maximum size allowed for transactions that
	// are considered standard and will therefore be relayed and considered
	// for mining.
	MaxStandardTxSize = 100000

	// DefaultMinRelayTxFee is the minimum fee in sompi that is required
	// for a transaction to be treated as free for relay and mining
	// purposes. It is also used to help determine if a transaction is
	// considered dust and as a base for calculating minimum required fees
	// for larger transactions. This value is in sompi/1000 bytes.
	DefaultMinRelayTxFee = util.Amount(1000)
)

// calcMinRequiredTxRelayFee returns the minimum transaction fee required for a
// transaction with the passed serialized size to be accepted into the memory
// pool and relayed.
func calcMinRequiredTxRelayFee(serializedSize int64, minRelayTxFee util.Amount) int64 {
	// Calculate the minimum fee for a transaction to be allowed into the
	// mempool and relayed by scaling the base fee. minTxRelayFee is in
	// sompi/kB so multiply by serializedSize (which is in bytes) and
	// divide by 1000 to get minimum sompis.
	minFee := (serializedSize * int64(minRelayTxFee)) / 1000

	if minFee == 0 && minRelayTxFee > 0 {
		minFee = int64(minRelayTxFee)
	}

	// Set the minimum fee to the maximum possible value if the calculated
	// fee is not in the valid range for monetary amounts.
	if minFee < 0 || minFee > util.MaxSompi {
		minFee = util.MaxSompi
	}

	return minFee
}

// checkInputsStandard performs a series of checks on a transaction's inputs
// to ensure they are "standard". A standard transaction input within the
// context of this function is one whose referenced public key script is of a
// standard form and, for pay-to-script-hash, does not have more than
// maxStandardP2SHSigOps signature operations.
func checkInputsStandard(tx *consensusexternalapi.DomainTransaction) error {
	// NOTE: The reference implementation also does a coinbase check here,
	// but coinbases have already been rejected prior to calling this
	// function so no need to recheck.

	for i, txIn := range tx.Inputs {
		// It is safe to elide existence and index checks here since
		// they have already been checked prior to calling this
		// function.
		entry := txIn.UTXOEntry
		originScriptPubKey := entry.ScriptPublicKey()
		switch txscript.GetScriptClass(originScriptPubKey.Script) {
		case txscript.ScriptHashTy:
			numSigOps := txscript.GetPreciseSigOpCount(
				txIn.SignatureScript, originScriptPubKey, true)
			if numSigOps > maxStandardP2SHSigOps {
				str := fmt.Sprintf("transaction input #%d has "+
					"%d signature operations which is more "+
					"than the allowed max amount of %d",
					i, numSigOps, maxStandardP2SHSigOps)
				return txRuleError(RejectNonstandard, str)
			}

		case txscript.NonStandardTy:
			str := fmt.Sprintf("transaction input #%d has a "+
				"non-standard script form", i)
			return txRuleError(RejectNonstandard, str)
		}
	}

	return nil
}

// isDust returns whether or not the passed transaction output amount is
// considered dust or not based on the passed minimum transaction relay fee.
// Dust is defined in terms of the minimum transaction relay fee. In
// particular, if the cost to the network to spend coins is more than 1/3 of the
// minimum transaction relay fee, it is considered dust.
func isDust(txOut *consensusexternalapi.DomainTransactionOutput, minRelayTxFee util.Amount) bool {
	// Unspendable outputs are considered dust.
	if txscript.IsUnspendable(txOut.ScriptPublicKey.Script) {
		return true
	}

	// The total serialized size consists of the output and the associated
	// input script to redeem it. Since there is no input script
	// to redeem it yet, use the minimum size of a typical input script.
	//
	// Pay-to-pubkey-hash bytes breakdown:
	//
	//  Output to hash (34 bytes):
	//   8 value, 1 script len, 25 script [1 OP_DUP, 1 OP_HASH_160,
	//   1 OP_DATA_20, 20 hash, 1 OP_EQUALVERIFY, 1 OP_CHECKSIG]
	//
	//  Input with compressed pubkey (148 bytes):
	//   36 prev outpoint, 1 script len, 107 script [1 OP_DATA_72, 72 sig,
	//   1 OP_DATA_33, 33 compressed pubkey], 4 sequence
	//
	//  Input with uncompressed pubkey (180 bytes):
	//   36 prev outpoint, 1 script len, 139 script [1 OP_DATA_72, 72 sig,
	//   1 OP_DATA_65, 65 compressed pubkey], 4 sequence
	//
	// Pay-to-pubkey bytes breakdown:
	//
	//  Output to compressed pubkey (44 bytes):
	//   8 value, 1 script len, 35 script [1 OP_DATA_33,
	//   33 compressed pubkey, 1 OP_CHECKSIG]
	//
	//  Output to uncompressed pubkey (76 bytes):
	//   8 value, 1 script len, 67 script [1 OP_DATA_65, 65 pubkey,
	//   1 OP_CHECKSIG]
	//
	//  Input (114 bytes):
	//   36 prev outpoint, 1 script len, 73 script [1 OP_DATA_72,
	//   72 sig], 4 sequence
	//
	// Theoretically this could examine the script type of the output script
	// and use a different size for the typical input script size for
	// pay-to-pubkey vs pay-to-pubkey-hash inputs per the above breakdowns,
	// but the only combination which is less than the value chosen is
	// a pay-to-pubkey script with a compressed pubkey, which is not very
	// common.
	//
	// The most common scripts are pay-to-pubkey-hash, and as per the above
	// breakdown, the minimum size of a p2pkh input script is 148 bytes. So
	// that figure is used.
	totalSize := estimatedsize.TransactionOutputEstimatedSerializedSize(txOut) + 148

	// The output is considered dust if the cost to the network to spend the
	// coins is more than 1/3 of the minimum free transaction relay fee.
	// minFreeTxRelayFee is in sompi/KB, so multiply by 1000 to
	// convert to bytes.
	//
	// Using the typical values for a pay-to-pubkey-hash transaction from
	// the breakdown above and the default minimum free transaction relay
	// fee of 1000, this equates to values less than 546 sompi being
	// considered dust.
	//
	// The following is equivalent to (value/totalSize) * (1/3) * 1000
	// without needing to do floating point math.
	return txOut.Value*1000/(3*totalSize) < uint64(minRelayTxFee)
}

// checkTransactionStandard performs a series of checks on a transaction to
// ensure it is a "standard" transaction. A standard transaction is one that
// conforms to several additional limiting cases over what is considered a
// "sane" transaction such as having a version in the supported range, being
// finalized, conforming to more stringent size constraints, having scripts
// of recognized forms, and not containing "dust" outputs (those that are
// so small it costs more to process them than they are worth).
func checkTransactionStandard(tx *consensusexternalapi.DomainTransaction, policy *policy) error {

	// The transaction must be a currently supported version.
	if tx.Version > policy.MaxTxVersion {
		str := fmt.Sprintf("transaction version %d is not in the "+
			"valid range of %d-%d", tx.Version, 0,
			policy.MaxTxVersion)
		return txRuleError(RejectNonstandard, str)
	}

	// Since extremely large transactions with a lot of inputs can cost
	// almost as much to process as the sender fees, limit the maximum
	// size of a transaction. This also helps mitigate CPU exhaustion
	// attacks.
	serializedLen := estimatedsize.TransactionEstimatedSerializedSize(tx)
	if serializedLen > MaxStandardTxSize {
		str := fmt.Sprintf("transaction size of %d is larger than max "+
			"allowed size of %d", serializedLen, MaxStandardTxSize)
		return txRuleError(RejectNonstandard, str)
	}

	for i, txIn := range tx.Inputs {
		// Each transaction input signature script must not exceed the
		// maximum size allowed for a standard transaction. See
		// the comment on maxStandardSigScriptSize for more details.
		sigScriptLen := len(txIn.SignatureScript)
		if sigScriptLen > maxStandardSigScriptSize {
			str := fmt.Sprintf("transaction input %d: signature "+
				"script size of %d bytes is large than max "+
				"allowed size of %d bytes", i, sigScriptLen,
				maxStandardSigScriptSize)
			return txRuleError(RejectNonstandard, str)
		}

		// Each transaction input signature script must only contain
		// opcodes which push data onto the stack.
		isPushOnly, err := txscript.IsPushOnlyScript(txIn.SignatureScript)
		if err != nil {
			str := fmt.Sprintf("transaction input %d: IsPushOnlyScript: %t. Error %s", i, isPushOnly, err)
			return txRuleError(RejectNonstandard, str)
		}
		if !isPushOnly {
			str := fmt.Sprintf("transaction input %d: signature "+
				"script is not push only", i)
			return txRuleError(RejectNonstandard, str)
		}
	}

	// None of the output public key scripts can be a non-standard script or
	// be "dust".
	for i, txOut := range tx.Outputs {
		if txOut.ScriptPublicKey.Version > constants.ScriptPublicKeyVersion {
			return txRuleError(RejectNonstandard, "The version of the scriptPublicKey is higher than the known version.")
		}
		scriptClass := txscript.GetScriptClass(txOut.ScriptPublicKey.Script)
		if scriptClass == txscript.NonStandardTy {
			str := fmt.Sprintf("transaction output %d: non-standard script form", i)
			return txRuleError(RejectNonstandard, str)
		}

		if isDust(txOut, policy.MinRelayTxFee) {
			str := fmt.Sprintf("transaction output %d: payment "+
				"of %d is dust", i, txOut.Value)
			return txRuleError(RejectDust, str)
		}
	}

	return nil
}
