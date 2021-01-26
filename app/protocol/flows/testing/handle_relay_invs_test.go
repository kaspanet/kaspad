package testing

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/blockrelay"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/mstime"
	"testing"
	"time"
)

var orphanBlock = &externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		constants.MaxBlockVersion,
		[]*externalapi.DomainHash{unknownBlockHash},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0,
		0,
		0,
	),
}

var validPruningPointBlock = &externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		constants.MaxBlockVersion,
		[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0,
		0,
		0,
	),
}

var invalidPruningPointBlock = &externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		constants.MaxBlockVersion,
		[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0,
		0,
		0,
	),
}

var unexpectedIBDBlock = &externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		constants.MaxBlockVersion,
		[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0,
		0,
		0,
	),
}

var invalidBlock = &externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		constants.MaxBlockVersion,
		[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4})},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0,
		0,
		0,
	),
}

var unknownBlockHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})
var knownInvalidBlockHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})
var validPruningPointHash = consensushashing.BlockHash(validPruningPointBlock)
var invalidBlockHash = consensushashing.BlockHash(invalidBlock)
var invalidPruningPointHash = consensushashing.BlockHash(invalidPruningPointBlock)
var orphanBlockHash = consensushashing.BlockHash(orphanBlock)

var fakeRelayInvsContextMap = map[externalapi.DomainHash]*externalapi.BlockInfo{
	*knownInvalidBlockHash: {
		Exists:      true,
		BlockStatus: externalapi.StatusInvalid,
	},
	*validPruningPointHash: {
		Exists:      true,
		BlockStatus: externalapi.StatusHeaderOnly,
	},
	*invalidPruningPointHash: {
		Exists:      true,
		BlockStatus: externalapi.StatusHeaderOnly,
	},
}

type fakeRelayInvsContext struct {
	params               *dagconfig.Params
	askedOrphanBlockInfo bool
}

func (f *fakeRelayInvsContext) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) ValidateAndInsertBlock(block *externalapi.DomainBlock) (*externalapi.BlockInsertionResult, error) {
	hash := consensushashing.BlockHash(block)
	if hash.Equal(orphanBlockHash) {
		return nil, ruleerrors.NewErrMissingParents(orphanBlock.Header.ParentHashes())
	}

	if hash.Equal(invalidBlockHash) {
		return nil, ruleerrors.ErrBadMerkleRoot
	}

	return nil, nil
}

func (f *fakeRelayInvsContext) ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) GetBlock(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) GetBlockHeader(blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) GetBlockInfo(blockHash *externalapi.DomainHash) (*externalapi.BlockInfo, error) {
	if info, ok := fakeRelayInvsContextMap[*blockHash]; ok {
		return info, nil
	}

	// Second time we ask for orphan block it's in the end of IBD, to
	// check if the IBD has finished.
	// Since we don't actually process headers, we just say the orphan
	// exists in the second time we're asked about it to indicate IBD
	// has finished.
	if blockHash.Equal(orphanBlockHash) {
		if f.askedOrphanBlockInfo {
			return &externalapi.BlockInfo{Exists: true}, nil
		}
		f.askedOrphanBlockInfo = true
	}

	return &externalapi.BlockInfo{
		Exists: false,
	}, nil
}

func (f *fakeRelayInvsContext) GetBlockAcceptanceData(blockHash *externalapi.DomainHash) (externalapi.AcceptanceData, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) GetHashesBetween(lowHash, highHash *externalapi.DomainHash, maxBlueScoreDifference uint64) ([]*externalapi.DomainHash, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	// This is done so we can test getting invalid block during IBD.
	return []*externalapi.DomainHash{invalidBlockHash}, nil
}

func (f *fakeRelayInvsContext) GetPruningPointUTXOs(expectedPruningPointHash *externalapi.DomainHash, fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) PruningPoint() (*externalapi.DomainHash, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) ClearImportedPruningPointData() error {
	return nil
}

func (f *fakeRelayInvsContext) AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) ValidateAndInsertImportedPruningPoint(newPruningPoint *externalapi.DomainBlock) error {
	if consensushashing.BlockHash(newPruningPoint).Equal(invalidPruningPointHash) {
		return ruleerrors.ErrBadMerkleRoot
	}

	return nil
}

func (f *fakeRelayInvsContext) GetVirtualSelectedParent() (*externalapi.DomainHash, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) CreateHeadersSelectedChainBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {
	return externalapi.BlockLocator{
		f.params.GenesisHash,
	}, nil
}

func (f *fakeRelayInvsContext) CreateFullHeadersSelectedChainBlockLocator() (externalapi.BlockLocator, error) {
	return externalapi.BlockLocator{
		f.params.GenesisHash,
	}, nil
}

