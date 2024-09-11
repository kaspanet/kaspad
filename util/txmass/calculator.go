package txmass

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

// Calculator exposes methods to calculate the mass of a transaction
type Calculator struct {
	massPerTxByte           uint64
	massPerScriptPubKeyByte uint64
	massPerSigOp            uint64

	// The parameter for scaling inverse KAS value to mass units (KIP-0009)
	storageMassParameter uint64
}

// NewCalculator creates a new instance of Calculator
func NewCalculator(massPerTxByte, massPerScriptPubKeyByte, massPerSigOp uint64) *Calculator {
	return &Calculator{
		massPerTxByte:           massPerTxByte,
		massPerScriptPubKeyByte: massPerScriptPubKeyByte,
		massPerSigOp:            massPerSigOp,
		storageMassParameter:    constants.SompiPerKaspa * 10_000,
	}
}

// MassPerTxByte returns the mass per transaction byte configured for this Calculator
func (c *Calculator) MassPerTxByte() uint64 { return c.massPerTxByte }

// MassPerScriptPubKeyByte returns the mass per ScriptPublicKey byte configured for this Calculator
func (c *Calculator) MassPerScriptPubKeyByte() uint64 { return c.massPerScriptPubKeyByte }

// MassPerSigOp returns the mass per SigOp byte configured for this Calculator
func (c *Calculator) MassPerSigOp() uint64 { return c.massPerSigOp }

// CalculateTransactionMass calculates the mass of the given transaction
func (c *Calculator) CalculateTransactionMass(transaction *externalapi.DomainTransaction) uint64 {
	if transactionhelper.IsCoinBase(transaction) {
		return 0
	}

	// calculate mass for size
	size := transactionEstimatedSerializedSize(transaction)
	massForSize := size * c.massPerTxByte

	// calculate mass for scriptPubKey
	totalScriptPubKeySize := uint64(0)
	for _, output := range transaction.Outputs {
		totalScriptPubKeySize += 2 //output.ScriptPublicKey.Version (uint16)
		totalScriptPubKeySize += uint64(len(output.ScriptPublicKey.Script))
	}
	massForScriptPubKey := totalScriptPubKeySize * c.massPerScriptPubKeyByte

	// calculate mass for SigOps
	totalSigOpCount := uint64(0)
	for _, input := range transaction.Inputs {
		totalSigOpCount += uint64(input.SigOpCount)
	}
	massForSigOps := totalSigOpCount * c.massPerSigOp

	// Sum all components of mass
	return massForSize + massForScriptPubKey + massForSigOps
}

// CalculateTransactionStorageMass calculates the storage mass of the given transaction (see KIP-0009)
func (c *Calculator) CalculateTransactionStorageMass(transaction *externalapi.DomainTransaction) uint64 {
	if transactionhelper.IsCoinBase(transaction) {
		return 0
	}

	outsLen := uint64(len(transaction.Outputs))
	insLen := uint64(len(transaction.Inputs))

	if insLen == 0 {
		panic("Storage mass calculation expects at least one input")
	}

	harmonicOuts := uint64(0)
	for _, output := range transaction.Outputs {
		inverseOut := c.storageMassParameter / output.Value
		if harmonicOuts+inverseOut < harmonicOuts {
			// Overflow detected. This requires 10^7 outputs so is unrealistic for wallet usages.
			// If this method is ever used for consensus, this case should be handled by returning an err
			panic("Unexpected overflow in storage mass calculation")
		}
		harmonicOuts += inverseOut
	}

	if outsLen == 1 || insLen == 1 || (outsLen == 2 && insLen == 2) {
		harmonicDiff := harmonicOuts
		for _, input := range transaction.Inputs {
			if input.UTXOEntry == nil {
				panic("Storage mass calculation expects a fully populated transaction")
			}
			inverseIn := c.storageMassParameter / input.UTXOEntry.Amount()
			if harmonicDiff < inverseIn {
				harmonicDiff = 0
			} else {
				harmonicDiff -= inverseIn
			}
		}
		return harmonicDiff
	}

	sumIns := uint64(0)
	for _, input := range transaction.Inputs {
		if input.UTXOEntry == nil {
			panic("Storage mass calculation expects a fully populated transaction")
		}
		// Total supply is bounded, so a sum of existing UTXO entries cannot overflow (nor can it be zero)
		sumIns += input.UTXOEntry.Amount()
	}
	meanIns := sumIns / insLen
	inverseMeanIns := c.storageMassParameter / meanIns
	arithmeticIns := insLen * inverseMeanIns

	if arithmeticIns < inverseMeanIns {
		// overflow (so subtraction would be negative)
		return 0
	}
	if harmonicOuts < arithmeticIns {
		// underflow
		return 0
	} else {
		return harmonicOuts - arithmeticIns
	}
}

// CalculateTransactionOverallMass calculates the overall mass of the transaction including compute and storage mass components (see KIP-0009)
func (c *Calculator) CalculateTransactionOverallMass(transaction *externalapi.DomainTransaction) uint64 {
	return max(c.CalculateTransactionMass(transaction), c.CalculateTransactionStorageMass(transaction))
}

// transactionEstimatedSerializedSize is the estimated size of a transaction in some
// serialization. This has to be deterministic, but not necessarily accurate, since
// it's only used as the size component in the transaction and block mass limit
// calculation.
func transactionEstimatedSerializedSize(tx *externalapi.DomainTransaction) uint64 {
	if transactionhelper.IsCoinBase(tx) {
		return 0
	}
	size := uint64(0)
	size += 2 // Txn Version
	size += 8 // number of inputs (uint64)
	for _, input := range tx.Inputs {
		size += transactionInputEstimatedSerializedSize(input)
	}

	size += 8 // number of outputs (uint64)
	for _, output := range tx.Outputs {
		size += TransactionOutputEstimatedSerializedSize(output)
	}

	size += 8 // lock time (uint64)
	size += externalapi.DomainSubnetworkIDSize
	size += 8                          // gas (uint64)
	size += externalapi.DomainHashSize // payload hash

	size += 8 // length of the payload (uint64)
	size += uint64(len(tx.Payload))

	return size
}

func transactionInputEstimatedSerializedSize(input *externalapi.DomainTransactionInput) uint64 {
	size := uint64(0)
	size += outpointEstimatedSerializedSize()

	size += 8 // length of signature script (uint64)
	size += uint64(len(input.SignatureScript))

	size += 8 // sequence (uint64)
	return size
}

func outpointEstimatedSerializedSize() uint64 {
	size := uint64(0)
	size += externalapi.DomainHashSize // ID
	size += 4                          // index (uint32)
	return size
}

// TransactionOutputEstimatedSerializedSize is the same as transactionEstimatedSerializedSize but for outputs only
func TransactionOutputEstimatedSerializedSize(output *externalapi.DomainTransactionOutput) uint64 {
	size := uint64(0)
	size += 8 // value (uint64)
	size += 2 // output.ScriptPublicKey.Version (uint 16)
	size += 8 // length of script public key (uint64)
	size += uint64(len(output.ScriptPublicKey.Script))
	return size
}
