package appmessage

import (
	"math/big"
)

// MsgBlockBlueWork represents a kaspa BlockBlueWork message
type MsgBlockBlueWork struct {
	baseMessage
	BlueWork *big.Int
}

// Command returns the protocol command string for the message
func (msg *MsgBlockBlueWork) Command() MessageCommand {
	return CmdBlockBlueWork
}

// NewBlockBlueWork returns a new kaspa BlockBlueWork message
func NewBlockBlueWork(blueWork *big.Int) *MsgBlockBlueWork {
	return &MsgBlockBlueWork{
		BlueWork: blueWork,
	}
}
