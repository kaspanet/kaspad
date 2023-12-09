package reachabilitydata

import (
	"github.com/zoomy-network/zoomyd/domain/consensus/model"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
)

type reachabilityData struct {
	children          []*externalapi.DomainHash
	parent            *externalapi.DomainHash
	interval          *model.ReachabilityInterval
	futureCoveringSet model.FutureCoveringTreeNodeSet
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ = &reachabilityData{
	[]*externalapi.DomainHash{},
	&externalapi.DomainHash{},
	&model.ReachabilityInterval{},
	model.FutureCoveringTreeNodeSet{},
}

// EmptyReachabilityData constructs an empty MutableReachabilityData object
func EmptyReachabilityData() model.MutableReachabilityData {
	return &reachabilityData{}
}

// New constructs a ReachabilityData object filled with given fields
func New(children []*externalapi.DomainHash,
	parent *externalapi.DomainHash,
	interval *model.ReachabilityInterval,
	futureCoveringSet model.FutureCoveringTreeNodeSet) model.ReachabilityData {

	return &reachabilityData{
		children:          children,
		parent:            parent,
		interval:          interval,
		futureCoveringSet: futureCoveringSet,
	}
}

func (rd *reachabilityData) Children() []*externalapi.DomainHash {
	return rd.children
}

func (rd *reachabilityData) Parent() *externalapi.DomainHash {
	return rd.parent
}

func (rd *reachabilityData) Interval() *model.ReachabilityInterval {
	return rd.interval
}

func (rd *reachabilityData) FutureCoveringSet() model.FutureCoveringTreeNodeSet {
	return rd.futureCoveringSet
}

func (rd *reachabilityData) CloneMutable() model.MutableReachabilityData {
	//return rd
	return &reachabilityData{
		children:          externalapi.CloneHashes(rd.children),
		parent:            rd.parent,
		interval:          rd.interval.Clone(),
		futureCoveringSet: rd.futureCoveringSet.Clone(),
	}
}

func (rd *reachabilityData) AddChild(child *externalapi.DomainHash) {
	rd.children = append(rd.children, child)
}

func (rd *reachabilityData) SetParent(parent *externalapi.DomainHash) {
	rd.parent = parent
}

func (rd *reachabilityData) SetInterval(interval *model.ReachabilityInterval) {
	rd.interval = interval
}

func (rd *reachabilityData) SetFutureCoveringSet(futureCoveringSet model.FutureCoveringTreeNodeSet) {
	rd.futureCoveringSet = futureCoveringSet
}

// Equal returns whether rd equals to other
func (rd *reachabilityData) Equal(other model.ReachabilityData) bool {
	otherReachabilityData, ok := other.(*reachabilityData)
	if !ok {
		return false
	}
	if rd == nil || otherReachabilityData == nil {
		return rd == otherReachabilityData
	}

	if !externalapi.HashesEqual(rd.children, otherReachabilityData.Children()) {
		return false
	}

	if !rd.parent.Equal(otherReachabilityData.Parent()) {
		return false
	}

	if !rd.interval.Equal(otherReachabilityData.Interval()) {
		return false
	}

	if !rd.futureCoveringSet.Equal(otherReachabilityData.FutureCoveringSet()) {
		return false
	}

	return true
}
