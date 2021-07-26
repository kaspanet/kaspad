package consensusreference

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusReference holds a reference for consensus.
// It is used when you want access to a consensus that
// can be swapped.
type ConsensusReference struct {
	consensus **externalapi.Consensus
}

func (ref ConsensusReference) Consensus() externalapi.Consensus {
	return **ref.consensus
}

// NewConsensusReference constructs a new ConsensusReference
func NewConsensusReference(consensus **externalapi.Consensus) ConsensusReference {
	return ConsensusReference{consensus: consensus}
}
