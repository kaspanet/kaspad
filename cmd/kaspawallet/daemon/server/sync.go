package server

import (
	"fmt"
	"sort"
	"time"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

var keyChains = []uint8{libkaspawallet.ExternalKeychain, libkaspawallet.InternalKeychain}

type walletAddressSet map[string]*walletAddress

func (was walletAddressSet) strings() []string {
	addresses := make([]string, 0, len(was))
	for addr := range was {
		addresses = append(addresses, addr)
	}
	return addresses
}

func (s *server) syncLoop() error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	err := s.collectRecentAddresses()
	if err != nil {
		return err
	}

	err = s.refreshUTXOs()
	if err != nil {
		return err
	}

	s.firstSyncDone.Store(true)
	log.Infof("Wallet is synced and ready for operation")

	for {
		select {
		case <-ticker.C:
		case <-s.forceSyncChan:
		}

		err := s.sync()
		if err != nil {
			return err
		}
	}
}

func (s *server) sync() error {
	err := s.collectFarAddresses()
	if err != nil {
		return err
	}

	err = s.collectRecentAddresses()
	if err != nil {
		return err
	}

	return s.refreshUTXOs()
}

const (
	numIndexesToQueryForFarAddresses    = 100
	numIndexesToQueryForRecentAddresses = 1000
)

// addressesToQuery scans the addresses in the given range. Because
// each cosigner in a multisig has its own unique path for generating
// addresses it goes over all the cosigners and add their addresses
// for each key chain.
func (s *server) addressesToQuery(start, end uint32) (walletAddressSet, error) {
	addresses := make(walletAddressSet)
	for index := start; index < end; index++ {
		for cosignerIndex := uint32(0); cosignerIndex < uint32(len(s.keysFile.ExtendedPublicKeys)); cosignerIndex++ {
			for _, keychain := range keyChains {
				address := &walletAddress{
					index:         index,
					cosignerIndex: cosignerIndex,
					keyChain:      keychain,
				}
				addressString, err := s.walletAddressString(address)
				if err != nil {
					return nil, err
				}
				addresses[addressString] = address
			}
		}
	}

	return addresses, nil
}

// collectFarAddresses collects numIndexesToQueryForFarAddresses addresses
// from the last point it stopped in the previous call.
func (s *server) collectFarAddresses() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.collectAddresses(s.nextSyncStartIndex, s.nextSyncStartIndex+numIndexesToQueryForFarAddresses)
	if err != nil {
		return err
	}

	s.nextSyncStartIndex += numIndexesToQueryForFarAddresses
	return nil
}

func (s *server) maxUsedIndexWithLock() uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.maxUsedIndex()
}

func (s *server) maxUsedIndex() uint32 {
	maxUsedIndex := s.keysFile.LastUsedExternalIndex()
	if s.keysFile.LastUsedInternalIndex() > maxUsedIndex {
		maxUsedIndex = s.keysFile.LastUsedInternalIndex()
	}

	return maxUsedIndex
}

// collectRecentAddresses collects addresses from used addresses until
// the address with the index of the last used address + numIndexesToQueryForRecentAddresses.
// collectRecentAddresses scans addresses in batches of numIndexesToQueryForRecentAddresses,
// and releases the lock between scans.
func (s *server) collectRecentAddresses() error {
	index := uint32(0)
	maxUsedIndex := uint32(0)
	for ; index < maxUsedIndex+numIndexesToQueryForRecentAddresses; index += numIndexesToQueryForRecentAddresses {
		err := s.collectAddressesWithLock(index, index+numIndexesToQueryForRecentAddresses)

		if err != nil {
			return err
		}
		maxUsedIndex = s.maxUsedIndexWithLock()

		s.updateSyncingProgressLog(index, maxUsedIndex)
	}

	s.lock.Lock()
	if index > s.nextSyncStartIndex {
		s.nextSyncStartIndex = index
	}
	s.lock.Unlock()

	return nil
}

func (s *server) collectAddressesWithLock(start, end uint32) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.collectAddresses(start, end)
}

func (s *server) collectAddresses(start, end uint32) error {
	addressSet, err := s.addressesToQuery(start, end)
	if err != nil {
		return err
	}

	getBalancesByAddressesResponse, err := s.rpcClient.GetBalancesByAddresses(addressSet.strings())
	if err != nil {
		return err
	}

	err = s.updateAddressesAndLastUsedIndexes(addressSet, getBalancesByAddressesResponse)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) updateAddressesAndLastUsedIndexes(requestedAddressSet walletAddressSet,
	getBalancesByAddressesResponse *appmessage.GetBalancesByAddressesResponseMessage) error {
	lastUsedExternalIndex := s.keysFile.LastUsedExternalIndex()
	lastUsedInternalIndex := s.keysFile.LastUsedInternalIndex()

	for _, entry := range getBalancesByAddressesResponse.Entries {
		walletAddress, ok := requestedAddressSet[entry.Address]
		if !ok {
			return errors.Errorf("Got result from address %s even though it wasn't requested", entry.Address)
		}

		if entry.Balance == 0 {
			continue
		}

		s.addressSet[entry.Address] = walletAddress

		if walletAddress.keyChain == libkaspawallet.ExternalKeychain {
			if walletAddress.index > lastUsedExternalIndex {
				lastUsedExternalIndex = walletAddress.index
			}
			continue
		}

		if walletAddress.index > lastUsedInternalIndex {
			lastUsedInternalIndex = walletAddress.index
		}
	}

	err := s.keysFile.SetLastUsedExternalIndex(lastUsedExternalIndex)
	if err != nil {
		return err
	}

	return s.keysFile.SetLastUsedInternalIndex(lastUsedInternalIndex)
}

