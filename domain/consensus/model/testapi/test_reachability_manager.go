package testapi

import (
	"github.com/c4ei/YunSeokYeol/domain/consensus/model"
	"github.com/c4ei/YunSeokYeol/domain/consensus/model/externalapi"
)

// TestReachabilityManager adds to the main ReachabilityManager methods required by tests
type TestReachabilityManager interface {
	model.ReachabilityManager
	SetReachabilityReindexWindow(reindexWindow uint64)
	SetReachabilityReindexSlack(reindexSlack uint64)
	ReachabilityReindexSlack() uint64
	ValidateIntervals(root *externalapi.DomainHash) error
	GetAllNodes(root *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
}
