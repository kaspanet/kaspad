package appmessage

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// MsgIBDRootUTXOSetChunk represents a kaspa IBDRootUTXOSetChunk message
type MsgIBDRootUTXOSetChunk struct {
	baseMessage
	OutpointAndUTXOEntryPairs []*OutpointAndUTXOEntryPair
}

// Command returns the protocol command string for the message
func (msg *MsgIBDRootUTXOSetChunk) Command() MessageCommand {
	return CmdIBDRootUTXOSetChunk
}

// NewMsgIBDRootUTXOSetChunk returns a new MsgIBDRootUTXOSetChunk.
func NewMsgIBDRootUTXOSetChunk(outpointAndUTXOEntryPairs []*OutpointAndUTXOEntryPair) *MsgIBDRootUTXOSetChunk {
	return &MsgIBDRootUTXOSetChunk{
		OutpointAndUTXOEntryPairs: outpointAndUTXOEntryPairs,
	}
}

// OutpointAndUTXOEntryPair is an outpoint along with its
// respective UTXO entry
type OutpointAndUTXOEntryPair struct {
	Outpoint  *Outpoint
	UTXOEntry *UTXOEntry
}

type UTXOEntry struct {
	Amount          uint64
	ScriptPublicKey *externalapi.ScriptPublicKey
	BlockBlueScore  uint64
	IsCoinbase      bool
}
