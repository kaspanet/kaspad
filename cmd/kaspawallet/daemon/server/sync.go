package server

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"time"
)

const (
	externalKeychain = 0
	internalKeychain = 1
)

var keyChains = []uint8{externalKeychain, internalKeychain}

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

// collectUTXOsFromFarAddresses collects UTXOs
// from s.nextSyncStartIndex to s.nextSyncStartIndex+numIndexesToQuery
// and increases s.nextSyncStartIndex to the last address it scanned.
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

	maxUsedIndex := s.keysFile.LastUsedExternalIndex
	if s.keysFile.LastUsedInternalIndex > maxUsedIndex {
		maxUsedIndex = s.keysFile.LastUsedInternalIndex
	}

	return maxUsedIndex
}

// collectUTXOsFromRecentAddresses collects UTXOs from used addresses until
// the address with the index of the last used address + 1000.
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

	for _, entry := range getUTXOsByAddressesResponse.Entries {
		walletAddress, ok := addressSet[entry.Address]
		if !ok {
			continue
		}

		if walletAddress.cosignerIndex != s.keysFile.CosignerIndex {
			continue
		}

		if walletAddress.keyChain == externalKeychain {
			if walletAddress.index > s.keysFile.LastUsedExternalIndex {
				s.keysFile.LastUsedExternalIndex = walletAddress.index
			}
			continue
		}

		if walletAddress.index > s.keysFile.LastUsedInternalIndex {
			s.keysFile.LastUsedInternalIndex = walletAddress.index
		}
	}

	for _, entry := range getUTXOsByAddressesResponse.Entries {
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

		s.utxos[*outpoint] = &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
			address:   address,
		}
	}

	// Save the file after changes in LastUsedExternalIndex and LastUsedInternalIndex
	err = s.keysFile.Sync(true)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) refreshExistingUTXOsWithLock() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.refreshExistingUTXOs()
}

func (s *server) refreshExistingUTXOs() error {
	addressSet := make(walletAddressSet, len(s.utxos))
	for _, utxo := range s.utxos {
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

	s.utxos = make(map[externalapi.DomainOutpoint]*walletUTXO, len(getUTXOsByAddressesResponse.Entries))
	for _, entry := range getUTXOsByAddressesResponse.Entries {
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

		s.utxos[*outpoint] = &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
			address:   address,
		}
	}
	return nil
}

func (s *server) validateIsSynced() error {
	isSynced := s.nextSyncStartIndex > s.keysFile.LastUsedInternalIndex && s.nextSyncStartIndex > s.keysFile.LastUsedExternalIndex
	if !isSynced {
		return errors.New("server is not synced")
	}

	return nil
}
