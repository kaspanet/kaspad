package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// UTXODiffReversalData is used by ConsensusStateManager to reverse the UTXODiffs during a re-org
type UTXODiffReversalData struct {
	SelectedParentHash     *externalapi.DomainHash
	SelectedParentUTXODiff externalapi.UTXODiff
}
