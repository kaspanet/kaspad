package server

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"time"
)

const (
	externalKeychain = 0
	internalKeychain = 1
)

var keyChains = []uint8{externalKeychain, internalKeychain}

func (s *server) sync() error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := s.collectUTXOs()
		if err != nil {
			return err
		}

		err = s.refreshExistingUTXOsWithLock()
		if err != nil {
			return err
		}

		err = s.syncKeysFile()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) syncKeysFile() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.keysFile.Sync(true)
}

type walletUTXO struct {
	Outpoint  *externalapi.DomainOutpoint
	UTXOEntry externalapi.UTXOEntry
	address   *walletAddress
}

type walletAddress struct {
	address       string
	index         uint32
	cosignerIndex uint32
	keyChain      uint8
}

func (wa *walletAddress) path() string {
	return fmt.Sprintf("m/%d/%d/%d", wa.cosignerIndex, wa.keyChain, wa.index)
}

type walletAddressSet map[string]*walletAddress

func (was walletAddressSet) strings() []string {
	addresses := make([]string, 0, len(was))
	for addr := range was {
		addresses = append(addresses, addr)
	}
	return addresses
}

const numAddressToQuery = 100

func (s *server) addressesToQuery() (walletAddressSet, error) {
	addresses := make(walletAddressSet, numAddressToQuery)
	for index := s.nextSyncStartIndex; len(addresses) < numAddressToQuery; index++ {
		for cosignerIndex := uint32(0); cosignerIndex < uint32(len(s.keysFile.ExtendedPublicKeys)); cosignerIndex++ {
			for _, keychain := range keyChains {
				path := fmt.Sprintf("m/%d/%d/%d", cosignerIndex, keychain, index)
				addr, err := libkaspawallet.Address(s.params, s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, path, s.keysFile.ECDSA)
				if err != nil {
					return nil, err
				}
				addresses[addr.String()] = &walletAddress{
					address:       addr.String(),
					index:         index,
					cosignerIndex: cosignerIndex,
					keyChain:      keychain,
				}
			}
		}
	}

	return addresses, nil
}

func (s *server) collectUTXOs() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	addressSet, err := s.addressesToQuery()
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
			return errors.Errorf("Got result from address %s even though it wasn't requested", address.address)
		}

		s.utxos[*outpoint] = &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
			address:   address,
		}
	}

	s.nextSyncStartIndex += numAddressToQuery
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
		addressSet[utxo.address.address] = utxo.address
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
			return errors.Errorf("Got result from address %s even though it wasn't requested", address.address)
		}

		s.utxos[*outpoint] = &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
			address:   address,
		}
	}
	return nil
}

func (s *server) IsSynced(_ context.Context, _ *pb.IsSyncedRequest) (*pb.IsSyncedResponse, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	isSynced := s.nextSyncStartIndex > s.keysFile.LastUsedInternalIndex && s.nextSyncStartIndex > s.keysFile.LastUsedExternalIndex
	return &pb.IsSyncedResponse{IsSynced: isSynced}, nil
}
