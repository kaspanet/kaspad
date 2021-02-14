package addressmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

type addressStore struct {
	database        database.Database
	addresses       map[addressKey]*appmessage.NetAddress
	bannedAddresses map[ipv6]*appmessage.NetAddress
}

func newAddressStore(database database.Database) *addressStore {
	return &addressStore{
		database:        database,
		addresses:       map[addressKey]*appmessage.NetAddress{},
		bannedAddresses: map[ipv6]*appmessage.NetAddress{},
	}
}

func (as *addressStore) add(key addressKey, address *appmessage.NetAddress) {
	_, ok := as.addresses[key]
	if !ok {
		as.addresses[key] = address
	}
}

func (as *addressStore) remove(key addressKey) {
	delete(as.addresses, key)
}

func (as *addressStore) getAllNotBanned() []*appmessage.NetAddress {
	addresses := make([]*appmessage.NetAddress, 0, len(as.addresses))
	for _, address := range as.addresses {
		addresses = append(addresses, address)
	}
	return addresses
}

func (as *addressStore) getAllNotBannedWithout(ignoredAddresses []*appmessage.NetAddress) []*appmessage.NetAddress {
	ignoredKeys := netAddressesKeys(ignoredAddresses)

	addresses := make([]*appmessage.NetAddress, 0, len(as.addresses))
	for key, address := range as.addresses {
		if !ignoredKeys[key] {
			addresses = append(addresses, address)
		}
	}
	return addresses
}

func (as *addressStore) isNotBanned(key addressKey) bool {
	_, ok := as.addresses[key]
	return ok
}

func (as *addressStore) addBanned(key addressKey, address *appmessage.NetAddress) {
	_, ok := as.bannedAddresses[key.address]
	if !ok {
		as.bannedAddresses[key.address] = address
	}
}

func (as *addressStore) removeBanned(key addressKey) {
	delete(as.bannedAddresses, key.address)
}

func (as *addressStore) getAllBanned() []*appmessage.NetAddress {
	bannedAddresses := make([]*appmessage.NetAddress, 0, len(as.bannedAddresses))
	for _, bannedAddress := range as.bannedAddresses {
		bannedAddresses = append(bannedAddresses, bannedAddress)
	}
	return bannedAddresses
}

func (as *addressStore) isBanned(key addressKey) bool {
	_, ok := as.bannedAddresses[key.address]
	return ok
}

func (as *addressStore) getBanned(key addressKey) (*appmessage.NetAddress, bool) {
	bannedAddress, ok := as.bannedAddresses[key.address]
	return bannedAddress, ok
}
