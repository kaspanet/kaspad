package ghostdag2

import (
	"sort"

	"math/big"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util"
)

type ghostdagHelper struct {
	k                  model.KType
	dataStore          model.GHOSTDAGDataStore
	dbAccess           model.DBReader
	dagTopologyManager model.DAGTopologyManager
	headerStore        model.BlockHeaderStore
}

// New creates a new instance of this alternative ghostdag impl
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	headerStore model.BlockHeaderStore,
	k model.KType) model.GHOSTDAGManager {

	return &ghostdagHelper{
		dbAccess:           databaseContext,
		dagTopologyManager: dagTopologyManager,
		dataStore:          ghostdagDataStore,
		headerStore:        headerStore,
		k:                  k,
	}
}

/* --------------------------------------------- */

func (gh *ghostdagHelper) GHOSTDAG(blockCandidate *externalapi.DomainHash) error {
	myWork := new(big.Int)
	maxWork := new(big.Int)
	var myScore uint64
	var spScore uint64
	/* find the selectedParent */
	blockParents, err := gh.dagTopologyManager.Parents(blockCandidate)
	if err != nil {
		return err
	}
	var selectedParent = blockParents[0]
	for _, parent := range blockParents {
		blockData, err := gh.dataStore.Get(gh.dbAccess, parent)
		if err != nil {
			return err
		}
		blockWork := blockData.BlueWork()
		blockScore := blockData.BlueScore()
		if blockWork.Cmp(maxWork) == 1 {
			selectedParent = parent
			maxWork = blockWork
			spScore = blockScore
		}
		if blockWork.Cmp(maxWork) == 0 && ismoreHash(parent, selectedParent) {
			selectedParent = parent
			maxWork = blockWork
			spScore = blockScore
		}
	}
	myWork.Set(maxWork)
	myScore = spScore

	/* Goal: iterate blockCandidate's mergeSet and divide it to : blue, blues, reds. */
	var mergeSetBlues = make([]*externalapi.DomainHash, 0)
	var mergeSetReds = make([]*externalapi.DomainHash, 0)
	var blueSet = make([]*externalapi.DomainHash, 0)

	mergeSetBlues = append(mergeSetBlues, selectedParent)

	mergeSetArr, err := gh.findMergeSet(blockParents, selectedParent)
	if err != nil {
		return err
	}

	err = gh.sortByBlueWork(mergeSetArr)
	if err != nil {
		return err
	}
	err = gh.findBlueSet(&blueSet, selectedParent)
	if err != nil {
		return err
	}

	for _, mergeSetBlock := range mergeSetArr {
		if mergeSetBlock.Equal(selectedParent) {
			if !contains(selectedParent, mergeSetBlues) {
				mergeSetBlues = append(mergeSetBlues, selectedParent)
				blueSet = append(blueSet, selectedParent)
			}
			continue
		}
		err := gh.divideBlueRed(selectedParent, mergeSetBlock, &mergeSetBlues, &mergeSetReds, &blueSet)
		if err != nil {
			return err
		}
	}
	myScore += uint64(len(mergeSetBlues))

	// We add up all the *work*(not blueWork) that all our blues and selected parent did
	for _, blue := range mergeSetBlues {
		header, err := gh.headerStore.BlockHeader(gh.dbAccess, blue)
		if err != nil {
			return err
		}
		myWork.Add(myWork, util.CalcWork(header.Bits()))
	}

	e := model.NewBlockGHOSTDAGData(myScore, myWork, selectedParent, mergeSetBlues, mergeSetReds, nil)
	gh.dataStore.Stage(blockCandidate, e)
	return nil
}

/* --------isMoreHash(w, selectedParent)----------------*/
func ismoreHash(parent *externalapi.DomainHash, selectedParent *externalapi.DomainHash) bool {
	parentByteArray := parent.ByteArray()
	selectedParentByteArray := selectedParent.ByteArray()
	//Check if parentHash is more then selectedParentHash
	for i := len(parentByteArray) - 1; i >= 0; i-- {
		switch {
		case parentByteArray[i] < selectedParentByteArray[i]:
			return false
		case parentByteArray[i] > selectedParentByteArray[i]:
			return true
		}
	}
	return false
}

