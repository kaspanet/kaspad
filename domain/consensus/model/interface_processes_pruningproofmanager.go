package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type PruningProofManager interface {
	BuildPruningPointProof(stagingArea *StagingArea) (*externalapi.PruningPointProof, error)
	ValidatePruningPointProof(pruningPointProof *externalapi.PruningPointProof) error
	ApplyPruningPointProof(stagingArea *StagingArea, pruningPointProof *externalapi.PruningPointProof) error
}
