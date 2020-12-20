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
	x.NotifyUtxosChangedRequest = &NotifyUtxosChangedRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *KaspadMessage_NotifyUtxosChangedResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyUtxosChangedResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyUtxosChangedResponse.Error.Message}
	}
	return &appmessage.NotifyUTXOsChangedResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_NotifyUtxosChangedResponse) fromAppMessage(message *appmessage.NotifyUTXOsChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyUtxosChangedResponse = &NotifyUtxosChangedResponseMessage{
		Error: err,
	}
	return nil
}

func (x *KaspadMessage_UtxosChangedNotification) toAppMessage() (appmessage.Message, error) {
	added := make([]*appmessage.UTXOsByAddressesEntry, len(x.UtxosChangedNotification.Added))
	for i, entry := range x.UtxosChangedNotification.Added {
		added[i] = entry.toAppMessage()
	}

	removed := make([]*appmessage.UTXOsByAddressesEntry, len(x.UtxosChangedNotification.Removed))
	for i, entry := range x.UtxosChangedNotification.Removed {
		removed[i] = entry.toAppMessage()
	}

	return &appmessage.UTXOsChangedNotificationMessage{
		Added:   added,
		Removed: removed,
	}, nil
}

func (x *KaspadMessage_UtxosChangedNotification) fromAppMessage(message *appmessage.UTXOsChangedNotificationMessage) error {
	added := make([]*UtxosByAddressesEntry, len(message.Added))
	for i, entry := range message.Added {
		added[i] = &UtxosByAddressesEntry{}
		added[i].fromAppMessage(entry)
	}

	removed := make([]*UtxosByAddressesEntry, len(message.Removed))
	for i, entry := range message.Removed {
		removed[i] = &UtxosByAddressesEntry{}
		removed[i].fromAppMessage(entry)
	}

	x.UtxosChangedNotification = &UtxosChangedNotificationMessage{
		Added:   added,
		Removed: removed,
	}
	return nil
}

func (x *UtxosByAddressesEntry) toAppMessage() *appmessage.UTXOsByAddressesEntry {
	outpoint := &appmessage.RPCOutpoint{
		TransactionID: x.Outpoint.TransactionId,
		Index:         x.Outpoint.Index,
	}
	var utxoEntry *appmessage.RPCUTXOEntry
	if x.UtxoEntry != nil {
		utxoEntry = &appmessage.RPCUTXOEntry{
			Amount:         x.UtxoEntry.Amount,
			ScriptPubKey:   x.UtxoEntry.ScriptPubKey,
			BlockBlueScore: x.UtxoEntry.BlockBlueScore,
			IsCoinbase:     x.UtxoEntry.IsCoinbase,
		}
	}
	return &appmessage.UTXOsByAddressesEntry{
		Address:   x.Address,
		Outpoint:  outpoint,
		UTXOEntry: utxoEntry,
	}
}

func (x *UtxosByAddressesEntry) fromAppMessage(entry *appmessage.UTXOsByAddressesEntry) {
	outpoint := &RpcOutpoint{
		TransactionId: entry.Outpoint.TransactionID,
		Index:         entry.Outpoint.Index,
	}
	var utxoEntry *RpcUtxoEntry
	if entry.UTXOEntry != nil {
		utxoEntry = &RpcUtxoEntry{
			Amount:         entry.UTXOEntry.Amount,
			ScriptPubKey:   entry.UTXOEntry.ScriptPubKey,
			BlockBlueScore: entry.UTXOEntry.BlockBlueScore,
			IsCoinbase:     entry.UTXOEntry.IsCoinbase,
		}
	}
	*x = UtxosByAddressesEntry{
		Address:   entry.Address,
		Outpoint:  outpoint,
		UtxoEntry: utxoEntry,
	}
}