func (f *fakeRelayInvsContext) GetSyncInfo() (*externalapi.SyncInfo, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) Tips() ([]*externalapi.DomainHash, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) GetVirtualInfo() (*externalapi.VirtualInfo, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) IsValidPruningPoint(blockHash *externalapi.DomainHash) (bool, error) {
	return true, nil
}

func (f *fakeRelayInvsContext) GetVirtualSelectedParentChainFromBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) GetHeadersSelectedTip() (*externalapi.DomainHash, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) MiningManager() miningmanager.MiningManager {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) Consensus() externalapi.Consensus {
	return f
}

func (f *fakeRelayInvsContext) Domain() domain.Domain {
	return f
}

func (f *fakeRelayInvsContext) Config() *config.Config {
	return &config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: f.params,
			},
		},
	}
}

func (f *fakeRelayInvsContext) OnNewBlock(block *externalapi.DomainBlock, blockInsertionResult *externalapi.BlockInsertionResult) error {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) SharedRequestedBlocks() *blockrelay.SharedRequestedBlocks {
	return blockrelay.NewSharedRequestedBlocks()
}

func (f *fakeRelayInvsContext) Broadcast(message appmessage.Message) error {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) AddOrphan(orphanBlock *externalapi.DomainBlock) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) GetOrphanRoots(orphanHash *externalapi.DomainHash) ([]*externalapi.DomainHash, bool, error) {
	panic("unimplemented")
}

func (f *fakeRelayInvsContext) IsOrphan(blockHash *externalapi.DomainHash) bool {
	return false
}

func (f *fakeRelayInvsContext) IsIBDRunning() bool {
	return false
}

func (f *fakeRelayInvsContext) TrySetIBDRunning(ibdPeer *peerpkg.Peer) bool {
	return true
}

func (f *fakeRelayInvsContext) UnsetIBDRunning() {
}

