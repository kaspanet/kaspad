package domain

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
)

// Domain provides a reference to the domain's external aps
type Domain interface {
	MiningManager() miningmanager.MiningManager
	Consensus() consensus.Consensus
}

type domain struct {
	miningManager miningmanager.MiningManager
	consensus     consensus.Consensus
}

func (d domain) Consensus() consensus.Consensus {
	return d.consensus
}

func (d domain) MiningManager() miningmanager.MiningManager {
	return d.miningManager
}

// New instantiates a new instance of a Domain object
func New(dagParams *dagconfig.Params, db infrastructuredatabase.Database) (Domain, error) {
	consensusFactory := consensus.NewFactory()
	consensusInstance, err := consensusFactory.NewConsensus(dagParams, db)
	if err != nil {
		return nil, err
	}

	miningManagerFactory := miningmanager.NewFactory()
	miningManager := miningManagerFactory.NewMiningManager(consensusInstance, constants.MaxMassAcceptedByBlock)

	return &domain{
		consensus:     consensusInstance,
		miningManager: miningManager,
	}, nil
}
