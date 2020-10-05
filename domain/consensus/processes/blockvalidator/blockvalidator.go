package blockvalidator

import "github.com/kaspanet/kaspad/app/appmessage"

// BlockValidator ...
type BlockValidator struct {
}

// New instantiates a new BlockValidator
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

// ValidateAgainstPastUTXO ...
func (bv *BlockValidator) ValidateAgainstPastUTXO(block *appmessage.MsgBlock) error {
	return nil
}

// ValidateFinality ...
func (bv *BlockValidator) ValidateFinality(block *appmessage.MsgBlock) error {
	return nil
}
