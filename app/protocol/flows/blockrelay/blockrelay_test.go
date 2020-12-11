package blockrelay

import (
	"errors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type mocRelayBlockRequestsContext struct {
	domain domain.Domain
}

func (m *mocRelayBlockRequestsContext) Domain() domain.Domain {
	return m.domain
}

func newMocRelayBlockRequestsContext(testName string) (moc *mocRelayBlockRequestsContext, teardown func(), err error) {
	domainInstance, teardown, err := setupTestDomain(testName)
	if err != nil {
		return nil, nil, err
	}

	return &mocRelayBlockRequestsContext{
		domain: domainInstance,
	}, teardown, err
}

type mocRelayInvsContext struct {
	domain                domain.Domain
	adapter               *netadapter.NetAdapter
	sharedRequestedBlocks *SharedRequestedBlocks
	ibdRunning            bool
	orphanBlocks          []*externalapi.DomainBlock
}

func (m *mocRelayInvsContext) NetAdapter() *netadapter.NetAdapter {
	return m.adapter
}

func (m *mocRelayInvsContext) Domain() domain.Domain {
	return m.domain
}

func (m *mocRelayInvsContext) Config() *config.Config {
	return config.DefaultConfig()
}

func (m *mocRelayInvsContext) OnNewBlock(block *externalapi.DomainBlock) error {
	return nil
}

func (m *mocRelayInvsContext) SharedRequestedBlocks() *SharedRequestedBlocks {
	return m.sharedRequestedBlocks
}

func (m *mocRelayInvsContext) StartIBDIfRequired() {
}

func (m *mocRelayInvsContext) IsInIBD() bool {
	return false
}

func (m *mocRelayInvsContext) Broadcast(message appmessage.Message) error {
	return nil
}

func (m *mocRelayInvsContext) AddOrphan(orphanBlock *externalapi.DomainBlock) {
	m.orphanBlocks = append(m.orphanBlocks, orphanBlock)
}

func (m *mocRelayInvsContext) IsOrphan(blockHash *externalapi.DomainHash) bool {
	for _, block := range m.orphanBlocks {
		if *consensushashing.BlockHash(block) == *blockHash {
			return true
		}
	}
	return false
}

func (m *mocRelayInvsContext) IsIBDRunning() bool {
	return m.ibdRunning
}
func (m *mocRelayInvsContext) TrySetIBDRunning() bool {
	m.ibdRunning = true
	return true
}
func (m *mocRelayInvsContext) UnsetIBDRunning() {
	m.ibdRunning = false
}

func newMocRelayInvsContext(testName string) (moc *mocRelayInvsContext, teardown func(), err error) {
	adapter, _ := netadapter.NewNetAdapter(config.DefaultConfig())
	domainInstance, teardown, err := setupTestDomain(testName)
	if err != nil {
		return nil, nil, err
	}

	return &mocRelayInvsContext{
		domain:                domainInstance,
		adapter:               adapter,
		sharedRequestedBlocks: NewSharedRequestedBlocks(),
	}, teardown, nil
}

func setupTestDomain(testName string) (domainInstance domain.Domain, teardown func(), err error) {
	dataDir, err := ioutil.TempDir("", testName)
	if err != nil {
		return nil, nil, err
	}
	db, err := ldb.NewLevelDB(dataDir)
	if err != nil {
		return nil, nil, err
	}
	teardown = func() {
		db.Close()
		os.RemoveAll(dataDir)
	}

	params := dagconfig.SimnetParams
	domainInstance, err = domain.New(&params, db)
	if err != nil {
		teardown()
		return nil, nil, err
	}

	return domainInstance, teardown, nil
}

func initTestBaseTransactions() []*externalapi.DomainTransaction {
	testTx := []*externalapi.DomainTransaction{{
		Version:      1,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      []*externalapi.DomainTransactionOutput{},
		LockTime:     1,
		SubnetworkID: subnetworks.SubnetworkIDCoinbase,
		Gas:          1,
		PayloadHash: externalapi.DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Payload: []byte{0x01},
		Fee:     0,
		Mass:    1,
		ID: &externalapi.DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
	}}
	return testTx
}

