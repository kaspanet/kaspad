package processes

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdag2"
	"testing"
)

func TestGHOSTDA(t *testing.T){
	t.Errorf("helo") //keep running
	if false {
		t.Fatalf("The test failed") // string - //stop
	}

	g := ghostdag2.New()


}

type GHOSTDAGDataStoreImpl struct {
	dagMap map[*model.DomainHash] *model.BlockGHOSTDAGData
}


func (ds *GHOSTDAGDataStoreImpl) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData){
	ds.dagMap[blockHash] = blockGHOSTDAGData
}
func (ds *GHOSTDAGDataStoreImpl) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) *model.BlockGHOSTDAGData{
	v, ok := ds.dagMap[blockHash]
	if ok{
		return v
	}
	return nil
}






//candidateBluesAnticoneSizes = make(map[model.DomainHash]model.KType, gm.k)
type DAGTopologyManagerImpl struct{
	//dagMap map[*model.DomainHash] *model.BlockGHOSTDAGData
	parentsMap map[*model.DomainHash] []*model.DomainHash
}

//Implemented//
func (dt *DAGTopologyManagerImpl) Parents(blockHash *model.DomainHash) []*model.DomainHash{
	v, ok:= dt.parentsMap[blockHash]
	if !ok{
		return make([]*model.DomainHash, 0)
	}else{
		return v
	}
}

func (dt *DAGTopologyManagerImpl) Children(blockHash *model.DomainHash) []*model.DomainHash{
	return nil
}
func (dt *DAGTopologyManagerImpl) IsParentOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool{
	return true
}
func (dt *DAGTopologyManagerImpl) IsChildOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool{
	return true
}
//Implemented//
func (dt *DAGTopologyManagerImpl) IsAncestorOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool{
	v, ok:= dt.parentsMap[blockHashB]
	if !ok{
		return false
	}
	for _, r := range v{
		if r == blockHashA{
			return true
		}
	}
	for _, y := range v{
		if dt.IsAncestorOf(blockHashA, y){
			return true
		}
	}
	return false

}

func (dt *DAGTopologyManagerImpl) IsDescendantOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool{
	return true
}