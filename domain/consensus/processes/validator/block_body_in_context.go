package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (v *validator) ValidateBodyInContext(block *model.DomainBlock) error {
	return v.checkBlockTransactionsFinalized(block)
}

func (v *validator) checkBlockTransactionsFinalized(block *model.DomainBlock) error {
	panic("unimplemented")
}
