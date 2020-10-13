package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (bv *BlockValidator) ValidateBodyInContext(block *model.DomainBlock) error {
	return bv.checkBlockTransactionsFinalized(block)
}

func (bv *BlockValidator) checkBlockTransactionsFinalized(block *model.DomainBlock) error {
	panic("unimplemented")
}
