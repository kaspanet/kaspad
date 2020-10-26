package processes

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"testing"
)

/*----------------contains-------------------------- */
func contains(s *externalapi.DomainHash, g []*externalapi.DomainHash) bool {
	for _, r := range g {
		if *r == *s {
			return true
		}
	}
	return false
}

func DeepEqualHashArrays(runtime, expected []*externalapi.DomainHash) bool {
	if len(runtime) != len(expected) {
		return false
	}
	for _, hash := range runtime {
		if !contains(hash, expected) {
			return false
		}
	}
	return true
}

func TestGHOSTDA(t *testing.T) {
	//t.Errorf("helo") //keep running
	//if false {
	//	t.Fatalf("The test failed") // string - //stop
	//}

	type testGhostdagData struct {
		hash                   *externalapi.DomainHash
		parents                []*externalapi.DomainHash
		expectedBlueScore      uint64
		expectedSelectedParent *externalapi.DomainHash
		expectedMergeSetBlues  []*externalapi.DomainHash
		expectedMergeSetReds   []*externalapi.DomainHash
	}

	type isolatedTest struct {
		k        model.KType
		subTests []testGhostdagData
	}

	genesisHash := &externalapi.DomainHash{}
	dagTopology := &DAGTopologyManagerImpl{
		parentsMap: make(map[externalapi.DomainHash][]*externalapi.DomainHash),
	}
	dagTopology.parentsMap[*genesisHash] = nil

	ghostdagDataStore := &GHOSTDAGDataStoreImpl{
		dagMap: make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData),
	}
	ghostdagDataStore.dagMap[*genesisHash] = &model.BlockGHOSTDAGData{
		BlueScore:          1,
		SelectedParent:     nil,
		MergeSetBlues:      nil,
		MergeSetReds:       nil,
		BluesAnticoneSizes: nil,
	}

	// ****************************** TESTS ****************************** //
	// test1: Graph form is a chain. K = 0
	dag1 := isolatedTest{
		k: 0,
		subTests: []testGhostdagData{
			{
				hash:                   &externalapi.DomainHash{1},
				parents:                []*externalapi.DomainHash{genesisHash},
				expectedBlueScore:      1,
				expectedSelectedParent: genesisHash,
				expectedMergeSetBlues:  []*externalapi.DomainHash{genesisHash},
				expectedMergeSetReds:   []*externalapi.DomainHash{},
			},
			{
				hash:                   &externalapi.DomainHash{2},
				parents:                []*externalapi.DomainHash{{1}},
				expectedBlueScore:      1,
				expectedSelectedParent: &externalapi.DomainHash{1},
				expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
				expectedMergeSetReds:   []*externalapi.DomainHash{},
			},
			{
				hash:                   &externalapi.DomainHash{3},
				parents:                []*externalapi.DomainHash{{2}},
				expectedBlueScore:      1,
				expectedSelectedParent: &externalapi.DomainHash{2},
				expectedMergeSetBlues:  []*externalapi.DomainHash{{2}},
				expectedMergeSetReds:   []*externalapi.DomainHash{},
			},
		},
	}

	//test2 : The graph’s longest chain was created by malicious miners (not the “heaviest”). K = 3
	dag2 := isolatedTest{k: 3, subTests: []testGhostdagData{
		{ /* 1*/
			hash:                   &externalapi.DomainHash{1},
			parents:                []*externalapi.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*externalapi.DomainHash{genesisHash},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 2 */
			hash:                   &externalapi.DomainHash{2},
			parents:                []*externalapi.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{1},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 3 */
			hash:                   &externalapi.DomainHash{3},
			parents:                []*externalapi.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{1},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 4 */
			hash:                   &externalapi.DomainHash{4},
			parents:                []*externalapi.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{1},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 5 */
			hash:                   &externalapi.DomainHash{5},
			parents:                []*externalapi.DomainHash{{2}, {3}, {4}},
			expectedBlueScore:      6,
			expectedSelectedParent: &externalapi.DomainHash{2},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{2}, {3}, {4}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 6 */
			hash:                   &externalapi.DomainHash{6},
			parents:                []*externalapi.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*externalapi.DomainHash{genesisHash},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 7 */
			hash:                   &externalapi.DomainHash{7},
			parents:                []*externalapi.DomainHash{{6}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{6},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{6}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 8 */
			hash:                   &externalapi.DomainHash{8},
			parents:                []*externalapi.DomainHash{{7}},
			expectedBlueScore:      4,
			expectedSelectedParent: &externalapi.DomainHash{7},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{7}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 9 */
			hash:                   &externalapi.DomainHash{9},
			parents:                []*externalapi.DomainHash{{8}},
			expectedBlueScore:      5,
			expectedSelectedParent: &externalapi.DomainHash{8},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{8}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 10 */
			hash:                   &externalapi.DomainHash{10},
			parents:                []*externalapi.DomainHash{{5}, {9}},
			expectedBlueScore:      7,
			expectedSelectedParent: &externalapi.DomainHash{5},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{5}},
			expectedMergeSetReds:   []*externalapi.DomainHash{{9}, {8}, {7}, {6}},
		},
	}}

	// test3: Selected Parent choice: same score – decide by hashes. K = 3
	dag3 := isolatedTest{k: 3, subTests: []testGhostdagData{
		{ /* 1*/
			hash:                   &externalapi.DomainHash{1},
			parents:                []*externalapi.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*externalapi.DomainHash{genesisHash},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 2 */
			hash:                   &externalapi.DomainHash{2},
			parents:                []*externalapi.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{1},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 3 */
			hash:                   &externalapi.DomainHash{3},
			parents:                []*externalapi.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{1},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 4 */
			hash:                   &externalapi.DomainHash{4},
			parents:                []*externalapi.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{1},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 5 */
			hash:                   &externalapi.DomainHash{5},
			parents:                []*externalapi.DomainHash{{2}, {3}, {4}},
			expectedBlueScore:      6,
			expectedSelectedParent: &externalapi.DomainHash{2},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{2}, {3}, {4}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{ /* 6 */
			hash:                   &externalapi.DomainHash{6},
			parents:                []*externalapi.DomainHash{{5}},
			expectedBlueScore:      7,
			expectedSelectedParent: &externalapi.DomainHash{5},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{5}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
	}}

	//test 4: mergeSetReds is not empty, one of the block in the mergeSet is not connected to more than k . K = 1
	dag4 := isolatedTest{k: 1, subTests: []testGhostdagData{
		{
			hash:                   &externalapi.DomainHash{1},
			parents:                []*externalapi.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*externalapi.DomainHash{genesisHash},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{2},
			parents:                []*externalapi.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{1},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{3},
			parents:                []*externalapi.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*externalapi.DomainHash{genesisHash},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{4},
			parents:                []*externalapi.DomainHash{{2}, {3}},
			expectedBlueScore:      4,
			expectedSelectedParent: &externalapi.DomainHash{2},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{2}},
			expectedMergeSetReds:   []*externalapi.DomainHash{{3}},
		},
	}}

	//test 5: Adding a block to the mergeSet will destroy one of the blue block K-cluster.(the block is keeping K-cluster )
	dag5 := isolatedTest{k: 2, subTests: []testGhostdagData{
		{
			hash:                   &externalapi.DomainHash{1},
			parents:                []*externalapi.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*externalapi.DomainHash{genesisHash},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{2},
			parents:                []*externalapi.DomainHash{genesisHash},
			expectedBlueScore:      2,
			expectedSelectedParent: genesisHash,
			expectedMergeSetBlues:  []*externalapi.DomainHash{genesisHash},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{3},
			parents:                []*externalapi.DomainHash{{1}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{1},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{1}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{4},
			parents:                []*externalapi.DomainHash{{2}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{2},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{2}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{5},
			parents:                []*externalapi.DomainHash{{2}},
			expectedBlueScore:      3,
			expectedSelectedParent: &externalapi.DomainHash{2},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{2}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{6},
			parents:                []*externalapi.DomainHash{{3}, {5}},
			expectedBlueScore:      6,
			expectedSelectedParent: &externalapi.DomainHash{3},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{2}, {3}, {5}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{7},
			parents:                []*externalapi.DomainHash{{3}, {4}},
			expectedBlueScore:      6,
			expectedSelectedParent: &externalapi.DomainHash{3},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{3}, {2}, {4}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{8},
			parents:                []*externalapi.DomainHash{{5}},
			expectedBlueScore:      4,
			expectedSelectedParent: &externalapi.DomainHash{5},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{5}},
			expectedMergeSetReds:   []*externalapi.DomainHash{},
		},
		{
			hash:                   &externalapi.DomainHash{9},
			parents:                []*externalapi.DomainHash{{6}, {7}, {8}},
			expectedBlueScore:      7,
			expectedSelectedParent: &externalapi.DomainHash{6},
			expectedMergeSetBlues:  []*externalapi.DomainHash{{6}},
			expectedMergeSetReds:   []*externalapi.DomainHash{{4}, {8}, {7}},
		},
	}}

	//testsArr := []*isolatedTest{&dag1, &dag2, &dag3, &dag4, &dag5}
	//for testIndex, testInfo := range testsArr {
	//	g := ghostdag2.New(nil, dagTopology, ghostdagDataStore, (testInfo.k))
	//	for i, testBlockData := range testInfo.subTests {
	//		dagTopology.parentsMap[*testBlockData.hash] = testBlockData.parents
	//		err := g.GHOSTDAG(testBlockData.hash)
	//		if err != nil {
	//			t.Fatalf("test #%d failed: GHOSTDAG error: %s", i, err)
	//		}
	//		ghostdagData, err := ghostdagDataStore.Get(nil,  testBlockData.hash)
	//		if err != nil{
	//			t.Fatalf("test #%d failed: ghostdagDataStore error: %s", i, err)
	//		}
	//		if testBlockData.expectedBlueScore != ghostdagData.BlueScore {
	//			t.Fatalf("test #%d failed: expected blue score %d but got %d", i, testBlockData.expectedBlueScore, ghostdagData.BlueScore)
	//		}
	//
	//		if *testBlockData.expectedSelectedParent != *ghostdagData.SelectedParent {
	//			t.Fatalf("test #%d failed: expected selected parent %v but got %v", i, testBlockData.expectedSelectedParent, ghostdagData.SelectedParent)
	//		}
	//
	//		if !DeepEqualHashArrays(testBlockData.expectedMergeSetBlues, ghostdagData.MergeSetBlues) {
	//			t.Fatalf("test #%d failed: expected merge set blues %v but got %v", i, testBlockData.expectedMergeSetBlues, ghostdagData.MergeSetBlues)
	//		}
	//
	//		if !DeepEqualHashArrays(testBlockData.expectedMergeSetReds, ghostdagData.MergeSetReds) {
	//			t.Fatalf("test #%d failed: expected merge set reds %v but got %v", i, testBlockData.expectedMergeSetReds, ghostdagData.MergeSetReds)
	//		}
	//
	//		err = ghostdagDataStore.Insert(nil, testBlockData.hash, ghostdagData)
	//		if err != nil {
	//			t.Fatalf("test #%d failed: Insert error: %s", i, err)
	//		}
	//	}
	//	fmt.Printf("test %d finished \n", testIndex)
	//
	//	dagTopology := &DAGTopologyManagerImpl{
	//		parentsMap: make(map[externalapi.DomainHash][]*externalapi.DomainHash),
	//	}
	//	dagTopology.parentsMap[*genesisHash] = nil
	//
	//	ghostdagDataStore := &GHOSTDAGDataStoreImpl{
	//		dagMap: make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData),
	//	}
	//	ghostdagDataStore.dagMap[*genesisHash] = &model.BlockGHOSTDAGData{
	//		BlueScore:          1,
	//		SelectedParent:     nil,
	//		MergeSetBlues:      nil,
	//		MergeSetReds:       nil,
	//		BluesAnticoneSizes: nil,
	//	}
	//}

	testsArr := []*isolatedTest{&dag1, &dag2, &dag3, &dag4, &dag5}
	//testsArr := []*isolatedTest{&dag5, &dag1, &dag2, &dag3, &dag4 }
	for testIndex, testInfo := range testsArr {
		g := ghostdagmanager.New(nil, dagTopology, ghostdagDataStore, testInfo.k)
		for i, testBlockData := range testInfo.subTests {
			dagTopology.parentsMap[*testBlockData.hash] = testBlockData.parents
			err := g.GHOSTDAG(testBlockData.hash)
			if err != nil {
				t.Fatalf("test #%d failed: GHOSTDAG error: %s", i, err)
			}
			ghostdagData, err := ghostdagDataStore.Get(nil, testBlockData.hash)
			if err != nil {
				t.Fatalf("test #%d failed: ghostdagDataStore error: %s", i, err)
			}
			if testBlockData.expectedBlueScore != (ghostdagData.BlueScore) {
				t.Fatalf("test #%d failed: expected blue score %d but got %d", i, testBlockData.expectedBlueScore, ghostdagData.BlueScore)
			}

			if *testBlockData.expectedSelectedParent != *ghostdagData.SelectedParent {
				t.Fatalf("test #%d failed: expected selected parent %v but got %v", i, testBlockData.expectedSelectedParent, ghostdagData.SelectedParent)
			}

			if !DeepEqualHashArrays(testBlockData.expectedMergeSetBlues, ghostdagData.MergeSetBlues) {
				t.Fatalf("test #%d failed: expected merge set blues %v but got %v", i, testBlockData.expectedMergeSetBlues, ghostdagData.MergeSetBlues)
			}

			if !DeepEqualHashArrays(testBlockData.expectedMergeSetReds, ghostdagData.MergeSetReds) {
				t.Fatalf("test #%d failed: expected merge set reds %v but got %v", i, testBlockData.expectedMergeSetReds, ghostdagData.MergeSetReds)
			}

		}
		fmt.Printf("test %d finished \n", testIndex)

		dagTopology := &DAGTopologyManagerImpl{
			parentsMap: make(map[externalapi.DomainHash][]*externalapi.DomainHash),
		}
		dagTopology.parentsMap[*genesisHash] = nil

		ghostdagDataStore := &GHOSTDAGDataStoreImpl{
			dagMap: make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData),
		}
		ghostdagDataStore.dagMap[*genesisHash] = &model.BlockGHOSTDAGData{
			BlueScore:          1,
			SelectedParent:     nil,
			MergeSetBlues:      nil,
			MergeSetReds:       nil,
			BluesAnticoneSizes: nil,
		}
	}

}

