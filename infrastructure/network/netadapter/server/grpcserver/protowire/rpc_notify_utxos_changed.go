package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_NotifyUtxosChangedRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyUtxosChangedRequest is nil")
	}
	return x.NotifyUtxosChangedRequest.toAppMessage()
}

func (x *KaspadMessage_NotifyUtxosChangedRequest) fromAppMessage(message *appmessage.NotifyUTXOsChangedRequestMessage) error {
	x.NotifyUtxosChangedRequest = &NotifyUtxosChangedRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *NotifyUtxosChangedRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyUtxosChangedRequestMessage is nil")
	}
	return &appmessage.NotifyUTXOsChangedRequestMessage{
		Addresses: x.Addresses,
	}, nil
}

func (x *KaspadMessage_NotifyUtxosChangedResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyUtxosChangedResponseMessage is nil")
	}
	return x.NotifyUtxosChangedResponse.toAppMessage()
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

func (x *NotifyUtxosChangedResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyUtxosChangedResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.NotifyUTXOsChangedResponseMessage{
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_UtxosChangedNotification) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_UtxosChangedNotification is nil")
	}
	return x.UtxosChangedNotification.toAppMessage()
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

func (x *UtxosChangedNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "UtxosChangedNotificationMessage is nil")
	}
	added := make([]*appmessage.UTXOsByAddressesEntry, len(x.Added))
	for i, entry := range x.Added {
		entryAsAppMessage, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		// UTXOEntry is optional in other places, but here it's required.
		if entryAsAppMessage.UTXOEntry == nil {
			return nil, errors.Wrapf(errorNil, "UTXOEntry is nil in UTXOsByAddressesEntry.Added")
		}
		added[i] = entryAsAppMessage
	}

	removed := make([]*appmessage.UTXOsByAddressesEntry, len(x.Removed))
	for i, entry := range x.Removed {
		entryAsAppMessage, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		removed[i] = entryAsAppMessage
	}

	return &appmessage.UTXOsChangedNotificationMessage{
		Added:   added,
		Removed: removed,
	}, nil
}

func (x *UtxosByAddressesEntry) toAppMessage() (*appmessage.UTXOsByAddressesEntry, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "UtxosByAddressesEntry is nil")
	}
	outpoint, err := x.Outpoint.toAppMessage()
	if err != nil {
		return nil, err
	}
	entry, err := x.UtxoEntry.toAppMessage()
	// entry is an optional field sometimes
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.UTXOsByAddressesEntry{
		Address:   x.Address,
		Outpoint:  outpoint,
		UTXOEntry: entry,
	}, nil
}

func (x *UtxosByAddressesEntry) fromAppMessage(entry *appmessage.UTXOsByAddressesEntry) {
	outpoint := &RpcOutpoint{
		TransactionId: entry.Outpoint.TransactionID,
		Index:         entry.Outpoint.Index,
	}
	var utxoEntry *RpcUtxoEntry
	if entry.UTXOEntry != nil {
		utxoEntry = &RpcUtxoEntry{
			Amount:          entry.UTXOEntry.Amount,
			ScriptPublicKey: ConvertFromRPCScriptPubKeyToAppMsgRPCScriptPubKey(entry.UTXOEntry.ScriptPublicKey),
			BlockDaaScore:   entry.UTXOEntry.BlockDAAScore,
			IsCoinbase:      entry.UTXOEntry.IsCoinbase,
		}
	}
	*x = UtxosByAddressesEntry{
		Address:   entry.Address,
		Outpoint:  outpoint,
		UtxoEntry: utxoEntry,
	}
}

func (x *RpcOutpoint) toAppMessage() (*appmessage.RPCOutpoint, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcOutpoint is nil")
	}
	return &appmessage.RPCOutpoint{
		TransactionID: x.TransactionId,
		Index:         x.Index,
	}, nil
}

func (x *RpcUtxoEntry) toAppMessage() (*appmessage.RPCUTXOEntry, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcUtxoEntry is nil")
	}
	scriptPubKey, err := x.ScriptPublicKey.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.RPCUTXOEntry{
		Amount:          x.Amount,
		ScriptPublicKey: scriptPubKey,
		BlockDAAScore:   x.BlockDaaScore,
		IsCoinbase:      x.IsCoinbase,
	}, nil
}
