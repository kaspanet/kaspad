// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TestRequstIBDBlocks tests the MsgRequestHeaders API.
func TestRequstIBDBlocks(t *testing.T) {
	hashStr := "000000000002e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	lowHash, err := externalapi.NewDomainHashFromString(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	hashStr = "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"
	highHash, err := externalapi.NewDomainHashFromString(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// Ensure we get the same data back out.
	msg := NewMsgRequstHeaders(lowHash, highHash)
	if !msg.HighHash.Equal(highHash) {
		t.Errorf("NewMsgRequstHeaders: wrong high hash - got %v, want %v",
			msg.HighHash, highHash)
	}

	// Ensure the command is expected value.
	wantCmd := MessageCommand(4)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgRequstHeaders: wrong command - got %v want %v",
			cmd, wantCmd)
	}
}
