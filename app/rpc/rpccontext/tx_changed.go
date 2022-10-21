package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/txindex"
)

// TXsConfirmationChanged represents information for the  TXsConfirmationChanged listener.
// This type is meant to be used in TXsChanged notifications
type TXsConfirmationChangedNotificationState struct {
	RequiredConfirmations uint32
	IncludePending bool
	RegisteredTxsToBlueScore txindex.TxIDsToBlueScores
	UnregesiteredTxsBlueScore map[externalapi.DomainTransactionID]uint64 //this is bluescore when txid was either a) inserted into listener, or b) removed from listener
}

func (ctx *Context) NewTXsConfirmationChangedNotificationState(txIds []*externalapi.DomainTransactionID, requiredConfirmations uint32, 
	includePending bool) (*TXsConfirmationChangedNotificationState, error) {
	registeredTxsToBlueScore, NotFound, err := ctx.TXIndex.GetTXsBlueScores(txIds)
	if err != nil {
		return nil, err
	}
	virtualInfo, err := ctx.Domain.Consensus().GetVirtualInfo()
	if err != nil {
		return nil, err
	}

	unregesiteredTxsBlueScore := make(txindex.TxIDsToBlueScores, len(NotFound))
	for _, txID := range NotFound {
		unregesiteredTxsBlueScore[*txID] = virtualInfo.BlueScore
	}
	return &TXsConfirmationChangedNotificationState{
		RequiredConfirmations: requiredConfirmations,
		IncludePending: includePending,
		RegisteredTxsToBlueScore: registeredTxsToBlueScore,
		UnregesiteredTxsBlueScore: unregesiteredTxsBlueScore,
	}, nil
}

func (tcc *TXsConfirmationChangedNotificationState) updateStateAndExtractConfirmations(txAcceptanceChange *txindex.TXAcceptanceChange) (
	pending []*appmessage.TxIDConfirmationsPair, confirmed []*appmessage.TxIDConfirmationsPair, unconfirmed []*appmessage.TxIDConfirmationsPair) {

	pending = make([]*appmessage.TxIDConfirmationsPair, 0)
	confirmed = make([]*appmessage.TxIDConfirmationsPair, 0)
	unconfirmed = make([]*appmessage.TxIDConfirmationsPair, 0)

	for txID := range txAcceptanceChange.Removed {
		_, found := tcc.RegisteredTxsToBlueScore[txID]
		if found  {
			delete(tcc.RegisteredTxsToBlueScore, txID)
			tcc.UnregesiteredTxsBlueScore[txID] = txAcceptanceChange.VirtualBlueScore
		}
	}
	for txID := range txAcceptanceChange.Added {
		_, found := tcc.UnregesiteredTxsBlueScore[txID]
		if !found  {
			delete(tcc.UnregesiteredTxsBlueScore, txID)
			tcc.RegisteredTxsToBlueScore[txID] = txAcceptanceChange.VirtualBlueScore
		} 
	}

	for txID, txBluescore := range tcc.RegisteredTxsToBlueScore {
		confirmations := uint32(txAcceptanceChange.VirtualBlueScore - txBluescore)
		if confirmations >= tcc.RequiredConfirmations {
			confirmed = append(confirmed, &appmessage.TxIDConfirmationsPair{TxID: txID.String(), Confirmations: int64(confirmations)})
		} else if tcc.IncludePending {
			pending = append(pending, &appmessage.TxIDConfirmationsPair{TxID: txID.String(), Confirmations: int64(confirmations)})
		}
	}

	for txID, txBluescore := range tcc.UnregesiteredTxsBlueScore {
		unconfirmations := uint32(txAcceptanceChange.VirtualBlueScore - txBluescore)
		if unconfirmations >= tcc.RequiredConfirmations {
			unconfirmed = append(unconfirmed, &appmessage.TxIDConfirmationsPair{TxID: txID.String(), Confirmations: int64(unconfirmations)})
			delete(tcc.UnregesiteredTxsBlueScore, txID)
		}
	}

	if tcc.IncludePending {
		return pending, confirmed, unconfirmed
	}

	return nil, confirmed, unconfirmed
}

// TXsConfirmationChanged represents information for the  TXsConfirmationChanged listener.
// This type is meant to be used in TXsChanged notifications
type AddressesTxsNotificationState struct {
	RequiredConfirmations uint32
	IncludePending bool
	IncludeSpending bool
	IncludeReciving bool
	RegisteredScriptPublicKeysToTxIdsToBlueScores map[string]txindex.TxIDsToBlueScores
	UnregesiteredcriptPublicKeysToTxIdsToBlueScores map[string]txindex.TxIDsToBlueScores //this is bluescore when txid was either a) inserted into listener, or b) removed from listener
}

func (ctx *Context) AddressesTxsNotificationState(txIds []*externalapi.DomainTransactionID, requiredConfirmations uint32, 
	includePending bool) (*TXsConfirmationChangedNotificationState, error) {
	registeredTxsToBlueScore, NotFound, err := ctx.TXIndex.GetTXsBlueScores(txIds)
	if err != nil {
		return nil, err
	}
	virtualInfo, err := ctx.Domain.Consensus().GetVirtualInfo()
	if err != nil {
		return nil, err
	}

	unregesiteredTxsBlueScore := make(txindex.TxIDsToBlueScores, len(NotFound))
	for _, txID := range NotFound {
		unregesiteredTxsBlueScore[*txID] = virtualInfo.BlueScore
	}
	return &TXsConfirmationChangedNotificationState{
		RequiredConfirmations: requiredConfirmations,
		IncludePending: includePending,
		RegisteredTxsToBlueScore: registeredTxsToBlueScore,
		UnregesiteredTxsBlueScore: unregesiteredTxsBlueScore,
	}, nil
}
