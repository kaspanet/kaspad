package processes

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdag2"
	"reflect"
	"testing"
)

func TestGHOSTDA(t *testing.T) {
	//t.Errorf("helo") //keep running
	//if false {
	//	t.Fatalf("The test failed") // string - //stop
	//}

	type testGhostdagData struct {
		hash    *model.DomainHash
		parents []*model.DomainHash

		expectedBlueScore      uint64
		expectedSelectedParent *model.DomainHash
		expectedMergeSetBlues  []*model.DomainHash
		expectedMergeSetReds   []*model.DomainHash
	}

	genesisHash := &model.DomainHash{}
	dagTopology := &DAGTopologyManagerImpl{
		parentsMap: make(map[model.DomainHash][]*model.DomainHash),
	}
	dagTopology.parentsMap[*genesisHash] = nil

	ghostdagDataStore := &GHOSTDAGDataStoreImpl{
		dagMap: make(map[model.DomainHash]*model.BlockGHOSTDAGData),
	}
	ghostdagDataStore.dagMap[*genesisHash] = &model.BlockGHOSTDAGData{
		BlueScore:          1,
		SelectedParent:     nil,
		MergeSetBlues:      nil,
		MergeSetReds:       nil,
		BluesAnticoneSizes: nil,
	}
	// test2: Graph form is a chain. K = 3
	//dag := []testGhostdagData{
	//	{
	//		hash:                   &model.DomainHash{1},
	//		parents:                []*model.DomainHash{genesisHash},
	//		expectedBlueScore:      2,
	//		expectedSelectedParent: genesisHash,
	//		expectedMergeSetBlues:  []*model.DomainHash{genesisHash},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{
	//		hash:                   &model.DomainHash{2},
	//		parents:                []*model.DomainHash{{1}},
	//		expectedBlueScore:      3,
	//		expectedSelectedParent: &model.DomainHash{1},
	//		expectedMergeSetBlues:  []*model.DomainHash{{1}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{
	//		hash:                   &model.DomainHash{3},
	//		parents:                []*model.DomainHash{genesisHash},
	//		expectedBlueScore:      2,
	//		expectedSelectedParent: genesisHash,
	//		expectedMergeSetBlues:  []*model.DomainHash{genesisHash},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//}

	//test4 : The graph’s longest chain was created by malicious miners (not the “heaviest”). K = 3
	//dag4 := []testGhostdagData{
	//	{ /* 1*/
	//		hash:                   &model.DomainHash{1},
	//		parents:                []*model.DomainHash{genesisHash},
	//		expectedBlueScore:      2,
	//		expectedSelectedParent: genesisHash,
	//		expectedMergeSetBlues:  []*model.DomainHash{genesisHash},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 2 */
	//		hash:                   &model.DomainHash{2},
	//		parents:                []*model.DomainHash{{1}},
	//		expectedBlueScore:      3,
	//		expectedSelectedParent: &model.DomainHash{1},
	//		expectedMergeSetBlues:  []*model.DomainHash{{1}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 3 */
	//		hash:                   &model.DomainHash{3},
	//		parents:                []*model.DomainHash{{1}},
	//		expectedBlueScore:      3,
	//		expectedSelectedParent: &model.DomainHash{1},
	//		expectedMergeSetBlues:  []*model.DomainHash{{1}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 4 */
	//		hash:                   &model.DomainHash{4},
	//		parents:                []*model.DomainHash{{1}},
	//		expectedBlueScore:      3,
	//		expectedSelectedParent: &model.DomainHash{1},
	//		expectedMergeSetBlues:  []*model.DomainHash{{1}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 5 */
	//		hash:                   &model.DomainHash{5},
	//		parents:                []*model.DomainHash{{2}, {3}, {4}},
	//		expectedBlueScore:      6,
	//		expectedSelectedParent: &model.DomainHash{2},
	//		expectedMergeSetBlues:  []*model.DomainHash{{2}, {3}, {4}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 6 */
	//		hash:                   &model.DomainHash{6},
	//		parents:                []*model.DomainHash{genesisHash},
	//		expectedBlueScore:      2,
	//		expectedSelectedParent: genesisHash,
	//		expectedMergeSetBlues:  []*model.DomainHash{genesisHash},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 7 */
	//		hash:                   &model.DomainHash{7},
	//		parents:                []*model.DomainHash{{6}},
	//		expectedBlueScore:      3,
	//		expectedSelectedParent: &model.DomainHash{6},
	//		expectedMergeSetBlues:  []*model.DomainHash{{6}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 8 */
	//		hash:                   &model.DomainHash{8},
	//		parents:                []*model.DomainHash{{7}},
	//		expectedBlueScore:      4,
	//		expectedSelectedParent: &model.DomainHash{7},
	//		expectedMergeSetBlues:  []*model.DomainHash{{7}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 9 */
	//		hash:                   &model.DomainHash{9},
	//		parents:                []*model.DomainHash{{8}},
	//		expectedBlueScore:      5,
	//		expectedSelectedParent: &model.DomainHash{8},
	//		expectedMergeSetBlues:  []*model.DomainHash{{8}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 10 */
	//		hash:                   &model.DomainHash{10},
	//		parents:                []*model.DomainHash{{5}, {9}},
	//		expectedBlueScore:      7,
	//		expectedSelectedParent: &model.DomainHash{5},
	//		expectedMergeSetBlues:  []*model.DomainHash{{5}},
	//		expectedMergeSetReds:   []*model.DomainHash{{9}, {8}, {7}, {6}},
	//	},
	//}

	// test5: Selected Parent choice: same score – decide by hashes. K = 3
	//dag2 := []testGhostdagData{
	//	{ /* 1*/
	//		hash:                   &model.DomainHash{1},
	//		parents:                []*model.DomainHash{genesisHash},
	//		expectedBlueScore:      2,
	//		expectedSelectedParent: genesisHash,
	//		expectedMergeSetBlues:  []*model.DomainHash{genesisHash},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 2 */
	//		hash:                   &model.DomainHash{2},
	//		parents:                []*model.DomainHash{{1}},
	//		expectedBlueScore:      3,
	//		expectedSelectedParent: &model.DomainHash{1},
	//		expectedMergeSetBlues:  []*model.DomainHash{{1}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 3 */
	//		hash:                   &model.DomainHash{3},
	//		parents:                []*model.DomainHash{{1}},
	//		expectedBlueScore:      3,
	//		expectedSelectedParent: &model.DomainHash{1},
	//		expectedMergeSetBlues:  []*model.DomainHash{{1}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 4 */
	//		hash:                   &model.DomainHash{4},
	//		parents:                []*model.DomainHash{{1}},
	//		expectedBlueScore:      3,
	//		expectedSelectedParent: &model.DomainHash{1},
	//		expectedMergeSetBlues:  []*model.DomainHash{{1}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 5 */
	//		hash:                   &model.DomainHash{5},
	//		parents:                []*model.DomainHash{{2}, {3}, {4}},
	//		expectedBlueScore:      6,
	//		expectedSelectedParent: &model.DomainHash{2},
	//		expectedMergeSetBlues:  []*model.DomainHash{{2}, {3}, {4}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//	{ /* 6 */
	//		hash:                   &model.DomainHash{6},
	//		parents:                []*model.DomainHash{{5}},
	//		expectedBlueScore:      7,
	//		expectedSelectedParent: &model.DomainHash{5},
	//		expectedMergeSetBlues:  []*model.DomainHash{{5}},
	//		expectedMergeSetReds:   []*model.DomainHash{},
	//	},
	//}

	//test 6: mergeSetReds is not empty, one of the block in the mergeSet is not connected to more than k . K = 1
	dag6 := []testGhostdagData{
		{
			hash:                   &model.DomainHash{1},
			parents:                []*model.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*model.DomainHash{genesisHash},
			expectedMergeSetReds:   []*model.DomainHash{},
		},
		{
			hash:                   &model.DomainHash{2},
			parents:                []*model.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &model.DomainHash{1},
			expectedMergeSetBlues:  []*model.DomainHash{{1}},
			expectedMergeSetReds:   []*model.DomainHash{},
		},
		{
			hash:                   &model.DomainHash{3},
			parents:                []*model.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*model.DomainHash{genesisHash},
			expectedMergeSetReds:   []*model.DomainHash{},
		},
		{
			hash:                   &model.DomainHash{4},
			parents:                []*model.DomainHash{{2}, {3}},
			expectedBlueScore:      4,
			expectedSelectedParent: &model.DomainHash{2},
			expectedMergeSetBlues:  []*model.DomainHash{{2}},
			expectedMergeSetReds:   []*model.DomainHash{{3}},
		},
	}

	g := ghostdag2.New(nil, dagTopology, ghostdagDataStore, 1)
	for i, testBlockData := range dag6 {
		dagTopology.parentsMap[*testBlockData.hash] = testBlockData.parents
		ghostdagData, err := g.GHOSTDAG(testBlockData.parents)
		if err != nil {
			t.Fatalf("test #%d failed: GHOSTDAG error: %s", i, err)
		}

		if testBlockData.expectedBlueScore != ghostdagData.BlueScore {
			t.Fatalf("test #%d failed: expected blue score %d but got %d", i, testBlockData.expectedBlueScore, ghostdagData.BlueScore)
		}

		if *testBlockData.expectedSelectedParent != *ghostdagData.SelectedParent {
			t.Fatalf("test #%d failed: expected selected parent %v but got %v", i, testBlockData.expectedSelectedParent, ghostdagData.SelectedParent)
		}

		if !reflect.DeepEqual(testBlockData.expectedMergeSetBlues, ghostdagData.MergeSetBlues) {
			t.Fatalf("test #%d failed: expected merge set blues %v but got %v", i, testBlockData.expectedMergeSetBlues, ghostdagData.MergeSetBlues)
		}

		if !reflect.DeepEqual(testBlockData.expectedMergeSetReds, ghostdagData.MergeSetReds) {
			t.Fatalf("test #%d failed: expected merge set reds %v but got %v", i, testBlockData.expectedMergeSetReds, ghostdagData.MergeSetReds)
		}

		err = ghostdagDataStore.Insert(nil, testBlockData.hash, ghostdagData)
		if err != nil {
			t.Fatalf("test #%d failed: Insert error: %s", i, err)
		}
	}
}

