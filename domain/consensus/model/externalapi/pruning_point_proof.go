package externalapi

import "math/big"

// PruningPointProof is the data structure holding the pruning point proof
type PruningPointProof struct {
	Headers              []BlockHeader
	PruningPointBlueWork *big.Int
}
