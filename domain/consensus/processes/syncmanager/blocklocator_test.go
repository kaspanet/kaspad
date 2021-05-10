package syncmanager_test

import (
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

func TestCreateBlockLocator(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig,
			"TestCreateBlockLocator")
		if err != nil {
			t.Fatalf("NewTestConsensus: %+v", err)
		}
		defer tearDown(false)

		chain := []*externalapi.DomainHash{consensusConfig.GenesisHash}
		tipHash := consensusConfig.GenesisHash
		for i := 0; i < 20; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		sideChainTipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Check a situation where low hash is not on the exact step blue score
		locator, err := tc.CreateBlockLocator(consensusConfig.GenesisHash, tipHash, 0)
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			chain[20],
			chain[19],
			chain[17],
			chain[13],
			chain[5],
			chain[0],
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check a situation where low hash is on the exact step blue score
		locator, err = tc.CreateBlockLocator(chain[5], tipHash, 0)
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			chain[20],
			chain[19],
			chain[17],
			chain[13],
			chain[5],
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check block locator with limit
		locator, err = tc.CreateBlockLocator(consensusConfig.GenesisHash, tipHash, 3)
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			chain[20],
			chain[19],
			chain[17],
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check a block locator from genesis to genesis
		locator, err = tc.CreateBlockLocator(consensusConfig.GenesisHash, consensusConfig.GenesisHash, 0)
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			consensusConfig.GenesisHash,
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check a block locator from one block to the same block
		locator, err = tc.CreateBlockLocator(chain[7], chain[7], 0)
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			chain[7],
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check block locator with incompatible blocks
		_, err = tc.CreateBlockLocator(sideChainTipHash, tipHash, 0)
		expectedErr := "highHash and lowHash are not in the same selected parent chain"
		if err == nil || !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("expected error '%s' but got '%s'", expectedErr, err)
		}

		// Check block locator with non exist blocks
		_, err = tc.CreateBlockLocator(&externalapi.DomainHash{}, tipHash, 0)
		expectedErr = "does not exist"
		if err == nil || !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("expected error '%s' but got '%s'", expectedErr, err)
		}

		_, err = tc.CreateBlockLocator(tipHash, &externalapi.DomainHash{}, 0)
		expectedErr = "does not exist"
		if err == nil || !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("expected error '%s' but got '%s'", expectedErr, err)
		}
	})
}

func TestCreateHeadersSelectedChainBlockLocator(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig,
			"TestCreateHeadersSelectedChainBlockLocator")
		if err != nil {
			t.Fatalf("NewTestConsensus: %+v", err)
		}
		defer tearDown(false)

		chain := []*externalapi.DomainHash{consensusConfig.GenesisHash}
		tipHash := consensusConfig.GenesisHash
		for i := 0; i < 20; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		sideChainTipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Check a situation where low hash is not on the exact step
		locator, err := tc.CreateHeadersSelectedChainBlockLocator(consensusConfig.GenesisHash, tipHash)
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			chain[20],
			chain[19],
			chain[17],
			chain[13],
			chain[5],
			chain[0],
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check a situation where low hash is on the exact step
		locator, err = tc.CreateHeadersSelectedChainBlockLocator(chain[5], tipHash)
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			chain[20],
			chain[19],
			chain[17],
			chain[13],
			chain[5],
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check a block locator from genesis to genesis
		locator, err = tc.CreateHeadersSelectedChainBlockLocator(consensusConfig.GenesisHash, consensusConfig.GenesisHash)
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			consensusConfig.GenesisHash,
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check a block locator from one block to the same block
		locator, err = tc.CreateHeadersSelectedChainBlockLocator(chain[7], chain[7])
		if err != nil {
			t.Fatalf("CreateBlockLocator: %+v", err)
		}

		if !externalapi.HashesEqual(locator, []*externalapi.DomainHash{
			chain[7],
		}) {
			t.Fatalf("unexpected block locator %s", locator)
		}

		// Check block locator with low hash higher than high hash
		_, err = tc.CreateHeadersSelectedChainBlockLocator(chain[20], chain[19])
		expectedErr := "cannot build block locator while highHash is lower than lowHash"
		if err == nil || !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("expected error '%s' but got '%s'", expectedErr, err)
		}

		// Check block locator with non chain blocks
		_, err = tc.CreateHeadersSelectedChainBlockLocator(consensusConfig.GenesisHash, sideChainTipHash)
		if !errors.Is(err, model.ErrBlockNotInSelectedParentChain) {
			t.Fatalf("expected error '%s' but got '%s'", database.ErrNotFound, err)
		}
	})
}
