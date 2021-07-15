package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager/blocktemplatebuilder"
	mempoolpkg "github.com/kaspanet/kaspad/domain/miningmanager/mempool"
)

// Factory instantiates new mining managers
type Factory interface {
	NewMiningManager(consensus externalapi.ConsensusReference, params *dagconfig.Params, mempoolConfig *mempoolpkg.Config) MiningManager
}

type factory struct{}

// NewMiningManager instantiate a new mining manager
func (f *factory) NewMiningManager(consensusReference externalapi.ConsensusReference, params *dagconfig.Params,
	mempoolConfig *mempoolpkg.Config) MiningManager {

	mempool := mempoolpkg.New(mempoolConfig, consensusReference)
	blockTemplateBuilder := blocktemplatebuilder.New(consensusReference, mempool, params.MaxBlockMass)

	return &miningManager{
		mempool:              mempool,
		blockTemplateBuilder: blockTemplateBuilder,
	}
}

// NewFactory creates a new mining manager factory
func NewFactory() Factory {
	return &factory{}
}
