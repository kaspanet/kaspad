package appmessage

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// NotifyTransactionAddedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyTransactionAddedRequestMessage struct {
	baseMessage
	Addresses []string
}

// Command returns the protocol command string for the message
func (msg *NotifyTransactionAddedRequestMessage) Command() MessageCommand {
	return CmdNotifyTransactionAddedRequestMessage
}

// NewNotifyTransactionAddedRequestMessage returns a instance of the message
func NewNotifyTransactionAddedRequestMessage(addresses []string) *NotifyTransactionAddedRequestMessage {
	return &NotifyTransactionAddedRequestMessage{
		Addresses: addresses,
	}
}

// NotifyTransactionAddedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyTransactionAddedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyTransactionAddedResponseMessage) Command() MessageCommand {
	return CmdNotifyTransactionAddedResponseMessage
}

// NewNotifyTransactionAddedResponseMessage returns a instance of the message
func NewNotifyTransactionAddedResponseMessage() *NotifyTransactionAddedResponseMessage {
	return &NotifyTransactionAddedResponseMessage{}
}

// TransactionAddedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type TransactionAddedNotificationMessage struct {
	baseMessage
	Addresses        []string
	BlockHash        string
	UTXOsVerboseData []*UTXOVerboseData
	Transaction      *MsgTx
	Status           uint32
}

// Command returns the protocol command string for the message
func (msg *TransactionAddedNotificationMessage) Command() MessageCommand {
	return CmdTransactionAddedNotificationMessage
}

// NewTransactionAddedNotificationMessage returns a instance of the message
func NewTransactionAddedNotificationMessage(addresses []string, blockHash *externalapi.DomainHash, utxosVerboseData []*UTXOVerboseData, transaction *MsgTx, status externalapi.TransactionStatus) *TransactionAddedNotificationMessage {
	return &TransactionAddedNotificationMessage{
		Addresses:        addresses,
		BlockHash:        blockHash.String(),
		UTXOsVerboseData: utxosVerboseData,
		Transaction:      transaction,
		Status:           uint32(status),
	}
}