/*  1. blue = selectedParent.blue + blues
    2. not connected to at most K blocks (from the blue group)
    3. for each block at blue , check if not destroy
*/

/* ---------------divideBluesReds--------------------- */
func (gh *ghostdagHelper) divideBlueRed(selectedParent *externalapi.DomainHash, desiredBlock *externalapi.DomainHash,
	blues *[]*externalapi.DomainHash, reds *[]*externalapi.DomainHash, blueSet *[]*externalapi.DomainHash) error {
	var k = int(gh.k)
	counter := 0

	var suspectsBlues = make([]*externalapi.DomainHash, 0)
	isMergeBlue := true
	//check that not-connected to at most k.
	for _, block := range *blueSet {
		isAnticone, err := gh.isAnticone(block, desiredBlock)
		if err != nil {
			return err
		}
		if isAnticone {
			counter++
			suspectsBlues = append(suspectsBlues, block)
		}
		if counter > k {
			isMergeBlue = false
			break
		}
	}
	if !isMergeBlue {
		if !contains(desiredBlock, *reds) {
			*reds = append(*reds, desiredBlock)
		}
		return nil
	}

	// check that the k-cluster of each blue is still valid.
	for _, blue := range suspectsBlues {
		isDestroyed, err := gh.checkIfDestroy(blue, blueSet)
		if err != nil {
			return err
		}
		if isDestroyed {
			isMergeBlue = false
			break
		}
	}
	if !isMergeBlue {
		if !contains(desiredBlock, *reds) {
			*reds = append(*reds, desiredBlock)
		}
		return nil
	}
	if !contains(desiredBlock, *blues) {
		*blues = append(*blues, desiredBlock)
	}
	if !contains(desiredBlock, *blueSet) {
		*blueSet = append(*blueSet, desiredBlock)
	}
	return nil
}

/* ---------------isAnticone-------------------------- */
func (gh *ghostdagHelper) isAnticone(blockA, blockB *externalapi.DomainHash) (bool, error) {
	isAAncestorOfAB, err := gh.dagTopologyManager.IsAncestorOf(blockA, blockB)
	if err != nil {
		return false, err
	}
	isBAncestorOfA, err := gh.dagTopologyManager.IsAncestorOf(blockB, blockA)
	if err != nil {
		return false, err
	}
	return !isAAncestorOfAB && !isBAncestorOfA, nil

}

/* ----------------validateKCluster------------------- */
func (gh *ghostdagHelper) validateKCluster(chain *externalapi.DomainHash, checkedBlock *externalapi.DomainHash, counter *int, blueSet *[]*externalapi.DomainHash) (bool, error) {
	var k = int(gh.k)
	isAnticone, err := gh.isAnticone(chain, checkedBlock)
	if err != nil {
		return false, err
	}
	if isAnticone {
		if *counter > k {
			return false, nil
		}
		ifDestroy, err := gh.checkIfDestroy(chain, blueSet)
		if err != nil {
			return false, err
		}
		if ifDestroy {
			return false, nil
		}
		*counter++
		return true, nil
	}
	isAncestorOf, err := gh.dagTopologyManager.IsAncestorOf(checkedBlock, chain)
	if err != nil {
		return false, err
	}
	if isAncestorOf {
		dataStore, err := gh.BlockData(chain)
		if err != nil {
			return false, err
		}
		if mergeSetReds := dataStore.MergeSetReds(); contains(checkedBlock, mergeSetReds) {
			return false, nil
		}
	} else {
		return true, nil
	}

	return false, nil
}

/*----------------contains-------------------------- */
func contains(item *externalapi.DomainHash, items []*externalapi.DomainHash) bool {
	for _, r := range items {
		if r.Equal(item) {
			return true
		}
	}
	return false
}

