package ghostdag2


import (
	"sort"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type ghostdagHelper struct{
	k 		      model.KType
	dataStore   model.GHOSTDAGDataStore
	dbAccess 	  *database.DomainDBContext
	dagTopologyManager model.DAGTopologyManager
}


// New instantiates a new GHOSTDAGHelper -like a factory
func New(
	databaseContext *dbaccess.DatabaseContext,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	k model.KType) model.GHOSTDAGManager {

	return &ghostdagHelper{
		dbAccess:    database.NewDomainDBContext(databaseContext),
		dagTopologyManager: dagTopologyManager,
		dataStore:  ghostdagDataStore,
		k:                  k,

	}
}


/* --------------------------------------------- */
//K, GHOSTDAGDataStore
func(gh *ghostdagHelper) GHOSTDAG (blockParents []*model.DomainHash) (*model.BlockGHOSTDAGData, error){
	var maxNum uint64 = 0
	var score uint64
	var myScore uint64
	var selectedParent *model.DomainHash
	/* find the selectedParent */
	for _,w := range blockParents{
		score = gh.dataStore.Get(gh.dbAccess, w).BlueScore
		//GHOSTDAGDataStore.get(w).blueScore()
		if score > maxNum{
			selectedParent = w
			maxNum = score
		}
	}
	myScore = maxNum

	/* Goal: iterate h's past and divide it to : blue, blues, reds.
	   Notes:
	   1. If block A is in B's reds group (for block B that belong to the blue group) it will never be in blue(or blues).
	*/



	var blues []*model.DomainHash = make([]*model.DomainHash, 0)
	var reds []*model.DomainHash = make([]*model.DomainHash, 0)
	var mergeSetArr []*model.DomainHash = make([]*model.DomainHash, 0)
	var blueSet []*model.DomainHash = make([]*model.DomainHash, 0)

	gh.findMergeSet(blockParents, &mergeSetArr, selectedParent)
	gh.sortByBlueScore(&mergeSetArr)
	gh.findBlueSet(&blueSet, selectedParent)
	for _,d := range mergeSetArr{
		gh.divideBlueRed(selectedParent, d, &blues, &reds, &blueSet)
	}
	myScore+= uint64(len(blues))
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

	return &e,nil
}



/*  1. blue = selectedParent.blue + blues
    2. h is not connected to at most K blocks (from the blue group)
    3. for each block at blue , check if not destroy
*/

/* ---------------divideBluesReds--------------------- */
func (gh *ghostdagHelper) divideBlueRed(selectedParent *model.DomainHash, desiredBlock *model.DomainHash, blues *[]*model.DomainHash, reds *[]*model.DomainHash, blueSet *[]*model.DomainHash ) {
	counter := 0
	var chain = selectedParent
	var stop = false
	for chain != nil { /*nil -> after genesis*/
		if !gh.validateKCluster(chain, desiredBlock, &counter, blueSet) {/* ret false*/
			stop = true
			break
		}
		/* Check valid for the blues of the chain */
		for _,b := range gh.dataStore.Get(gh.dbAccess, chain).MergeSetBlues{
				if !gh.validateKCluster(b, desiredBlock, &counter, blueSet) {/* ret false*/
					stop = true
					break
				}
		}
		for _,e := range *blues{
			if !gh.validateKCluster(e, desiredBlock, &counter, blueSet) {/* ret false*/
				stop = true
				break
			}
		}
		if stop {
			break
		}
		chain =gh.dataStore.Get(gh.dbAccess, chain).SelectedParent
	}
	if stop{
		*reds = append(*reds,desiredBlock )
	}
	if !stop{
		*blues = append(*blues, desiredBlock)
		*blueSet = append(*blueSet, desiredBlock)
	}
	return
}

/* ---------------isAnticone-------------------------- */
func (gh *ghostdagHelper) isAnticone(h1, h2 *model.DomainHash) bool{
	//return !isInPast(h1, h2) && !isInPast(h1,h2)
	return !gh.dagTopologyManager.IsAncestorOf(h1, h2) && ! gh.dagTopologyManager.IsAncestorOf(h2, h1)
}


/* ----------------validateKCluster------------------- */
func(gh *ghostdagHelper) validateKCluster(chain *model.DomainHash, s1 *model.DomainHash, counter *int, blueSet *[]*model.DomainHash) bool{
	var k int = int(gh.k)
	if n := gh.isAnticone(chain, s1);n {
		if *counter > k{
			return false;
		}
		if gh.checkIfDestroy(chain, blueSet){
			return false;
		}
		*counter ++
		return true
	} else{
		if gh.dagTopologyManager.IsAncestorOf(s1, chain){
			if g:= gh.dataStore.Get(gh.dbAccess, chain).MergeSetReds; contains(s1, g){
				return false
			}
		} else{
			return true
		}
	}
	return false
}

/*----------------contains-------------------------- */
func contains(s *model.DomainHash, g []*model.DomainHash) bool{
	for _,r := range g{
		if r == s{
			return true
		}
	}
	return false
}

/* ----------------checkIfDestroy------------------- */
/* find number of not-connected in his blue*/
func (gh *ghostdagHelper)checkIfDestroy(chain *model.DomainHash, blueSet *[]*model.DomainHash) bool{
	var k int = int(gh.k)
	counter := 0
	for _,s2 := range *blueSet{
		if gh.isAnticone(s2, chain) {
			counter++
		}
		if counter > k{
			return false
		}
	}
	return true
}


/* ----------------findMergeSet------------------- */
func (gh *ghostdagHelper)findMergeSet (h []*model.DomainHash, allMergeSet *[]*model.DomainHash, selectedParent *model.DomainHash) {

	var nodeQueue = make([]*model.DomainHash, 0)
	for _,g := range h {
		nodeQueue = append(nodeQueue, g)
	}
	for len(nodeQueue) > 0 {/*return boolean */
		ha := nodeQueue[0]
		nodeQueue = nodeQueue[1:]
		if selectedParent == ha{
			continue
		}
		//if isInPast(ha, selectedParent){
		if gh.dagTopologyManager.IsAncestorOf(ha, selectedParent){
			continue
		}

		*allMergeSet = append(*allMergeSet, ha)
		gh.insertParent(ha, &nodeQueue)

	}

}

/* ----------------insertParent------------------- */
/* Insert all parents to the queue*/
func (gh *ghostdagHelper) insertParent (h *model.DomainHash, q1 *[]*model.DomainHash) {
	for _,v := range gh.dagTopologyManager.Parents(h){
		if contains(v, *q1){
			continue
		}
		*q1 = append(*q1, v)
	}
}

/* ----------------findBlueSet------------------- */
func(gh *ghostdagHelper) findBlueSet(blueSet *[]*model.DomainHash, h *model.DomainHash) {
	for h!= nil{
		*blueSet = append( *blueSet, h)
		//blueSet = append(gh.dataStore.Get(gh.dbAccess, h).MergeSetBlues, blueSet) //change
		for _,v := range gh.dataStore.Get(gh.dbAccess, h).MergeSetBlues{
			if contains(v, *blueSet){
				continue
			}
			*blueSet = append(*blueSet, v)
		}
		h = gh.dataStore.Get(gh.dbAccess, h).SelectedParent
	}
}



/* ----------------sortByBlueScore------------------- */
func (gh *ghostdagHelper)sortByBlueScore(arr *[]*model.DomainHash) {
	sort.SliceStable(*arr, func(i, j int) bool{
		isSmaller := gh.dataStore.Get(gh.dbAccess,(*arr)[i] ).BlueScore < gh.dataStore.Get(gh.dbAccess,(*arr)[j]).BlueScore
		return isSmaller
	})
}


/* --------------------------------------------- */

func (gh *ghostdagHelper)BlockData(blockHash *model.DomainHash) *model.BlockGHOSTDAGData{
	return gh.dataStore.Get(gh.dbAccess, blockHash)
	//last
}

