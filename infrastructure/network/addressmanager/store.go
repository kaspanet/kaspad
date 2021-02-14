package addressmanager

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/util/mstime"
	"net"
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
	serializedSize := 16 + 2 // ipv6 + port
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

func (as *addressStore) serializeNetAddress(netAddress *appmessage.NetAddress) []byte {
	serializedSize := 16 + 2 + 8 + 8 // ipv6 + port + timestamp + services
	serializedNetAddress := make([]byte, serializedSize)

	copy(serializedNetAddress[:], netAddress.IP[:])
	binary.LittleEndian.PutUint16(serializedNetAddress[16:], netAddress.Port)
	binary.LittleEndian.PutUint64(serializedNetAddress[18:], uint64(netAddress.Timestamp.UnixMilliseconds()))
	binary.LittleEndian.PutUint64(serializedNetAddress[26:], uint64(netAddress.Services))

	return serializedNetAddress
}

func (as *addressStore) deserializeNetAddress(serializedNetAddress []byte) *appmessage.NetAddress {
	ip := make(net.IP, 16)
	copy(ip[:], serializedNetAddress[:])

	port := binary.LittleEndian.Uint16(serializedNetAddress[16:])
	timestamp := mstime.UnixMilliseconds(int64(binary.LittleEndian.Uint64(serializedNetAddress[18:])))
	services := appmessage.ServiceFlag(binary.LittleEndian.Uint64(serializedNetAddress[26:]))

	return &appmessage.NetAddress{
		IP:        ip,
		Port:      port,
		Timestamp: timestamp,
		Services:  services,
	}
}
