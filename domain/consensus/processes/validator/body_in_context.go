package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (bv *Validator) ValidateBodyInContext(block *model.DomainBlock) error {
	return bv.checkBlockTransactionsFinalized(block)
}

func (bv *Validator) checkBlockTransactionsFinalized(block *model.DomainBlock) error {
	panic("unimplemented")
}
