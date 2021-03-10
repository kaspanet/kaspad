package consensushashing

import (
	"encoding/binary"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/pkg/errors"
)

// SigHashType represents hash type bits at the end of a signature.
type SigHashType uint32

// Hash type bits from the end of a signature.
const (
	SigHashAll          SigHashType = 0x1
	SigHashNone         SigHashType = 0x2
	SigHashSingle       SigHashType = 0x3
	SigHashAnyOneCanPay SigHashType = 0x80

	// SigHashMask defines the number of bits of the hash type which is used
	// to identify which outputs are signed.
	SigHashMask = 0x1f
)

// CalcSignatureHash will, given a script and hash type for the current script
// engine instance, calculate the signature hash to be used for signing and
// verification.
func CalcSignatureHash(prevScriptPublicKey *externalapi.ScriptPublicKey, hashType SigHashType,
	tx *externalapi.DomainTransaction, idx int) (*externalapi.DomainHash, error) {

	if prevScriptPublicKey.Version > constants.MaxScriptPublicKeyVersion {
		return nil, errors.Errorf("Script version is unkown.")
	}
	return calcSignatureHash(prevScriptPublicKey, hashType, tx, idx)
}

// calcSignatureHash will, given a script and hash type for the current script
// engine instance, calculate the signature hash to be used for signing and
// verification.
func calcSignatureHash(prevScriptPublicKey *externalapi.ScriptPublicKey, hashType SigHashType,
	tx *externalapi.DomainTransaction, idx int) (*externalapi.DomainHash, error) {

	// The SigHashSingle signature type signs only the corresponding input
	// and output (the output with the same index number as the input).
	//
	// Since transactions can have more inputs than outputs, this means it
	// is improper to use SigHashSingle on input indices that don't have a
	// corresponding output.
	if hashType&SigHashMask == SigHashSingle && idx >= len(tx.Outputs) {
		return nil, errors.New("sigHashSingle index out of bounds")
	}

	// Make a shallow copy of the transaction, zeroing out the payload and the
	// script for all inputs that are not currently being processed.
	txCopy := shallowCopyTx(tx)
	txCopy.Payload = []byte{}
	for i := range txCopy.Inputs {
		if i == idx {
			sigScript := prevScriptPublicKey.Script
			var version [2]byte
			binary.LittleEndian.PutUint16(version[:], prevScriptPublicKey.Version)
			txCopy.Inputs[idx].SignatureScript = append(version[:], sigScript...)
		} else {
			txCopy.Inputs[i].SignatureScript = nil
		}
	}

	switch hashType & SigHashMask {
	case SigHashNone:
		txCopy.Outputs = txCopy.Outputs[0:0] // Empty slice.
		for i := range txCopy.Inputs {
			if i != idx {
				txCopy.Inputs[i].Sequence = 0
			}
		}

	case SigHashSingle:
		// Resize output array to up to and including requested index.
		txCopy.Outputs = txCopy.Outputs[:idx+1]

		// All but current output get zeroed out.
		for i := 0; i < idx; i++ {
			txCopy.Outputs[i].Value = 0
			txCopy.Outputs[i].ScriptPublicKey.Script = nil
			txCopy.Outputs[i].ScriptPublicKey.Version = 0
		}

		// Sequence on all other inputs is 0, too.
		for i := range txCopy.Inputs {
			if i != idx {
				txCopy.Inputs[i].Sequence = 0
			}
		}

	default:
		// Consensus treats undefined hashtypes like normal SigHashAll
		// for purposes of hash generation.
		fallthrough
	case SigHashAll:
		// Nothing special here.
	}
	if hashType&SigHashAnyOneCanPay != 0 {
		txCopy.Inputs = txCopy.Inputs[idx : idx+1]
	}

	// The final hash is the hash of both the serialized modified
	// transaction and the hash type (encoded as a 4-byte little-endian
	// value) appended.
	return TransactionHashForSigning(&txCopy, uint32(hashType)), nil
}

// shallowCopyTx creates a shallow copy of the transaction for use when
// calculating the signature hash. It is used over the Copy method on the
// transaction itself since that is a deep copy and therefore does more work and
// allocates much more space than needed.
func shallowCopyTx(tx *externalapi.DomainTransaction) externalapi.DomainTransaction {
	// As an additional memory optimization, use contiguous backing arrays
	// for the copied inputs and outputs and point the final slice of
	// pointers into the contiguous arrays. This avoids a lot of small
	// allocations.
	txCopy := externalapi.DomainTransaction{
		Version:      tx.Version,
		Inputs:       make([]*externalapi.DomainTransactionInput, len(tx.Inputs)),
		Outputs:      make([]*externalapi.DomainTransactionOutput, len(tx.Outputs)),
		LockTime:     tx.LockTime,
		SubnetworkID: tx.SubnetworkID,
		Gas:          tx.Gas,
		Payload:      tx.Payload,
	}
	txIns := make([]externalapi.DomainTransactionInput, len(tx.Inputs))
	for i, oldTxIn := range tx.Inputs {
		txIns[i] = *oldTxIn
		txCopy.Inputs[i] = &txIns[i]
	}
	txOuts := make([]externalapi.DomainTransactionOutput, len(tx.Outputs))
	for i, oldTxOut := range tx.Outputs {
		txOuts[i] = *oldTxOut
		txCopy.Outputs[i] = &txOuts[i]
	}
	return txCopy
}
