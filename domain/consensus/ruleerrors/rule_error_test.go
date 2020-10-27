package ruleerrors

import (
	"errors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"testing"
)

func TestNewErrMissingTxOut(t *testing.T) {
	outer := NewErrMissingTxOut([]externalapi.DomainOutpoint{{TransactionID: externalapi.DomainTransactionID{255, 255, 255}, Index: 5}})
	expectedOuterErr := "ErrMissingTxOut: [ffffff0000000000000000000000000000000000000000000000000000000000:5]"
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
