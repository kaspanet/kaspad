package consensusreference

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusReference holds a reference to a consensus object.
// The consensus object may be swapped with a new one entirely
// during the IBD process. Before an atomic consensus operation,
// callers are expected to call Consensus() once and work against
// that instance throughout.
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
