package server

import (
	"fmt"
	"sort"
	"time"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

var keyChains = []uint8{libkaspawallet.ExternalKeychain, libkaspawallet.InternalKeychain}
type mempoolTransactionsMap map[string]bool

func (s *server) utxosSortedByAmount() []*walletUTXO {
	utxos := make([]*walletUTXO, len(s.utxoSet))
	i := 0
	for _, walletUtxo := range s.utxoSet {
		utxos[i] = walletUtxo
		i = i + 1
	}

	sort.Slice(utxos, func(i, j int) bool { return utxos[i].UTXOEntry.Amount() > utxos[j].UTXOEntry.Amount() })
	return utxos
}

func (s *server) availableUtxosSortedByAmount() []*walletUTXO {
	utxos := make([]*walletUTXO, 0)
	for _, walletUtxo := range s.utxoSet {
		if s.tracker.isOutpointAvailable(walletUtxo.Outpoint) {
			utxos = append(utxos, walletUtxo)
		}
	}

	sort.Slice(utxos, func(i, j int) bool { return utxos[i].UTXOEntry.Amount() > utxos[j].UTXOEntry.Amount() })
	return utxos
}

func (was walletAddressSet) strings() []string {
	addresses := make([]string, 0, len(was))
	for addr := range was {
		addresses = append(addresses, addr)
	}
	return addresses
}

func (s *server) intialize() error {
	err := s.collectRecentAddresses()
	if err != nil {
		return err
	}

	err = s.update()
	if err != nil {
		return err
	}

	err = s.trackMempool()
	if err != nil {
		return err
	}
	return nil
}

func (s *server) sync() error {
	err := s.intialize()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := s.collectFarAddresses()
		if err != nil {
			return err
		}

		err = s.collectRecentAddresses()
		if err != nil {
			return err
		}
		
		err = s.update()
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
// the address with the index of the last used address + 1000.
// collectRecentAddresses scans addresses in batches of numIndexesToQuery,
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

// updateUTXOSet clears the current UTXO set, and re-fills it with the given entries
func (s *server) refreshUTXOs(utxosByAddresses []*appmessage.UTXOsByAddressesEntry) error {

	newWalletUTXOSet := make(walletUTXOSet, len(utxosByAddresses))

	for _, utxosByAddress := range utxosByAddresses {
		outpoint, err := appmessage.RPCOutpointToDomainOutpoint(utxosByAddress.Outpoint)
		if err != nil {
			return err
		}

		utxoEntry, err := appmessage.RPCUTXOEntryToUTXOEntry(utxosByAddress.UTXOEntry)
		if err != nil {
			return err
		}
		walletAddress, found := s.addressSet[utxosByAddress.Address]
		if !found {
			return errors.Errorf("Got result from address %s even though it wasn't requested", utxosByAddress.Address)

		}
		newWalletUTXOSet[*outpoint] = &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
			address:   walletAddress,
		}
	}
	s.utxoSet = newWalletUTXOSet

	fmt.Println("total", len(s.utxoSet))
	fmt.Println("reserved", len(s.tracker.reservedOutpoints))
	fmt.Println("followed txIDS", len(s.tracker.sentTransactions))
	fmt.Println("UtxosSent", s.tracker.countOutpointsInmempool())

	return nil
}

func (s *server) update() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	
	err := s.untrackMempoolTransactions()
	if err != nil {
		return err
	}
	err = s.collectAndRefreshUTXOs()
	if err != nil {
		return err
	}

	s.tracker.untrackExpiredOutpointsAsReserved()

	return nil
}

func (s *server) untrackMempoolTransactions() error {
	getMempoolEntriesResponse, err := s.rpcClient.GetMempoolEntries()
	if err != nil {
		return err
	}
	if getMempoolEntriesResponse.Error != nil {
		return errors.Errorf(getMempoolEntriesResponse.Error.Message)
	}
	mapMempoolTransactions := make(mempoolTransactionsMap)
	for _, mempoolEntry := range getMempoolEntriesResponse.Entries {
		transaction, err := appmessage.RPCTransactionToDomainTransaction(mempoolEntry.Transaction)
		if err != nil {
			return err
		}
		if transaction.ID != nil{
			fmt.Println(transaction.ID.String())
			mapMempoolTransactions[transaction.ID.String()] = true
		}
	}

	for transactionID := range s.tracker.sentTransactions {
		if _, found := mapMempoolTransactions[transactionID]; !found {
			s.tracker.untrackSentTransactionID(transactionID)
		}
	}
	return nil
}

func (s *server) trackMempool() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	getMempoolEntriesResponse, err := s.rpcClient.GetMempoolEntries()
	if err != nil {
		return err
	}
	if getMempoolEntriesResponse.Error != nil {
		return errors.Errorf(getMempoolEntriesResponse.Error.Message)
	}
	for _, mempoolEntry := range getMempoolEntriesResponse.Entries {
		transaction, err := appmessage.RPCTransactionToDomainTransaction(mempoolEntry.Transaction)
		if err != nil {
			return err
		}
		for _, input := range transaction.Inputs {
			_, address, err := txscript.ExtractScriptPubKeyAddress(input.UTXOEntry.ScriptPublicKey(), s.params)
			if err != nil {
				return err
			}
			if _, found := s.addressSet[address.String()]; found {
				s.tracker.trackTransaction(transaction)
				break
			}
		}
	}

	return nil

}

func (s *server) collectAndRefreshUTXOs() error {
	getUTXOsByAddressesResponse, err := s.rpcClient.GetUTXOsByAddresses(s.addressSet.strings())
	if err != nil {
		return err
	}
	return s.refreshUTXOs(getUTXOsByAddressesResponse.Entries)
}

func (s *server) isSynced() bool {
	return s.nextSyncStartIndex > s.keysFile.LastUsedInternalIndex() && s.nextSyncStartIndex > s.keysFile.LastUsedExternalIndex()
}
