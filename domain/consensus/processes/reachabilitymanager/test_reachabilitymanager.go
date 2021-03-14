package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
)

type testReachabilityManager struct {
	*reachabilityManager
}

func (t *testReachabilityManager) ReachabilityReindexSlack() uint64 {
	return t.reachabilityManager.reindexSlack
}

func (t *testReachabilityManager) SetReachabilityReindexSlack(reindexSlack uint64) {
	t.reachabilityManager.reindexSlack = reindexSlack
}

func (t *testReachabilityManager) SetReachabilityReindexWindow(reindexWindow uint64) {
	t.reachabilityManager.reindexWindow = reindexWindow
}

func (t *testReachabilityManager) ValidateIntervals(root *externalapi.DomainHash) error {
	return t.reachabilityManager.validateIntervals(root)
}

func (t *testReachabilityManager) GetAllNodes(root *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	return t.reachabilityManager.getAllNodes(root)
}

// NewTestReachabilityManager creates an instance of a TestReachabilityManager
func NewTestReachabilityManager(manager model.ReachabilityManager) testapi.TestReachabilityManager {
	return &testReachabilityManager{reachabilityManager: manager.(*reachabilityManager)}
}
