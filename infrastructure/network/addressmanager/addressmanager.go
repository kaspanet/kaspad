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

// netAddressKeys returns a key of the ip address to use it in maps.
func netAddressesKeys(netAddresses []*appmessage.NetAddress) map[addressKey]bool {
	result := make(map[addressKey]bool, len(netAddresses))
	for _, netAddress := range netAddresses {
		key := netAddressKey(netAddress)
		result[key] = true
	}

	return result
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
	localAddresses, err := newLocalAddressManager(cfg)
	if err != nil {
		return nil, err
	}

	return &AddressManager{
		store:          newAddressStore(database),
		localAddresses: localAddresses,
		random:         NewAddressRandomize(),
		cfg:            cfg,
	}, nil
}

func (am *AddressManager) addAddressNoLock(address *appmessage.NetAddress) {
	if !IsRoutable(address, am.cfg.AcceptUnroutable) {
		return
	}

	key := netAddressKey(address)
	am.store.add(key, address)
}

// AddAddress adds address to the address manager
func (am *AddressManager) AddAddress(address *appmessage.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.addAddressNoLock(address)
}

// AddAddresses adds addresses to the address manager
func (am *AddressManager) AddAddresses(addresses ...*appmessage.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, address := range addresses {
		am.addAddressNoLock(address)
	}
}

// RemoveAddress removes addresses from the address manager
func (am *AddressManager) RemoveAddress(address *appmessage.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	am.store.remove(key)
}

// Addresses returns all addresses
func (am *AddressManager) Addresses() []*appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.store.getAllNotBanned()
}

// BannedAddresses returns all banned addresses
func (am *AddressManager) BannedAddresses() []*appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.store.getAllBanned()
}

// notBannedAddressesWithException returns all not banned addresses with excpetion
func (am *AddressManager) notBannedAddressesWithException(exceptions []*appmessage.NetAddress) []*appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.store.getAllNotBannedWithout(exceptions)
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
func (am *AddressManager) Ban(addressToBan *appmessage.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	keyToBan := netAddressKey(addressToBan)
	keysToDelete := make([]addressKey, 0)
	for _, address := range am.store.getAllNotBanned() {
		key := netAddressKey(address)
		if key.address.equal(keyToBan.address) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	for _, key := range keysToDelete {
		am.store.remove(key)
	}

	am.store.addBanned(keyToBan, addressToBan)
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

	am.store.removeBanned(key)
	return nil
}

// IsBanned returns true if the given address is marked as banned
func (am *AddressManager) IsBanned(address *appmessage.NetAddress) (bool, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	am.unbanIfOldEnough(key)
	if !am.store.isBanned(key) {
		if !am.store.isNotBanned(key) {
			return false, errors.Wrapf(ErrAddressNotFound, "address %s "+
				"is not registered with the address manager", address.TCPAddress())
		}
		return false, nil
	}

	return true, nil

}

func (am *AddressManager) unbanIfOldEnough(key addressKey) {
	address, ok := am.store.getBanned(key)
	if !ok {
		return
	}

	const maxBanTime = 24 * time.Hour
	if mstime.Since(address.Timestamp) > maxBanTime {
		am.store.removeBanned(key)
	}
}
