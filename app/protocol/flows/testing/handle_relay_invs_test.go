package testing

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/blockrelay"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus"
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
	"github.com/pkg/errors"
)

var headerOnlyBlock = &externalapi.DomainBlock{
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
	Transactions: []*externalapi.DomainTransaction{{}},
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
	Transactions: []*externalapi.DomainTransaction{{}},
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
	Transactions: []*externalapi.DomainTransaction{{}},
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
	Transactions: []*externalapi.DomainTransaction{{}},
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
	Transactions: []*externalapi.DomainTransaction{{}},
}

var unknownBlockHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})
var knownInvalidBlockHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})
var validPruningPointHash = consensushashing.BlockHash(validPruningPointBlock)
var invalidBlockHash = consensushashing.BlockHash(invalidBlock)
var invalidPruningPointHash = consensushashing.BlockHash(invalidPruningPointBlock)
var orphanBlockHash = consensushashing.BlockHash(orphanBlock)
var headerOnlyBlockHash = consensushashing.BlockHash(headerOnlyBlock)

type fakeRelayInvsContext struct {
	testName             string
	params               *dagconfig.Params
	askedOrphanBlockInfo bool
	finishedIBD          chan struct{}

	trySetIBDRunningResponse                      bool
	isValidPruningPointResponse                   bool
	validateAndInsertImportedPruningPointResponse error
	getBlockInfoResponse                          *externalapi.BlockInfo
	validateAndInsertBlockResponse                error
	virtualBlueScore                              uint64
	rwLock                                        sync.RWMutex
}

func (f *fakeRelayInvsContext) PopulateMass(*externalapi.DomainTransaction) {
	panic("implement me")
}

func (f *fakeRelayInvsContext) IsRecoverableError(err error) bool {
	panic("implement me")
}

func (f *fakeRelayInvsContext) Init(skipAddingGenesis bool) error {
	panic("implement me")
}

func (f *fakeRelayInvsContext) ValidateAndInsertBlockWithTrustedData(block *externalapi.BlockWithTrustedData, validateUTXO bool) (*externalapi.BlockInsertionResult, error) {
	panic("implement me")
}

func (f *fakeRelayInvsContext) PruningPointAndItsAnticoneWithTrustedData() ([]*externalapi.BlockWithTrustedData, error) {
	panic("implement me")
}

func (f *fakeRelayInvsContext) DeleteStagingConsensus() error {
	panic("implement me")
}

func (f *fakeRelayInvsContext) StagingConsensus() externalapi.Consensus {
	panic("implement me")
}

func (f *fakeRelayInvsContext) InitStagingConsensus() error {
	panic("implement me")
}

func (f *fakeRelayInvsContext) CommitStagingConsensus() error {
	panic("implement me")
}

func (f *fakeRelayInvsContext) EstimateNetworkHashesPerSecond(startHash *externalapi.DomainHash, windowSize int) (uint64, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) GetBlockEvenIfHeaderOnly(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	panic("implement me")
}

func (f *fakeRelayInvsContext) GetBlockRelations(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, *externalapi.DomainHash, []*externalapi.DomainHash, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) OnPruningPointUTXOSetOverride() error {
	return nil
}

func (f *fakeRelayInvsContext) GetVirtualUTXOs(expectedVirtualParents []*externalapi.DomainHash, fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) Anticone(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) ValidateAndInsertBlock(*externalapi.DomainBlock, bool) (*externalapi.BlockInsertionResult, error) {
	return nil, f.validateAndInsertBlockResponse
}

func (f *fakeRelayInvsContext) ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) GetBlock(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) GetBlockHeader(blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) GetBlockInfo(blockHash *externalapi.DomainHash) (*externalapi.BlockInfo, error) {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	if f.getBlockInfoResponse != nil {
		return f.getBlockInfoResponse, nil
	}

	return &externalapi.BlockInfo{
		Exists: false,
	}, nil
}

func (f *fakeRelayInvsContext) GetBlockAcceptanceData(blockHash *externalapi.DomainHash) (externalapi.AcceptanceData, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) GetHashesBetween(lowHash, highHash *externalapi.DomainHash, maxBlocks uint64) (hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	// This is done so we can test getting invalid block during IBD.
	return []*externalapi.DomainHash{invalidBlockHash}, nil
}

func (f *fakeRelayInvsContext) GetPruningPointUTXOs(expectedPruningPointHash *externalapi.DomainHash, fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) PruningPoint() (*externalapi.DomainHash, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) ClearImportedPruningPointData() error {
	return nil
}

func (f *fakeRelayInvsContext) AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) ValidateAndInsertImportedPruningPoint(newPruningPoint *externalapi.DomainHash) error {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return f.validateAndInsertImportedPruningPointResponse
}

func (f *fakeRelayInvsContext) GetVirtualSelectedParent() (*externalapi.DomainHash, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) CreateBlockLocatorFromPruningPoint(highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) CreateHeadersSelectedChainBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return externalapi.BlockLocator{
		f.params.GenesisHash,
	}, nil
}

