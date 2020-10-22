package validator

import "github.com/kaspanet/kaspad/domain/consensus/model"

// transactionEstimatedSerializedSize is the estimated size of a transaction in some
// serialization. This has to be deterministic, but not necessarily accurate, since
// it's only used as the size component in the transaction mass and block size limit
// calculation.
func (v *validator) transactionEstimatedSerializedSize(tx *model.DomainTransaction) uint64 {
	size := uint64(0)

	size += 8 // number of inputs (uint64)
	for _, input := range tx.Inputs {
		size += v.transactionInputEstimatedSerializedSize(input)
	}

	size += 8 // number of outputs (uint64)
	for _, output := range tx.Outputs {
		size += v.transactionOutputEstimatedSerializedSize(output)
	}

	size += 8 // lock time (uint64)
	size += model.SubnetworkIDSize
	size += 8              // gas (uint64)
	size += model.HashSize // payload hash

	size += 8 // length of the payload (uint64)
	size += uint64(len(tx.Payload))

	return size
}

func (v *validator) transactionInputEstimatedSerializedSize(input *model.DomainTransactionInput) uint64 {
	size := uint64(0)
	size += v.outpointEstimatedSerializedSize()

	size += 8 // length of signature script (uint64)
	size += uint64(len(input.SignatureScript))

	size += 8 // sequence (uint64)
	return size
}

func (v *validator) outpointEstimatedSerializedSize() uint64 {
	size := uint64(0)
	size += model.HashSize // ID
	size += 4              // index (uint32)
	return size
}

func (v *validator) transactionOutputEstimatedSerializedSize(output *model.DomainTransactionOutput) uint64 {
	size := uint64(0)
	size += 8 // value (uint64)
	size += 8 // length of script public key (uint64)
	size += uint64(len(output.ScriptPublicKey))
	return size
}
