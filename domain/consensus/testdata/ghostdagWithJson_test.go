package testdata

import (
	"encoding/json"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdag2"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"os"
	"testing"
)

// TestGHOSTDAG iterates over several dag simulations, and checks
// that the blue score, blue set and selected parent of each
// block are calculated as expected.
func TestGHOSTDAG(t *testing.T) {

	type block struct {
		ID             string   `json:"ID"`
		Score          uint64   `json:"ExpectedScore"`
		SelectedParent string   `json:"ExpectedSelectedParent"`
		MergeSetReds   []string `json:"ExpectedReds"`
		MergeSetBlues  []string `json:"ExpectedBlues"`
		Parents        []string `json:"Parents"`
	}

	// json struct:
	type testDag struct {
		K                    dagconfig.KType `json:"K"`
		GenesisID            string          `json:"GenesisID"`
		ExpectedMergeSetReds []string        `json:"ExpectedReds"`
		Blocks               []block         `json:"Blocks"`
	}

	type implManager struct {
		function func(
			databaseContext model.DBReader,
			dagTopologyManager model.DAGTopologyManager,
			ghostdagDataStore model.GHOSTDAGDataStore,
			k model.KType) model.GHOSTDAGManager
		implName string
	}

	type testGhostdagData struct {
		hash                   *externalapi.DomainHash
		parents                []*externalapi.DomainHash
		expectedBlueScore      uint64
		expectedSelectedParent *externalapi.DomainHash
		expectedMergeSetBlues  []*externalapi.DomainHash
		expectedMergeSetReds   []*externalapi.DomainHash
	}

	dagTopology := &DAGTopologyManagerImpl{
		parentsMap: make(map[externalapi.DomainHash][]*externalapi.DomainHash),
	}

	ghostdagDataStore := &GHOSTDAGDataStoreImpl{
		dagMap: make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData),
	}

	dagArray := []string{"./dags/dag0.json"}
	jsonFile, err := os.Open(dagArray[0])
	if err != nil {
		t.Fatalf("failed opening the json file %s: %v", dagArray[0], err)
	}
	defer jsonFile.Close()
	var test testDag
	decoder := json.NewDecoder(jsonFile)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&test)
	if err != nil {
		t.Fatalf("TestGHOSTDAG:failed decoding json: %v", err)
	}

	var genesisId [32]byte //change genesis ID from type string to type []byte
	copy(genesisId[:], test.GenesisID)
	genesisHash := &externalapi.DomainHash{0}
	*genesisHash = genesisId

	dagTopology.parentsMap[*genesisHash] = nil
	ghostdagDataStore.dagMap[*genesisHash] = &model.BlockGHOSTDAGData{
		BlueScore:          0,
		SelectedParent:     nil,
		MergeSetBlues:      nil,
		MergeSetReds:       nil,
		BluesAnticoneSizes: nil,
	}

	//NOTE: FOR ADDING/REMOVING AN IMPLEMENTATION CHANGE BELOW:
	implementationFactories := []implManager{
		{ghostdagmanager.New, "Original"},
		{ghostdag2.New, "Tal's impl"},
	}

	for _, factory := range implementationFactories {
		fmt.Printf("____________________________\n")
		//	for testNum, test := range testsArr {
		testNum := 0
		g := factory.function(nil, dagTopology, ghostdagDataStore, model.KType(test.K))
		fmt.Printf("Impl:%s,  TestNum:%d\n", factory.implName, testNum+1)
		for _, testBlockData := range test.Blocks {
			testNum++
			blockID := infoStrToByte(testBlockData.ID)
			dagTopology.parentsMap[*blockID] = infoStrToByteArray(testBlockData.Parents)
			err := g.GHOSTDAG(blockID)
			if err != nil {
				t.Fatalf("Test #%d failed:on  GHOSTDAG error: %s.", testNum+1, err)
			}
			ghostdagData, err := ghostdagDataStore.Get(nil, blockID)
			if err != nil {
				t.Fatalf("Test #%d failed: ghostdagDataStore error: %s.", testNum+1, err)
			}
			if testBlockData.Score != (ghostdagData.BlueScore) {
				t.Fatalf("Test #%d failed: expected blue score %d but got %d.", testNum+1, testBlockData.Score, ghostdagData.BlueScore)
			}

			if *infoStrToByte(testBlockData.SelectedParent) != *ghostdagData.SelectedParent {
				t.Fatalf("Test #%d failed: expected selected parent %v but got %v.", testNum+1, *infoStrToByte(testBlockData.SelectedParent), ghostdagData.SelectedParent)
			}

			if !DeepEqualHashArrays(infoStrToByteArray(testBlockData.MergeSetBlues), ghostdagData.MergeSetBlues) {
				t.Fatalf("Test #%d failed: expected merge set blues %v but got %v.", testNum+1, infoStrToByteArray(testBlockData.MergeSetBlues), ghostdagData.MergeSetBlues)
			}

			if !DeepEqualHashArrays(infoStrToByteArray(testBlockData.MergeSetReds), ghostdagData.MergeSetReds) {
				t.Fatalf("Test #%d failed: expected merge set reds %v but got %v.", testNum+1, infoStrToByteArray(testBlockData.MergeSetReds), ghostdagData.MergeSetReds)
			}

		}
		fmt.Printf("    Test success!\n\n")

		//fmt.Printf("Test %d successfully finished. \n", testStruct.testNum+1)

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
		//}
	}

}

func infoStrToByte(strID string) *externalapi.DomainHash {
	var domainHash externalapi.DomainHash
	var genesisId [32]byte //change genesis ID from type string to type []byte
	copy(genesisId[:], strID)
	//genesisHash := &externalapi.DomainHash{0}
	domainHash = genesisId
	return &domainHash
}

func infoStrToByteArray(strIDArr []string) []*externalapi.DomainHash {
	var domainHashArr []*externalapi.DomainHash
	for _, strID := range strIDArr {
		domainHashArr = append(domainHashArr, infoStrToByte(strID))
	}
	return domainHashArr
}

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

/* ---------------------- */
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

func (ds *GHOSTDAGDataStoreImpl) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

//func (ds *GHOSTDAGDataStoreImpl) Insert(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) error {
//	ds.dagMap[*blockHash] = blockGHOSTDAGData
//	return nil
//}
func (ds *GHOSTDAGDataStoreImpl) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
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
	}

	return v, nil
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

func (dt *DAGTopologyManagerImpl) IsAncestorOfAny(blockHash *externalapi.DomainHash, potentialDescendants []*externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}
func (dt *DAGTopologyManagerImpl) IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) SetParents(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) error {
	panic("unimplemented")
}