// updateUTXOSet clears the current UTXO set, and re-fills it with the given entries
func (s *server) updateUTXOSet(entries []*appmessage.UTXOsByAddressesEntry, mempoolEntries []*appmessage.MempoolEntryByAddress, refreshStart time.Time) error {
	utxos := make([]*walletUTXO, 0, len(entries))

	exclude := make(map[appmessage.RPCOutpoint]struct{})
	for _, entriesByAddress := range mempoolEntries {
		for _, entry := range entriesByAddress.Sending {
			for _, input := range entry.Transaction.Inputs {
				exclude[*input.PreviousOutpoint] = struct{}{}
			}
		}
	}

	for _, entry := range entries {
		if _, ok := exclude[*entry.Outpoint]; ok {
			continue
		}

		outpoint, err := appmessage.RPCOutpointToDomainOutpoint(entry.Outpoint)
		if err != nil {
			return err
		}

		utxoEntry, err := appmessage.RPCUTXOEntryToUTXOEntry(entry.UTXOEntry)
		if err != nil {
			return err
		}

		address, ok := s.addressSet[entry.Address]
		if !ok {
			return errors.Errorf("Got result from address %s even though it wasn't requested", entry.Address)
		}
		utxos = append(utxos, &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
			address:   address,
		})
	}

	sort.Slice(utxos, func(i, j int) bool { return utxos[i].UTXOEntry.Amount() > utxos[j].UTXOEntry.Amount() })

	s.lock.Lock()
	s.startTimeOfLastCompletedRefresh = refreshStart
	s.utxosSortedByAmount = utxos
	s.lock.Unlock()

	return nil
}

func (s *server) refreshUTXOs() error {
	refreshStart := time.Now()
	s.lock.Lock()
	addresses := s.addressSet.strings()
	s.lock.Unlock()
	// It's important to check the mempool before calling `GetUTXOsByAddresses`:
	// If we would do it the other way around an output can be spent in the mempool
	// and not in consensus, and between the calls its spending transaction will be
	// added to consensus and removed from the mempool, so `getUTXOsByAddressesResponse`
	// will include an obsolete output.
	mempoolEntriesByAddresses, err := s.rpcClient.GetMempoolEntriesByAddresses(addresses, true, true)
	if err != nil {
		return err
	}

	getUTXOsByAddressesResponse, err := s.rpcClient.GetUTXOsByAddresses(addresses)
	if err != nil {
		return err
	}

	return s.updateUTXOSet(getUTXOsByAddressesResponse.Entries, mempoolEntriesByAddresses.Entries, refreshStart)
}

func (s *server) forceSync() {
	// Technically if two callers check the `if` simultaneously they will both spawn a
	// goroutine, but we don't care about the small redundancy in such a rare case.
	if len(s.forceSyncChan) == 0 {
		go func() {
			s.forceSyncChan <- struct{}{}
		}()
	}
}

func (s *server) isSynced() bool {
	return s.nextSyncStartIndex > s.maxUsedIndex() && s.firstSyncDone.Load()
}

func (s *server) formatSyncStateReport() string {
	maxUsedIndex := s.maxUsedIndex()

	if s.nextSyncStartIndex > maxUsedIndex {
		maxUsedIndex = s.nextSyncStartIndex
	}

	if s.nextSyncStartIndex < s.maxUsedIndex() {
		return fmt.Sprintf("scanned %d out of %d addresses (%.2f%%)",
			s.nextSyncStartIndex, maxUsedIndex, float64(s.nextSyncStartIndex)*100.0/float64(maxUsedIndex))
	}
	return "loading the wallet UTXO set"
}

func (s *server) updateSyncingProgressLog(currProcessedAddresses, currMaxUsedAddresses uint32) {
	if currMaxUsedAddresses > s.maxUsedAddressesForLog {
		s.maxUsedAddressesForLog = currMaxUsedAddresses
		if s.isLogFinalProgressLineShown {
			log.Infof("An additional set of previously used addresses found, processing...")
			s.maxProcessedAddressesForLog = 0
			s.isLogFinalProgressLineShown = false
		}
	}

	if currProcessedAddresses > s.maxProcessedAddressesForLog {
		s.maxProcessedAddressesForLog = currProcessedAddresses
	}

	if s.maxProcessedAddressesForLog >= s.maxUsedAddressesForLog {
		if !s.isLogFinalProgressLineShown {
			log.Infof("Finished scanning recent addresses")
			s.isLogFinalProgressLineShown = true
		}
	} else {
		percentProcessed := float64(s.maxProcessedAddressesForLog) / float64(s.maxUsedAddressesForLog) * 100.0

		log.Infof("%d addresses of %d processed (%.2f%%)...",
			s.maxProcessedAddressesForLog, s.maxUsedAddressesForLog, percentProcessed)
	}
}
