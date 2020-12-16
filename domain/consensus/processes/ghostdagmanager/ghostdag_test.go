package ghostdagmanager_test

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdag2"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

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
	K                    model.KType `json:"K"`
	GenesisID            string      `json:"GenesisID"`
	ExpectedMergeSetReds []string    `json:"ExpectedReds"`
	Blocks               []block     `json:"Blocks"`
}

type implManager struct {
	function func(
		databaseContext model.DBReader,
		dagTopologyManager model.DAGTopologyManager,
		ghostdagDataStore model.GHOSTDAGDataStore,
		headerStore model.BlockHeaderStore,
		k model.KType) model.GHOSTDAGManager
	implName string
}

// TestGHOSTDAG iterates over several dag simulations, and checks
// that the blue score, blue set and selected parent of each
// block are calculated as expected.
func TestGHOSTDAG(t *testing.T) {
	//NOTE: FOR ADDING/REMOVING AN IMPLEMENTATION CHANGE BELOW:
	implementationFactories := []implManager{
		{ghostdagmanager.New, "Original"},
		{ghostdag2.New, "Tal's impl"},
	}
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		dagTopology := &DAGTopologyManagerImpl{
			parentsMap: make(map[externalapi.DomainHash][]*externalapi.DomainHash),
		}

		ghostdagDataStore := &GHOSTDAGDataStoreImpl{
			dagMap: make(map[externalapi.DomainHash]model.BlockGHOSTDAGData),
		}

		blockHeadersStore := &blockHeadersStore{
			dagMap: make(map[externalapi.DomainHash]*externalapi.DomainBlockHeader),
		}

		blockGHOSTDAGDataGenesis := ghostdagmanager.NewBlockGHOSTDAGData(0, new(big.Int), nil, nil, nil, nil)
		genesisHeader := params.GenesisBlock.Header
		genesisWork := util.CalcWork(genesisHeader.Bits)

		var testsCounter int
		err := filepath.Walk("../../testdata/dags", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			jsonFile, err := os.Open(path)
			if err != nil {
				t.Fatalf("TestGHOSTDAG : failed opening the json file %s: %v", info.Name(), err)
			}
			defer jsonFile.Close()
			var test testDag
			decoder := json.NewDecoder(jsonFile)
			decoder.DisallowUnknownFields()
			err = decoder.Decode(&test)
			if err != nil {
				t.Fatalf("TestGHOSTDAG:failed decoding json: %v", err)
			}
			params.K = test.K

			var genesisHash externalapi.DomainHash
			copy(genesisHash[:], test.GenesisID)

			dagTopology.parentsMap[genesisHash] = nil

			ghostdagDataStore.dagMap[genesisHash] = blockGHOSTDAGDataGenesis
			blockHeadersStore.dagMap[genesisHash] = genesisHeader

			for _, factory := range implementationFactories {

				g := factory.function(nil, dagTopology, ghostdagDataStore, blockHeadersStore, test.K)
				for _, testBlockData := range test.Blocks {

					blockID := StringToByte(testBlockData.ID)
					dagTopology.parentsMap[*blockID] = StringToByteArray(testBlockData.Parents)
					blockHeadersStore.dagMap[*blockID] = &externalapi.DomainBlockHeader{
						ParentHashes: StringToByteArray(testBlockData.Parents),
						Bits:         genesisHeader.Bits,
					}

					err := g.GHOSTDAG(blockID)
					if err != nil {
						t.Fatalf("Test failed: \n Impl: %s,FileName: %s \n error on GHOSTDAG - block %s: %s.",
							factory.implName, info.Name(), testBlockData.ID, err)
					}
					ghostdagData, err := ghostdagDataStore.Get(nil, blockID)
					if err != nil {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: ghostdagDataStore error: %v.",
							factory.implName, info.Name(), testBlockData.ID, err)
					}

					// because the difficulty is constant and equal to genesis the work should be blueScore*genesisWork.
					expectedWork := new(big.Int).Mul(genesisWork, new(big.Int).SetUint64(testBlockData.Score))
					if expectedWork.Cmp(ghostdagData.BlueWork()) != 0 {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: expected blue work %d but got %d.",
							factory.implName, info.Name(), testBlockData.ID, expectedWork, ghostdagData.BlueWork())
					}
					if testBlockData.Score != (ghostdagData.BlueScore()) {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: expected blue score %d but got %d.",
							factory.implName, info.Name(), testBlockData.ID, testBlockData.Score, ghostdagData.BlueScore())
					}

					if *StringToByte(testBlockData.SelectedParent) != *ghostdagData.SelectedParent() {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: expected selected parent %v but got %v.",
							factory.implName, info.Name(), testBlockData.ID, testBlockData.SelectedParent, string(ghostdagData.SelectedParent()[:]))
					}

					if !reflect.DeepEqual(StringToByteArray(testBlockData.MergeSetBlues), ghostdagData.MergeSetBlues()) {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: expected merge set blues %v but got %v.",
							factory.implName, info.Name(), testBlockData.ID, testBlockData.MergeSetBlues, hashesToStrings(ghostdagData.MergeSetBlues()))
					}

					if !reflect.DeepEqual(StringToByteArray(testBlockData.MergeSetReds), ghostdagData.MergeSetReds()) {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: expected merge set reds %v but got %v.",
							factory.implName, info.Name(), testBlockData.ID, testBlockData.MergeSetReds, hashesToStrings(ghostdagData.MergeSetReds()))
					}
				}
				dagTopology.parentsMap = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
				dagTopology.parentsMap[genesisHash] = nil
				ghostdagDataStore.dagMap = make(map[externalapi.DomainHash]model.BlockGHOSTDAGData)
				ghostdagDataStore.dagMap[genesisHash] = blockGHOSTDAGDataGenesis
				blockHeadersStore.dagMap = make(map[externalapi.DomainHash]*externalapi.DomainBlockHeader)
				blockHeadersStore.dagMap[genesisHash] = genesisHeader
			}

			testsCounter++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if testsCounter != 3 {
			t.Fatalf("Expected 3 test files, ran %d instead", testsCounter)
		}
	})
}

