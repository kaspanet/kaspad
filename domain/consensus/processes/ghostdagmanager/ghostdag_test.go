package ghostdagmanager_test

import (
	"encoding/json"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/util/difficulty"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdag2"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
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
			dagMap: make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData),
		}

		blockHeadersStore := &blockHeadersStore{
			dagMap: make(map[externalapi.DomainHash]externalapi.BlockHeader),
		}

		blockGHOSTDAGDataGenesis := model.NewBlockGHOSTDAGData(0, new(big.Int), nil, nil, nil, nil)
		genesisHeader := params.GenesisBlock.Header
		genesisWork := difficulty.CalcWork(genesisHeader.Bits())

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

			genesisHash := *StringToDomainHash(test.GenesisID)

			dagTopology.parentsMap[genesisHash] = nil

			ghostdagDataStore.dagMap[genesisHash] = blockGHOSTDAGDataGenesis
			blockHeadersStore.dagMap[genesisHash] = genesisHeader

			for _, factory := range implementationFactories {

				g := factory.function(nil, dagTopology, ghostdagDataStore, blockHeadersStore, test.K)
				for _, testBlockData := range test.Blocks {

					blockID := StringToDomainHash(testBlockData.ID)
					dagTopology.parentsMap[*blockID] = StringToDomainHashSlice(testBlockData.Parents)
					blockHeadersStore.dagMap[*blockID] = blockheader.NewImmutableBlockHeader(
						constants.MaxBlockVersion,
						StringToDomainHashSlice(testBlockData.Parents),
						nil,
						nil,
						nil,
						0,
						genesisHeader.Bits(),
						0,
					)

					err := g.GHOSTDAG(blockID)
					if err != nil {
						t.Fatalf("Test failed: \n Impl: %s,FileName: %s \n error on GHOSTDAG - block %s: %s.",
							factory.implName, info.Name(), testBlockData.ID, err)
					}
					ghostdagData, err := ghostdagDataStore.Get(nil, nil, blockID)
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

					if !StringToDomainHash(testBlockData.SelectedParent).Equal(ghostdagData.SelectedParent()) {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: expected selected parent %v but got %s.",
							factory.implName, info.Name(), testBlockData.ID, testBlockData.SelectedParent, ghostdagData.SelectedParent())
					}

					if !reflect.DeepEqual(StringToDomainHashSlice(testBlockData.MergeSetBlues), ghostdagData.MergeSetBlues()) {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: expected merge set blues %v but got %v.",
							factory.implName, info.Name(), testBlockData.ID, testBlockData.MergeSetBlues, hashesToStrings(ghostdagData.MergeSetBlues()))
					}

					if !reflect.DeepEqual(StringToDomainHashSlice(testBlockData.MergeSetReds), ghostdagData.MergeSetReds()) {
						t.Fatalf("\nTEST FAILED:\n Impl: %s, FileName: %s \nBlock: %s, \nError: expected merge set reds %v but got %v.",
							factory.implName, info.Name(), testBlockData.ID, testBlockData.MergeSetReds, hashesToStrings(ghostdagData.MergeSetReds()))
					}
				}
				dagTopology.parentsMap = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
				dagTopology.parentsMap[genesisHash] = nil
				ghostdagDataStore.dagMap = make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData)
				ghostdagDataStore.dagMap[genesisHash] = blockGHOSTDAGDataGenesis
				blockHeadersStore.dagMap = make(map[externalapi.DomainHash]externalapi.BlockHeader)
				blockHeadersStore.dagMap[genesisHash] = genesisHeader
			}

			testsCounter++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if testsCounter != 6 {
			t.Fatalf("Expected 6 test files, ran %d instead", testsCounter)
		}
	})
}

