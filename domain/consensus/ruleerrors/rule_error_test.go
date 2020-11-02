package ruleerrors

import (
	"errors"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func TestNewErrMissingTxOut(t *testing.T) {
	outer := NewErrMissingTxOut([]externalapi.DomainOutpoint{{TransactionID: externalapi.DomainTransactionID{255, 255, 255}, Index: 5}})
	expectedOuterErr := "ErrMissingTxOut: [(0000000000000000000000000000000000000000000000000000000000ffffff: 5)]"
	inner := &ErrMissingTxOut{}
	if !errors.As(outer, inner) {
		t.Fatal("TestWrapInRuleError: Outer should contain ErrMissingTxOut in it")
	}

	if len(inner.MissingOutpoints) != 1 {
		t.Fatalf("TestWrapInRuleError: Expected len(inner.MissingOutpoints) 1, found: %d", len(inner.MissingOutpoints))
	}
	if inner.MissingOutpoints[0].Index != 5 {
		t.Fatalf("TestWrapInRuleError: Expected 5. found: %d", inner.MissingOutpoints[0].Index)
	}

	rule := &RuleError{}
	if !errors.As(outer, rule) {
		t.Fatal("TestWrapInRuleError: Outer should contain RuleError in it")
	}
	if rule.message != "ErrMissingTxOut" {
		t.Fatalf("TestWrapInRuleError: Expected message = 'ErrMissingTxOut', found: '%s'", rule.message)
	}
	if errors.Is(rule.inner, inner) {
		t.Fatal("TestWrapInRuleError: rule.inner should contain the ErrMissingTxOut in it")
	}

	if outer.Error() != expectedOuterErr {
		t.Fatalf("TestWrapInRuleError: Expected %s. found: %s", expectedOuterErr, outer.Error())
	}
}

func TestNewErrInvalidTransactionsInNewBlock(t *testing.T) {
	outer := NewErrInvalidTransactionsInNewBlock([]InvalidTransaction{{&externalapi.DomainTransaction{Fee: 1337}, ErrNoTxInputs}})
	//TODO: Implement Stringer for `DomainTransaction`
	expectedOuterErr := "ErrInvalidTransactionsInNewBlock: [(3a464e1e43410c7add1dd81c3f10486f41eb473bb43e8d64feca3c7f0c8028d3: ErrNoTxInputs)]"
	inner := &ErrInvalidTransactionsInNewBlock{}
	if !errors.As(outer, inner) {
		t.Fatal("TestNewErrInvalidTransactionsInNewBlock: Outer should contain ErrInvalidTransactionsInNewBlock in it")
	}

	if len(inner.InvalidTransactions) != 1 {
		t.Fatalf("TestNewErrInvalidTransactionsInNewBlock: Expected len(inner.MissingOutpoints) 1, found: %d", len(inner.InvalidTransactions))
	}
	if inner.InvalidTransactions[0].err != ErrNoTxInputs {
		t.Fatalf("TestNewErrInvalidTransactionsInNewBlock: Expected ErrNoTxInputs. found: %v", inner.InvalidTransactions[0].err)
	}
	if inner.InvalidTransactions[0].Transaction.Fee != 1337 {
		t.Fatalf("TestNewErrInvalidTransactionsInNewBlock: Expected 1337. found: %v", inner.InvalidTransactions[0].Transaction.Fee)
	}

	rule := &RuleError{}
	if !errors.As(outer, rule) {
		t.Fatal("TestNewErrInvalidTransactionsInNewBlock: Outer should contain RuleError in it")
	}
	if rule.message != "ErrInvalidTransactionsInNewBlock" {
		t.Fatalf("TestNewErrInvalidTransactionsInNewBlock: Expected message = 'ErrInvalidTransactionsInNewBlock', found: '%s'", rule.message)
	}
	if errors.Is(rule.inner, inner) {
		t.Fatal("TestNewErrInvalidTransactionsInNewBlock: rule.inner should contain the ErrInvalidTransactionsInNewBlock in it")
	}

	if outer.Error() != expectedOuterErr {
		t.Fatalf("TestNewErrInvalidTransactionsInNewBlock: Expected %s. found: %s", expectedOuterErr, outer.Error())
	}
}
