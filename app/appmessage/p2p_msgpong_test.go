// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"testing"
)

// TestPongLatest tests the MsgPong API against the latest protocol version.
func TestPongLatest(t *testing.T) {
	nonce := uint64(0x1a05b581a5182c)
	msg := NewMsgPong(nonce)
	if msg.Nonce != nonce {
		t.Errorf("NewMsgPong: wrong nonce - got %v, want %v",
			msg.Nonce, nonce)
	}

	// Ensure the command is expected value.
	wantCmd := MessageCommand(8)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgPong: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
