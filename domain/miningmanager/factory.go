package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensusreference"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager/blocktemplatebuilder"
	mempoolpkg "github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// Factory instantiates new mining managers
type Factory interface {
	NewMiningManager(consensus consensusreference.ConsensusReference, params *dagconfig.Params, mempoolConfig *mempoolpkg.Config) MiningManager
}

type factory struct{}

// NewMiningManager instantiate a new mining manager
func (f *factory) NewMiningManager(consensusReference consensusreference.ConsensusReference, params *dagconfig.Params,
	mempoolConfig *mempoolpkg.Config) MiningManager {

	mempool := mempoolpkg.New(mempoolConfig, consensusReference)
	// In the current pruning window (according to 06/04/2022) the header with the most mass weighted 5294 grams.
	// We take a 10x factor for safety.
	// TODO: Remove this behaviour once `ignoreHeaderMass` is set to true in all networks.
	const estimatedHeaderUpperBound = 60000
	if estimatedHeaderUpperBound > params.MaxBlockMass {
		panic(errors.Errorf("Estimated header mass upper bound is higher than the max block mass allowed"))
	}
	maxBlockMass := params.MaxBlockMass - estimatedHeaderUpperBound
	blockTemplateBuilder := blocktemplatebuilder.New(consensusReference, mempool, maxBlockMass, params.CoinbasePayloadScriptPublicKeyMaxLength)

	return &miningManager{
		consensusReference:   consensusReference,
		mempool:              mempool,
		blockTemplateBuilder: blockTemplateBuilder,
		cachingTime:          time.Now(),
		cacheLock:            &sync.Mutex{},
	}
}

// NewFactory creates a new mining manager factory
func NewFactory() Factory {
	return &factory{}
}
