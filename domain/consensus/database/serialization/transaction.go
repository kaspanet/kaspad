package serialization

import (
	"math"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// DomainTransactionToDbTransaction converts DomainTransaction to DbTransaction
func DomainTransactionToDbTransaction(domainTransaction *externalapi.DomainTransaction) *DbTransaction {
	dbInputs := make([]*DbTransactionInput, len(domainTransaction.Inputs))
	for i, domainTransactionInput := range domainTransaction.Inputs {
		dbInputs[i] = &DbTransactionInput{
			PreviousOutpoint: DomainOutpointToDbOutpoint(&domainTransactionInput.PreviousOutpoint),
			SignatureScript:  domainTransactionInput.SignatureScript,
			Sequence:         domainTransactionInput.Sequence,
			SigOpCount:       uint32(domainTransactionInput.SigOpCount),
		}
	}

	dbOutputs := make([]*DbTransactionOutput, len(domainTransaction.Outputs))
	for i, domainTransactionOutput := range domainTransaction.Outputs {
		dbScriptPublicKey := ScriptPublicKeyToDBScriptPublicKey(domainTransactionOutput.ScriptPublicKey)
		dbOutputs[i] = &DbTransactionOutput{
			Value:           domainTransactionOutput.Value,
			ScriptPublicKey: dbScriptPublicKey,
		}
	}

	return &DbTransaction{
		Version:      uint32(domainTransaction.Version),
		Inputs:       dbInputs,
		Outputs:      dbOutputs,
		LockTime:     domainTransaction.LockTime,
		SubnetworkID: DomainSubnetworkIDToDbSubnetworkID(&domainTransaction.SubnetworkID),
		Gas:          domainTransaction.Gas,
		Payload:      domainTransaction.Payload,
	}
}

// DbTransactionToDomainTransaction converts DbTransaction to DomainTransaction
func DbTransactionToDomainTransaction(dbTransaction *DbTransaction) (*externalapi.DomainTransaction, error) {
	domainSubnetworkID, err := DbSubnetworkIDToDomainSubnetworkID(dbTransaction.SubnetworkID)
	if err != nil {
		return nil, err
	}

	domainInputs := make([]*externalapi.DomainTransactionInput, len(dbTransaction.Inputs))
	for i, dbTransactionInput := range dbTransaction.Inputs {
		domainPreviousOutpoint, err := DbOutpointToDomainOutpoint(dbTransactionInput.PreviousOutpoint)
		if err != nil {
			return nil, err
		}
		domainInputs[i] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: *domainPreviousOutpoint,
			SignatureScript:  dbTransactionInput.SignatureScript,
			Sequence:         dbTransactionInput.Sequence,
			SigOpCount:       byte(dbTransactionInput.SigOpCount),
		}
	}

	domainOutputs := make([]*externalapi.DomainTransactionOutput, len(dbTransaction.Outputs))
	for i, dbTransactionOutput := range dbTransaction.Outputs {
		scriptPublicKey, err := DBScriptPublicKeyToScriptPublicKey(dbTransactionOutput.ScriptPublicKey)
		if err != nil {
			return nil, err
		}
		domainOutputs[i] = &externalapi.DomainTransactionOutput{
			Value:           dbTransactionOutput.Value,
			ScriptPublicKey: scriptPublicKey,
		}
	}

	if dbTransaction.Version > math.MaxUint16 {
		return nil, errors.Errorf("The transaction version is bigger then uint16.")
	}
	return &externalapi.DomainTransaction{
		Version:      uint16(dbTransaction.Version),
		Inputs:       domainInputs,
		Outputs:      domainOutputs,
		LockTime:     dbTransaction.LockTime,
		SubnetworkID: *domainSubnetworkID,
		Gas:          dbTransaction.Gas,
		Payload:      dbTransaction.Payload,
	}, nil
}
