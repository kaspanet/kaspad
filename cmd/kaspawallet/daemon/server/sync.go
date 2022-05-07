package server

import (
	"fmt"
	"sort"
	"time"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"

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

	for i := range ticker.C {
		fmt.Println(i)
		err := s.collectRecentAddresses()
		if err != nil {
			return err
		}

		err = s.collectFarAddresses()
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

// collectFarAddresses collects numIndexesToQuery addresses
// from the last point it stopped in the previous call.
func (s *server) collectFarAddresses() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.collectAddresses(s.nextSyncStartIndex, s.nextSyncStartIndex+numIndexesToQuery)
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

// collectRecentAddresses collects addresses from used addresses until
// the address with the index of the last used address + 1000.
// collectRecentAddresses scans addresses in batches of numIndexesToQuery,
// and releases the lock between scans.
func (s *server) collectRecentAddresses() error {
	maxUsedIndex := s.maxUsedIndex()
	for i := uint32(0); i < maxUsedIndex+1000; i += numIndexesToQuery {
		err := s.collectAddressesWithLock(i, i+numIndexesToQuery)
		if err != nil {
			return err
		}
	}

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

	getMemepoolEntriesResponse, err := s.rpcClient.GetMempoolEntries()
	if err != nil {
		return err
	}

	mempoolWalletAddressesOutpoints := make(mempoolOutpoints)
	mempoolTransactions := make([]*externalapi.DomainTransaction, 0)

	for _, memepoolEntry := range getMemepoolEntriesResponse.Entries {
		transaction, err := appmessage.RPCTransactionToDomainTransaction(memepoolEntry.Transaction)
		if err != nil {
			return err
		}
		mempoolTransactions = append(
			mempoolTransactions,
			transaction,
		)
		if s.tracker.isTransactionTracked(transaction) {
			for _, input := range transaction.Inputs {
				scriptPubKey := input.UTXOEntry.ScriptPublicKey()
				_, address, err := txscript.ExtractScriptPubKeyAddress(scriptPubKey, s.params)
				if err != nil {
					return err
				}
				if _, found := s.addressSet[address.String()]; found {
					mempoolWalletAddressesOutpoints[input.PreviousOutpoint] = true
				}
			}
		}
	}

	s.tracker.untrackTransactionDifference(mempoolTransactions)   //clean up transaction tracker
	s.tracker.mempoolOutpoints = mempoolWalletAddressesOutpoints //clean up sent outpoint tracker

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
	fmt.Println("utxos mempool", len(s.tracker.mempoolOutpoints))
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
