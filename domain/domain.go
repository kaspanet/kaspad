package domain

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
)

type Domain interface {
	miningmanager.MiningManager
	consensus.Consensus
}

type domain struct {
	miningmanager.MiningManager
	consensus.Consensus
}

func New(dagParams *dagconfig.Params, db infrastructuredatabase.Database) (Domain, error) {
	consensusFactory := consensus.NewFactory()
	consensusInstance, err := consensusFactory.NewConsensus(dagParams, db)
	if err != nil {
		return nil, err
	}

	miningManagerFactory := miningmanager.NewFactory()
	miningManager := miningManagerFactory.NewMiningManager(consensusInstance, constants.MaxMassAcceptedByBlock)

	return &domain{
		Consensus:     consensusInstance,
		MiningManager: miningManager,
	}, nil
}
