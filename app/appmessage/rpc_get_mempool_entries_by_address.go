package appmessage

// GetMempoolEntriesByAddressesRequestMessage  is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntriesByAddressesRequestMessage struct {
	Addresses []string

	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntriesByAddressesRequestMessage) Command() MessageCommand {
	return CmdGetMempoolEntriesByAddressesRequestMessage
}

// NewGetMempoolEntriesByAddressesRequestMessage returns a instance of the message
func NewGetMempoolEntriesByAddressesRequestMessage(addresses []string) *GetMempoolEntriesByAddressesRequestMessage {
	return &GetMempoolEntriesByAddressesRequestMessage{Addresses: addresses}
}

// GetMempoolEntriesByAddressesResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntriesByAddressesResponseMessage struct {
	baseMessage
	SpendingEntries  []*MempoolEntry
	ReceivingEntries []*MempoolEntry

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntriesByAddressesResponseMessage) Command() MessageCommand {
	return CmdGetMempoolEntriesResponseMessage
}

// NewGetMempoolEntriesResponseMessage returns a instance of the message
func NewGetMempoolEntriesByAddressesResponseMessage(spendingEntries []*MempoolEntry, receivingEntries []*MempoolEntry,
) *GetMempoolEntriesByAddressesResponseMessage {

	return &GetMempoolEntriesByAddressesResponseMessage{
		SpendingEntries:  spendingEntries,
		ReceivingEntries: receivingEntries,
	}
}
