package reachabilitymanager

import "github.com/kaspanet/kaspad/domain/consensus/model"

type testReachabilityManager struct {
	*reachabilityManager
}

func (t testReachabilityManager) ReachabilityReindexSlack() uint64 {
	return reachabilityReindexSlack
}

func (t testReachabilityManager) ReachabilityReindexWindow() uint64 {
	return reachabilityReindexWindow
}

func NewTestReachabilityManager(manager model.ReachabilityManager) model.TestReachabilityManager {
	return &testReachabilityManager{reachabilityManager: manager.(*reachabilityManager)}
}
