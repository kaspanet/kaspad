package miningmanager

import (
	"sync"
	"time"

	"github.com/c4ei/yunseokyeol/domain/consensusreference"
	"github.com/c4ei/yunseokyeol/domain/dagconfig"
	"github.com/c4ei/yunseokyeol/domain/miningmanager/blocktemplatebuilder"
	mempoolpkg "github.com/c4ei/yunseokyeol/domain/miningmanager/mempool"
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
	blockTemplateBuilder := blocktemplatebuilder.New(consensusReference, mempool, params.MaxBlockMass, params.CoinbasePayloadScriptPublicKeyMaxLength)

	return &miningManager{
		consensusReference:   consensusReference,
		mempool:              mempool,
		blockTemplateBuilder: blockTemplateBuilder,
		cachingTime:          time.Time{},
		cacheLock:            &sync.Mutex{},
	}
}

// NewFactory creates a new mining manager factory
func NewFactory() Factory {
	return &factory{}
}
