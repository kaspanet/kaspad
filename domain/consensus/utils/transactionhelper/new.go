package transactionhelper

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

// NewSubnetworkTransaction returns a new trsnactions in the specified subnetwork with specified gas and payload
func NewSubnetworkTransaction(version int32, inputs []*externalapi.DomainTransactionInput,
	outputs []*externalapi.DomainTransactionOutput, subnetworkID *externalapi.DomainSubnetworkID,
	gas uint64, payload []byte) *externalapi.DomainTransaction {

	payloadHash := hashes.HashData(payload)
	return &externalapi.DomainTransaction{
		Version:      version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     0,
		SubnetworkID: *subnetworkID,
		Gas:          gas,
		PayloadHash:  *payloadHash,
		Payload:      payload,
		Fee:          0,
		Mass:         0,
	}
}
