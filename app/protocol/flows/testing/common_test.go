package testing

import (
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/pkg/errors"
	"strings"
	"testing"
)

func checkFlowError(t *testing.T, err error, isProtocolError bool, shouldBan bool, contains string) {
	pErr := &protocolerrors.ProtocolError{}
	if errors.As(err, &pErr) != isProtocolError {
		t.Fatalf("Unexepcted error %+v", err)
	}

	if pErr.ShouldBan != shouldBan {
		t.Fatalf("Exepcted shouldBan %t but got %t", shouldBan, pErr.ShouldBan)
	}

	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("Unexpected error. Expected error to contain '%s' but got: %+v", contains, err)
	}
}
