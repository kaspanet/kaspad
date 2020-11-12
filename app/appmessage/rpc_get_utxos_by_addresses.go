package appmessage

// GetUTXOsByAddressRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetUTXOsByAddressRequestMessage struct {
	baseMessage
	Address string
}

// Command returns the protocol command string for the message
func (msg *GetUTXOsByAddressRequestMessage) Command() MessageCommand {
	return CmdGetUTXOsByAddressRequestMessage
}

// NewGetUTXOsByAddressRequestMessage returns a instance of the message
func NewGetUTXOsByAddressRequestMessage(address string) *GetUTXOsByAddressRequestMessage {
	return &GetUTXOsByAddressRequestMessage{
		Address: address,
	}
}

// BlockVerboseData holds verbose data about a UTXO
type UTXOVerboseData struct {
	Amount         uint64
	ScriptPubKey   []byte
	BlockBlueScore uint64
	IsCoinbase     bool
}

// GetUTXOsByAddressResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetUTXOsByAddressResponseMessage struct {
	baseMessage
	Address          string
	UTXOsVerboseData []*UTXOVerboseData

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetUTXOsByAddressResponseMessage) Command() MessageCommand {
	return CmdGetUTXOsByAddressResponseMessage
}

// NewGetUTXOsByAddressResponseMessage returns a instance of the message
func NewGetUTXOsByAddressResponseMessage(address string, utxosVerboseData []*UTXOVerboseData) *GetUTXOsByAddressResponseMessage {
	return &GetUTXOsByAddressResponseMessage{
		Address:          address,
		UTXOsVerboseData: utxosVerboseData,
	}
}
