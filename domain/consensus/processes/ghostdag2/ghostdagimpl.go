package ghostdag2

import (
	"github.com/kaspanet/kaspad/util/difficulty"
	"sort"

	"math/big"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type ghostdagHelper struct {
	k                  model.KType
	dataStore          model.GHOSTDAGDataStore
	dbAccess           model.DBReader
	dagTopologyManager model.DAGTopologyManager
	headerStore        model.BlockHeaderStore
}

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

func (gh *ghostdagHelper) GHOSTDAG(stagingArea *model.StagingArea, blockCandidate *externalapi.DomainHash) error {
	var blueScore uint64 = 0
	var blueWork = new(big.Int)
	blueWork.SetUint64(0)
	var selectedParent *externalapi.DomainHash = nil
	var mergeSetBlues = make([]*externalapi.DomainHash, 0)
	var mergeSetReds = make([]*externalapi.DomainHash, 0)

	parents, err := gh.dagTopologyManager.Parents(stagingArea, blockCandidate)
	if err != nil {
		return err
	}
	// If genesis:
	if len(parents) == 0 {
		blockGHOSTDAGData := model.NewBlockGHOSTDAGData(blueScore, blueWork, selectedParent, mergeSetBlues, mergeSetReds, nil)
		gh.dataStore.Stage(stagingArea, blockCandidate, blockGHOSTDAGData)
		return nil
	}

	maxBlueWork := new(big.Int)
	maxBlueWork.SetUint64(0)
	maxBlueScore := uint64(0)
	// Find the selected parent.
	for _, parent := range parents {
		parentBlockData, err := gh.dataStore.Get(gh.dbAccess, stagingArea, parent)
		if err != nil {
			return err
		}
		parentBlueWork := parentBlockData.BlueWork()
		switch parentBlueWork.Cmp(maxBlueWork) {
		case 0:
			if isMoreHash(parent, selectedParent) {
				selectedParent = parent
				maxBlueWork = parentBlueWork
				maxBlueScore = parentBlockData.BlueScore()
			}
		case 1:
			selectedParent = parent
			maxBlueWork = parentBlueWork
			maxBlueScore = parentBlockData.BlueScore()
		}
	}
	blueWork.Set(maxBlueWork)
	blueScore = maxBlueScore

	blueSet := make([]*externalapi.DomainHash, 0)
	blueSet = append(blueSet, selectedParent)
	mergeSetBlues = append(mergeSetBlues, selectedParent)
	mergeSet, err := gh.findMergeSet(stagingArea, parents, selectedParent)
	if err != nil {
		return err
	}
	err = gh.sortByBlueWork(stagingArea, mergeSet)
	if err != nil {
		return err
	}
	err = gh.findBlueSet(stagingArea, &blueSet, selectedParent)
	if err != nil {
		return err
	}
	// Iterate on mergeSet and divide it to mergeSetBlues and mergeSetReds.
	for _, mergeSetBlock := range mergeSet {
		if mergeSetBlock.Equal(selectedParent) {
			continue
		}
		err := gh.divideToBlueAndRed(stagingArea, mergeSetBlock, &mergeSetBlues, &mergeSetReds, &blueSet)
		if err != nil {
			return err
		}
	}
	blueScore += uint64(len(mergeSetBlues))

	// Calculation of blue work
	for _, blue := range mergeSetBlues {
		header, err := gh.headerStore.BlockHeader(gh.dbAccess, stagingArea, blue)
		if err != nil {
			return err
		}
		blueWork.Add(blueWork, difficulty.CalcWork(header.Bits()))
	}

	blockGHOSTDAGData := model.NewBlockGHOSTDAGData(blueScore, blueWork, selectedParent, mergeSetBlues, mergeSetReds, nil)
	gh.dataStore.Stage(stagingArea, blockCandidate, blockGHOSTDAGData)
	return nil
}

func isMoreHash(parent *externalapi.DomainHash, selectedParent *externalapi.DomainHash) bool {
	parentByteArray := parent.ByteArray()
	selectedParentByteArray := selectedParent.ByteArray()

	for i := 0; i < len(parentByteArray); i++ {
		switch {
		case parentByteArray[i] < selectedParentByteArray[i]:
			return false
		case parentByteArray[i] > selectedParentByteArray[i]:
			return true
		}
	}
	return false
}

func (gh *ghostdagHelper) divideToBlueAndRed(stagingArea *model.StagingArea, blockToCheck *externalapi.DomainHash,
	blues *[]*externalapi.DomainHash, reds *[]*externalapi.DomainHash, blueSet *[]*externalapi.DomainHash) error {

	var k = int(gh.k)
	anticoneBlocksCounter := 0
	anticoneBlues := make([]*externalapi.DomainHash, 0)
	isMergeBlue := true

	//check that not-connected to at most k.
	for _, blueblock := range *blueSet {
		isAnticone, err := gh.isAnticone(stagingArea, blueblock, blockToCheck)
		if err != nil {
			return err
		}
		if isAnticone {
			anticoneBlocksCounter++
			anticoneBlues = append(anticoneBlues, blueblock)
		}
		if anticoneBlocksCounter > k {
			isMergeBlue = false
			break
		}
	}
	if !isMergeBlue {
		if !contains(blockToCheck, *reds) {
			*reds = append(*reds, blockToCheck)
		}
		return nil
	}

	// check that the k-cluster of each anticone blue block is still valid.
	for _, blueBlock := range anticoneBlues {
		isDestroyed, err := gh.checkIfDestroy(stagingArea, blueBlock, blueSet)
		if err != nil {
			return err
		}
		if isDestroyed {
			isMergeBlue = false
			break
		}
	}
	if !isMergeBlue {
		if !contains(blockToCheck, *reds) {
			*reds = append(*reds, blockToCheck)
		}
		return nil
	}
	if !contains(blockToCheck, *blues) {
		*blues = append(*blues, blockToCheck)
	}
	if !contains(blockToCheck, *blueSet) {
		*blueSet = append(*blueSet, blockToCheck)
	}
	return nil
}

func (gh *ghostdagHelper) isAnticone(stagingArea *model.StagingArea, blockA, blockB *externalapi.DomainHash) (bool, error) {
	isBlockAAncestorOfBlockB, err := gh.dagTopologyManager.IsAncestorOf(stagingArea, blockA, blockB)
	if err != nil {
		return false, err
	}
	isBlockBAncestorOfBlockA, err := gh.dagTopologyManager.IsAncestorOf(stagingArea, blockB, blockA)
	if err != nil {
		return false, err
	}

	return !isBlockAAncestorOfBlockB && !isBlockBAncestorOfBlockA, nil
}

func (gh *ghostdagHelper) validateKCluster(stagingArea *model.StagingArea, chain *externalapi.DomainHash, checkedBlock *externalapi.DomainHash,
	counter *int, blueSet *[]*externalapi.DomainHash) (bool, error) {

	k := int(gh.k)
	isAnticone, err := gh.isAnticone(stagingArea, chain, checkedBlock)
	if err != nil {
		return false, err
	}
	if isAnticone {
		if *counter > k {
			return false, nil
		}
		ifDestroy, err := gh.checkIfDestroy(stagingArea, chain, blueSet)
		if err != nil {
			return false, err
		}
		if ifDestroy {
			return false, nil
		}
		*counter++
		return true, nil
	}

	isAncestorOf, err := gh.dagTopologyManager.IsAncestorOf(stagingArea, checkedBlock, chain)
	if err != nil {
		return false, err
	}
	if isAncestorOf {
		dataStore, err := gh.BlockData(stagingArea, chain)
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

func contains(desiredItem *externalapi.DomainHash, items []*externalapi.DomainHash) bool {
	for _, item := range items {
		if item.Equal(desiredItem) {
			return true
		}
	}
	return false
}

// find number of not-connected in blueSet
func (gh *ghostdagHelper) checkIfDestroy(stagingArea *model.StagingArea, blockBlue *externalapi.DomainHash,
	blueSet *[]*externalapi.DomainHash) (bool, error) {

	// check that the K-cluster of each block in the blueSet is not destroyed when adding the block to the mergeSet.
	k := int(gh.k)
	counter := 0
	for _, blue := range *blueSet {
		isAnticone, err := gh.isAnticone(stagingArea, blue, blockBlue)
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

func (gh *ghostdagHelper) findMergeSet(stagingArea *model.StagingArea, parents []*externalapi.DomainHash,
	selectedParent *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	mergeSet := make([]*externalapi.DomainHash, 0)
	blocksQueue := make([]*externalapi.DomainHash, 0)
	for _, parent := range parents {
		if !contains(parent, blocksQueue) {
			blocksQueue = append(blocksQueue, parent)
		}
	}
	for len(blocksQueue) > 0 {
		block := blocksQueue[0]
		blocksQueue = blocksQueue[1:]
		if selectedParent.Equal(block) {
			if !contains(block, mergeSet) {
				mergeSet = append(mergeSet, block)
			}
			continue
		}
		isAncestorOf, err := gh.dagTopologyManager.IsAncestorOf(stagingArea, block, selectedParent)
		if err != nil {
			return nil, err
		}
		if isAncestorOf {
			continue
		}
		if !contains(block, mergeSet) {
			mergeSet = append(mergeSet, block)
		}
		err = gh.insertParent(stagingArea, block, &blocksQueue)
		if err != nil {
			return nil, err
		}
	}
	return mergeSet, nil
}

// Insert all parents to the queue
func (gh *ghostdagHelper) insertParent(stagingArea *model.StagingArea, child *externalapi.DomainHash, queue *[]*externalapi.DomainHash) error {
	parents, err := gh.dagTopologyManager.Parents(stagingArea, child)
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

func (gh *ghostdagHelper) findBlueSet(stagingArea *model.StagingArea, blueSet *[]*externalapi.DomainHash,
	selectedParent *externalapi.DomainHash) error {

	for selectedParent != nil {
		if !contains(selectedParent, *blueSet) {
			*blueSet = append(*blueSet, selectedParent)
		}
		blockData, err := gh.dataStore.Get(gh.dbAccess, stagingArea, selectedParent)
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

func (gh *ghostdagHelper) sortByBlueWork(stagingArea *model.StagingArea, arr []*externalapi.DomainHash) error {
	var err error
	sort.Slice(arr, func(i, j int) bool {

		blockLeft, err := gh.dataStore.Get(gh.dbAccess, stagingArea, arr[i])
		if err != nil {
			return false
		}
		blockRight, err := gh.dataStore.Get(gh.dbAccess, stagingArea, arr[j])
		if err != nil {
			return false
		}
		if blockLeft.BlueWork().Cmp(blockRight.BlueWork()) == -1 {
			return true
		}
		if blockLeft.BlueWork().Cmp(blockRight.BlueWork()) == 0 {
			return isMoreHash(arr[j], arr[i])
		}
		return false
	})
	return err
}

func (gh *ghostdagHelper) BlockData(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
	return gh.dataStore.Get(gh.dbAccess, stagingArea, blockHash)
}

func (gh *ghostdagHelper) ChooseSelectedParent(*model.StagingArea, ...*externalapi.DomainHash) (*externalapi.DomainHash, error) {
	panic("implement me")
}

func (gh *ghostdagHelper) Less(*externalapi.DomainHash, *model.BlockGHOSTDAGData, *externalapi.DomainHash, *model.BlockGHOSTDAGData) bool {
	panic("implement me")
}

func (gh *ghostdagHelper) GetSortedMergeSet(*model.StagingArea, *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}
