package addressmanager

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

var notBannedAddressBucket = database.MakeBucket([]byte("not-banned-addresses"))
var bannedAddressBucket = database.MakeBucket([]byte("banned-addresses"))

type addressStore struct {
	database           database.Database
	notBannedAddresses map[addressKey]*appmessage.NetAddress
	bannedAddresses    map[ipv6]*appmessage.NetAddress
}

func newAddressStore(database database.Database) *addressStore {
	return &addressStore{
		database:           database,
		notBannedAddresses: map[addressKey]*appmessage.NetAddress{},
		bannedAddresses:    map[ipv6]*appmessage.NetAddress{},
	}
}

func (as *addressStore) add(key addressKey, address *appmessage.NetAddress) {
	_, ok := as.notBannedAddresses[key]
	if !ok {
		as.notBannedAddresses[key] = address
	}
}

func (as *addressStore) remove(key addressKey) {
	delete(as.notBannedAddresses, key)
}

func (as *addressStore) getAllNotBanned() []*appmessage.NetAddress {
	addresses := make([]*appmessage.NetAddress, 0, len(as.notBannedAddresses))
	for _, address := range as.notBannedAddresses {
		addresses = append(addresses, address)
	}
	return addresses
}

func (as *addressStore) getAllNotBannedWithout(ignoredAddresses []*appmessage.NetAddress) []*appmessage.NetAddress {
	ignoredKeys := netAddressesKeys(ignoredAddresses)

	addresses := make([]*appmessage.NetAddress, 0, len(as.notBannedAddresses))
	for key, address := range as.notBannedAddresses {
		if !ignoredKeys[key] {
			addresses = append(addresses, address)
		}
	}
	return addresses
}

func (as *addressStore) isNotBanned(key addressKey) bool {
	_, ok := as.notBannedAddresses[key]
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

func (as *addressStore) serializeAddressKey(key addressKey) []byte {
	serializedSize := 18
	serializedKey := make([]byte, serializedSize)

	copy(serializedKey[:], key.address[:])
	binary.LittleEndian.PutUint16(serializedKey[16:], key.port)

	return serializedKey
}

func (as *addressStore) deserializeAddressKey(serializedKey []byte) addressKey {
	var ip ipv6
	copy(ip[:], serializedKey[:])

	port := binary.LittleEndian.Uint16(serializedKey[16:])

	return addressKey{
		port:    port,
		address: ip,
	}
}