type GHOSTDAGDataStoreImpl struct {
	dagMap map[model.DomainHash]*model.BlockGHOSTDAGData
}

func (ds *GHOSTDAGDataStoreImpl) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) error {
	ds.dagMap[*blockHash] = blockGHOSTDAGData
	return nil
}
func (ds *GHOSTDAGDataStoreImpl) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) (*model.BlockGHOSTDAGData, error) {
	v, ok := ds.dagMap[*blockHash]
	if ok {
		return v, nil
	}
	return nil, nil
}

//candidateBluesAnticoneSizes = make(map[model.DomainHash]model.KType, gm.k)
type DAGTopologyManagerImpl struct {
	//dagMap map[*model.DomainHash] *model.BlockGHOSTDAGData
	parentsMap map[model.DomainHash][]*model.DomainHash
}

//Implemented//
func (dt *DAGTopologyManagerImpl) Parents(blockHash *model.DomainHash) ([]*model.DomainHash, error) {
	v, ok := dt.parentsMap[*blockHash]
	if !ok {
		return make([]*model.DomainHash, 0), nil
	} else {
		return v, nil
	}
}

func (dt *DAGTopologyManagerImpl) Children(blockHash *model.DomainHash) ([]*model.DomainHash, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) IsParentOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) IsChildOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) (bool, error) {
	panic("unimplemented")
}

//Implemented//
func (dt *DAGTopologyManagerImpl) IsAncestorOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) (bool, error) {
	bParents, ok := dt.parentsMap[*blockHashB]
	if !ok {
		return false, nil
	}
	for _, parent := range bParents {
		if *parent == *blockHashA {
			return true, nil
		}
	}
	for _, y := range bParents {
		isAnc, err := dt.IsAncestorOf(blockHashA, y)
		if err != nil {
			return false, err
		}
		if isAnc {
			return true, nil
		}
	}
	return false, nil

}

func (dt *DAGTopologyManagerImpl) IsDescendantOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (gh *DAGTopologyManagerImpl) IsAncestorOfAny(blockHash *model.DomainHash, potentialDescendants []*model.DomainHash) (bool, error) {
	panic("unimplemented")
}
func (gh *DAGTopologyManagerImpl) IsInSelectedParentChainOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) (bool, error) {
	panic("unimplemented")
}