// TestBlueWork tests if GHOSTDAG picks as selected parent the parent
// with the most blue work, even if its blue score is not the greatest.
// To do that it creates one chain of 3 blocks over genesis, and another
// chain of 2 blocks with more blue work than the 3 blocks chain, and
// checks that a block that points to both chain tips will have the
// 2 blocks chain tip as its selected parent.
func TestBlueWork(t *testing.T) {
	dagTopology := &DAGTopologyManagerImpl{
		parentsMap: make(map[externalapi.DomainHash][]*externalapi.DomainHash),
	}

	ghostdagDataStore := &GHOSTDAGDataStoreImpl{
		dagMap: make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData),
	}

	blockHeadersStore := &blockHeadersStore{
		dagMap: make(map[externalapi.DomainHash]externalapi.BlockHeader),
	}

	fakeGenesisHash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0})
	longestChainBlock1Hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})
	longestChainBlock2Hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})
	longestChainBlock3Hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})
	heaviestChainBlock1Hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4})
	heaviestChainBlock2Hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{5})
	tipHash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{6})

	lowDifficultyHeader := blockheader.NewImmutableBlockHeader(
		0,
		nil,
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0,
		0,
		0,
	)

	dagTopology.parentsMap[*fakeGenesisHash] = nil
	ghostdagDataStore.dagMap[*fakeGenesisHash] = model.NewBlockGHOSTDAGData(0, new(big.Int), nil, nil, nil, nil)
	blockHeadersStore.dagMap[*fakeGenesisHash] = lowDifficultyHeader

	dagTopology.parentsMap[*longestChainBlock1Hash] = []*externalapi.DomainHash{fakeGenesisHash}
	blockHeadersStore.dagMap[*longestChainBlock1Hash] = lowDifficultyHeader

	dagTopology.parentsMap[*longestChainBlock2Hash] = []*externalapi.DomainHash{longestChainBlock1Hash}
	blockHeadersStore.dagMap[*longestChainBlock2Hash] = lowDifficultyHeader

	dagTopology.parentsMap[*longestChainBlock3Hash] = []*externalapi.DomainHash{longestChainBlock2Hash}
	blockHeadersStore.dagMap[*longestChainBlock3Hash] = lowDifficultyHeader

	dagTopology.parentsMap[*heaviestChainBlock1Hash] = []*externalapi.DomainHash{fakeGenesisHash}
	blockHeadersStore.dagMap[*heaviestChainBlock1Hash] = blockheader.NewImmutableBlockHeader(
		0,
		nil,
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0,
		math.MaxUint32, // Put a very high difficulty so the chain that contains this block will have a very high blue work
		0,
	)

	dagTopology.parentsMap[*heaviestChainBlock2Hash] = []*externalapi.DomainHash{heaviestChainBlock1Hash}
	blockHeadersStore.dagMap[*heaviestChainBlock2Hash] = lowDifficultyHeader

	dagTopology.parentsMap[*tipHash] = []*externalapi.DomainHash{heaviestChainBlock2Hash, longestChainBlock3Hash}
	blockHeadersStore.dagMap[*tipHash] = lowDifficultyHeader

	manager := ghostdagmanager.New(nil, dagTopology, ghostdagDataStore, blockHeadersStore, 18)
	blocksForGHOSTDAG := []*externalapi.DomainHash{
		longestChainBlock1Hash,
		longestChainBlock2Hash,
		longestChainBlock3Hash,
		heaviestChainBlock1Hash,
		heaviestChainBlock2Hash,
		tipHash,
	}

	for _, blockHash := range blocksForGHOSTDAG {
		err := manager.GHOSTDAG(blockHash)
		if err != nil {
			t.Fatalf("GHOSTDAG: %+v", err)
		}
	}

	if ghostdagDataStore.dagMap[*longestChainBlock3Hash].BlueScore() <= ghostdagDataStore.dagMap[*heaviestChainBlock2Hash].BlueScore() {
		t.Fatalf("Expected longestChainBlock3Hash to have greater blue score than heaviestChainBlock2Hash")
	}

	if !ghostdagDataStore.dagMap[*tipHash].SelectedParent().Equal(heaviestChainBlock2Hash) {
		t.Fatalf("Expected the block with the most blue work to be the selected parent of the tip")
	}
}

