package dagtraversalmanager_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

func TestBlockWindow(t *testing.T) {
	tests := map[string][]*struct {
		parents        []string
		id             string //id is a virtual entity that is used only for tests so we can define relations between blocks without knowing their hash
		expectedWindow []string
	}{
		dagconfig.MainnetParams.Name: {
			{
				parents:        []string{"A"},
				id:             "B",
				expectedWindow: []string{},
			},
			{
				parents:        []string{"B"},
				id:             "C",
				expectedWindow: []string{"B"},
			},
			{
				parents:        []string{"B"},
				id:             "D",
				expectedWindow: []string{"B"},
			},
			{
				parents:        []string{"C", "D"},
				id:             "E",
				expectedWindow: []string{"D", "C", "B"},
			},
			{
				parents:        []string{"C", "D"},
				id:             "F",
				expectedWindow: []string{"D", "C", "B"},
			},
			{
				parents:        []string{"A"},
				id:             "G",
				expectedWindow: []string{},
			},
			{
				parents:        []string{"G"},
				id:             "H",
				expectedWindow: []string{"G"},
			},
			{
				parents:        []string{"H", "F"},
				id:             "I",
				expectedWindow: []string{"F", "D", "C", "H", "B", "G"},
			},
			{
				parents:        []string{"I"},
				id:             "J",
				expectedWindow: []string{"I", "F", "D", "C", "H", "B", "G"},
			},
			{
				parents:        []string{"J"},
				id:             "K",
				expectedWindow: []string{"J", "I", "F", "D", "C", "H", "B", "G"},
			},
			{
				parents:        []string{"K"},
				id:             "L",
				expectedWindow: []string{"K", "J", "I", "F", "D", "C", "H", "B", "G"},
			},
			{
				parents:        []string{"L"},
				id:             "M",
				expectedWindow: []string{"L", "K", "J", "I", "F", "D", "C", "H", "B", "G"},
			},
			{
				parents:        []string{"M"},
				id:             "N",
				expectedWindow: []string{"M", "L", "K", "J", "I", "F", "D", "C", "H", "B"},
			},
			{
				parents:        []string{"N"},
				id:             "O",
				expectedWindow: []string{"N", "M", "L", "K", "J", "I", "F", "D", "C", "H"},
			},
		},
		dagconfig.TestnetParams.Name: {
			{
				parents:        []string{"A"},
				id:             "B",
				expectedWindow: []string{},
			},
			{
				parents:        []string{"B"},
				id:             "C",
				expectedWindow: []string{"B"},
			},
			{
				parents:        []string{"B"},
				id:             "D",
				expectedWindow: []string{"B"},
			},
			{
				parents:        []string{"C", "D"},
				id:             "E",
				expectedWindow: []string{"C", "D", "B"},
			},
			{
				parents:        []string{"C", "D"},
				id:             "F",
				expectedWindow: []string{"C", "D", "B"},
			},
			{
				parents:        []string{"A"},
				id:             "G",
				expectedWindow: []string{},
			},
			{
				parents:        []string{"G"},
				id:             "H",
				expectedWindow: []string{"G"},
			},
			{
				parents:        []string{"H", "F"},
				id:             "I",
				expectedWindow: []string{"F", "C", "H", "D", "B", "G"},
			},
			{
				parents:        []string{"I"},
				id:             "J",
				expectedWindow: []string{"I", "F", "C", "H", "D", "B", "G"},
			},
			{
				parents:        []string{"J"},
				id:             "K",
				expectedWindow: []string{"J", "I", "F", "C", "H", "D", "B", "G"},
			},
			{
				parents:        []string{"K"},
				id:             "L",
				expectedWindow: []string{"K", "J", "I", "F", "C", "H", "D", "B", "G"},
			},
			{
				parents:        []string{"L"},
				id:             "M",
				expectedWindow: []string{"L", "K", "J", "I", "F", "C", "H", "D", "B", "G"},
			},
			{
				parents:        []string{"M"},
				id:             "N",
				expectedWindow: []string{"M", "L", "K", "J", "I", "F", "C", "H", "D", "B"},
			},
			{
				parents:        []string{"N"},
				id:             "O",
				expectedWindow: []string{"N", "M", "L", "K", "J", "I", "F", "C", "H", "D"},
			},
		},
		dagconfig.DevnetParams.Name: {
			{
				parents:        []string{"A"},
				id:             "B",
				expectedWindow: []string{},
			},
			{
				parents:        []string{"B"},
				id:             "C",
				expectedWindow: []string{"B"},
			},
			{
				parents:        []string{"B"},
				id:             "D",
				expectedWindow: []string{"B"},
			},
			{
				parents:        []string{"C", "D"},
				id:             "E",
				expectedWindow: []string{"D", "C", "B"},
			},
			{
				parents:        []string{"C", "D"},
				id:             "F",
				expectedWindow: []string{"D", "C", "B"},
			},
			{
				parents:        []string{"A"},
				id:             "G",
				expectedWindow: []string{},
			},
			{
				parents:        []string{"G"},
				id:             "H",
				expectedWindow: []string{"G"},
			},
			{
				parents:        []string{"H", "F"},
				id:             "I",
				expectedWindow: []string{"F", "D", "H", "C", "B", "G"},
			},
			{
				parents:        []string{"I"},
				id:             "J",
				expectedWindow: []string{"I", "F", "D", "H", "C", "B", "G"},
			},
			{
				parents:        []string{"J"},
				id:             "K",
				expectedWindow: []string{"J", "I", "F", "D", "H", "C", "B", "G"},
			},
			{
				parents:        []string{"K"},
				id:             "L",
				expectedWindow: []string{"K", "J", "I", "F", "D", "H", "C", "B", "G"},
			},
			{
				parents:        []string{"L"},
				id:             "M",
				expectedWindow: []string{"L", "K", "J", "I", "F", "D", "H", "C", "B", "G"},
			},
			{
				parents:        []string{"M"},
				id:             "N",
				expectedWindow: []string{"M", "L", "K", "J", "I", "F", "D", "H", "C", "B"},
			},
			{
				parents:        []string{"N"},
				id:             "O",
				expectedWindow: []string{"N", "M", "L", "K", "J", "I", "F", "D", "H", "C"},
			},
		},
		dagconfig.SimnetParams.Name: {
			{
				parents:        []string{"A"},
				id:             "B",
				expectedWindow: []string{},
			},
			{
				parents:        []string{"B"},
				id:             "C",
				expectedWindow: []string{"B"},
			},
			{
				parents:        []string{"B"},
				id:             "D",
				expectedWindow: []string{"B"},
			},
			{
				parents:        []string{"D", "C"},
				id:             "E",
				expectedWindow: []string{"C", "D", "B"},
			},
			{
				parents:        []string{"D", "C"},
				id:             "F",
				expectedWindow: []string{"C", "D", "B"},
			},
			{
				parents:        []string{"A"},
				id:             "G",
				expectedWindow: []string{},
			},
			{
				parents:        []string{"G"},
				id:             "H",
				expectedWindow: []string{"G"},
			},
			{
				parents:        []string{"H", "F"},
				id:             "I",
				expectedWindow: []string{"F", "H", "C", "D", "B", "G"},
			},
			{
				parents:        []string{"I"},
				id:             "J",
				expectedWindow: []string{"I", "F", "H", "C", "D", "B", "G"},
			},
			{
				parents:        []string{"J"},
				id:             "K",
				expectedWindow: []string{"J", "I", "F", "H", "C", "D", "B", "G"},
			},
			{
				parents:        []string{"K"},
				id:             "L",
				expectedWindow: []string{"K", "J", "I", "F", "H", "C", "D", "B", "G"},
			},
			{
				parents:        []string{"L"},
				id:             "M",
				expectedWindow: []string{"L", "K", "J", "I", "F", "H", "C", "D", "B", "G"},
			},
			{
				parents:        []string{"M"},
				id:             "N",
				expectedWindow: []string{"M", "L", "K", "J", "I", "F", "H", "C", "D", "B"},
			},
			{
				parents:        []string{"N"},
				id:             "O",
				expectedWindow: []string{"N", "M", "L", "K", "J", "I", "F", "H", "C", "D"},
			},
		},
	}
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.K = 1
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig, "TestBlockWindow")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		windowSize := 10
		blockByIDMap := make(map[string]*externalapi.DomainHash)
		idByBlockMap := make(map[externalapi.DomainHash]string)
		blockByIDMap["A"] = consensusConfig.GenesisHash
		idByBlockMap[*consensusConfig.GenesisHash] = "A"

		blocksData := tests[consensusConfig.Name]

		for _, blockData := range blocksData {
			parents := hashset.New()
			for _, parentID := range blockData.parents {
				parent := blockByIDMap[parentID]
				parents.Add(parent)
			}

			block, _, err := tc.AddBlock(parents.ToSlice(), nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			blockByIDMap[blockData.id] = block
			idByBlockMap[*block] = blockData.id

			stagingArea := model.NewStagingArea()

			window, err := tc.DAGTraversalManager().BlockWindow(stagingArea, block, windowSize)
			if err != nil {
				t.Fatalf("BlockWindow: %s", err)
			}
			sort.Sort(testutils.NewTestGhostDAGSorter(stagingArea, window, tc, t))
			if err := checkWindowIDs(window, blockData.expectedWindow, idByBlockMap); err != nil {
				t.Errorf("Unexpected values for window for block %s: %s", blockData.id, err)
			}
		}
	})
}

func checkWindowIDs(window []*externalapi.DomainHash, expectedIDs []string, idByBlockMap map[externalapi.DomainHash]string) error {
	ids := make([]string, len(window))
	for i, node := range window {
		ids[i] = idByBlockMap[*node]
	}
	if !reflect.DeepEqual(ids, expectedIDs) {
		return errors.Errorf("window expected to have blocks %s but got %s", expectedIDs, ids)
	}
	return nil
}
