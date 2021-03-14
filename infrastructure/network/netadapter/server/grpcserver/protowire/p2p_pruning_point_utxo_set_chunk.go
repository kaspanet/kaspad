package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_PruningPointUtxoSetChunk) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_PruningPointUtxoSetChunk is nil")
	}
	outpointAndUTXOEntryPairs := make([]*appmessage.OutpointAndUTXOEntryPair, len(x.PruningPointUtxoSetChunk.OutpointAndUtxoEntryPairs))
	for i, outpointAndUTXOEntryPair := range x.PruningPointUtxoSetChunk.OutpointAndUtxoEntryPairs {
		outpointEntryPairAppMessage, err := outpointAndUTXOEntryPair.toAppMessage()
		if err != nil {
			return nil, err
		}
		outpointAndUTXOEntryPairs[i] = outpointEntryPairAppMessage
	}
	return &appmessage.MsgPruningPointUTXOSetChunk{
		OutpointAndUTXOEntryPairs: outpointAndUTXOEntryPairs,
	}, nil
}

func (x *OutpointAndUtxoEntryPair) toAppMessage() (*appmessage.OutpointAndUTXOEntryPair, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "OutpointAndUtxoEntryPair is nil")
	}
	outpoint, err := x.Outpoint.toAppMessage()
	if err != nil {
		return nil, err
	}
	utxoEntry, err := x.UtxoEntry.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.OutpointAndUTXOEntryPair{
		Outpoint:  outpoint,
		UTXOEntry: utxoEntry,
	}, nil
}

func (x *KaspadMessage_PruningPointUtxoSetChunk) fromAppMessage(message *appmessage.MsgPruningPointUTXOSetChunk) error {
	outpointAndUTXOEntryPairs := make([]*OutpointAndUtxoEntryPair, len(message.OutpointAndUTXOEntryPairs))
	for i, outpointAndUTXOEntryPair := range message.OutpointAndUTXOEntryPairs {
		transactionID := domainTransactionIDToProto(&outpointAndUTXOEntryPair.Outpoint.TxID)
		outpoint := &Outpoint{
			TransactionId: transactionID,
			Index:         outpointAndUTXOEntryPair.Outpoint.Index,
		}
		scriptPublicKey := &ScriptPublicKey{
			Script:  outpointAndUTXOEntryPair.UTXOEntry.ScriptPublicKey.Script,
			Version: uint32(outpointAndUTXOEntryPair.UTXOEntry.ScriptPublicKey.Version),
		}
		utxoEntry := &UtxoEntry{
			Amount:          outpointAndUTXOEntryPair.UTXOEntry.Amount,
			ScriptPublicKey: scriptPublicKey,
			BlockBlueScore:  outpointAndUTXOEntryPair.UTXOEntry.BlockBlueScore,
			IsCoinbase:      outpointAndUTXOEntryPair.UTXOEntry.IsCoinbase,
		}
		outpointAndUTXOEntryPairs[i] = &OutpointAndUtxoEntryPair{
			Outpoint:  outpoint,
			UtxoEntry: utxoEntry,
		}
	}
	x.PruningPointUtxoSetChunk = &PruningPointUtxoSetChunkMessage{
		OutpointAndUtxoEntryPairs: outpointAndUTXOEntryPairs,
	}
	return nil
}
