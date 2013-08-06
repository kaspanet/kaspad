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

func (s *server) onChainChanged(notification *appmessage.VirtualSelectedParentChainChangedNotificationMessage) {
	for _, transactionIDs := range notification.AcceptedTransactionIDs{
		for _, transactionID := range transactionIDs.AcceptedTransactionIDs {
			if s.tracker.isTransactionIDTracked(transactionID) {
				s.tracker.untrackSentTransactionID(transactionID)
			}
		}
	}
}

func (s *server) sync() error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	err := s.collectRecentAddresses()
	if err != nil {
		return err
	}

	err = s.refreshExistingUTXOsWithLock()
	if err != nil {
		return err
	}

	err = s.rpcClient.RegisterForVirtualSelectedParentChainChangedNotifications(true, s.onChainChanged)
	if err != nil {
		return err
	}

	for i := range ticker.C {
		fmt.Println(i)
		err = s.collectFarAddresses()
		if err != nil {
			return err
		}

		err = s.collectRecentAddresses()
		if err != nil {
			return err
		}

		err = s.refreshExistingUTXOsWithLock()
		if err != nil {
			return err
		}
	}

	return nil
}

const numIndexesToQueryForFarAddresses = 100
const numIndexesToQueryForRecentAddresses = 1000

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

func (s *server) maxUsedIndex() uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()

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
		maxUsedIndex = s.maxUsedIndex()
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

func (s *server) refreshExistingUTXOsWithLock() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.refreshUTXOs()
}

// updateUTXOSet clears the current UTXO set, and re-fills it with the given entries
func (s *server) updateUTXOSet(entries []*appmessage.UTXOsByAddressesEntry) error {

	utxos := make([]*walletUTXO, len(entries))

	s.tracker.untrackExpiredOutpointsAsReserved() //untrack all stale reserved outpoints, before comparing in loop
	availableUtxos := make([]*walletUTXO, 0)
	for i, entry := range entries {
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
		utxos[i] = &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
			address:   address,
		}

		if s.tracker.isOutpointAvailable(outpoint) {
			availableUtxos = append(availableUtxos, &walletUTXO{
				Outpoint:  outpoint,
				UTXOEntry: utxoEntry,
				address:   address,
			})
		}
	}

	sort.Slice(utxos, func(i, j int) bool { return utxos[i].UTXOEntry.Amount() > utxos[j].UTXOEntry.Amount() })

	s.utxosSortedByAmount = utxos

	sort.Slice(availableUtxos, func(i, j int) bool {
		return availableUtxos[i].UTXOEntry.Amount() > availableUtxos[j].UTXOEntry.Amount()
	})

	s.availableUtxosSortedByAmount = availableUtxos

	fmt.Println("utxos total", len(s.utxosSortedByAmount))
	fmt.Println("utxos available", len(s.availableUtxosSortedByAmount))
	fmt.Println("utxos reserved", len(s.tracker.reservedOutpoints))
	fmt.Println("transactions in mempool", len(s.tracker.sentTransactions))

	s.tracker.untrackOutpointDifferenceViaWalletUTXOs(utxos) //clean up reserved tracker

	return nil
}

func (s *server) refreshUTXOs() error {
	getUTXOsByAddressesResponse, err := s.rpcClient.GetUTXOsByAddresses(s.addressSet.strings())
	if err != nil {
		return err
	}

	return s.updateUTXOSet(getUTXOsByAddressesResponse.Entries)
}

func (s *server) isSynced() bool {
	return s.nextSyncStartIndex > s.keysFile.LastUsedInternalIndex() && s.nextSyncStartIndex > s.keysFile.LastUsedExternalIndex()
}
