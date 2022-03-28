package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"github.com/pkg/errors"
)

func (mp *mempool) fillInputsAndGetMissingParents(transaction *externalapi.DomainTransaction) (
	parents model.IDToTransactionMap, missingOutpoints []*externalapi.DomainOutpoint, err error) {

	parentsInPool := mp.transactionsPool.getParentTransactionsInPool(transaction)

	fillInputs(transaction, parentsInPool)

	err = mp.consensusReference.Consensus().ValidateTransactionAndPopulateWithConsensusData(transaction)
	if err != nil {
		errMissingOutpoints := ruleerrors.ErrMissingTxOut{}
		if errors.As(err, &errMissingOutpoints) {
			return parentsInPool, errMissingOutpoints.MissingOutpoints, nil
		}
		if errors.Is(err, ruleerrors.ErrImmatureSpend) {
			return nil, nil, transactionRuleError(
				RejectImmatureSpend, "one of the transaction inputs spends an immature UTXO")
		}
		if errors.As(err, &ruleerrors.RuleError{}) {
			return nil, nil, newRuleError(err)
		}
		return nil, nil, err
	}

	return parentsInPool, nil, nil
}

func fillInputs(transaction *externalapi.DomainTransaction, parentsInPool model.IDToTransactionMap) {
	for _, input := range transaction.Inputs {
		parent, ok := parentsInPool[input.PreviousOutpoint.TransactionID]
		if !ok {
			continue
		}
		relevantOutput := parent.Transaction().Outputs[input.PreviousOutpoint.Index]
		input.UTXOEntry = utxo.NewUTXOEntry(relevantOutput.Value, relevantOutput.ScriptPublicKey,
			false, constants.UnacceptedDAAScore)
	}
}
