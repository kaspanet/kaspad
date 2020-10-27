package ghostdag2

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"sort"
)

type ghostdagHelper struct {
	k                  model.KType
	dataStore          model.GHOSTDAGDataStore
	dbAccess           *database.DomainDBContext
	dagTopologyManager model.DAGTopologyManager
}

// New instantiates a new GHOSTDAGHelper -like a factory
func New(
	databaseContext *dbaccess.DatabaseContext,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	k model.KType) model.GHOSTDAGManager {

	return &ghostdagHelper{
		dbAccess:           database.NewDomainDBContext(databaseContext),
		dagTopologyManager: dagTopologyManager,
		dataStore:          ghostdagDataStore,
		k:                  k,
	}
}

/* --------------------------------------------- */
//K, GHOSTDAGDataStore
func (gh *ghostdagHelper) GHOSTDAG(blockCandidate *externalapi.DomainHash) error {
	var maxNum uint64 = 0
	var myScore uint64 = 0
	var selectedParent *externalapi.DomainHash
	/* find the selectedParent */
	blockParents, err := gh.dagTopologyManager.Parents(blockCandidate)
	if err != nil {
		return err
	}
	for _, w := range blockParents {
		blockData, err := gh.dataStore.Get(gh.dbAccess, w)
		if err != nil {
			return err
		}
		score := blockData.BlueScore
		//GHOSTDAGDataStore.get(w).blueScore()
		if score > maxNum {
			selectedParent = w
			maxNum = score
		}
		if score == maxNum && isLessHash(w, selectedParent) {
			selectedParent = w
		}
	}
	myScore = maxNum

	/* Goal: iterate h's past and divide it to : blue, blues, reds.
	   Notes:
	   1. If block A is in B's reds group (for block B that belong to the blue group) it will never be in blue(or blues).
	*/

	var blues = make([]*externalapi.DomainHash, 0)
	var reds = make([]*externalapi.DomainHash, 0)
	var blueSet = make([]*externalapi.DomainHash, 0)

	mergeSetArr, err := gh.findMergeSet(blockParents, selectedParent)
	if err != nil {
		return err
	}
	//STOP HERE
	err = gh.sortByBlueScore(mergeSetArr)
	if err != nil {
		return err
	}
	err = gh.findBlueSet(&blueSet, selectedParent)
	if err != nil {
		return err
	}

	for _, d := range mergeSetArr {
		if *d == *selectedParent {
			if !contains(selectedParent, blues) {
				blues = append(blues, selectedParent)
			}
			continue
		}
		err := gh.divideBlueRed(selectedParent, d, &blues, &reds, &blueSet)
		if err != nil {
			return err
		}
	}
	myScore += uint64(len(blues))
	/* Finial Data:
	   1. BlueScore => myScore
	   2. blues
	   3. reds
	   4. selectedParent

	*/
	e := model.BlockGHOSTDAGData{
		BlueScore:      myScore,
		SelectedParent: selectedParent,
		MergeSetBlues:  blues,
		MergeSetReds:   reds,
	}
	gh.dataStore.Stage(blockCandidate, &e)
	return nil
}

/* --------isLessHash(w, selectedParent)----------------*/
func isLessHash(w *externalapi.DomainHash, selectedParent *externalapi.DomainHash) bool {
	//Check if w is less then selectedParent
	for i := len(w) - 1; i >= 0; i-- {
		switch {
		case w[i] < selectedParent[i]:
			return true
		case w[i] > selectedParent[i]:
			return false
		}
	}
	return false
}

/*  1. blue = selectedParent.blue + blues
    2. h is not connected to at most K blocks (from the blue group)
    3. for each block at blue , check if not destroy
*/

