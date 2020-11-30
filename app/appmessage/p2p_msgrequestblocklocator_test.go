package appmessage

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

// TestRequestBlockLocator tests the MsgRequestBlockLocator API.
func TestRequestBlockLocator(t *testing.T) {
	hashStr := "000000000002e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	highHash, err := hashes.FromString(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// Ensure the command is expected value.
	wantCmd := MessageCommand(9)
	msg := NewMsgRequestBlockLocator(highHash, &externalapi.DomainHash{}, 0)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgRequestBlockLocator: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
