package appmessage

type MempoolEntryByAddress struct {
	Address		string
	Receiving	[]*MempoolEntry
	Sending		[]*MempoolEntry
}


// GetMempoolEntriesByAddressesRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntriesByAddressesRequestMessage struct {
	baseMessage
	Addresses 	[]string
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntriesByAddressesRequestMessage) Command() MessageCommand {
	return CmdGetMempoolEntriesByAddressesRequestMessage
}

// NewGetMempoolEntriesByAddressesRequestMessage returns a instance of the message
func NewGetMempoolEntriesByAddressesRequestMessage(addresses []string) *GetMempoolEntriesByAddressesRequestMessage {
	return &GetMempoolEntriesByAddressesRequestMessage{
		Addresses: addresses,
	}
}

// GetMempoolEntriesResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntriesByAddressesResponseMessage struct {
	baseMessage
	Entries []*MempoolEntryByAddress

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntriesByAddressesResponseMessage) Command() MessageCommand {
	return CmdGetMempoolEntriesByAddressesResponseMessage
}

// NewGetMempoolEntriesByAddressesResponseMessage returns a instance of the message
func NewGetMempoolEntriesByAddressesResponseMessage(entries []*MempoolEntryByAddress) *GetMempoolEntriesByAddressesResponseMessage {
	return &GetMempoolEntriesByAddressesResponseMessage{
		Entries: entries,
	}
}
