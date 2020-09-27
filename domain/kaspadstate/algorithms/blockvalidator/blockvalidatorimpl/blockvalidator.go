package blockvalidatorimpl

import "github.com/kaspanet/kaspad/app/appmessage"

type BlockValidator struct {
}

func New() *BlockValidator {
	return &BlockValidator{}
}

func (bv *BlockValidator) ValidateHeaderInIsolation(block *appmessage.MsgBlock) error {
	return nil
}

func (bv *BlockValidator) ValidateHeaderInContext(block *appmessage.MsgBlock) error {
	return nil
}

func (bv *BlockValidator) ValidateBodyInIsolation(block *appmessage.MsgBlock) error {
	return nil
}

func (bv *BlockValidator) ValidateBodyInContext(block *appmessage.MsgBlock) error {
	return nil
}