func (f *fakeRelayInvsContext) CreateFullHeadersSelectedChainBlockLocator() (externalapi.BlockLocator, error) {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return externalapi.BlockLocator{
		f.params.GenesisHash,
	}, nil
}

func (f *fakeRelayInvsContext) GetSyncInfo() (*externalapi.SyncInfo, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) Tips() ([]*externalapi.DomainHash, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) SetVirtualBlueScore(blueSCore uint64) {
	f.rwLock.Lock()
	defer f.rwLock.Unlock()
	f.virtualBlueScore = blueSCore
}

func (f *fakeRelayInvsContext) GetVirtualInfo() (*externalapi.VirtualInfo, error) {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return &externalapi.VirtualInfo{
		ParentHashes:   nil,
		Bits:           0,
		PastMedianTime: 0,
		BlueScore:      f.virtualBlueScore,
		DAAScore:       0,
	}, nil
}

func (f *fakeRelayInvsContext) GetVirtualDAAScore() (uint64, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) IsValidPruningPoint(blockHash *externalapi.DomainHash) (bool, error) {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return f.isValidPruningPointResponse, nil
}

func (f *fakeRelayInvsContext) GetVirtualSelectedParentChainFromBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) GetHeadersSelectedTip() (*externalapi.DomainHash, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) MiningManager() miningmanager.MiningManager {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) Consensus() externalapi.Consensus {
	return f
}

func (f *fakeRelayInvsContext) Domain() domain.Domain {
	return f
}

func (f *fakeRelayInvsContext) Config() *config.Config {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return &config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: f.params,
			},
		},
	}
}

func (f *fakeRelayInvsContext) OnNewBlock(block *externalapi.DomainBlock, blockInsertionResult *externalapi.BlockInsertionResult) error {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) SharedRequestedBlocks() *blockrelay.SharedRequestedBlocks {
	return blockrelay.NewSharedRequestedBlocks()
}

func (f *fakeRelayInvsContext) Broadcast(message appmessage.Message) error {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) AddOrphan(orphanBlock *externalapi.DomainBlock) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) GetOrphanRoots(orphanHash *externalapi.DomainHash) ([]*externalapi.DomainHash, bool, error) {
	panic(errors.Errorf("called unimplemented function from test '%s'", f.testName))
}

func (f *fakeRelayInvsContext) IsOrphan(blockHash *externalapi.DomainHash) bool {
	return false
}

func (f *fakeRelayInvsContext) IsIBDRunning() bool {
	return false
}

func (f *fakeRelayInvsContext) TrySetIBDRunning(ibdPeer *peerpkg.Peer) bool {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return f.trySetIBDRunningResponse
}

func (f *fakeRelayInvsContext) UnsetIBDRunning() {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	close(f.finishedIBD)
}

func (f *fakeRelayInvsContext) SetValidateAndInsertBlockResponse(err error) {
	f.rwLock.Lock()
	defer f.rwLock.Unlock()
	f.validateAndInsertBlockResponse = err
}

func (f *fakeRelayInvsContext) SetValidateAndInsertImportedPruningPointResponse(err error) {
	f.rwLock.Lock()
	defer f.rwLock.Unlock()
	f.validateAndInsertImportedPruningPointResponse = err
}

func (f *fakeRelayInvsContext) SetGetBlockInfoResponse(info externalapi.BlockInfo) {
	f.rwLock.Lock()
	defer f.rwLock.Unlock()
	f.getBlockInfoResponse = &info
}

func (f *fakeRelayInvsContext) SetTrySetIBDRunningResponse(b bool) {
	f.rwLock.Lock()
	defer f.rwLock.Unlock()
	f.trySetIBDRunningResponse = b
}

func (f *fakeRelayInvsContext) SetIsValidPruningPointResponse(b bool) {
	f.rwLock.Lock()
	defer f.rwLock.Unlock()
	f.isValidPruningPointResponse = b
}

func (f *fakeRelayInvsContext) GetGenesisHeader() externalapi.BlockHeader {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return f.params.GenesisBlock.Header
}

func (f *fakeRelayInvsContext) GetFinishedIBDChan() chan struct{} {
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	return f.finishedIBD
}

