package domain

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/kaspanet/kaspad/domain/consensusreference"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/kaspanet/kaspad/domain/prefixmanager"
	"github.com/kaspanet/kaspad/domain/prefixmanager/prefix"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

// Domain provides a reference to the domain's external aps
type Domain interface {
	MiningManager() miningmanager.MiningManager
	Consensus() externalapi.Consensus
	StagingConsensus() externalapi.Consensus
	InitStagingConsensusWithoutGenesis() error
	CommitStagingConsensus() error
	DeleteStagingConsensus() error
	ConsensusEventsChannel() chan externalapi.ConsensusEvent
}

type domain struct {
	miningManager          miningmanager.MiningManager
	consensus              *externalapi.Consensus
	stagingConsensus       *externalapi.Consensus
	stagingConsensusLock   sync.RWMutex
	consensusConfig        *consensus.Config
	db                     infrastructuredatabase.Database
	consensusEventsChannel chan externalapi.ConsensusEvent
}

func (d *domain) ConsensusEventsChannel() chan externalapi.ConsensusEvent {
	return d.consensusEventsChannel
}

func (d *domain) Consensus() externalapi.Consensus {
	return *d.consensus
}

func (d *domain) StagingConsensus() externalapi.Consensus {
	d.stagingConsensusLock.RLock()
	defer d.stagingConsensusLock.RUnlock()
	return *d.stagingConsensus
}

func (d *domain) MiningManager() miningmanager.MiningManager {
	return d.miningManager
}

func (d *domain) InitStagingConsensusWithoutGenesis() error {
	cfg := *d.consensusConfig
	cfg.SkipAddingGenesis = true
	return d.initStagingConsensus(&cfg)
}

func (d *domain) initStagingConsensus(cfg *consensus.Config) error {
	d.stagingConsensusLock.Lock()
	defer d.stagingConsensusLock.Unlock()

	_, hasInactivePrefix, err := prefixmanager.InactivePrefix(d.db)
	if err != nil {
		return err
	}

	if hasInactivePrefix {
		return errors.Errorf("cannot create staging consensus when a staging consensus already exists")
	}

	activePrefix, exists, err := prefixmanager.ActivePrefix(d.db)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("cannot create a staging consensus when there's " +
			"no active consensus")
	}

	inactivePrefix := activePrefix.Flip()
	err = prefixmanager.SetPrefixAsInactive(d.db, inactivePrefix)
	if err != nil {
		return err
	}

	consensusFactory := consensus.NewFactory()

	consensusInstance, shouldMigrate, err := consensusFactory.NewConsensus(cfg, d.db, inactivePrefix, d.consensusEventsChannel)
	if err != nil {
		return err
	}

	if shouldMigrate {
		return errors.Errorf("A fresh consensus should never return shouldMigrate=true")
	}

	d.stagingConsensus = &consensusInstance
	return nil
}

func (d *domain) CommitStagingConsensus() error {
	d.stagingConsensusLock.Lock()
	defer d.stagingConsensusLock.Unlock()

	dbTx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()

	inactivePrefix, hasInactivePrefix, err := prefixmanager.InactivePrefix(d.db)
	if err != nil {
		return err
	}

	if !hasInactivePrefix {
		return errors.Errorf("there's no inactive prefix to commit")
	}

	activePrefix, exists, err := prefixmanager.ActivePrefix(dbTx)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("cannot commit a staging consensus when there's " +
			"no active consensus")
	}

	err = prefixmanager.SetPrefixAsActive(dbTx, inactivePrefix)
	if err != nil {
		return err
	}

	err = prefixmanager.SetPrefixAsInactive(dbTx, activePrefix)
	if err != nil {
		return err
	}

	err = dbTx.Commit()
	if err != nil {
		return err
	}

	// We delete anything associated with the old prefix outside
	// of the transaction in order to save memory.
	err = prefixmanager.DeleteInactivePrefix(d.db)
	if err != nil {
		return err
	}

	tempConsensusPointer := unsafe.Pointer(d.stagingConsensus)
	consensusPointer := (*unsafe.Pointer)(unsafe.Pointer(&d.consensus))
	atomic.StorePointer(consensusPointer, tempConsensusPointer)
	d.stagingConsensus = nil
	return nil
}

func (d *domain) DeleteStagingConsensus() error {
	d.stagingConsensusLock.Lock()
	defer d.stagingConsensusLock.Unlock()

	err := prefixmanager.DeleteInactivePrefix(d.db)
	if err != nil {
		return err
	}

	d.stagingConsensus = nil
	return nil
}

// New instantiates a new instance of a Domain object
func New(consensusConfig *consensus.Config, mempoolConfig *mempool.Config, db infrastructuredatabase.Database) (Domain, error) {
	err := prefixmanager.DeleteInactivePrefix(db)
	if err != nil {
		return nil, err
	}

	activePrefix, exists, err := prefixmanager.ActivePrefix(db)
	if err != nil {
		return nil, err
	}

	if !exists {
		activePrefix = &prefix.Prefix{}
		err = prefixmanager.SetPrefixAsActive(db, activePrefix)
		if err != nil {
			return nil, err
		}
	}

	consensusEventsChan := make(chan externalapi.ConsensusEvent, 100e3)
	consensusFactory := consensus.NewFactory()
	consensusInstance, shouldMigrate, err := consensusFactory.NewConsensus(consensusConfig, db, activePrefix, consensusEventsChan)
	if err != nil {
		return nil, err
	}

	domainInstance := &domain{
		consensus:              &consensusInstance,
		consensusConfig:        consensusConfig,
		db:                     db,
		consensusEventsChannel: consensusEventsChan,
	}

	if shouldMigrate {
		err := domainInstance.migrate()
		if err != nil {
			return nil, err
		}
	}

	miningManagerFactory := miningmanager.NewFactory()

	// We create a consensus wrapper because the actual consensus might change
	consensusReference := consensusreference.NewConsensusReference(&domainInstance.consensus)
	domainInstance.miningManager = miningManagerFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempoolConfig)
	return domainInstance, nil
}
