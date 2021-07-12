package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
)

const (
	// maxStandardP2SHSigOps is the maximum number of signature operations
	// that are considered standard in a pay-to-script-hash script.
	maxStandardP2SHSigOps = 15

	// maximumStandardSignatureScriptSize is the maximum size allowed for a
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
	maximumStandardSignatureScriptSize = 1650

	// maximumStandardTransactionMass is the maximum mass allowed for transactions that
	// are considered standard and will therefore be relayed and considered for mining.
	maximumStandardTransactionMass = 100000
)

func (mp *mempool) checkTransactionStandardInIsolation(transaction *externalapi.DomainTransaction) error {
	// The transaction must be a currently supported version.
	//
	// This check is currently mirrored in consensus.
	// However, in a later version of Kaspa the consensus-valid transaction version range might diverge from the
	// standard transaction version range, and thus the validation should happen in both levels.
	if transaction.Version > mp.config.MaximumStandardTransactionVersion ||
		transaction.Version < mp.config.MinimumStandardTransactionVersion {
		str := fmt.Sprintf("transaction version %d is not in the valid range of %d-%d", transaction.Version,
			mp.config.MinimumStandardTransactionVersion, mp.config.MaximumStandardTransactionVersion)
		return transactionRuleError(RejectNonstandard, str)
	}

	// Since extremely large transactions with a lot of inputs can cost
	// almost as much to process as the sender fees, limit the maximum
	// size of a transaction. This also helps mitigate CPU exhaustion
	// attacks.
	serializedLength := estimatedsize.TransactionEstimatedSerializedSize(transaction)
	if serializedLength > maximumStandardTransactionMass {
		str := fmt.Sprintf("transaction size of %d is larger than max allowed size of %d",
			serializedLength, maximumStandardTransactionMass)
		return transactionRuleError(RejectNonstandard, str)
	}

	for i, input := range transaction.Inputs {
		// Each transaction input signature script must not exceed the
		// maximum size allowed for a standard transaction. See
		// the comment on maximumStandardSignatureScriptSize for more details.
		signatureScriptLen := len(input.SignatureScript)
		if signatureScriptLen > maximumStandardSignatureScriptSize {
			str := fmt.Sprintf("transaction input %d: signature script size of %d bytes is larger than the "+
				"maximum allowed size of %d bytes", i, signatureScriptLen, maximumStandardSignatureScriptSize)
			return transactionRuleError(RejectNonstandard, str)
		}
	}

	// None of the output public key scripts can be a non-standard script or be "dust".
	for i, output := range transaction.Outputs {
		if output.ScriptPublicKey.Version > constants.MaxScriptPublicKeyVersion {
			return transactionRuleError(RejectNonstandard, "The version of the scriptPublicKey is higher than the known version.")
		}
		scriptClass := txscript.GetScriptClass(output.ScriptPublicKey.Script)
		if scriptClass == txscript.NonStandardTy {
			str := fmt.Sprintf("transaction output %d: non-standard script form", i)
			return transactionRuleError(RejectNonstandard, str)
		}

		if mp.IsTransactionOutputDust(output) {
			str := fmt.Sprintf("transaction output %d: payment "+
				"of %d is dust", i, output.Value)
			return transactionRuleError(RejectDust, str)
		}
	}

	return nil
}

