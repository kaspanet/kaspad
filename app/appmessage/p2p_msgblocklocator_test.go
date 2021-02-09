package appmessage

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/davecgh/go-spew/spew"
)

// TestBlockLocator tests the MsgBlockLocator API.
func TestBlockLocator(t *testing.T) {
	hashStr := "000000000002e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	locatorHash, err := externalapi.NewDomainHashFromString(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	msg := NewMsgBlockLocator([]*externalapi.DomainHash{locatorHash})

	// Ensure the command is expected value.
	wantCmd := MessageCommand(10)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgBlockLocator: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure block locator hashes are added properly.
	if msg.BlockLocatorHashes[0] != locatorHash {
		t.Errorf("AddBlockLocatorHash: wrong block locator added - "+
			"got %v, want %v",
			spew.Sprint(msg.BlockLocatorHashes[0]),
			spew.Sprint(locatorHash))
	}
}