/* ---------------divideBluesReds--------------------- */
func (gh *ghostdagHelper) divideBlueRed(selectedParent *externalapi.DomainHash, desiredBlock *externalapi.DomainHash,
	blues *[]*externalapi.DomainHash, reds *[]*externalapi.DomainHash, blueSet *[]*externalapi.DomainHash) error {
	counter := 0
	var chain = selectedParent
	var stop = false
	for chain != nil { /*nil -> after genesis*/
		// iterate on the selected parent chain, for each node in the chain i check also for his mergeSet.
		isValid, err := gh.validateKCluster(chain, desiredBlock, &counter, blueSet)
		if err != nil {
			return err
		}
		if !isValid {
			stop = true
			break
		}
		/* Check valid for the blues of the chain */
		blockData, err := gh.dataStore.Get(gh.dbAccess, chain)
		if err != nil {
			return err
		}

		for _, b := range blockData.MergeSetBlues {
			//if !gh.validateKCluster(b, desiredBlock, &counter, blueSet) { /* ret false*/
			isValid2, err := gh.validateKCluster(b, desiredBlock, &counter, blueSet)
			if err != nil {
				return err
			}
			if !isValid2 {
				stop = true
				break
			}
		}

		if stop {
			break
		}
		//chain = gh.dataStore.Get(gh.dbAccess, chain).SelectedParent
		blockData2, err := gh.dataStore.Get(gh.dbAccess, chain)
		if err != nil {
			return err
		}
		chain = blockData2.SelectedParent
	}
	if stop {
		if !contains(desiredBlock, *reds) {
			*reds = append(*reds, desiredBlock)
		}
	} else {
		var isBlue bool = true

		for _, e := range *blues {
			isDestroyed, err := gh.checkIfDestroy(e, blues)
			if err != nil {
				return err
			}
			if isDestroyed {
				isBlue = false
				break
			}
		}
		if !isBlue {
			if !contains(desiredBlock, *reds) {
				*reds = append(*reds, desiredBlock)
			}
		} else {
			if !contains(desiredBlock, *blues) {
				*blues = append(*blues, desiredBlock)
			}
			if !contains(desiredBlock, *blueSet) {
				*blueSet = append(*blueSet, desiredBlock)
			}
		}
	}
	return nil
}

/* ---------------isAnticone-------------------------- */
func (gh *ghostdagHelper) isAnticone(h1, h2 *externalapi.DomainHash) (bool, error) {
	//return !isInPast(h1, h2) && !isInPast(h1,h2)
	//return !gh.dagTopologyManager.IsAncestorOf(h1, h2) && !gh.dagTopologyManager.IsAncestorOf(h2, h1)
	isAB, err := gh.dagTopologyManager.IsAncestorOf(h1, h2)
	if err != nil {
		return false, err
	}
	isBA, err := gh.dagTopologyManager.IsAncestorOf(h2, h1)
	if err != nil {
		return false, err
	}
	return !isAB && !isBA, nil

}

/* ----------------validateKCluster------------------- */
func (gh *ghostdagHelper) validateKCluster(chain *externalapi.DomainHash, checkedBlock *externalapi.DomainHash, counter *int, blueSet *[]*externalapi.DomainHash) (bool, error) {
	var k int = int(gh.k)
	isAnt, err := gh.isAnticone(chain, checkedBlock)
	if err != nil {
		return false, err
	}
	if isAnt {
		//if n := gh.isAnticone(chain, checkedBlock); n {
		if *counter > k {
			return false, nil
		}
		//if gh.checkIfDestroy(chain, blueSet) {
		ifDes, err := gh.checkIfDestroy(chain, blueSet)
		if err != nil {
			return false, err
		}
		if ifDes {
			return false, nil
		}
		*counter++
		return true, nil
	} else {
		isAnt2, err := gh.dagTopologyManager.IsAncestorOf(checkedBlock, chain)
		if err != nil {
			return false, err
		}
		//if gh.dagTopologyManager.IsAncestorOf(checkedBlock, chain) {
		if isAnt2 {
			dataStore, err := gh.BlockData(chain)
			if err != nil {
				return false, err
			}
			if g := dataStore.MergeSetReds; contains(checkedBlock, g) {
				//if g := gh.dataStore.Get(gh.dbAccess, chain).MergeSetReds; contains(checkedBlock, g) {
				return false, nil
			}
		} else {
			return true, nil
		}
	}
	return false, nil
}

/*----------------contains-------------------------- */
func contains(s *externalapi.DomainHash, g []*externalapi.DomainHash) bool {
	for _, r := range g {
		if *r == *s {
			return true
		}
	}
	return false
}

/* ----------------checkIfDestroy------------------- */
/* find number of not-connected in his blue*/
func (gh *ghostdagHelper) checkIfDestroy(chain *externalapi.DomainHash, blueSet *[]*externalapi.DomainHash) (bool, error) {
	// Goal: check that the K-cluster of each block in the blueSet is not destroyed when adding the block to the mergeSet.
	var k int = int(gh.k)
	counter := 0
	for _, s2 := range *blueSet {
		//if gh.isAnticone(s2, chain) {
		isAnt, err := gh.isAnticone(s2, chain)
		if err != nil {
			return true, err
		}
		if isAnt {
			counter++
		}
		if counter > k {
			return true, nil
		}
	}
	return false, nil
}

