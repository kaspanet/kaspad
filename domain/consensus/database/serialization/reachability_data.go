package serialization

import (
	"github.com/zoomy-network/zoomyd/domain/consensus/model"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/reachabilitydata"
)

// ReachablityDataToDBReachablityData converts ReachabilityData to DbReachabilityData
func ReachablityDataToDBReachablityData(reachabilityData model.ReachabilityData) *DbReachabilityData {
	parent := reachabilityData.Parent()
	var dbParent *DbHash
	if parent != nil {
		dbParent = DomainHashToDbHash(parent)
	}

	return &DbReachabilityData{
		Children:          DomainHashesToDbHashes(reachabilityData.Children()),
		Parent:            dbParent,
		Interval:          reachablityIntervalToDBReachablityInterval(reachabilityData.Interval()),
		FutureCoveringSet: DomainHashesToDbHashes(reachabilityData.FutureCoveringSet()),
	}
}

// DBReachablityDataToReachablityData converts DbReachabilityData to ReachabilityData
func DBReachablityDataToReachablityData(dbReachabilityData *DbReachabilityData) (model.ReachabilityData, error) {
	children, err := DbHashesToDomainHashes(dbReachabilityData.Children)
	if err != nil {
		return nil, err
	}

	var parent *externalapi.DomainHash
	if dbReachabilityData.Parent != nil {
		var err error
		parent, err = DbHashToDomainHash(dbReachabilityData.Parent)
		if err != nil {
			return nil, err
		}
	}

	interval := dbReachablityIntervalToReachablityInterval(dbReachabilityData.Interval)

	futureCoveringSet, err := DbHashesToDomainHashes(dbReachabilityData.FutureCoveringSet)
	if err != nil {
		return nil, err
	}

	return reachabilitydata.New(children, parent, interval, futureCoveringSet), nil
}
