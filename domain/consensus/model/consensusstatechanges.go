package model

// ConsensusStateChanges represents a set of changes that need to be made
// to transition the current consensus state to a new one
type ConsensusStateChanges struct {
	AcceptanceData *BlockAcceptanceData
	UTXODiff       *UTXODiff
}