func hashesToStrings(arr []*externalapi.DomainHash) []string {
	var strArr = make([]string, len(arr))
	for i, hash := range arr {
		strArr[i] = string(hash.ByteSlice())
	}
	return strArr
}

func StringToDomainHash(strID string) *externalapi.DomainHash {
	var genesisHashArray [externalapi.DomainHashSize]byte
	copy(genesisHashArray[:], strID)
	return externalapi.NewDomainHashFromByteArray(&genesisHashArray)
}

func StringToDomainHashSlice(stringIDArr []string) []*externalapi.DomainHash {
	domainHashArr := make([]*externalapi.DomainHash, len(stringIDArr))
	for i, strID := range stringIDArr {
		domainHashArr[i] = StringToDomainHash(strID)
	}
	return domainHashArr
}

/* ---------------------- */
type GHOSTDAGDataStoreImpl struct {
	dagMap map[externalapi.DomainHash]*model.BlockGHOSTDAGData
}

func (ds *GHOSTDAGDataStoreImpl) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {
	ds.dagMap[*blockHash] = blockGHOSTDAGData
}

func (ds *GHOSTDAGDataStoreImpl) IsStaged(*model.StagingArea) bool {
	panic("implement me")
}

func (ds *GHOSTDAGDataStoreImpl) Discard() {
	panic("implement me")
}

func (ds *GHOSTDAGDataStoreImpl) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

func (ds *GHOSTDAGDataStoreImpl) Get(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
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

func (dt *DAGTopologyManagerImpl) IsAncestorOf(stagingArea *model.StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	blockBParents, isOk := dt.parentsMap[*hashBlockB]
	if !isOk {
		return false, nil
	}

	for _, parentOfB := range blockBParents {
		if parentOfB.Equal(hashBlockA) {
			return true, nil
		}
	}

	for _, parentOfB := range blockBParents {
		isAncestorOf, err := dt.IsAncestorOf(nil, hashBlockA, parentOfB)
		if err != nil {
			return false, err
		}
		if isAncestorOf {
			return true, nil
		}
	}
	return false, nil

}

func (dt *DAGTopologyManagerImpl) IsAncestorOfAny(blockHash *externalapi.DomainHash, potentialDescendants []*externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}
func (dt *DAGTopologyManagerImpl) IsAnyAncestorOf([]*externalapi.DomainHash, *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}
func (dt *DAGTopologyManagerImpl) IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	panic("unimplemented")
}

func (dt *DAGTopologyManagerImpl) SetParents(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) error {
	panic("unimplemented")
}

type blockHeadersStore struct {
	dagMap map[externalapi.DomainHash]externalapi.BlockHeader
}

func (b *blockHeadersStore) Discard() { panic("unimplemented") }

func (b *blockHeadersStore) Commit(_ model.DBTransaction) error { panic("unimplemented") }

func (b *blockHeadersStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, blockHeader externalapi.BlockHeader) {
	b.dagMap[*blockHash] = blockHeader
}

func (b *blockHeadersStore) IsStaged(*model.StagingArea) bool { panic("unimplemented") }

func (b *blockHeadersStore) BlockHeader(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error) {
	header, ok := b.dagMap[*blockHash]
	if ok {
		return header, nil
	}
	return nil, errors.New("Header isn't in the store")
}

func (b *blockHeadersStore) HasBlockHeader(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	_, ok := b.dagMap[*blockHash]
	return ok, nil
}

func (b *blockHeadersStore) BlockHeaders(dbContext model.DBReader, stagingArea *model.StagingArea, blockHashes []*externalapi.DomainHash) ([]externalapi.BlockHeader, error) {
	res := make([]externalapi.BlockHeader, 0, len(blockHashes))
	for _, hash := range blockHashes {
		header, err := b.BlockHeader(nil, nil, hash)
		if err != nil {
			return nil, err
		}
		res = append(res, header)
	}
	return res, nil
}

func (b *blockHeadersStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
	delete(b.dagMap, *blockHash)
}

func (b *blockHeadersStore) Count(*model.StagingArea) uint64 {
	return uint64(len(b.dagMap))
}
