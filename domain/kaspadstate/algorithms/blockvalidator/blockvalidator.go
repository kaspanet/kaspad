package blockvalidator

import "github.com/kaspanet/kaspad/app/appmessage"

// BlockValidator ...
type BlockValidator struct {
}

// New ...
func New() *BlockValidator {
	return &BlockValidator{}
}

// ValidateHeaderInIsolation ...
func (bv *BlockValidator) ValidateHeaderInIsolation(block *appmessage.MsgBlock) error {
	return nil
}

// ValidateHeaderInContext ...
func (bv *BlockValidator) ValidateHeaderInContext(block *appmessage.MsgBlock) error {
	return nil
}

// ValidateBodyInIsolation ...
func (bv *BlockValidator) ValidateBodyInIsolation(block *appmessage.MsgBlock) error {
	return nil
}

// ValidateBodyInContext ...
func (bv *BlockValidator) ValidateBodyInContext(block *appmessage.MsgBlock) error {
	return nil
}
