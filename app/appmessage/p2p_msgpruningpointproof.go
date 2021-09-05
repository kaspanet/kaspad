package appmessage

import "math/big"

// MsgPruningPointProof represents a kaspa PruningPointProof message
type MsgPruningPointProof struct {
	baseMessage

	Headers              []*MsgBlockHeader
	PruningPointBlueWork *big.Int
}

// Command returns the protocol command string for the message
func (msg *MsgPruningPointProof) Command() MessageCommand {
	return CmdPruningPointProof
}

// NewMsgPruningPointProof returns a new MsgPruningPointProof.
func NewMsgPruningPointProof(headers []*MsgBlockHeader, pruningPointBlueWork *big.Int) *MsgPruningPointProof {
	return &MsgPruningPointProof{
		Headers:              headers,
		PruningPointBlueWork: pruningPointBlueWork,
	}
}
