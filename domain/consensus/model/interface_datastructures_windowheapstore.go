package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// WindowHeapSliceStore caches the slices that are needed for the heap implementation of DAGTraversalManager.BlockWindow
type WindowHeapSliceStore interface {
	Store
	Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, windowSize int, pairs []*externalapi.BlockGHOSTDAGDataHashPair)
	IsStaged(stagingArea *StagingArea) bool
	Get(stagingArea *StagingArea, blockHash *externalapi.DomainHash, windowSize int) ([]*externalapi.BlockGHOSTDAGDataHashPair, error)
}
