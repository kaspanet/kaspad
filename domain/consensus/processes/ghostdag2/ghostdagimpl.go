package ghostdag2

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
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
func (gh *ghostdagHelper) GHOSTDAG(blockParents []*model.DomainHash) (*model.BlockGHOSTDAGData, error) {
	var maxNum uint64 = 0
	var myScore uint64
	var selectedParent *model.DomainHash
	/* find the selectedParent */
	for _, w := range blockParents {
		blockData, err := gh.dataStore.Get(gh.dbAccess, w)
		if err != nil {
			return nil, err
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
	myScore = maxNum + 1

	/* Goal: iterate h's past and divide it to : blue, blues, reds.
	   Notes:
	   1. If block A is in B's reds group (for block B that belong to the blue group) it will never be in blue(or blues).
	*/

	var blues []*model.DomainHash = make([]*model.DomainHash, 0)
	var reds []*model.DomainHash = make([]*model.DomainHash, 0)
	var blueSet []*model.DomainHash = make([]*model.DomainHash, 0)

	mergeSetArr, err := gh.findMergeSet(blockParents, selectedParent)
	if err != nil {
		return nil, err
	}
	//STOP HERE
	err = gh.sortByBlueScore(mergeSetArr)
	if err != nil {
		return nil, err
	}
	err = gh.findBlueSet(&blueSet, selectedParent)
	if err != nil {
		return nil, err
	}

	for _, d := range mergeSetArr {
		err := gh.divideBlueRed(selectedParent, d, &blues, &reds, &blueSet)
		if err != nil {
			return nil, err
		}
	}
	myScore += uint64(len(blues))
	/* Finial Data:
	   1. BlueScore => myScore
	   2. blues
	   3. reds
	   4. selectedParent

	*/
	e := model.BlockGHOSTDAGData{BlueScore: myScore,
		SelectedParent: selectedParent,
		MergeSetBlues:  blues,
		MergeSetReds:   reds,
	}

	return &e, nil
}

/* --------isLessHash(w, selectedParent)----------------*/
func isLessHash(w *model.DomainHash, selectedParent *model.DomainHash) bool {
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
func (gh *ghostdagHelper) divideBlueRed(selectedParent *model.DomainHash, desiredBlock *model.DomainHash,
	blues *[]*model.DomainHash, reds *[]*model.DomainHash, blueSet *[]*model.DomainHash) error {
	counter := 0
	var chain = selectedParent
	var stop = false
	for chain != nil { /*nil -> after genesis*/
		//if !gh.validateKCluster(chain, desiredBlock, &counter, blueSet) { /* ret false*/
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
			isValid2, err2 := gh.validateKCluster(b, desiredBlock, &counter, blueSet)
			if err2 != nil {
				return err2
			}
			if !isValid2 {
				stop = true
				break
			}
		}
		for _, e := range *blues {
			//if !gh.validateKCluster(e, desiredBlock, &counter, blueSet) { /* ret false*/
			isValid3, err3 := gh.validateKCluster(e, desiredBlock, &counter, blueSet)
			if err3 != nil {
				return err3
			}
			if !isValid3 {
				stop = true
				break
			}
		}
		if stop {
			break
		}
		//chain = gh.dataStore.Get(gh.dbAccess, chain).SelectedParent
		blockData2, err2 := gh.dataStore.Get(gh.dbAccess, chain)
		if err2 != nil {
			return err2
		}
		chain = blockData2.SelectedParent
	}
	if stop {
		*reds = append(*reds, desiredBlock)
	}
	if !stop {
		*blues = append(*blues, desiredBlock)
		*blueSet = append(*blueSet, desiredBlock)
	}
	return nil
}

/* ---------------isAnticone-------------------------- */
func (gh *ghostdagHelper) isAnticone(h1, h2 *model.DomainHash) (bool, error) {
	//return !isInPast(h1, h2) && !isInPast(h1,h2)
	//return !gh.dagTopologyManager.IsAncestorOf(h1, h2) && !gh.dagTopologyManager.IsAncestorOf(h2, h1)
	isAB, err := gh.dagTopologyManager.IsAncestorOf(h1, h2)
	if err != nil {
		return false, err
	}
	isBA, err := gh.dagTopologyManager.IsAncestorOf(h1, h2)
	if err != nil {
		return false, err
	}
	return !isAB && !isBA, nil

}

/* ----------------validateKCluster------------------- */
func (gh *ghostdagHelper) validateKCluster(chain *model.DomainHash, s1 *model.DomainHash, counter *int, blueSet *[]*model.DomainHash) (bool, error) {
	var k int = int(gh.k)
	isAnt, err := gh.isAnticone(chain, s1)
	if err != nil {
		return false, err
	}
	if isAnt {
		//if n := gh.isAnticone(chain, s1); n {
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
		isAnt2, err := gh.dagTopologyManager.IsAncestorOf(s1, chain)
		if err != nil {
			return false, err
		}
		//if gh.dagTopologyManager.IsAncestorOf(s1, chain) {
		if isAnt2 {
			dataStore, err2 := gh.BlockData(chain)
			if err2 != nil {
				return false, err
			}
			if g := dataStore.MergeSetReds; contains(s1, g) {
				//if g := gh.dataStore.Get(gh.dbAccess, chain).MergeSetReds; contains(s1, g) {
				return false, nil
			}
		} else {
			return true, nil
		}
	}
	return false, nil
}

/*----------------contains-------------------------- */
func contains(s *model.DomainHash, g []*model.DomainHash) bool {
	for _, r := range g {
		if r == s {
			return true
		}
	}
	return false
}

/* ----------------checkIfDestroy------------------- */
/* find number of not-connected in his blue*/
func (gh *ghostdagHelper) checkIfDestroy(chain *model.DomainHash, blueSet *[]*model.DomainHash) (bool, error) {
	var k int = int(gh.k)
	counter := 0
	for _, s2 := range *blueSet {
		//if gh.isAnticone(s2, chain) {
		isAnt, err := gh.isAnticone(s2, chain)
		if err != nil {
			return false, err
		}
		if isAnt {
			counter++
		}
		if counter > k {
			return false, nil
		}
	}
	return true, nil
}

/* ----------------findMergeSet------------------- */
func (gh *ghostdagHelper) findMergeSet(h []*model.DomainHash, selectedParent *model.DomainHash) ([]*model.DomainHash, error) {

	allMergeSet := make([]*model.DomainHash, 0)
	var nodeQueue = make([]*model.DomainHash, 0)
	for _, g := range h {
		nodeQueue = append(nodeQueue, g)
	}
	for len(nodeQueue) > 0 { /*return boolean */
		ha := nodeQueue[0]
		nodeQueue = nodeQueue[1:]
		if selectedParent == ha {
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

		allMergeSet = append(allMergeSet, ha)
		err = gh.insertParent(ha, &nodeQueue)
		if err != nil {
			return nil, err
		}

	}
	return allMergeSet, nil
}

/* ----------------insertParent------------------- */
/* Insert all parents to the queue*/
func (gh *ghostdagHelper) insertParent(h *model.DomainHash, q1 *[]*model.DomainHash) error {
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
func (gh *ghostdagHelper) findBlueSet(blueSet *[]*model.DomainHash, h *model.DomainHash) error {
	for h != nil {
		*blueSet = append(*blueSet, h)
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
func (gh *ghostdagHelper) sortByBlueScore(arr []*model.DomainHash) error {
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

func (gh *ghostdagHelper) BlockData(blockHash *model.DomainHash) (*model.BlockGHOSTDAGData, error) {
	return gh.dataStore.Get(gh.dbAccess, blockHash)
	//last
}

func (gh *ghostdagHelper) ChooseSelectedParent(
	blockHashA *model.DomainHash, blockAGHOSTDAGData *model.BlockGHOSTDAGData,
	blockHashB *model.DomainHash, blockBGHOSTDAGData *model.BlockGHOSTDAGData) *model.DomainHash {
	panic("unimplemented")
}
