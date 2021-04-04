// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addressmanager

import (
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/util/mstime"
	"net"
	"sync"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

const maxAddresses = 4096

// addressRandomizer is the interface for the randomizer needed for the AddressManager.
type addressRandomizer interface {
	RandomAddress(addresses []*appmessage.NetAddress) *appmessage.NetAddress
	RandomAddresses(addresses []*appmessage.NetAddress, count int) []*appmessage.NetAddress
}

// addressKey represents a pair of IP and port, the IP is always in V6 representation
type addressKey struct {
	port    uint16
	address ipv6
}

type address struct {
	netAddress            *appmessage.NetAddress
	connectionFailedCount uint64
}

type ipv6 [net.IPv6len]byte

func (i ipv6) equal(other ipv6) bool {
	return i == other
}

// ErrAddressNotFound is an error returned from some functions when a
// given address is not found in the address manager
var ErrAddressNotFound = errors.New("address not found")

// NetAddressKey returns a key of the ip address to use it in maps.
func netAddressKey(netAddress *appmessage.NetAddress) addressKey {
	key := addressKey{port: netAddress.Port}
	// all IPv4 can be represented as IPv6.
	copy(key.address[:], netAddress.IP.To16())
	return key
}

// AddressManager provides a concurrency safe address manager for caching potential
// peers on the Kaspa network.
type AddressManager struct {
	store          *addressStore
	localAddresses *localAddressManager
	mutex          sync.Mutex
	cfg            *Config
	random         addressRandomizer
}

// New returns a new Kaspa address manager.
func New(cfg *Config, database database.Database) (*AddressManager, error) {
	addressStore, err := newAddressStore(database)
	if err != nil {
		return nil, err
	}
	localAddresses, err := newLocalAddressManager(cfg)
	if err != nil {
		return nil, err
	}

	return &AddressManager{
		store:          addressStore,
		localAddresses: localAddresses,
		random:         NewAddressRandomize(),
		cfg:            cfg,
	}, nil
}

func (am *AddressManager) addAddressNoLock(netAddress *appmessage.NetAddress) error {
	if !IsRoutable(netAddress, am.cfg.AcceptUnroutable) {
		return nil
	}

	key := netAddressKey(netAddress)
	address := &address{netAddress: netAddress, connectionFailedCount: 0}
	err := am.store.add(key, address)
	if err != nil {
		return err
	}

	if am.store.notBannedCount() > maxAddresses {
		allAddresses := am.store.getAllNotBanned()

		maxConnectionFailedCount := uint64(0)
		toRemove := allAddresses[0]
		for _, address := range allAddresses[1:] {
			if address.connectionFailedCount > maxConnectionFailedCount {
				maxConnectionFailedCount = address.connectionFailedCount
				toRemove = address
			}
		}

		toRemoveKey := netAddressKey(toRemove.netAddress)
		err := am.store.remove(toRemoveKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (am *AddressManager) removeAddressNoLock(address *appmessage.NetAddress) error {
	key := netAddressKey(address)
	return am.store.remove(key)
}

// AddAddress adds address to the address manager
func (am *AddressManager) AddAddress(address *appmessage.NetAddress) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.addAddressNoLock(address)
}

// AddAddresses adds addresses to the address manager
func (am *AddressManager) AddAddresses(addresses ...*appmessage.NetAddress) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, address := range addresses {
		err := am.addAddressNoLock(address)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveAddress removes addresses from the address manager
func (am *AddressManager) RemoveAddress(address *appmessage.NetAddress) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.removeAddressNoLock(address)
}

// MarkConnectionFailure notifies the address manager that the given address
// has failed to connect
func (am *AddressManager) MarkConnectionFailure(address *appmessage.NetAddress) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	entry, ok := am.store.getNotBanned(key)
	if !ok {
		return errors.Errorf("address %s is not registered with the address manager", address.TCPAddress())
	}
	entry.connectionFailedCount = entry.connectionFailedCount + 1
	return am.store.updateNotBanned(key, entry)
}

// MarkConnectionSuccess notifies the address manager that the given address
// has successfully connected
func (am *AddressManager) MarkConnectionSuccess(address *appmessage.NetAddress) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	entry, ok := am.store.getNotBanned(key)
	if !ok {
		return errors.Errorf("address %s is not registered with the address manager", address.TCPAddress())
	}
	entry.connectionFailedCount = 0
	return am.store.updateNotBanned(key, entry)
}

// Addresses returns all addresses
func (am *AddressManager) Addresses() []*appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.store.getAllNotBannedNetAddresses()
}

// BannedAddresses returns all banned addresses
func (am *AddressManager) BannedAddresses() []*appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.store.getAllBannedNetAddresses()
}

// notBannedAddressesWithException returns all not banned addresses with excpetion
func (am *AddressManager) notBannedAddressesWithException(exceptions []*appmessage.NetAddress) []*appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.store.getAllNotBannedNetAddressesWithout(exceptions)
}

// RandomAddress returns a random address that isn't banned and isn't in exceptions
func (am *AddressManager) RandomAddress(exceptions []*appmessage.NetAddress) *appmessage.NetAddress {
	validAddresses := am.notBannedAddressesWithException(exceptions)
	return am.random.RandomAddress(validAddresses)
}

// RandomAddresses returns count addresses at random that aren't banned and aren't in exceptions
func (am *AddressManager) RandomAddresses(count int, exceptions []*appmessage.NetAddress) []*appmessage.NetAddress {
	validAddresses := am.notBannedAddressesWithException(exceptions)
	return am.random.RandomAddresses(validAddresses, count)
}

// BestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (am *AddressManager) BestLocalAddress(remoteAddress *appmessage.NetAddress) *appmessage.NetAddress {
	return am.localAddresses.bestLocalAddress(remoteAddress)
}

// Ban marks the given address as banned
func (am *AddressManager) Ban(addressToBan *appmessage.NetAddress) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	keyToBan := netAddressKey(addressToBan)
	keysToDelete := make([]addressKey, 0)
	for _, address := range am.store.getAllNotBannedNetAddresses() {
		key := netAddressKey(address)
		if key.address.equal(keyToBan.address) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	for _, key := range keysToDelete {
		err := am.store.remove(key)
		if err != nil {
			return err
		}
	}

	address := &address{netAddress: addressToBan}
	return am.store.addBanned(keyToBan, address)
}

// Unban unmarks the given address as banned
func (am *AddressManager) Unban(address *appmessage.NetAddress) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	if !am.store.isBanned(key) {
		return errors.Wrapf(ErrAddressNotFound, "address %s "+
			"is not registered with the address manager as banned", address.TCPAddress())
	}

	return am.store.removeBanned(key)
}

// IsBanned returns true if the given address is marked as banned
func (am *AddressManager) IsBanned(address *appmessage.NetAddress) (bool, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	err := am.unbanIfOldEnough(key)
	if err != nil {
		return false, err
	}
	if !am.store.isBanned(key) {
		if !am.store.isNotBanned(key) {
			return false, errors.Wrapf(ErrAddressNotFound, "address %s "+
				"is not registered with the address manager", address.TCPAddress())
		}
		return false, nil
	}

	return true, nil
}

func (am *AddressManager) unbanIfOldEnough(key addressKey) error {
	address, ok := am.store.getBanned(key)
	if !ok {
		return nil
	}

	const maxBanTime = 24 * time.Hour
	if mstime.Since(address.netAddress.Timestamp) > maxBanTime {
		err := am.store.removeBanned(key)
		if err != nil {
			return err
		}
	}
	return nil
}
