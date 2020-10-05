package model

// ConsensusStateChanges ...
type ConsensusStateChanges struct {
	AcceptanceData *BlockAcceptanceData
	UTXODiff       *UTXODiff
}
