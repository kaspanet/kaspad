package model

import "github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"

// PruningProofManager builds, validates and applies pruning proofs.
type PruningProofManager interface {
	BuildPruningPointProof(stagingArea *StagingArea) (*externalapi.PruningPointProof, error)
	ValidatePruningPointProof(pruningPointProof *externalapi.PruningPointProof) error
	ApplyPruningPointProof(pruningPointProof *externalapi.PruningPointProof) error
}
