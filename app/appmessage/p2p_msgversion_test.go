// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"net"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"
)

// TestVersion tests the MsgVersion API.
func TestVersion(t *testing.T) {
	pver := ProtocolVersion

	// Create version message data.
	selectedTipHash := &externalapi.DomainHash{12, 34}
	tcpAddrMe := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 16111}
	me := NewNetAddress(tcpAddrMe, SFNodeNetwork)
	generatedID, err := id.GenerateID()
	if err != nil {
		t.Fatalf("id.GenerateID: %s", err)
	}

	// Ensure we get the correct data back out.
	msg := NewMsgVersion(me, generatedID, "mainnet", selectedTipHash, nil)
	if msg.ProtocolVersion != pver {
		t.Errorf("NewMsgVersion: wrong protocol version - got %v, want %v",
			msg.ProtocolVersion, pver)
	}
	if !reflect.DeepEqual(msg.Address, me) {
		t.Errorf("NewMsgVersion: wrong me address - got %v, want %v",
			spew.Sdump(&msg.Address), spew.Sdump(me))
	}
	if msg.ID.String() != generatedID.String() {
		t.Errorf("NewMsgVersion: wrong nonce - got %s, want %s",
			msg.ID, generatedID)
	}
	if msg.UserAgent != DefaultUserAgent {
		t.Errorf("NewMsgVersion: wrong user agent - got %v, want %v",
			msg.UserAgent, DefaultUserAgent)
	}
	if !msg.SelectedTipHash.Equal(selectedTipHash) {
		t.Errorf("NewMsgVersion: wrong selected tip hash - got %s, want %s",
			msg.SelectedTipHash, selectedTipHash)
	}
	if msg.DisableRelayTx {
		t.Errorf("NewMsgVersion: disable relay tx is not false by "+
			"default - got %v, want %v", msg.DisableRelayTx, false)
	}

	msg.AddUserAgent("myclient", "1.2.3", "optional", "comments")
	customUserAgent := DefaultUserAgent + "myclient:1.2.3(optional; comments)/"
	if msg.UserAgent != customUserAgent {
		t.Errorf("AddUserAgent: wrong user agent - got %s, want %s",
			msg.UserAgent, customUserAgent)
	}

	msg.AddUserAgent("mygui", "3.4.5")
	customUserAgent += "mygui:3.4.5/"
	if msg.UserAgent != customUserAgent {
		t.Errorf("AddUserAgent: wrong user agent - got %s, want %s",
			msg.UserAgent, customUserAgent)
	}

	// Version message should not have any services set by default.
	if msg.Services != 0 {
		t.Errorf("NewMsgVersion: wrong default services - got %v, want %v",
			msg.Services, 0)

	}
	if msg.HasService(SFNodeNetwork) {
		t.Errorf("HasService: SFNodeNetwork service is set")
	}

	// Ensure the command is expected value.
	wantCmd := MessageCommand(0)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgVersion: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure adding the full service node flag works.
	msg.AddService(SFNodeNetwork)
	if msg.Services != SFNodeNetwork {
		t.Errorf("AddService: wrong services - got %v, want %v",
			msg.Services, SFNodeNetwork)
	}
	if !msg.HasService(SFNodeNetwork) {
		t.Errorf("HasService: SFNodeNetwork service not set")
	}
}
