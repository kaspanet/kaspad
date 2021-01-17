package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"math"
)

func (x *KaspadMessage_IbdRootUtxoSetChunk) toAppMessage() (appmessage.Message, error) {
	outpointAndUTXOEntryPairs := make([]*appmessage.OutpointAndUTXOEntryPair, len(x.IbdRootUtxoSetChunk.OutpointAndUtxoEntryPairs))
	for i, outpointAndUTXOEntryPair := range x.IbdRootUtxoSetChunk.OutpointAndUtxoEntryPairs {
		transactionId, err := outpointAndUTXOEntryPair.Outpoint.TransactionId.toDomain()
		if err != nil {
			return nil, err
		}
		outpoint := &appmessage.Outpoint{
			TxID:  *transactionId,
			Index: outpointAndUTXOEntryPair.Outpoint.Index,
		}
		if outpointAndUTXOEntryPair.UtxoEntry.ScriptPublicKey.Version > math.MaxUint16 {
			return nil, errors.Errorf("ScriptPublicKey version is bigger then uint16.")
		}
		scriptPublicKey := &externalapi.ScriptPublicKey{
			Script:  outpointAndUTXOEntryPair.UtxoEntry.ScriptPublicKey.Script,
			Version: uint16(outpointAndUTXOEntryPair.UtxoEntry.ScriptPublicKey.Version),
		}
		utxoEntry := &appmessage.UTXOEntry{
			Amount:          outpointAndUTXOEntryPair.UtxoEntry.Amount,
			ScriptPublicKey: scriptPublicKey,
			BlockBlueScore:  outpointAndUTXOEntryPair.UtxoEntry.BlockBlueScore,
			IsCoinbase:      outpointAndUTXOEntryPair.UtxoEntry.IsCoinbase,
		}
		outpointAndUTXOEntryPairs[i] = &appmessage.OutpointAndUTXOEntryPair{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
		}
	}
	return &appmessage.MsgIBDRootUTXOSetChunk{
		OutpointAndUTXOEntryPairs: outpointAndUTXOEntryPairs,
	}, nil
}

func (x *KaspadMessage_IbdRootUtxoSetChunk) fromAppMessage(message *appmessage.MsgIBDRootUTXOSetChunk) error {
	outpointAndUTXOEntryPairs := make([]*OutpointAndUtxoEntryPair, len(message.OutpointAndUTXOEntryPairs))
	for i, outpointAndUTXOEntryPair := range message.OutpointAndUTXOEntryPairs {
		transactionId := domainTransactionIDToProto(&outpointAndUTXOEntryPair.Outpoint.TxID)
		outpoint := &Outpoint{
			TransactionId: transactionId,
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
	x.IbdRootUtxoSetChunk = &IbdRootUtxoSetChunkMessage{
		OutpointAndUtxoEntryPairs: outpointAndUTXOEntryPairs,
	}
	return nil
}
