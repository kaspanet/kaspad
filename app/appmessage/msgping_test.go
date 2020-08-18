// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"testing"

	"github.com/kaspanet/kaspad/util/random"
)

// TestPing tests the MsgPing API against the latest protocol version.
func TestPing(t *testing.T) {
	// Ensure we get the same nonce back out.
	nonce, err := random.Uint64()
	if err != nil {
		t.Errorf("random.Uint64: Error generating nonce: %v", err)
	}
	msg := NewMsgPing(nonce)
	if msg.Nonce != nonce {
		t.Errorf("NewMsgPing: wrong nonce - got %v, want %v",
			msg.Nonce, nonce)
	}

	// Ensure the command is expected value.
	wantCmd := MessageCommand(7)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgPing: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
