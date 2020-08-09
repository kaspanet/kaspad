package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"testing"
)

// TestSelectedTip tests the MsgSelectedTip API.
func TestSelectedTip(t *testing.T) {

	// Ensure the command is expected value.
	wantCmd := MessageCommand(11)
	msg := NewMsgSelectedTip(&daghash.ZeroHash)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgSelectedTip: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
