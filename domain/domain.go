package domain

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
)

// Domain provides a reference to the domain's external aps
type Domain interface {
	MiningManager() miningmanager.MiningManager
	Consensus() externalapi.Consensus
}

type domain struct {
	miningManager miningmanager.MiningManager
	consensus     externalapi.Consensus
}

func (d domain) Consensus() externalapi.Consensus {
	return d.consensus
}

func (d domain) MiningManager() miningmanager.MiningManager {
	return d.miningManager
}

// New instantiates a new instance of a Domain object
func New(consensusConfig *consensus.Config, mempoolConfig *mempool.Config,
	db infrastructuredatabase.Database) (Domain, error) {

	consensusFactory := consensus.NewFactory()
	consensusInstance, err := consensusFactory.NewConsensus(consensusConfig, db)
	if err != nil {
		return nil, err
	}

	miningManagerFactory := miningmanager.NewFactory()
	miningManager := miningManagerFactory.NewMiningManager(consensusInstance, &consensusConfig.Params, mempoolConfig)

	return &domain{
		consensus:     consensusInstance,
		miningManager: miningManager,
	}, nil
}
