package reachabilitydata

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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

func EmptyReachabilityData() model.ReachabilityData {
	return &reachabilityData{}
}

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

func (rd *reachabilityData) CloneWritable() model.ReachabilityData {
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

func (rd *reachabilityData) AddToFutureCoveringSet(futureHash *externalapi.DomainHash) {
	rd.futureCoveringSet = append(rd.futureCoveringSet, futureHash)
}

// Equal returns whether rd equals to other
func (rd *reachabilityData) Equal(other model.ReadOnlyReachabilityData) bool {
	if rd == nil || other == nil {
		return rd == nil && other == nil
	}

	if !externalapi.HashesEqual(rd.children, other.Children()) {
		return false
	}

	if !rd.parent.Equal(other.Parent()) {
		return false
	}

	if !rd.interval.Equal(other.Interval()) {
		return false
	}

	if !rd.futureCoveringSet.Equal(other.FutureCoveringSet()) {
		return false
	}

	return true
}
