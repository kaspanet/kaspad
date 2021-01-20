package appmessage

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// MsgPruningPointUTXOSetChunk represents a kaspa PruningPointUTXOSetChunk message
type MsgPruningPointUTXOSetChunk struct {
	baseMessage
	OutpointAndUTXOEntryPairs []*OutpointAndUTXOEntryPair
}

// Command returns the protocol command string for the message
func (msg *MsgPruningPointUTXOSetChunk) Command() MessageCommand {
	return CmdPruningPointUTXOSetChunk
}

// NewMsgPruningPointUTXOSetChunk returns a new MsgPruningPointUTXOSetChunk.
func NewMsgPruningPointUTXOSetChunk(outpointAndUTXOEntryPairs []*OutpointAndUTXOEntryPair) *MsgPruningPointUTXOSetChunk {
	return &MsgPruningPointUTXOSetChunk{
		OutpointAndUTXOEntryPairs: outpointAndUTXOEntryPairs,
	}
}

// OutpointAndUTXOEntryPair is an outpoint along with its
// respective UTXO entry
type OutpointAndUTXOEntryPair struct {
	Outpoint  *Outpoint
	UTXOEntry *UTXOEntry
}

// UTXOEntry houses details about an individual transaction output in a UTXO
type UTXOEntry struct {
	Amount          uint64
	ScriptPublicKey *externalapi.ScriptPublicKey
	BlockBlueScore  uint64
	IsCoinbase      bool
}
