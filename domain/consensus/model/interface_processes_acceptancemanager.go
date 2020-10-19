package model

// AcceptanceManager manages transaction acceptance
// and related data
type AcceptanceManager interface {
	CalculateAcceptanceDataAndUTXOMultiset(blockGHOSTDAGData *BlockGHOSTDAGData) (*BlockAcceptanceData, Multiset, error)
}
