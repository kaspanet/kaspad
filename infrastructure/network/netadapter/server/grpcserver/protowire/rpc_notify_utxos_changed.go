package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_NotifyUtxosChangedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyUTXOsChangedRequestMessage{
		Addresses: x.NotifyUtxosChangedRequest.Addresses,
	}, nil
}

func (x *KaspadMessage_NotifyUtxosChangedRequest) fromAppMessage(message *appmessage.NotifyUTXOsChangedRequestMessage) error {
	x.NotifyUtxosChangedRequest = &NotifyUTXOsChangedRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *KaspadMessage_NotifyUtxosChangedResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyUtxosChangedResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyUtxosChangedResponse.Error.Message}
	}
	return &appmessage.NotifyBlockAddedResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_NotifyUtxosChangedResponse) fromAppMessage(message *appmessage.NotifyUTXOsChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyUtxosChangedResponse = &NotifyUTXOsChangedResponseMessage{
		Error: err,
	}
	return nil
}

func (x *KaspadMessage_UtxosChangedNotification) toAppMessage() (appmessage.Message, error) {
	added := make([]*appmessage.UTXOsByAddressesEntry, len(x.UtxosChangedNotification.Added))
	for i, entry := range x.UtxosChangedNotification.Added {
		outpoint := &appmessage.RPCOutpoint{
			TransactionID: entry.Outpoint.TransactionId,
			Index:         entry.Outpoint.Index,
		}
		utxoEntry := &appmessage.RPCUTXOEntry{
			Amount:         entry.UtxoEntry.Amount,
			ScriptPubKey:   entry.UtxoEntry.ScriptPubKey,
			BlockBlueScore: entry.UtxoEntry.BlockBlueScore,
			IsCoinbase:     entry.UtxoEntry.IsCoinbase,
		}
		added[i] = &appmessage.UTXOsByAddressesEntry{
			Address:   entry.Address,
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
		}
	}

	removed := make([]*appmessage.RPCOutpoint, len(x.UtxosChangedNotification.Removed))
	for i, outpoint := range x.UtxosChangedNotification.Removed {
		removed[i] = &appmessage.RPCOutpoint{
			TransactionID: outpoint.TransactionId,
			Index:         outpoint.Index,
		}
	}

	return &appmessage.UTXOsChangedNotificationMessage{
		Added:   added,
		Removed: removed,
	}, nil
}

func (x *KaspadMessage_UtxosChangedNotification) fromAppMessage(message *appmessage.UTXOsChangedNotificationMessage) error {
	added := make([]*UTXOsByAddressesEntry, len(message.Added))
	for i, entry := range message.Added {
		outpoint := &RPCOutpoint{
			TransactionId: entry.Outpoint.TransactionID,
			Index:         entry.Outpoint.Index,
		}
		utxoEntry := &RPCUTXOEntry{
			Amount:         entry.UTXOEntry.Amount,
			ScriptPubKey:   entry.UTXOEntry.ScriptPubKey,
			BlockBlueScore: entry.UTXOEntry.BlockBlueScore,
			IsCoinbase:     entry.UTXOEntry.IsCoinbase,
		}
		added[i] = &UTXOsByAddressesEntry{
			Address:   entry.Address,
			Outpoint:  outpoint,
			UtxoEntry: utxoEntry,
		}
	}

	removed := make([]*RPCOutpoint, len(message.Removed))
	for i, outpoint := range message.Removed {
		removed[i] = &RPCOutpoint{
			TransactionId: outpoint.TransactionID,
			Index:         outpoint.Index,
		}
	}

	x.UtxosChangedNotification = &UTXOsChangedNotificationMessage{
		Added:   added,
		Removed: removed,
	}
	return nil
}
