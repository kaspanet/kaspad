package reachabilitymanager

import "github.com/kaspanet/kaspad/domain/consensus/model"

type testReachabilityManager struct {
	*reachabilityManager
}

func (t testReachabilityManager) ReachabilityReindexSlack() uint64 {
	return t.reachabilityManager.reindexSlack
}

func (t testReachabilityManager) SetReachabilityReindexSlack(reindexSlack uint64) {
	t.reachabilityManager.reindexSlack = reindexSlack
}

func (t testReachabilityManager) SetReachabilityReindexWindow(reindexWindow uint64) {
	t.reachabilityManager.reindexWindow = reindexWindow
}

func NewTestReachabilityManager(manager model.ReachabilityManager) model.TestReachabilityManager {
	return &testReachabilityManager{reachabilityManager: manager.(*reachabilityManager)}
}