func TestHandleRelayInvs(t *testing.T) {
	triggerIBD := func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
		err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(consensushashing.BlockHash(orphanBlock)))
		if err != nil {
			t.Fatalf("Enqueue: %+v", err)
		}

		msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
		if err != nil {
			t.Fatalf("DequeueWithTimeout: %+v", err)
		}
		_ = msg.(*appmessage.MsgRequestRelayBlocks)

		context.SetValidateAndInsertBlockResponse(ruleerrors.NewErrMissingParents(orphanBlock.Header.ParentHashes()))

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

	checkNoActivity := func(t *testing.T, outgoingRoute *router.Route) {
		msg, err := outgoingRoute.DequeueWithTimeout(5 * time.Second)
		if !errors.Is(err, router.ErrTimeout) {
			t.Fatalf("Expected to time out, but got message %s and error %+v", msg.Command(), err)
		}
	}

	tests := []struct {
		name                 string
		funcToExecute        func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext)
		expectsProtocolError bool
		expectsBan           bool
		expectsIBDToFinish   bool
		expectsErrToContain  string
	}{
		{
			name: "sending unexpected message type",
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
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
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {

				context.SetGetBlockInfoResponse(externalapi.BlockInfo{
					Exists:      true,
					BlockStatus: externalapi.StatusInvalid,
				})

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
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
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
			name: "sending header only block on relay",
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
				err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(headerOnlyBlockHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestRelayBlocks)

				err = incomingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(headerOnlyBlock))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsErrToContain:  "block where expected block with body",
		},
		{
			name: "sending invalid block",
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
				err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(invalidBlockHash))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestRelayBlocks)

				context.SetValidateAndInsertBlockResponse(ruleerrors.ErrBadMerkleRoot)
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
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
				err := incomingRoute.Enqueue(appmessage.NewMsgInvBlock(consensushashing.BlockHash(orphanBlock)))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}

				msg, err := outgoingRoute.DequeueWithTimeout(time.Second)
				if err != nil {
					t.Fatalf("DequeueWithTimeout: %+v", err)
				}
				_ = msg.(*appmessage.MsgRequestRelayBlocks)

				context.SetValidateAndInsertBlockResponse(ruleerrors.NewErrMissingParents(orphanBlock.Header.ParentHashes()))
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
			name: "starting IBD when peer is already in IBD",
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
				context.SetTrySetIBDRunningResponse(false)
				triggerIBD(t, incomingRoute, outgoingRoute, context)

				checkNoActivity(t, outgoingRoute)
			},
			expectsIBDToFinish: false,
		},
		{
			name: "sending unknown highest hash",
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
				triggerIBD(t, incomingRoute, outgoingRoute, context)

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
			expectsIBDToFinish:   true,
			expectsErrToContain:  "is not in the original blockLocator",
		},
		{
			name: "sending unexpected type instead of highest hash",
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
				triggerIBD(t, incomingRoute, outgoingRoute, context)

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
			expectsIBDToFinish:   true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. expected: %s",
				appmessage.CmdIBDBlockLocatorHighestHash),
		},
		{
			name: "sending unexpected type instead of a header",
			funcToExecute: func(t *testing.T, incomingRoute, outgoingRoute *router.Route, context *fakeRelayInvsContext) {
				triggerIBD(t, incomingRoute, outgoingRoute, context)

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
				_ = msg.(*appmessage.MsgRequestIBDBlocks)

				// Sending unrequested block locator
				err = incomingRoute.Enqueue(appmessage.NewMsgBlockLocator(nil))
				if err != nil {
					t.Fatalf("Enqueue: %+v", err)
				}
			},
			expectsProtocolError: true,
			expectsBan:           true,
			expectsIBDToFinish:   true,
			expectsErrToContain: fmt.Sprintf("received unexpected message type. expected: %s or %s",
				appmessage.CmdIBDBlocks, appmessage.CmdDoneIBDBlocks),
		},
	}

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		for _, test := range tests {

			// This is done to avoid race condition
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				incomingRoute := router.NewRoute("incoming")
				outgoingRoute := router.NewRoute("outgoing")
				peer := peerpkg.New(nil)
				errChan := make(chan error)
				context := &fakeRelayInvsContext{
					testName:    test.name,
					params:      &consensusConfig.Params,
					finishedIBD: make(chan struct{}),

					trySetIBDRunningResponse:    true,
					isValidPruningPointResponse: true,
				}
				go func() {
					errChan <- blockrelay.HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
				}()

				test.funcToExecute(t, incomingRoute, outgoingRoute, context)

				if test.expectsErrToContain != "" {
					select {
					case err := <-errChan:
						checkFlowError(t, err, test.expectsProtocolError, test.expectsBan, test.expectsErrToContain)
					case <-time.After(10 * time.Second):
						t.Fatalf("waiting for error timed out after %s", 10*time.Second)
					}
				}

				select {
				case <-context.GetFinishedIBDChan():
					if !test.expectsIBDToFinish {
						t.Fatalf("IBD unexpecetedly finished")
					}
				case <-time.After(10 * time.Second):
					if test.expectsIBDToFinish {
						t.Fatalf("IBD didn't finished after %d", time.Second)
					}
				}

				if test.expectsErrToContain == "" {
					// Close the route to stop the flow
					incomingRoute.Close()

					select {
					case err := <-errChan:
						if !errors.Is(err, router.ErrRouteClosed) {
							t.Fatalf("unexpected error %+v", err)
						}
					case <-time.After(10 * time.Second):
						t.Fatalf("waiting for flow to finish timed out after %s", time.Second)
					}
				}
			})
		}
	})
}
