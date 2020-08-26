// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"testing"
)

// TestVerAck tests the MsgVerAck API.
func TestVerAck(t *testing.T) {
	// Ensure the command is expected value.
	wantCmd := MessageCommand(1)
	msg := NewMsgVerAck()
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgVerAck: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