/* ----------------findMergeSet------------------- */
func (gh *ghostdagHelper) findMergeSet(h []*externalapi.DomainHash, selectedParent *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	allMergeSet := make([]*externalapi.DomainHash, 0)
	var nodeQueue = make([]*externalapi.DomainHash, 0)
	for _, g := range h {
		if !contains(g, nodeQueue) {
			nodeQueue = append(nodeQueue, g)
		}

	}
	for len(nodeQueue) > 0 { /*return boolean */
		ha := nodeQueue[0]
		nodeQueue = nodeQueue[1:]
		if *selectedParent == *ha {
			if !contains(ha, allMergeSet) {
				allMergeSet = append(allMergeSet, ha)
			}
			continue
		}
		//if isInPast(ha, selectedParent){
		//if gh.dagTopologyManager.IsAncestorOf(ha, selectedParent) {
		isanc, err := gh.dagTopologyManager.IsAncestorOf(ha, selectedParent)
		if err != nil {
			return nil, err
		}
		if isanc {
			continue
		}
		if !contains(ha, allMergeSet) {
			allMergeSet = append(allMergeSet, ha)
		}
		err = gh.insertParent(ha, &nodeQueue)
		if err != nil {
			return nil, err
		}

	}
	return allMergeSet, nil
}

/* ----------------insertParent------------------- */
/* Insert all parents to the queue*/
func (gh *ghostdagHelper) insertParent(h *externalapi.DomainHash, q1 *[]*externalapi.DomainHash) error {
	parents, err := gh.dagTopologyManager.Parents(h)
	if err != nil {
		return err
	}
	for _, v := range parents {
		if contains(v, *q1) {
			continue
		}
		*q1 = append(*q1, v)
	}
	return nil
}

/* ----------------findBlueSet------------------- */
func (gh *ghostdagHelper) findBlueSet(blueSet *[]*externalapi.DomainHash, h *externalapi.DomainHash) error {
	for h != nil {
		if !contains(h, *blueSet) {
			*blueSet = append(*blueSet, h)
		}
		//blueSet = append(gh.dataStore.Get(gh.dbAccess, h).MergeSetBlues, blueSet) //change
		//for _, v := range gh.dataStore.Get(gh.dbAccess, h).MergeSetBlues {
		blockData, err := gh.dataStore.Get(gh.dbAccess, h)
		if err != nil {
			return err
		}
		mergeSetBlue := blockData.MergeSetBlues
		for _, v := range mergeSetBlue {
			if contains(v, *blueSet) {
				continue
			}
			*blueSet = append(*blueSet, v)
		}
		//h = gh.dataStore.Get(gh.dbAccess, h).SelectedParent
		h = blockData.SelectedParent
	}
	return nil
}

/* ----------------sortByBlueScore------------------- */
func (gh *ghostdagHelper) sortByBlueScore(arr []*externalapi.DomainHash) error {
	//var err error
	//
	//sort.SliceStable(*arr, func(i, j int) bool {
	//
	//	isSmaller := gh.dataStore.Get(gh.dbAccess, (*arr)[i]).BlueScore < gh.dataStore.Get(gh.dbAccess, (*arr)[j]).BlueScore
	//	return isSmaller
	//})
	//return nil
	blockData := make([]*model.BlockGHOSTDAGData, len(arr))

	for i, h := range arr {
		var err error
		blockData[i], err = gh.dataStore.Get(gh.dbAccess, h)
		if err != nil {
			return err
		}
	}
	sort.SliceStable(blockData, func(i, j int) bool {
		isSmaller := blockData[i].BlueScore < blockData[j].BlueScore
		return isSmaller
	})
	return nil
}

/* --------------------------------------------- */

func (gh *ghostdagHelper) BlockData(blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
	return gh.dataStore.Get(gh.dbAccess, blockHash)
	//last
}

func (gh *ghostdagHelper) ChooseSelectedParent(blockHashA *externalapi.DomainHash,
	blockHashB *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	panic("unimplemented")
}
