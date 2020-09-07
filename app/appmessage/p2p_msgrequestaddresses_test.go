// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"testing"
)

// TestRequestAddresses tests the MsgRequestAddresses API.
func TestRequestAddresses(t *testing.T) {
	// Ensure the command is expected value.
	wantCmd := MessageCommand(2)
	msg := NewMsgRequestAddresses(false, nil)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgRequestAddresses: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
