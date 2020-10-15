package model

// PastMedianTimeManager provides a method to resolve the
// past median time of a block
type PastMedianTimeManager interface {
	PastMedianTime(blockGHOSTDAGData *BlockGHOSTDAGData) (int64, error)
}