// IsTransactionOutputDust returns whether or not the passed transaction output amount
// is considered dust or not based on the configured minimum transaction relay fee.
// Dust is defined in terms of the minimum transaction relay fee. In
// particular, if the cost to the network to spend coins is more than 1/3 of the
// minimum transaction relay fee, it is considered dust.
//
// It is exported for use by transaction generators and wallets
func (mp *mempool) IsTransactionOutputDust(output *externalapi.DomainTransactionOutput) bool {
	// Unspendable outputs are considered dust.
	if txscript.IsUnspendable(output.ScriptPublicKey.Script) {
		return true
	}

	// The total serialized size consists of the output and the associated
	// input script to redeem it. Since there is no input script
	// to redeem it yet, use the minimum size of a typical input script.
	//
	// Pay-to-pubkey bytes breakdown:
	//
	//  Output to pubkey (43 bytes):
	//   8 value, 1 script len, 34 script [1 OP_DATA_32,
	//   32 pubkey, 1 OP_CHECKSIG]
	//
	//  Input (105 bytes):
	//   36 prev outpoint, 1 script len, 64 script [1 OP_DATA_64,
	//   64 sig], 4 sequence
	//
	// The most common scripts are pay-to-pubkey, and as per the above
	// breakdown, the minimum size of a p2pk input script is 148 bytes. So
	// that figure is used.
	totalSerializedSize := estimatedsize.TransactionOutputEstimatedSerializedSize(output) + 148

	// The output is considered dust if the cost to the network to spend the
	// coins is more than 1/3 of the minimum free transaction relay fee.
	// mp.config.MinimumRelayTransactionFee is in sompi/KB, so multiply
	// by 1000 to convert to bytes.
	//
	// Using the typical values for a pay-to-pubkey transaction from
	// the breakdown above and the default minimum free transaction relay
	// fee of 1000, this equates to values less than 546 sompi being
	// considered dust.
	//
	// The following is equivalent to (value/totalSerializedSize) * (1/3) * 1000
	// without needing to do floating point math.
	return output.Value*1000/(3*totalSerializedSize) < uint64(mp.config.MinimumRelayTransactionFee)
}

// checkTransactionStandardInContext performs a series of checks on a transaction's
// inputs to ensure they are "standard". A standard transaction input within the
// context of this function is one whose referenced public key script is of a
// standard form and, for pay-to-script-hash, does not have more than
// maxStandardP2SHSigOps signature operations.
// In addition, makes sure that the transaction's fee is above the minimum for acceptance
// into the mempool and relay
func (mp *mempool) checkTransactionStandardInContext(transaction *externalapi.DomainTransaction) error {
	for i, input := range transaction.Inputs {
		// It is safe to elide existence and index checks here since
		// they have already been checked prior to calling this
		// function.
		utxoEntry := input.UTXOEntry
		originScriptPubKey := utxoEntry.ScriptPublicKey()
		switch txscript.GetScriptClass(originScriptPubKey.Script) {
		case txscript.ScriptHashTy:
			numSigOps := txscript.GetPreciseSigOpCount(
				input.SignatureScript, originScriptPubKey, true)
			if numSigOps > maxStandardP2SHSigOps {
				str := fmt.Sprintf("transaction input #%d has %d signature operations which is more "+
					"than the allowed max amount of %d", i, numSigOps, maxStandardP2SHSigOps)
				return transactionRuleError(RejectNonstandard, str)
			}

		case txscript.NonStandardTy:
			str := fmt.Sprintf("transaction input #%d has a non-standard script form", i)
			return transactionRuleError(RejectNonstandard, str)
		}
	}

	serializedSize := estimatedsize.TransactionEstimatedSerializedSize(transaction)
	minimumFee := mp.minimumRequiredTransactionRelayFee(serializedSize)
	if transaction.Fee < minimumFee {
		str := fmt.Sprintf("transaction %s has %d fees which is under the required amount of %d",
			consensushashing.TransactionID(transaction), transaction.Fee, minimumFee)
		return transactionRuleError(RejectInsufficientFee, str)
	}

	return nil
}

// minimumRequiredTransactionRelayFee returns the minimum transaction fee required for a
// transaction with the passed mass to be accepted into the mampool and relayed.
func (mp *mempool) minimumRequiredTransactionRelayFee(mass uint64) uint64 {
	// Calculate the minimum fee for a transaction to be allowed into the
	// mempool and relayed by scaling the base fee. MinimumRelayTransactionFee is in
	// sompi/kg so multiply by mass (which is in grams) and divide by 1000 to get minimum sompis.
	minimumFee := (mass * uint64(mp.config.MinimumRelayTransactionFee)) / 1000

	if minimumFee == 0 && mp.config.MinimumRelayTransactionFee > 0 {
		minimumFee = uint64(mp.config.MinimumRelayTransactionFee)
	}

	// Set the minimum fee to the maximum possible value if the calculated
	// fee is not in the valid range for monetary amounts.
	if minimumFee > util.MaxSompi {
		minimumFee = util.MaxSompi
	}

	return minimumFee
}
