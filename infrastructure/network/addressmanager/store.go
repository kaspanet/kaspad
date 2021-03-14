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

func newAddressStore(database database.Database) (*addressStore, error) {
	addressStore := &addressStore{
		database:           database,
		notBannedAddresses: map[addressKey]*appmessage.NetAddress{},
		bannedAddresses:    map[ipv6]*appmessage.NetAddress{},
	}
	err := addressStore.restoreNotBannedAddresses()
	if err != nil {
		return nil, err
	}
	err = addressStore.restoreBannedAddresses()
	if err != nil {
		return nil, err
	}

	log.Infof("Loaded %d addresses and %d banned addresses",
		len(addressStore.notBannedAddresses), len(addressStore.bannedAddresses))

	return addressStore, nil
}

func (as *addressStore) restoreNotBannedAddresses() error {
	cursor, err := as.database.Cursor(notBannedAddressBucket)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for ok := cursor.First(); ok; ok = cursor.Next() {
		databaseKey, err := cursor.Key()
		if err != nil {
			return err
		}
		serializedKey := databaseKey.Suffix()
		key := as.deserializeAddressKey(serializedKey)

		serializedNetAddress, err := cursor.Value()
		if err != nil {
			return err
		}
		netAddress := as.deserializeNetAddress(serializedNetAddress)
		as.notBannedAddresses[key] = netAddress
	}
	return nil
}

func (as *addressStore) restoreBannedAddresses() error {
	cursor, err := as.database.Cursor(bannedAddressBucket)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for ok := cursor.First(); ok; ok = cursor.Next() {
		databaseKey, err := cursor.Key()
		if err != nil {
			return err
		}
		var ipv6 ipv6
		copy(ipv6[:], databaseKey.Suffix())

		serializedNetAddress, err := cursor.Value()
		if err != nil {
			return err
		}
		netAddress := as.deserializeNetAddress(serializedNetAddress)
		as.bannedAddresses[ipv6] = netAddress
	}
	return nil
}

func (as *addressStore) add(key addressKey, address *appmessage.NetAddress) error {
	if _, ok := as.notBannedAddresses[key]; ok {
		return nil
	}

	as.notBannedAddresses[key] = address

	databaseKey := as.notBannedDatabaseKey(key)
	serializedAddress := as.serializeNetAddress(address)
	return as.database.Put(databaseKey, serializedAddress)
}

func (as *addressStore) remove(key addressKey) error {
	delete(as.notBannedAddresses, key)

	databaseKey := as.notBannedDatabaseKey(key)
	return as.database.Delete(databaseKey)
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

func (as *addressStore) addBanned(key addressKey, address *appmessage.NetAddress) error {
	if _, ok := as.bannedAddresses[key.address]; ok {
		return nil
	}

	as.bannedAddresses[key.address] = address

	databaseKey := as.bannedDatabaseKey(key)
	serializedAddress := as.serializeNetAddress(address)
	return as.database.Put(databaseKey, serializedAddress)
}

func (as *addressStore) removeBanned(key addressKey) error {
	delete(as.bannedAddresses, key.address)

	databaseKey := as.bannedDatabaseKey(key)
	return as.database.Delete(databaseKey)
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

func (as *addressStore) notBannedDatabaseKey(key addressKey) *database.Key {
	serializedKey := as.serializeAddressKey(key)
	return notBannedAddressBucket.Key(serializedKey)
}

func (as *addressStore) bannedDatabaseKey(key addressKey) *database.Key {
	return bannedAddressBucket.Key(key.address[:])
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