/* ----------------checkIfDestroy------------------- */
/* find number of not-connected in his blue*/
func (gh *ghostdagHelper) checkIfDestroy(blockBlue *externalapi.DomainHash, blueSet *[]*externalapi.DomainHash) (bool, error) {
	// Goal: check that the K-cluster of each block in the blueSet is not destroyed when adding the block to the mergeSet.
	var k = int(gh.k)
	counter := 0
	for _, blue := range *blueSet {
		isAnticone, err := gh.isAnticone(blue, blockBlue)
		if err != nil {
			return true, err
		}
		if isAnticone {
			counter++
		}
		if counter > k {
			return true, nil
		}
	}
	return false, nil
}

/* ----------------findMergeSet------------------- */
func (gh *ghostdagHelper) findMergeSet(parents []*externalapi.DomainHash, selectedParent *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	allMergeSet := make([]*externalapi.DomainHash, 0)
	blockQueue := make([]*externalapi.DomainHash, 0)
	for _, parent := range parents {
		if !contains(parent, blockQueue) {
			blockQueue = append(blockQueue, parent)
		}

	}
	for len(blockQueue) > 0 {
		block := blockQueue[0]
		blockQueue = blockQueue[1:]
		if selectedParent.Equal(block) {
			if !contains(block, allMergeSet) {
				allMergeSet = append(allMergeSet, block)
			}
			continue
		}
		isancestorOf, err := gh.dagTopologyManager.IsAncestorOf(block, selectedParent)
		if err != nil {
			return nil, err
		}
		if isancestorOf {
			continue
		}
		if !contains(block, allMergeSet) {
			allMergeSet = append(allMergeSet, block)
		}
		err = gh.insertParent(block, &blockQueue)
		if err != nil {
			return nil, err
		}

	}
	return allMergeSet, nil
}

/* ----------------insertParent------------------- */
/* Insert all parents to the queue*/
func (gh *ghostdagHelper) insertParent(child *externalapi.DomainHash, queue *[]*externalapi.DomainHash) error {
	parents, err := gh.dagTopologyManager.Parents(child)
	if err != nil {
		return err
	}
	for _, parent := range parents {
		if contains(parent, *queue) {
			continue
		}
		*queue = append(*queue, parent)
	}
	return nil
}

/* ----------------findBlueSet------------------- */
func (gh *ghostdagHelper) findBlueSet(blueSet *[]*externalapi.DomainHash, selectedParent *externalapi.DomainHash) error {
	for selectedParent != nil {
		if !contains(selectedParent, *blueSet) {
			*blueSet = append(*blueSet, selectedParent)
		}
		blockData, err := gh.dataStore.Get(gh.dbAccess, selectedParent)
		if err != nil {
			return err
		}
		mergeSetBlue := blockData.MergeSetBlues()
		for _, blue := range mergeSetBlue {
			if contains(blue, *blueSet) {
				continue
			}
			*blueSet = append(*blueSet, blue)
		}
		selectedParent = blockData.SelectedParent()
	}
	return nil
}

/* ----------------sortByBlueScore------------------- */
func (gh *ghostdagHelper) sortByBlueWork(arr []*externalapi.DomainHash) error {

	var err error = nil
	sort.Slice(arr, func(i, j int) bool {

		blockLeft, error := gh.dataStore.Get(gh.dbAccess, arr[i])
		if error != nil {
			err = error
			return false
		}

		blockRight, error := gh.dataStore.Get(gh.dbAccess, arr[j])
		if error != nil {
			err = error
			return false
		}

		if blockLeft.BlueWork().Cmp(blockRight.BlueWork()) == -1 {
			return true
		}
		if blockLeft.BlueWork().Cmp(blockRight.BlueWork()) == 0 {
			return ismoreHash(arr[j], arr[i])
		}
		return false
	})
	return err
}

/* --------------------------------------------- */

func (gh *ghostdagHelper) BlockData(blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
	return gh.dataStore.Get(gh.dbAccess, blockHash)
}
func (gh *ghostdagHelper) ChooseSelectedParent(blockHashes ...*externalapi.DomainHash) (*externalapi.DomainHash, error) {
	panic("implement me")
}

func (gh *ghostdagHelper) Less(blockHashA *externalapi.DomainHash, ghostdagDataA *model.BlockGHOSTDAGData, blockHashB *externalapi.DomainHash, ghostdagDataB *model.BlockGHOSTDAGData) bool {
	panic("implement me")
}
