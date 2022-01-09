package server

import (
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

func (s *server) sync() error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := s.collectUTXOsFromRecentAddresses()
		if err != nil {
			return err
		}

		err = s.collectUTXOsFromFarAddresses()
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

const numIndexesToQuery = 100

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

// collectUTXOsFromFarAddresses collects numIndexesToQuery UTXOs
// from the last point it stopped in the previous call.
func (s *server) collectUTXOsFromFarAddresses() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.collectUTXOs(s.nextSyncStartIndex, s.nextSyncStartIndex+numIndexesToQuery)
	if err != nil {
		return err
	}

	s.nextSyncStartIndex += numIndexesToQuery
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

// collectUTXOsFromRecentAddresses collects UTXOs from used addresses until
// the address with the index of the last used address + 1000.
// collectUTXOsFromRecentAddresses scans addresses in batches of numIndexesToQuery,
// and releases the lock between scans.
func (s *server) collectUTXOsFromRecentAddresses() error {
	maxUsedIndex := s.maxUsedIndex()
	for i := uint32(0); i < maxUsedIndex+1000; i += numIndexesToQuery {
		err := s.collectUTXOsWithLock(i, i+numIndexesToQuery)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) collectUTXOsWithLock(start, end uint32) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.collectUTXOs(start, end)
}

func (s *server) collectUTXOs(start, end uint32) error {
	addressSet, err := s.addressesToQuery(start, end)
	if err != nil {
		return err
	}

	getUTXOsByAddressesResponse, err := s.rpcClient.GetUTXOsByAddresses(addressSet.strings())
	if err != nil {
		return err
	}

	err = s.updateLastUsedIndexes(addressSet, getUTXOsByAddressesResponse)
	if err != nil {
		return err
	}

	err = s.updateUTXOs(addressSet, getUTXOsByAddressesResponse)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) updateUTXOs(addressSet walletAddressSet,
	getUTXOsByAddressesResponse *appmessage.GetUTXOsByAddressesResponseMessage) error {

	return s.addEntriesToUTXOSet(getUTXOsByAddressesResponse.Entries, addressSet)
}

func (s *server) updateLastUsedIndexes(addressSet walletAddressSet,
	getUTXOsByAddressesResponse *appmessage.GetUTXOsByAddressesResponseMessage) error {

	lastUsedExternalIndex := s.keysFile.LastUsedExternalIndex()
	lastUsedInternalIndex := s.keysFile.LastUsedInternalIndex()

	for _, entry := range getUTXOsByAddressesResponse.Entries {
		walletAddress, ok := addressSet[entry.Address]
		if !ok {
			return errors.Errorf("Got result from address %s even though it wasn't requested", entry.Address)
		}

		if walletAddress.cosignerIndex != s.keysFile.CosignerIndex {
			continue
		}

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

	return s.refreshExistingUTXOs()
}

func (s *server) addEntriesToUTXOSet(entries []*appmessage.UTXOsByAddressesEntry, addressSet walletAddressSet) error {
	utxos := make([]*walletUTXO, len(entries))
	for i, entry := range entries {
		outpoint, err := appmessage.RPCOutpointToDomainOutpoint(entry.Outpoint)
		if err != nil {
			return err
		}

		utxoEntry, err := appmessage.RPCUTXOEntryToUTXOEntry(entry.UTXOEntry)
		if err != nil {
			return err
		}

		address, ok := addressSet[entry.Address]
		if !ok {
			return errors.Errorf("Got result from address %s even though it wasn't requested", entry.Address)
		}
		utxos[i] = &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
			address:   address,
		}
	}

	sort.Slice(utxos, func(i, j int) bool { return utxos[i].UTXOEntry.Amount() > utxos[i].UTXOEntry.Amount() })

	s.utxosSortedByAmount = mergeUTXOSlices(s.utxosSortedByAmount, utxos)

	return nil
}

func mergeUTXOSlices(left, right []*walletUTXO) []*walletUTXO {
	result := make([]*walletUTXO, len(left)+len(right))

	i := 0
	for len(left) > 0 && len(right) > 0 {
		if left[0].UTXOEntry.Amount() > right[0].UTXOEntry.Amount() {
			result[i] = left[0]
			left = left[1:]
		} else {
			result[i] = right[0]
			right = right[1:]
		}
		i++
	}

	for j := 0; j < len(left); j++ {
		result[i] = left[j]
		i++
	}
	for j := 0; j < len(right); j++ {
		result[i] = right[j]
		i++
	}

	return result
}

// insertUTXO inserts the given utxo into s.utxosSortedByAmount, while keeping it sorted.
func (s *server) insertUTXO(utxo *walletUTXO) {
	s.utxosSortedByAmount = append(s.utxosSortedByAmount, utxo)
	// bubble up the new UTXO to keep the UTXOs sorted by value
	index := len(s.utxosSortedByAmount) - 1
	for index > 0 && utxo.UTXOEntry.Amount() > s.utxosSortedByAmount[index-1].UTXOEntry.Amount() {
		s.utxosSortedByAmount[index] = s.utxosSortedByAmount[index-1]
		index--
	}
	s.utxosSortedByAmount[index] = utxo
}

func (s *server) refreshExistingUTXOs() error {
	addressSet := make(walletAddressSet, len(s.utxosSortedByAmount))
	for _, utxo := range s.utxosSortedByAmount {
		addressString, err := s.walletAddressString(utxo.address)
		if err != nil {
			return err
		}

		addressSet[addressString] = utxo.address
	}

	getUTXOsByAddressesResponse, err := s.rpcClient.GetUTXOsByAddresses(addressSet.strings())
	if err != nil {
		return err
	}

	s.utxosSortedByAmount = make([]*walletUTXO, 0)

	return s.addEntriesToUTXOSet(getUTXOsByAddressesResponse.Entries, addressSet)
}

func (s *server) isSynced() bool {
	return s.nextSyncStartIndex > s.keysFile.LastUsedInternalIndex() && s.nextSyncStartIndex > s.keysFile.LastUsedExternalIndex()
}