func TestHandleRelayInvsErrors(t *testing.T) {
	triggerIBD := func(incomingRoute, outgoingRoute *router.Route) {
		err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(consensushashing.BlockHash(orphanBlock)))
		if err != nil {
			t.Fatalf("Enqueue: %+v", err)
		}

		msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
		if err != nil {
			t.Fatalf("DequeueWithTimeout: %+v", err)
		}
		_ = msg.(*appmessage.MsgRequestRelayBlocks)

		err = incomingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(orphanBlock))
		if err != nil {
			t.Fatalf("Enqueue: %+v", err)
		}

		msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
		if err != nil {
			t.Fatalf("DequeueWithTimeout: %+v", err)
		}
		_ = msg.(*appmessage.MsgRequestBlockLocator)

		err = incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
		if err != nil {
			t.Fatalf("Enqueue: %+v", err)
		}
	}

	tests := []struct {
		name                 string
		funcToExecute        func(incomingRoute, outgoingRoute *router.Route)
		expectsProtocolError bool
		expectsBan           bool
		expectsErrToContain  string
	}{
		{
			name: "sending unexpected message type",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				err := incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "message in the block relay handleRelayInvsFlow while expecting an inv message",
		},
		{
			name: "sending a known invalid inv",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(knownInvalidBlockHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "sent inv of an invalid block",
		},
		{
			name: "sending unrequested block",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(unknownBlockHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestRelayBlocks)

				err = incomingRoute.Enqueue(appmessage.NewMsgBlock(&appmessage.MsgBlockHeader{
					Version:              0,
					ParentHashes:         nil,
					HashMerkleRoot:       &externalapi.DomainHash{},
					AcceptedIDMerkleRoot: &externalapi.DomainHash{},
					UTXOCommitment:       &externalapi.DomainHash{},
					Timestamp:            mstime.Time{},
					Bits:                 0,
					Nonce:                0,
				}))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "got unrequested block",
		},
		{
			name: "sending invalid block",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(invalidBlockHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestRelayBlocks)

				err = incomingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(invalidBlock))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "got invalid block",
		},
		{
			name: "sending unexpected message instead of block locator",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(consensushashing.BlockHash(orphanBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestRelayBlocks)

				err = incomingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(orphanBlock))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestBlockLocator)

				// Sending a block while expected a block locator
				err = incomingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(orphanBlock))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. expected: %s",
				appmessage.CmdBlockLocator),
		},
		{
			name: "sending unknown highest hash",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgIBDBlockLocator)

				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(unknownBlockHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "is not in the original blockLocator",
		},
		{
			name: "sending unexpected type instead of highest hash",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgIBDBlockLocator)

				err = incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. expected: %s",
				appmessage.CmdIBDBlockLocatorHighestHash),
		},
		{
			name: "sending unexpected type instead of a header",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}

				highestHash := msg.(*appmessage.MsgIBDBlockLocator).BlockLocatorHashes[0]
				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(highestHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestHeaders)

				// Sending unrequested block locator
				err = incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. expected: %s or %s",
				appmessage.CmdHeader, appmessage.CmdDoneHeaders),
		},
		{
			name: "sending unexpected type instead of pruning point hash",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}

				highestHash := msg.(*appmessage.MsgIBDBlockLocator).BlockLocatorHashes[0]
				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(highestHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestHeaders)

				err = incomingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointHashMessage)

				// Sending unrequested block locator
				err = incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. expected: %s",
				appmessage.CmdPruningPointHash),
		},
		{
			name: "sending unexpected type instead of pruning point block",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}

				highestHash := msg.(*appmessage.MsgIBDBlockLocator).BlockLocatorHashes[0]
				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(highestHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestHeaders)

				err = incomingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointHashMessage)

				err = incomingRoute.Enqueue(appmessage.NewPruningPointHashMessage(validPruningPointHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointUTXOSetAndBlock)

				// Sending unrequested block locator
				err = incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. expected: %s",
				appmessage.CmdIBDBlock),
		},
		{
			name: "sending unexpected type instead of UTXO chunk",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}

				highestHash := msg.(*appmessage.MsgIBDBlockLocator).BlockLocatorHashes[0]
				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(highestHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestHeaders)

				err = incomingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointHashMessage)

				err = incomingRoute.Enqueue(appmessage.NewPruningPointHashMessage(validPruningPointHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointUTXOSetAndBlock)

				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(validPruningPointBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				// Sending unrequested block locator
				err = incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. "+
				"expected: %s or %s or %s", appmessage.CmdPruningPointUTXOSetChunk,
				appmessage.CmdDonePruningPointUTXOSetChunks, appmessage.CmdUnexpectedPruningPoint),
		},
		{
			name: "sending invalid pruning point",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}

				highestHash := msg.(*appmessage.MsgIBDBlockLocator).BlockLocatorHashes[0]
				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(highestHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestHeaders)

				err = incomingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointHashMessage)

				err = incomingRoute.Enqueue(appmessage.NewPruningPointHashMessage(invalidPruningPointHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointUTXOSetAndBlock)

				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(invalidPruningPointBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				err = incomingRoute.Enqueue(appmessage.NewMsgDonePruningPointUTXOSetChunks())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "error with pruning point UTXO set",
		},
		{
			name: "sending unexpected type instead of IBD block",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}

				highestHash := msg.(*appmessage.MsgIBDBlockLocator).BlockLocatorHashes[0]
				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(highestHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestHeaders)

				err = incomingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointHashMessage)

				err = incomingRoute.Enqueue(appmessage.NewPruningPointHashMessage(validPruningPointHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointUTXOSetAndBlock)

				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(validPruningPointBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				err = incomingRoute.Enqueue(appmessage.NewMsgDonePruningPointUTXOSetChunks())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestIBDBlocks)

				// Sending unrequested block locator
				err = incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. expected: %s",
				appmessage.CmdIBDBlock),
		},
		{
			name: "sending unexpected IBD block",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}

				highestHash := msg.(*appmessage.MsgIBDBlockLocator).BlockLocatorHashes[0]
				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(highestHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestHeaders)

				err = incomingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointHashMessage)

				err = incomingRoute.Enqueue(appmessage.NewPruningPointHashMessage(validPruningPointHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointUTXOSetAndBlock)

				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(validPruningPointBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				err = incomingRoute.Enqueue(appmessage.NewMsgDonePruningPointUTXOSetChunks())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestIBDBlocks)

				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(unexpectedIBDBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "expected block",
		},
		{
			name: "sending invalid IBD block",
			funcToExecute: func(incomingRoute, outgoingRoute *router.Route) {
				triggerIBD(incomingRoute, outgoingRoute)

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}

				highestHash := msg.(*appmessage.MsgIBDBlockLocator).BlockLocatorHashes[0]
				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlockLocatorHighestHash(highestHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestHeaders)

				err = incomingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointHashMessage)

				err = incomingRoute.Enqueue(appmessage.NewPruningPointHashMessage(validPruningPointHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestPruningPointUTXOSetAndBlock)

				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(validPruningPointBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				err = incomingRoute.Enqueue(appmessage.NewMsgDonePruningPointUTXOSetChunks())
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err = outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestIBDBlocks)

				err = incomingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(invalidBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "invalid block",
		},
	}

	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		for _, test := range tests {
			incomingRoute := router.NewRoute()
			outgoingRoute := router.NewRoute()
			peer := peerpkg.New(nil)
			errChan := make(chan error)
			go func() {
				errChan <- blockrelay.HandleRelayInvs(&fakeRelayInvsContext{
					params: params,
				}, incomingRoute, outgoingRoute, peer)
			}()

			test.funcToExecute(incomingRoute, outgoingRoute)

			select {
			case err := <-errChan:
				checkFlowError(t, err, test.expectsProtocolError, test.expectsBan, test.expectsErrToContain)
			case <-time.After(time.Second):
				t.Fatalf("timed out after %s", time.Second)
			}
		}
	})
}
