package processes

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdag2"
	"testing"
)

func TestGHOSTDA(t *testing.T) {
	t.Errorf("helo") //keep running
	if false {
		t.Fatalf("The test failed") // string - //stop
	}

	dagTopology := &DAGTopologyManagerImpl{}
	ghostdagDataStore := &GHOSTDAGDataStoreImpl{}
	g := ghostdag2.New(nil, dagTopology, ghostdagDataStore, 10)

}

type GHOSTDAGDataStoreImpl struct {
	dagMap map[*model.DomainHash]*model.BlockGHOSTDAGData
}

func (ds *GHOSTDAGDataStoreImpl) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) error {
	ds.dagMap[blockHash] = blockGHOSTDAGData
	return nil
}
func (ds *GHOSTDAGDataStoreImpl) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) (*model.BlockGHOSTDAGData, error) {
	v, ok := ds.dagMap[blockHash]
	if ok {
		return v, nil
	}
	return nil, nil
}

//candidateBluesAnticoneSizes = make(map[model.DomainHash]model.KType, gm.k)
type DAGTopologyManagerImpl struct {
	//dagMap map[*model.DomainHash] *model.BlockGHOSTDAGData
	parentsMap map[*model.DomainHash][]*model.DomainHash
}

//Implemented//
func (dt *DAGTopologyManagerImpl) Parents(blockHash *model.DomainHash) ([]*model.DomainHash, error) {
	v, ok := dt.parentsMap[blockHash]
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
	bParents, ok := dt.parentsMap[blockHashB]
	if !ok {
		return false, nil
	}
	for _, r := range bParents {
		if r == blockHashA {
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
