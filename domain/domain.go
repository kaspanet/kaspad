package domain

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/kaspanet/kaspad/domain/prefixmanager"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
	"sync/atomic"
	"unsafe"
)

// Domain provides a reference to the domain's external aps
type Domain interface {
	MiningManager() miningmanager.MiningManager
	Consensus() externalapi.Consensus
	TemporaryConsensus() externalapi.Consensus
	CreateTemporaryConsensus() error
	CommitTemporaryConsensus() error
}
type domain struct {
	miningManager   miningmanager.MiningManager
	consensus       *externalapi.Consensus
	tempConsensus   *externalapi.Consensus // We assume there's no concurrent access to tempConsensus
	consensusConfig *consensus.Config
	db              infrastructuredatabase.Database
}

func (d *domain) Consensus() externalapi.Consensus {
	return *d.consensus
}

func (d *domain) TemporaryConsensus() externalapi.Consensus {
	return *d.tempConsensus
}

func (d *domain) MiningManager() miningmanager.MiningManager {
	return d.miningManager
}

func (d *domain) CreateTemporaryConsensus() error {
	_, hasInactivePrefix, err := prefixmanager.InactivePrefix(d.db)
	if err != nil {
		return err
	}

	if hasInactivePrefix {
		return errors.Errorf("cannot have more than one inactive prefix")
	}

	activePrefix, exists, err := prefixmanager.ActivePrefix(d.db)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("cannot create a temporary consensus when there's " +
			"no active consensus")
	}

	inactivePrefix := prefixmanager.NewPrefix(0)
	if activePrefix.Equal(prefixmanager.NewPrefix(0)) {
		inactivePrefix = prefixmanager.NewPrefix(1)
	}

	err = prefixmanager.SetPrefixAsInactive(d.db, inactivePrefix)
	if err != nil {
		return err
	}

	consensusFactory := consensus.NewFactory()
	consensusInstance, err := consensusFactory.NewConsensus(d.consensusConfig, d.db, inactivePrefix)
	if err != nil {
		return err
	}

	d.tempConsensus = &consensusInstance
	return nil
}

func (d *domain) CommitTemporaryConsensus() error {
	transaction, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer transaction.RollbackUnlessClosed()

	inactivePrefix, hasInactivePrefix, err := prefixmanager.InactivePrefix(d.db)
	if err != nil {
		return err
	}

	if !hasInactivePrefix {
		return errors.Errorf("there's no inactive prefix to commit")
	}

	activePrefix, exists, err := prefixmanager.ActivePrefix(transaction)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("cannot commit a temporary consensus when there's " +
			"no active consensus")
	}

	err = prefixmanager.SetPrefixAsActive(transaction, inactivePrefix)
	if err != nil {
		return err
	}

	err = prefixmanager.SetPrefixAsInactive(transaction, activePrefix)
	if err != nil {
		return err
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}

	// We delete anything associated with the old prefix outside
	// of the transaction in order to save memory.
	err = prefixmanager.DeleteInactivePrefix(d.db)
	if err != nil {
		return err
	}

	tempConsensusPointer := unsafe.Pointer(d.tempConsensus)
	consensusPointer := (*unsafe.Pointer)(unsafe.Pointer(&d.consensus))
	atomic.StorePointer(consensusPointer, tempConsensusPointer)
	d.tempConsensus = nil
	return nil
}

func (d *domain) DeleteTemporaryConsensus() error {
	err := prefixmanager.DeleteInactivePrefix(d.db)
	if err != nil {
		return err
	}

	d.tempConsensus = nil
	return nil
}

// New instantiates a new instance of a Domain object
func New(consensusConfig *consensus.Config, db infrastructuredatabase.Database) (Domain, error) {
	err := prefixmanager.DeleteInactivePrefix(db)
	if err != nil {
		return nil, err
	}

	activePrefix, exists, err := prefixmanager.ActivePrefix(db)
	if err != nil {
		return nil, err
	}

	if !exists {
		const defaultActivePrefix = 0
		activePrefix = prefixmanager.NewPrefix(defaultActivePrefix)
		err = prefixmanager.SetPrefixAsActive(db, activePrefix)
		if err != nil {
			return nil, err
		}
	}

	consensusFactory := consensus.NewFactory()
	consensusInstance, err := consensusFactory.NewConsensus(consensusConfig, db, activePrefix)
	if err != nil {
		return nil, err
	}

	miningManagerFactory := miningmanager.NewFactory()
	miningManager := miningManagerFactory.NewMiningManager(consensusInstance, &consensusConfig.Params)

	return &domain{
		consensus:       &consensusInstance,
		miningManager:   miningManager,
		consensusConfig: consensusConfig,
		db:              db,
	}, nil
}
