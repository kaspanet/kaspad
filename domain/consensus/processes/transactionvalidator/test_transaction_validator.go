package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
)

type testTransactionValidator struct {
	*transactionValidator
}

// NewTestTransactionValidator creates an instance of a TestTransactionValidator
func NewTestTransactionValidator(baseTransactionValidator model.TransactionValidator) testapi.TestTransactionValidator {
	return &testTransactionValidator{transactionValidator: baseTransactionValidator.(*transactionValidator)}
}

func (tbv *testTransactionValidator) SigCache() *txscript.SigCache {
	return tbv.sigCache
}

func (tbv *testTransactionValidator) SetSigCache(sigCache *txscript.SigCache) {
	tbv.sigCache = sigCache
}