type GHOSTDAGDataStoreImpl struct {
	dagMap map[externalapi.DomainHash]*model.BlockGHOSTDAGData
}

func (ds *GHOSTDAGDataStoreImpl) Stage(blockHash *externalapi.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {
	ds.dagMap[*blockHash] = blockGHOSTDAGData
	return
}

func (ds *GHOSTDAGDataStoreImpl) IsStaged() bool {
	panic("implement me")
}

func (ds *GHOSTDAGDataStoreImpl) Discard() {
	panic("implement me")
}

func (ds *GHOSTDAGDataStoreImpl) Commit(dbTx model.DBTxProxy) error {
	panic("implement me")
}

//func (ds *GHOSTDAGDataStoreImpl) Insert(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) error {
//	ds.dagMap[*blockHash] = blockGHOSTDAGData
//	return nil
//}
func (ds *GHOSTDAGDataStoreImpl) Get(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
	v, ok := ds.dagMap[*blockHash]
	if ok {
		return v, nil
	}
	return nil, nil
}

//candidateBluesAnticoneSizes = make(map[externalapi.DomainHash]model.KType, gm.k)
type DAGTopologyManagerImpl struct {
	//dagMap map[*externalapi.DomainHash] *model.BlockGHOSTDAGData
	parentsMap map[externalapi.DomainHash][]*externalapi.DomainHash
}

func (dt *DAGTopologyManagerImpl) Tips() ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (dt *DAGTopologyManagerImpl) AddTip(tipHash *externalapi.DomainHash) error {
	panic("implement me")
}

//Implemented//
func (dt *DAGTopologyManagerImpl) Parents(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	v, ok := dt.parentsMap[*blockHash]
	if !ok {
		return make([]*externalapi.DomainHash, 0), nil
	} else {
		return v, nil
	}
}

func (dt *DAGTopologyManagerImpl) Children(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) IsParentOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) IsChildOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}

//Implemented//
func (dt *DAGTopologyManagerImpl) IsAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
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

func (dt *DAGTopologyManagerImpl) IsDescendantOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (gh *DAGTopologyManagerImpl) IsAncestorOfAny(blockHash *externalapi.DomainHash, potentialDescendants []*externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}
func (gh *DAGTopologyManagerImpl) IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}
