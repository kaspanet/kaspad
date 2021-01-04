package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"sort"
	"testing"
)

type testGhostDAGSorter struct {
	slice []*externalapi.DomainHash
	tc    testapi.TestConsensus
	test  testing.TB
}

// NewTestGhostDAGSorter returns a sort.Interface over the slice, so you can sort it via GhostDAG ordering
func NewTestGhostDAGSorter(slice []*externalapi.DomainHash, tc testapi.TestConsensus, t testing.TB) sort.Interface {
	return testGhostDAGSorter{
		slice: slice,
		tc:    tc,
		test:  t,
	}
}

func (sorter testGhostDAGSorter) Len() int {
	return len(sorter.slice)
}

func (sorter testGhostDAGSorter) Less(i, j int) bool {
	ghostdagDataI, err := sorter.tc.GHOSTDAGDataStore().Get(sorter.tc.DatabaseContext(), sorter.slice[i])
	if err != nil {
		sorter.test.Fatalf("TestGhostDAGSorter: Failed getting ghostdag data for %s", err)
	}
	ghostdagDataJ, err := sorter.tc.GHOSTDAGDataStore().Get(sorter.tc.DatabaseContext(), sorter.slice[j])
	if err != nil {
		sorter.test.Fatalf("TestGhostDAGSorter: Failed getting ghostdag data for %s", err)
	}
	return !sorter.tc.GHOSTDAGManager().Less(sorter.slice[i], ghostdagDataI, sorter.slice[j], ghostdagDataJ)
}

func (sorter testGhostDAGSorter) Swap(i, j int) {
	sorter.slice[i], sorter.slice[j] = sorter.slice[j], sorter.slice[i]
}
