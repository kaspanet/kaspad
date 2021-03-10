package consensushashing

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/serialization"
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

type SighashReusedValues struct {
	previousOutputsHash *externalapi.DomainHash
	sequencesHash       *externalapi.DomainHash
	outputsHash         *externalapi.DomainHash
	payloadHash         *externalapi.DomainHash
}

// CalcSignatureHash will, given a script and hash type calculate the signature hash
// to be used for signing and verification.
// This returns error only if one of the provided parameters are invalid.
func CalcSignatureHash_BIP143(prevScriptPublicKey *externalapi.ScriptPublicKey, hashType SigHashType,
	tx *externalapi.DomainTransaction, idx int, reusedValues *SighashReusedValues) (
	*externalapi.DomainHash, *SighashReusedValues, error) {

	if prevScriptPublicKey.Version > constants.MaxScriptPublicKeyVersion {
		return nil, nil, errors.Errorf("Script version is unkown.")
	}
	return calcSignatureHash_BIP143(prevScriptPublicKey, hashType, tx, idx, reusedValues)
}

func calcSignatureHash_BIP143(prevScriptPublicKey *externalapi.ScriptPublicKey, hashType SigHashType,
	tx *externalapi.DomainTransaction, idx int, reusedValues *SighashReusedValues) (
	*externalapi.DomainHash, *SighashReusedValues, error) {

	txIn := tx.Inputs[idx]
	hashWriter := hashes.NewTransactionSigningHashWriter()
	infallibleWriteElement(hashWriter, tx.Version)

	// TODO: PreviousOutputsHash

	// TODO: SequencesHash

	infallibleWriteElement(hashWriter, txIn.PreviousOutpoint.TransactionID)
	infallibleWriteElement(hashWriter, txIn.PreviousOutpoint.Index)

	infallibleWriteElement(hashWriter, prevScriptPublicKey.Version)
	infallibleWriteElement(hashWriter, prevScriptPublicKey.Script)

	// TODO: Previous output value

	infallibleWriteElement(hashWriter, txIn.Sequence)

	// TODO: OutputsHash

	infallibleWriteElement(hashWriter, tx.LockTime)

	payloadHash := getPayloadHash(tx, reusedValues)
	infallibleWriteElement(hashWriter, payloadHash)

	infallibleWriteElement(hashWriter, hashType)

	return hashWriter.Finalize(), nil, nil
}

func getPayloadHash(tx *externalapi.DomainTransaction, reusedValues *SighashReusedValues) *externalapi.DomainHash {
	if reusedValues.payloadHash == nil {
		hashWriter := hashes.NewTransactionSigningHashWriter()
		hashWriter.InfallibleWrite(tx.Payload)
		reusedValues.payloadHash = hashWriter.Finalize()
	}
	return reusedValues.payloadHash
}

func infallibleWriteElement(hashWriter hashes.HashWriter, element interface{}) {
	err := serialization.WriteElement(hashWriter, element)
	if err != nil {
		// It seems like this could only happen if the writer returned an error.
		// and this writer should never return an error (no allocations or possible failures)
		// the only non-writer error path here is unknown types in `WriteElement`
		panic(errors.Wrap(err, "TransactionHashForSigning() failed. this should never fail for structurally-valid transactions"))
	}
}
