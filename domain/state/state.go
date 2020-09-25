package state

import "github.com/kaspanet/kaspad/app/appmessage"

type State interface {
	ValidateAndInsertBlock(block *appmessage.MsgBlock) error
}

type state struct {
}

func (s *state) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return nil
}