func TestHandleRelayBlockRequests(t *testing.T) {
	peer := peerpkg.New(nil)

	context, teardown, err := newMocRelayBlockRequestsContext(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup new MocRelayBlockRequestsContext instance: %v", err)
	}
	defer teardown()

	block, err := context.Domain().Consensus().BuildBlock(&externalapi.DomainCoinbaseData{}, nil)
	if err != nil {
		t.Fatalf("consensus.BuildBlock with an empty coinbase shouldn't fail: %v", err)
	}

	err = context.Domain().Consensus().ValidateAndInsertBlock(block)
	if err != nil {
		t.Fatalf("consensus.ValidateAndInsertBlock with a block straight from consensus.BuildBlock should not fail: %v", err)
	}

	msgRequestRelayBlocks := appmessage.MsgRequestRelayBlocks{
		Hashes: []*externalapi.DomainHash{consensushashing.BlockHash(block)},
	}

	t.Run("Simple call test", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		go func() {
			outgoingRoute.Dequeue()
			incomingRoute.Close()
		}()

		incomingRoute.Enqueue(&msgRequestRelayBlocks)
		HandleRelayBlockRequests(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test on wrong message type", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayBlockRequests(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test on closed route", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Close()
		HandleRelayBlockRequests(context, incomingRoute, outgoingRoute, peer)

		err := HandleRelayBlockRequests(context, incomingRoute, outgoingRoute, peer)
		if err.Error() != routerpkg.ErrRouteClosed.Error() {
			t.Fatalf("HandleRelayBlockRequests: expected ErrRouteClosed, got %s", err)
		}
	})

	t.Run("Test block requesting", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		go func() {
			for _, hash := range msgRequestRelayBlocks.Hashes {
				message, err := outgoingRoute.Dequeue()
				if err != nil {
					t.Fatalf("HandleRelayBlockRequests: %s", err)
				}

				msgBlock := message.(*appmessage.MsgBlock)
				block, err := context.Domain().Consensus().GetBlock(hash)
				if err != nil {
					t.Fatalf("HandleRelayBlockRequests: %s", err)
				}

				blockHash := consensushashing.BlockHash(block).String()
				msgBlockHash := msgBlock.Header.BlockHash().String()
				if blockHash != msgBlockHash {
					t.Fatalf("HandleRelayBlockRequests: expected equal blocks hash %s != %s", blockHash, msgBlockHash)
				}
			}
			incomingRoute.Close()
			outgoingRoute.Close()
		}()

		incomingRoute.Enqueue(&msgRequestRelayBlocks)
		err := HandleRelayBlockRequests(context, incomingRoute, outgoingRoute, peer)
		if !errors.Is(err, routerpkg.ErrRouteClosed) {
			t.Fatalf("HandleRelayBlockRequests: expected ErrRouteClosed, got %s", err)
		}
	})
}

func TestHandleRelayInvs(t *testing.T) {
	peer := peerpkg.New(nil)
	context, teardown, err := newMocRelayInvsContext(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup new MocRelayBlockRequestsContext instance: %v", err)
	}
	defer teardown()

	block, err := context.Domain().Consensus().BuildBlock(&externalapi.DomainCoinbaseData{}, nil)
	blockHash := consensushashing.BlockHash(block)
	if err != nil {
		t.Fatalf("consensus.BuildBlock with an empty coinbase shouldn't fail: %v", err)
	}

	msgInvRelayBlock := appmessage.MsgInvRelayBlock{
		Hash: consensushashing.BlockHash(block),
	}

	t.Run("Test on wrong message type 1", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&msgInvRelayBlock)
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test on wrong message type 2", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test on closed route", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Close()

		err := HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
		if err.Error() != routerpkg.ErrRouteClosed.Error() {
			t.Fatalf("TestHandleRelayInvs: expected ErrRouteClosed, got %s", err)
		}
	})

	t.Run("Test handle invalid block", func(t *testing.T) {
		invalidBlock := &externalapi.DomainBlock{
			Header: &externalapi.DomainBlockHeader{
				ParentHashes:         []*externalapi.DomainHash{{1}},
				HashMerkleRoot:       externalapi.DomainHash{100},
				AcceptedIDMerkleRoot: externalapi.DomainHash{3},
				UTXOCommitment:       externalapi.DomainHash{4},
				TimeInMilliseconds:   5,
				Bits:                 6,
				Nonce:                7,
			},
			Transactions: initTestBaseTransactions(),
		}

		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&msgInvRelayBlock)
		incomingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(invalidBlock))

		err := HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
		if err == nil {
			t.Fatal("TestHandleRelayInvs: expected err, got nil")
		}
	})

	t.Run("Test handle valid block", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&msgInvRelayBlock)

		go func() {
			message, err := outgoingRoute.Dequeue()
			msgRequestRelayBlocks := message.(*appmessage.MsgRequestRelayBlocks)
			if err != nil {
				t.Fatalf("TestHandleRelayInvs: %s", err)
			}

			for _, hash := range msgRequestRelayBlocks.Hashes {
				if hash.String() != blockHash.String() {
					incomingRoute.Close()
					t.Fatalf("TestHandleRelayInvs: expected equal blocks hash %s != %s", blockHash.String(), hash.String())
				}
			}

			incomingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(block))
			incomingRoute.Close()
		}()

		err := HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
		if err != nil && !errors.Is(err, routerpkg.ErrRouteClosed) {
			t.Fatalf("TestHandleRelayInvs: %s", err)
		}

		blockInfo, err := context.Domain().Consensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %s", err)
		}

		if !blockInfo.Exists {
			t.Fatalf("TestHandleRelayInvs: valid block wasn't inserted %s", blockHash.String())
		}
	})

}