func hashesToStrings(arr []*externalapi.DomainHash) []string {
	var strArr = make([]string, len(arr))
	for i, hash := range arr {
		strArr[i] = string(hash[:])
	}
	return strArr
}

func StringToByte(strID string) *externalapi.DomainHash {
	var domainHash externalapi.DomainHash
	copy(domainHash[:], strID)
	return &domainHash
}

func StringToByteArray(stringIDArr []string) []*externalapi.DomainHash {
	domainHashArr := make([]*externalapi.DomainHash, len(stringIDArr))
	for i, strID := range stringIDArr {
		domainHashArr[i] = StringToByte(strID)
	}
	return domainHashArr
}

/* ---------------------- */
type GHOSTDAGDataStoreImpl struct {
	dagMap map[externalapi.DomainHash]model.BlockGHOSTDAGData
}

func (ds *GHOSTDAGDataStoreImpl) Stage(blockHash *externalapi.DomainHash, blockGHOSTDAGData model.BlockGHOSTDAGData) {
	ds.dagMap[*blockHash] = blockGHOSTDAGData
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

func (ds *GHOSTDAGDataStoreImpl) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (model.BlockGHOSTDAGData, error) {
	v, ok := ds.dagMap[*blockHash]
	if ok {
		return v, nil
	}
	return nil, nil
}

type DAGTopologyManagerImpl struct {
	parentsMap map[externalapi.DomainHash][]*externalapi.DomainHash
}

func (dt *DAGTopologyManagerImpl) ChildInSelectedParentChainOf(context, highHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	panic("implement me")
}

func (dt *DAGTopologyManagerImpl) Tips() ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (dt *DAGTopologyManagerImpl) AddTip(tipHash *externalapi.DomainHash) error {
	panic("implement me")
}

func (dt *DAGTopologyManagerImpl) Parents(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	v, ok := dt.parentsMap[*blockHash]
	if !ok {
		return []*externalapi.DomainHash{}, nil
	}

	return v, nil
}

func (dt *DAGTopologyManagerImpl) Children(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) IsParentOf(hashBlockA *externalapi.DomainHash, hashBlockB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) IsChildOf(hashBlockA *externalapi.DomainHash, hashBlockB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) IsAncestorOf(hashBlockA *externalapi.DomainHash, hashBlockB *externalapi.DomainHash) (bool, error) {
	blockBParents, isOk := dt.parentsMap[*hashBlockB]
	if !isOk {
		return false, nil
	}

	for _, parentOfB := range blockBParents {
		if *parentOfB == *hashBlockA {
			return true, nil
		}
	}

	for _, parentOfB := range blockBParents {
		isAncestorOf, err := dt.IsAncestorOf(hashBlockA, parentOfB)
		if err != nil {
			return false, err
		}
		if isAncestorOf {
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

type blockHeadersStore struct {
	dagMap map[externalapi.DomainHash]*externalapi.DomainBlockHeader
}

func (b *blockHeadersStore) Discard() { panic("unimplemented") }

func (b *blockHeadersStore) Commit(_ model.DBTransaction) error { panic("unimplemented") }

func (b *blockHeadersStore) Stage(blockHash *externalapi.DomainHash, blockHeader *externalapi.DomainBlockHeader) {
	b.dagMap[*blockHash] = blockHeader
}

func (b *blockHeadersStore) IsStaged() bool { panic("unimplemented") }

func (b *blockHeadersStore) BlockHeader(_ model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error) {
	header, ok := b.dagMap[*blockHash]
	if ok {
		return header, nil
	}
	return nil, errors.New("Header isn't in the store")
}

func (b *blockHeadersStore) HasBlockHeader(_ model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	_, ok := b.dagMap[*blockHash]
	return ok, nil
}

func (b *blockHeadersStore) BlockHeaders(_ model.DBReader, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlockHeader, error) {
	res := make([]*externalapi.DomainBlockHeader, 0, len(blockHashes))
	for _, hash := range blockHashes {
		header, err := b.BlockHeader(nil, hash)
		if err != nil {
			return nil, err
		}
		res = append(res, header)
	}
	return res, nil
}

func (b *blockHeadersStore) Delete(blockHash *externalapi.DomainHash) {
	delete(b.dagMap, *blockHash)
}

func (b *blockHeadersStore) Count() uint64 {
	return uint64(len(b.dagMap))
}
