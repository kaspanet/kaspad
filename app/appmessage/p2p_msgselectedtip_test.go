package appmessage

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TestSelectedTip tests the MsgSelectedTip API.
func TestSelectedTip(t *testing.T) {
	// Ensure the command is expected value.
	wantCmd := MessageCommand(11)
	msg := NewMsgSelectedTip(&externalapi.DomainHash{})
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgSelectedTip: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
