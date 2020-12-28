package estimatedsize

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

// TransactionEstimatedSerializedSize is the estimated size of a transaction in some
// serialization. This has to be deterministic, but not necessarily accurate, since
// it's only used as the size component in the transaction mass and block size limit
// calculation.
func TransactionEstimatedSerializedSize(tx *externalapi.DomainTransaction) uint64 {
	if transactionhelper.IsCoinBase(tx) {
		return 0
	}
	size := uint64(0)

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

// TransactionOutputEstimatedSerializedSize is the same as TransactionEstimatedSerializedSize but for outputs only
func TransactionOutputEstimatedSerializedSize(output *externalapi.DomainTransactionOutput) uint64 {
	size := uint64(0)
	size += 8 // value (uint64)
	size += 2 //output.ScriptPublicKey.Version (uint 16)
	size += 8 // length of script public key (uint64)
	size += uint64(len(output.ScriptPublicKey.Script))
	return size
}
