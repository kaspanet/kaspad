package model

type ConsensusStateChanges struct {
	acceptanceData *AcceptanceData
	utxoDiff       *UTXODiff
}
