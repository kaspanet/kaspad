package testutils

import (
	"sort"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
)

type testGhostDAGSorter struct {
	slice       []*externalapi.DomainHash
	tc          testapi.TestConsensus
	test        testing.TB
	stagingArea *model.StagingArea
}

// NewTestGhostDAGSorter returns a sort.Interface over the slice, so you can sort it via GhostDAG ordering
func NewTestGhostDAGSorter(stagingArea *model.StagingArea, slice []*externalapi.DomainHash, tc testapi.TestConsensus,
	t testing.TB) sort.Interface {

	return testGhostDAGSorter{
		slice:       slice,
		tc:          tc,
		test:        t,
		stagingArea: stagingArea,
	}
}

func (sorter testGhostDAGSorter) Len() int {
	return len(sorter.slice)
}

func (sorter testGhostDAGSorter) Less(i, j int) bool {
	ghostdagDataI, err := sorter.tc.GHOSTDAGDataStore().Get(sorter.tc.DatabaseContext(), sorter.stagingArea, sorter.slice[i], false)
	if err != nil {
		sorter.test.Fatalf("TestGhostDAGSorter: Failed getting ghostdag data for %s", err)
	}
	ghostdagDataJ, err := sorter.tc.GHOSTDAGDataStore().Get(sorter.tc.DatabaseContext(), sorter.stagingArea, sorter.slice[j], false)
	if err != nil {
		sorter.test.Fatalf("TestGhostDAGSorter: Failed getting ghostdag data for %s", err)
	}
	return !sorter.tc.GHOSTDAGManager().Less(sorter.slice[i], ghostdagDataI, sorter.slice[j], ghostdagDataJ)
}

func (sorter testGhostDAGSorter) Swap(i, j int) {
	sorter.slice[i], sorter.slice[j] = sorter.slice[j], sorter.slice[i]
}
