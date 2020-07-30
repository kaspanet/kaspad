// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"testing"
)

// TestGetSelectedTip tests the MsgRequestSelectedTip API.
func TestGetSelectedTip(t *testing.T) {
	// Ensure the command is expected value.
	wantCmd := MessageCommand(12)
	msg := NewMsgGetSelectedTip()
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgGetSelectedTip: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
