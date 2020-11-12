// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addressmanager

import (
	"bytes"
	crand "crypto/rand" // for seeding
	"encoding/binary"
	"encoding/gob"
	"io"
	"math/rand"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

// AddressKey represents a "string" key in the form of ip:port for IPv4 addresses
// or [ip]:port for IPv6 addresses for use as keys in maps.
type AddressKey string
type newAddressBucketArray [NewBucketCount]map[AddressKey]*KnownAddress
type triedAddressBucketArray [TriedBucketCount][]*KnownAddress

// AddressManager provides a concurrency safe address manager for caching potential
// peers on the Kaspa network.
type AddressManager struct {
	cfg      *config.Config
	database database.Database

	mutex              sync.Mutex
	lookupFunc         func(string) ([]net.IP, error)
	random             *rand.Rand
	key                [32]byte
	addressIndex       map[AddressKey]*KnownAddress // address keys to known addresses for all addresses.
	started            int32
	shutdown           int32
	wg                 sync.WaitGroup
	quit               chan struct{}
	localAddressesLock sync.Mutex
	localAddresses     map[AddressKey]*localAddress
	localSubnetworkID  *externalapi.DomainSubnetworkID

	fullNodeNewAddressBucketArray     newAddressBucketArray
	fullNodeNewAddressCount           int
	fullNodeTriedAddressBucketArray   triedAddressBucketArray
	fullNodeTriedAddressCount         int
	subnetworkNewAddressBucketArrays  map[externalapi.DomainSubnetworkID]*newAddressBucketArray
	subnetworkNewAddressCounts        map[externalapi.DomainSubnetworkID]int
	subnetworkTriedAddresBucketArrays map[externalapi.DomainSubnetworkID]*triedAddressBucketArray
	subnetworkTriedAddressCounts      map[externalapi.DomainSubnetworkID]int
}

type serializedKnownAddress struct {
	Address       AddressKey
	SourceAddress AddressKey
	SubnetworkID  string
	Attempts      int
	TimeStamp     int64
	LastAttempt   int64
	LastSuccess   int64
	IsBanned      bool
	BannedTime    int64
	// no refcount or tried, that is available from context.
}

type serializedNewAddressBucketArray [NewBucketCount][]AddressKey
type serializedTriedAddressBucketArray [TriedBucketCount][]AddressKey

// PeersStateForSerialization is the data model that is used to
// serialize the peers state to any encoding.
type PeersStateForSerialization struct {
	Version   int
	Key       [32]byte
	Addresses []*serializedKnownAddress

	SubnetworkNewAddressBucketArrays   map[string]*serializedNewAddressBucketArray // string is Subnetwork ID
	FullNodeNewAddressBucketArray      serializedNewAddressBucketArray
	SubnetworkTriedAddressBucketArrays map[string]*serializedTriedAddressBucketArray // string is Subnetwork ID
	FullNodeTriedAddressBucketArray    serializedTriedAddressBucketArray
}

type localAddress struct {
	netAddress *appmessage.NetAddress
	score      AddressPriority
}

// AddressPriority type is used to describe the hierarchy of local address
// discovery methods.
type AddressPriority int

const (
	// InterfacePrio signifies the address is on a local interface
	InterfacePrio AddressPriority = iota

	// BoundPrio signifies the address has been explicitly bounded to.
	BoundPrio

	// UpnpPrio signifies the address was obtained from UPnP.
	UpnpPrio

	// HTTPPrio signifies the address was obtained from an external HTTP service.
	HTTPPrio

	// ManualPrio signifies the address was provided by --externalip.
	ManualPrio
)

const (
	// needAddressThreshold is the number of addresses under which the
	// address manager will claim to need more addresses.
	needAddressThreshold = 1000

	// dumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	dumpAddressInterval = time.Minute * 10

	// triedBucketSize is the maximum number of addresses in each
	// tried address bucket.
	triedBucketSize = 256

	// TriedBucketCount is the number of buckets we split tried
	// addresses over.
	TriedBucketCount = 64

	// newBucketSize is the maximum number of addresses in each new address
	// bucket.
	newBucketSize = 64

	// NewBucketCount is the number of buckets that we spread new addresses
	// over.
	NewBucketCount = 1024

	// triedBucketsPerGroup is the number of tried buckets over which an
	// address group will be spread.
	triedBucketsPerGroup = 8

	// newBucketsPerGroup is the number of new buckets over which an
	// source address group will be spread.
	newBucketsPerGroup = 64

	// newBucketsPerAddress is the number of buckets a frequently seen new
	// address may end up in.
	newBucketsPerAddress = 8

	// numMissingDays is the number of days before which we assume an
	// address has vanished if we have not seen it announced  in that long.
	numMissingDays = 30

	// numRetries is the number of tried without a single success before
	// we assume an address is bad.
	numRetries = 3

	// maxFailures is the maximum number of failures we will accept without
	// a success before considering an address bad.
	maxFailures = 10

	// minBadDays is the number of days since the last success before we
	// will consider evicting an address.
	minBadDays = 7

	// getAddrMin is the least addresses that we will send in response
	// to a getAddresses. If we have less than this amount, we send everything.
	getAddrMin = 50

	// GetAddressesMax is the most addresses that we will send in response
	// to a getAddress (in practise the most addresses we will return from a
	// call to AddressCache()).
	GetAddressesMax = 2500

	// getAddrPercent is the percentage of total addresses known that we
	// will share with a call to AddressCache.
	getAddrPercent = 23

	// serializationVersion is the current version of the on-disk format.
	serializationVersion = 1
)

var peersDBKey = database.MakeBucket().Key([]byte("peers"))

// ErrAddressNotFound is an error returned from some functions when a
// given address is not found in the address manager
var ErrAddressNotFound = errors.New("address not found")

// New returns a new Kaspa address manager.
func New(cfg *config.Config, database database.Database) (*AddressManager, error) {
	addressManager := AddressManager{
		cfg:               cfg,
		database:          database,
		lookupFunc:        cfg.Lookup,
		random:            rand.New(rand.NewSource(time.Now().UnixNano())),
		quit:              make(chan struct{}),
		localAddresses:    make(map[AddressKey]*localAddress),
		localSubnetworkID: cfg.SubnetworkID,
	}
	err := addressManager.initListeners()
	if err != nil {
		return nil, err
	}
	addressManager.reset()
	return &addressManager, nil
}

// updateAddress is a helper function to either update an address already known
// to the address manager, or to add the address if not already known.
func (am *AddressManager) updateAddress(netAddress, sourceAddress *appmessage.NetAddress, subnetworkID *externalapi.DomainSubnetworkID) {
	// Filter out non-routable addresses. Note that non-routable
	// also includes invalid and local addresses.
	if !am.IsRoutable(netAddress) {
		return
	}

	addressKey := NetAddressKey(netAddress)
	knownAddress := am.knownAddress(netAddress)
	if knownAddress != nil {
		// Update the last seen time and services.
		// note that to prevent causing excess garbage on getaddr
		// messages the netaddresses in addrmaanger are *immutable*,
		// if we need to change them then we replace the pointer with a
		// new copy so that we don't have to copy every netAddress for getaddress.
		if netAddress.Timestamp.After(knownAddress.netAddress.Timestamp) ||
			(knownAddress.netAddress.Services&netAddress.Services) !=
				netAddress.Services {

			netAddressCopy := *knownAddress.netAddress
			netAddressCopy.Timestamp = netAddress.Timestamp
			netAddressCopy.AddService(netAddress.Services)
			knownAddress.netAddress = &netAddressCopy
		}

		// If already in tried, we have nothing to do here.
		if knownAddress.tried {
			return
		}

		// Already at our max?
		if knownAddress.referenceCount == newBucketsPerAddress {
			return
		}

		// The more entries we have, the less likely we are to add more.
		// likelihood is 2N.
		factor := int32(2 * knownAddress.referenceCount)
		if am.random.Int31n(factor) != 0 {
			return
		}
	} else {
		// Make a copy of the net address to avoid races since it is
		// updated elsewhere in the addressManager code and would otherwise
		// change the actual netAddress on the peer.
		netAddressCopy := *netAddress
		knownAddress = &KnownAddress{netAddress: &netAddressCopy, sourceAddress: sourceAddress, subnetworkID: subnetworkID}
		am.addressIndex[addressKey] = knownAddress
		am.incrementNewAddressCount(subnetworkID)
	}

	// Already exists?
	newAddressBucketArray := am.newAddressBucketArray(knownAddress.subnetworkID)
	newAddressBucketIndex := am.newAddressBucketIndex(netAddress, sourceAddress)
	if newAddressBucketArray != nil {
		if _, ok := newAddressBucketArray[newAddressBucketIndex][addressKey]; ok {
			return
		}
	}

	// Enforce max addresses.
	if newAddressBucketArray != nil && len(newAddressBucketArray[newAddressBucketIndex]) > newBucketSize {
		log.Tracef("new bucket is full, expiring old")
		am.expireNew(knownAddress.subnetworkID, newAddressBucketIndex)
	}

	// Add to new bucket.
	knownAddress.referenceCount++
	am.updateAddrNew(newAddressBucketIndex, addressKey, knownAddress)

	totalAddressCount := am.newAddressCount(knownAddress.subnetworkID) + am.triedAddressCount(knownAddress.subnetworkID)
	log.Tracef("Added new address %s for a total of %d addresses", addressKey, totalAddressCount)

}

func (am *AddressManager) updateAddrNew(bucket int, addressKey AddressKey, knownAddress *KnownAddress) {
	if knownAddress.subnetworkID == nil {
		am.fullNodeNewAddressBucketArray[bucket][addressKey] = knownAddress
		return
	}

	if _, ok := am.subnetworkNewAddressBucketArrays[*knownAddress.subnetworkID]; !ok {
		am.subnetworkNewAddressBucketArrays[*knownAddress.subnetworkID] = &newAddressBucketArray{}
		for i := range am.subnetworkNewAddressBucketArrays[*knownAddress.subnetworkID] {
			am.subnetworkNewAddressBucketArrays[*knownAddress.subnetworkID][i] = make(map[AddressKey]*KnownAddress)
		}
	}
	am.subnetworkNewAddressBucketArrays[*knownAddress.subnetworkID][bucket][addressKey] = knownAddress
}

func (am *AddressManager) updateAddrTried(bucketIndex int, knownAddress *KnownAddress) {
	if knownAddress.subnetworkID == nil {
		am.fullNodeTriedAddressBucketArray[bucketIndex] = append(am.fullNodeTriedAddressBucketArray[bucketIndex], knownAddress)
		return
	}

	if _, ok := am.subnetworkTriedAddresBucketArrays[*knownAddress.subnetworkID]; !ok {
		am.subnetworkTriedAddresBucketArrays[*knownAddress.subnetworkID] = &triedAddressBucketArray{}
		for i := range am.subnetworkTriedAddresBucketArrays[*knownAddress.subnetworkID] {
			am.subnetworkTriedAddresBucketArrays[*knownAddress.subnetworkID][i] = nil
		}
	}
	am.subnetworkTriedAddresBucketArrays[*knownAddress.subnetworkID][bucketIndex] = append(am.subnetworkTriedAddresBucketArrays[*knownAddress.subnetworkID][bucketIndex], knownAddress)
}

// expireNew makes space in the new buckets by expiring the really bad entries.
// If no bad entries are available we look at a few and remove the oldest.
func (am *AddressManager) expireNew(subnetworkID *externalapi.DomainSubnetworkID, bucketIndex int) {
	// First see if there are any entries that are so bad we can just throw
	// them away. otherwise we throw away the oldest entry in the cache.
	// We keep track of oldest in the initial traversal and use that
	// information instead.
	var oldest *KnownAddress
	newAddressBucketArray := am.newAddressBucketArray(subnetworkID)
	for addressKey, knownAddress := range newAddressBucketArray[bucketIndex] {
		if knownAddress.isBad() {
			log.Tracef("expiring bad address %s", addressKey)
			delete(newAddressBucketArray[bucketIndex], addressKey)
			knownAddress.referenceCount--
			if knownAddress.referenceCount == 0 {
				am.decrementNewAddressCount(subnetworkID)
				delete(am.addressIndex, addressKey)
			}
			continue
		}
		if oldest == nil {
			oldest = knownAddress
		} else if !knownAddress.netAddress.Timestamp.After(oldest.netAddress.Timestamp) {
			oldest = knownAddress
		}
	}

	if oldest != nil {
		addressKey := NetAddressKey(oldest.netAddress)
		log.Tracef("expiring oldest address %s", addressKey)

		delete(newAddressBucketArray[bucketIndex], addressKey)
		oldest.referenceCount--
		if oldest.referenceCount == 0 {
			am.decrementNewAddressCount(subnetworkID)
			delete(am.addressIndex, addressKey)
		}
	}
}

// pickTried selects an address from the tried bucket to be evicted.
// We just choose the eldest.
func (am *AddressManager) pickTried(subnetworkID *externalapi.DomainSubnetworkID, bucketIndex int) (
	knownAddress *KnownAddress, knownAddressIndex int) {

	var oldest *KnownAddress
	oldestIndex := -1
	triedAddressBucketArray := am.triedAddressBucketArray(subnetworkID)
	for i, address := range triedAddressBucketArray[bucketIndex] {
		if oldest == nil || oldest.netAddress.Timestamp.After(address.netAddress.Timestamp) {
			oldestIndex = i
			oldest = address
		}
	}
	return oldest, oldestIndex
}

func (am *AddressManager) newAddressBucketIndex(netAddress, srcAddress *appmessage.NetAddress) int {
	// doublesha256(key + sourcegroup + int64(doublesha256(key + group + sourcegroup))%bucket_per_source_group) % num_new_buckets

	data1 := []byte{}
	data1 = append(data1, am.key[:]...)
	data1 = append(data1, []byte(am.GroupKey(netAddress))...)
	data1 = append(data1, []byte(am.GroupKey(srcAddress))...)
	hash1 := hashes.HashData(data1)
	hash64 := binary.LittleEndian.Uint64(hash1[:])
	hash64 %= newBucketsPerGroup
	var hashbuf [8]byte
	binary.LittleEndian.PutUint64(hashbuf[:], hash64)
	data2 := []byte{}
	data2 = append(data2, am.key[:]...)
	data2 = append(data2, am.GroupKey(srcAddress)...)
	data2 = append(data2, hashbuf[:]...)

	hash2 := hashes.HashData(data2)
	return int(binary.LittleEndian.Uint64(hash2[:]) % NewBucketCount)
}

func (am *AddressManager) triedAddressBucketIndex(netAddress *appmessage.NetAddress) int {
	// doublesha256(key + group + truncate_to_64bits(doublesha256(key)) % buckets_per_group) % num_buckets
	data1 := []byte{}
	data1 = append(data1, am.key[:]...)
	data1 = append(data1, []byte(NetAddressKey(netAddress))...)
	hash1 := hashes.HashData(data1)
	hash64 := binary.LittleEndian.Uint64(hash1[:])
	hash64 %= triedBucketsPerGroup
	var hashbuf [8]byte
	binary.LittleEndian.PutUint64(hashbuf[:], hash64)
	data2 := []byte{}
	data2 = append(data2, am.key[:]...)
	data2 = append(data2, am.GroupKey(netAddress)...)
	data2 = append(data2, hashbuf[:]...)

	hash2 := hashes.HashData(data2)
	return int(binary.LittleEndian.Uint64(hash2[:]) % TriedBucketCount)
}

// addressHandler is the main handler for the address manager. It must be run
// as a goroutine.
func (am *AddressManager) addressHandler() {
	dumpAddressTicker := time.NewTicker(dumpAddressInterval)
	defer dumpAddressTicker.Stop()

out:
	for {
		select {
		case <-dumpAddressTicker.C:
			err := am.savePeers()
			if err != nil {
				panic(errors.Wrap(err, "error saving peers"))
			}

		case <-am.quit:
			break out
		}
	}
	err := am.savePeers()
	if err != nil {
		panic(errors.Wrap(err, "error saving peers"))
	}
	am.wg.Done()
	log.Trace("Address handler done")
}

// savePeers saves all the known addresses to the database so they can be read back
// in at next run.
func (am *AddressManager) savePeers() error {
	serializedPeersState, err := am.serializePeersState()
	if err != nil {
		return err
	}

	return am.database.Put(peersDBKey, serializedPeersState)
}

func (am *AddressManager) serializePeersState() ([]byte, error) {
	peersState, err := am.PeersStateForSerialization()
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	err = encoder.Encode(&peersState)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode peers state")
	}

	return buffer.Bytes(), nil
}

// PeersStateForSerialization returns the data model that is used to serialize the peers state to any encoding.
func (am *AddressManager) PeersStateForSerialization() (*PeersStateForSerialization, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// First we make a serializable data structure so we can encode it to
	// gob.
	peersState := new(PeersStateForSerialization)
	peersState.Version = serializationVersion
	copy(peersState.Key[:], am.key[:])

	peersState.Addresses = make([]*serializedKnownAddress, len(am.addressIndex))
	i := 0
	for addressKey, knownAddress := range am.addressIndex {
		serializedAddress := new(serializedKnownAddress)
		serializedAddress.Address = addressKey
		if knownAddress.subnetworkID == nil {
			serializedAddress.SubnetworkID = ""
		} else {
			serializedAddress.SubnetworkID = knownAddress.subnetworkID.String()
		}
		serializedAddress.TimeStamp = knownAddress.netAddress.Timestamp.UnixMilliseconds()
		serializedAddress.SourceAddress = NetAddressKey(knownAddress.sourceAddress)
		serializedAddress.Attempts = knownAddress.attempts
		serializedAddress.LastAttempt = knownAddress.lastAttempt.UnixMilliseconds()
		serializedAddress.LastSuccess = knownAddress.lastSuccess.UnixMilliseconds()
		serializedAddress.IsBanned = knownAddress.isBanned
		serializedAddress.BannedTime = knownAddress.bannedTime.UnixMilliseconds()
		// Tried and referenceCount are implicit in the rest of the structure
		// and will be worked out from context on unserialisation.
		peersState.Addresses[i] = serializedAddress
		i++
	}

	peersState.SubnetworkNewAddressBucketArrays = make(map[string]*serializedNewAddressBucketArray)
	for subnetworkID := range am.subnetworkNewAddressBucketArrays {
		subnetworkIDStr := subnetworkID.String()
		peersState.SubnetworkNewAddressBucketArrays[subnetworkIDStr] = &serializedNewAddressBucketArray{}

		for i := range am.subnetworkNewAddressBucketArrays[subnetworkID] {
			peersState.SubnetworkNewAddressBucketArrays[subnetworkIDStr][i] = make([]AddressKey, len(am.subnetworkNewAddressBucketArrays[subnetworkID][i]))
			j := 0
			for k := range am.subnetworkNewAddressBucketArrays[subnetworkID][i] {
				peersState.SubnetworkNewAddressBucketArrays[subnetworkIDStr][i][j] = k
				j++
			}
		}
	}

	for i := range am.fullNodeNewAddressBucketArray {
		peersState.FullNodeNewAddressBucketArray[i] = make([]AddressKey, len(am.fullNodeNewAddressBucketArray[i]))
		j := 0
		for k := range am.fullNodeNewAddressBucketArray[i] {
			peersState.FullNodeNewAddressBucketArray[i][j] = k
			j++
		}
	}

	peersState.SubnetworkTriedAddressBucketArrays = make(map[string]*serializedTriedAddressBucketArray)
	for subnetworkID := range am.subnetworkTriedAddresBucketArrays {
		subnetworkIDStr := subnetworkID.String()
		peersState.SubnetworkTriedAddressBucketArrays[subnetworkIDStr] = &serializedTriedAddressBucketArray{}

		for i := range am.subnetworkTriedAddresBucketArrays[subnetworkID] {
			peersState.SubnetworkTriedAddressBucketArrays[subnetworkIDStr][i] = make([]AddressKey, len(am.subnetworkTriedAddresBucketArrays[subnetworkID][i]))
			j := 0
			for _, knownAddress := range am.subnetworkTriedAddresBucketArrays[subnetworkID][i] {
				peersState.SubnetworkTriedAddressBucketArrays[subnetworkIDStr][i][j] = NetAddressKey(knownAddress.netAddress)
				j++
			}
		}
	}

	for i := range am.fullNodeTriedAddressBucketArray {
		peersState.FullNodeTriedAddressBucketArray[i] = make([]AddressKey, len(am.fullNodeTriedAddressBucketArray[i]))
		j := 0
		for _, knownAddress := range am.fullNodeTriedAddressBucketArray[i] {
			peersState.FullNodeTriedAddressBucketArray[i][j] = NetAddressKey(knownAddress.netAddress)
			j++
		}
	}

	return peersState, nil
}

// loadPeers loads the known address from the database. If missing,
// just don't load anything and start fresh.
func (am *AddressManager) loadPeers() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	serializedPeerState, err := am.database.Get(peersDBKey)
	if database.IsNotFoundError(err) {
		am.reset()
		log.Info("No peers state was found in the database. Created a new one", am.totalNumAddresses())
		return nil
	}
	if err != nil {
		return err
	}

	err = am.deserializePeersState(serializedPeerState)
	if err != nil {
		return err
	}

	log.Infof("Loaded %d addresses from database", am.totalNumAddresses())
	return nil
}

func (am *AddressManager) deserializePeersState(serializedPeerState []byte) error {
	var peersState PeersStateForSerialization
	r := bytes.NewBuffer(serializedPeerState)
	dec := gob.NewDecoder(r)
	err := dec.Decode(&peersState)
	if err != nil {
		return errors.Wrap(err, "error deserializing peers state")
	}

	if peersState.Version != serializationVersion {
		return errors.Errorf("unknown version %d in serialized "+
			"peers state", peersState.Version)
	}
	copy(am.key[:], peersState.Key[:])

	for _, serializedKnownAddress := range peersState.Addresses {
		knownAddress := new(KnownAddress)
		knownAddress.netAddress, err = am.DeserializeNetAddress(serializedKnownAddress.Address)
		if err != nil {
			return errors.Errorf("failed to deserialize netaddress "+
				"%s: %s", serializedKnownAddress.Address, err)
		}
		knownAddress.sourceAddress, err = am.DeserializeNetAddress(serializedKnownAddress.SourceAddress)
		if err != nil {
			return errors.Errorf("failed to deserialize netaddress "+
				"%s: %s", serializedKnownAddress.SourceAddress, err)
		}
		if serializedKnownAddress.SubnetworkID != "" {
			knownAddress.subnetworkID, err = subnetworks.FromString(serializedKnownAddress.SubnetworkID)
			if err != nil {
				return errors.Errorf("failed to deserialize subnetwork id "+
					"%s: %s", serializedKnownAddress.SubnetworkID, err)
			}
		}
		knownAddress.attempts = serializedKnownAddress.Attempts
		knownAddress.lastAttempt = mstime.UnixMilliseconds(serializedKnownAddress.LastAttempt)
		knownAddress.lastSuccess = mstime.UnixMilliseconds(serializedKnownAddress.LastSuccess)
		knownAddress.isBanned = serializedKnownAddress.IsBanned
		knownAddress.bannedTime = mstime.UnixMilliseconds(serializedKnownAddress.BannedTime)
		am.addressIndex[NetAddressKey(knownAddress.netAddress)] = knownAddress
	}

	for subnetworkIDStr := range peersState.SubnetworkNewAddressBucketArrays {
		subnetworkID, err := subnetworks.FromString(subnetworkIDStr)
		if err != nil {
			return err
		}
		for i, subnetworkNewAddressBucket := range peersState.SubnetworkNewAddressBucketArrays[subnetworkIDStr] {
			for _, addressKey := range subnetworkNewAddressBucket {
				knownAddress, ok := am.addressIndex[addressKey]
				if !ok {
					return errors.Errorf("newbucket contains %s but "+
						"none in address list", addressKey)
				}

				if knownAddress.referenceCount == 0 {
					am.subnetworkNewAddressCounts[*subnetworkID]++
				}
				knownAddress.referenceCount++
				am.updateAddrNew(i, addressKey, knownAddress)
			}
		}
	}

	for i, fullNodeNewAddressBucket := range peersState.FullNodeNewAddressBucketArray {
		for _, addressKey := range fullNodeNewAddressBucket {
			knownAddress, ok := am.addressIndex[addressKey]
			if !ok {
				return errors.Errorf("full nodes newbucket contains %s but "+
					"none in address list", addressKey)
			}

			if knownAddress.referenceCount == 0 {
				am.fullNodeNewAddressCount++
			}
			knownAddress.referenceCount++
			am.updateAddrNew(i, addressKey, knownAddress)
		}
	}

	for subnetworkIDString := range peersState.SubnetworkTriedAddressBucketArrays {
		subnetworkID, err := subnetworks.FromString(subnetworkIDString)
		if err != nil {
			return err
		}
		for i, subnetworkTriedAddressBucket := range peersState.SubnetworkTriedAddressBucketArrays[subnetworkIDString] {
			for _, addressKey := range subnetworkTriedAddressBucket {
				knownAddress, ok := am.addressIndex[addressKey]
				if !ok {
					return errors.Errorf("Tried bucket contains %s but "+
						"none in address list", addressKey)
				}

				knownAddress.tried = true
				am.subnetworkTriedAddressCounts[*subnetworkID]++
				am.subnetworkTriedAddresBucketArrays[*subnetworkID][i] = append(am.subnetworkTriedAddresBucketArrays[*subnetworkID][i], knownAddress)
			}
		}
	}

	for i, fullNodeTriedAddressBucket := range peersState.FullNodeTriedAddressBucketArray {
		for _, addressKey := range fullNodeTriedAddressBucket {
			knownAddress, ok := am.addressIndex[addressKey]
			if !ok {
				return errors.Errorf("Full nodes tried bucket contains %s but "+
					"none in address list", addressKey)
			}

			knownAddress.tried = true
			am.fullNodeTriedAddressCount++
			am.fullNodeTriedAddressBucketArray[i] = append(am.fullNodeTriedAddressBucketArray[i], knownAddress)
		}
	}

	// Sanity checking.
	for addressKey, knownAddress := range am.addressIndex {
		if knownAddress.referenceCount == 0 && !knownAddress.tried {
			return errors.Errorf("address %s after serialisation "+
				"with no references", addressKey)
		}

		if knownAddress.referenceCount > 0 && knownAddress.tried {
			return errors.Errorf("address %s after serialisation "+
				"which is both new and tried!", addressKey)
		}
	}

	return nil
}

// DeserializeNetAddress converts a given address string to a *appmessage.NetAddress
func (am *AddressManager) DeserializeNetAddress(addressKey AddressKey) (*appmessage.NetAddress, error) {
	host, portString, err := net.SplitHostPort(string(addressKey))
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		return nil, err
	}

	return am.HostToNetAddress(host, uint16(port), appmessage.SFNodeNetwork)
}

// Start begins the core address handler which manages a pool of known
// addresses, timeouts, and interval based writes.
func (am *AddressManager) Start() error {
	// Already started?
	if atomic.AddInt32(&am.started, 1) != 1 {
		return nil
	}

	log.Trace("Starting address manager")

	// Load peers we already know about from the database.
	err := am.loadPeers()
	if err != nil {
		return err
	}

	// Start the address ticker to save addresses periodically.
	am.wg.Add(1)
	spawn("addressManager.addressHandler", am.addressHandler)
	return nil
}

// Stop gracefully shuts down the address manager by stopping the main handler.
func (am *AddressManager) Stop() error {
	if atomic.AddInt32(&am.shutdown, 1) != 1 {
		log.Warnf("Address manager is already in the process of " +
			"shutting down")
		return nil
	}

	log.Infof("Address manager shutting down")
	close(am.quit)
	am.wg.Wait()
	return nil
}

// AddAddresses adds new addresses to the address manager. It enforces a max
// number of addresses and silently ignores duplicate addresses. It is
// safe for concurrent access.
func (am *AddressManager) AddAddresses(addresses []*appmessage.NetAddress, sourceAddress *appmessage.NetAddress, subnetworkID *externalapi.DomainSubnetworkID) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, address := range addresses {
		am.updateAddress(address, sourceAddress, subnetworkID)
	}
}

// AddAddress adds a new address to the address manager. It enforces a max
// number of addresses and silently ignores duplicate addresses. It is
// safe for concurrent access.
func (am *AddressManager) AddAddress(address, sourceAddress *appmessage.NetAddress, subnetworkID *externalapi.DomainSubnetworkID) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.updateAddress(address, sourceAddress, subnetworkID)
}

// numAddresses returns the number of addresses that belongs to a specific subnetwork id
// which are known to the address manager.
func (am *AddressManager) numAddresses(subnetworkID *externalapi.DomainSubnetworkID) int {
	if subnetworkID == nil {
		return am.fullNodeNewAddressCount + am.fullNodeTriedAddressCount
	}
	return am.subnetworkTriedAddressCounts[*subnetworkID] + am.subnetworkNewAddressCounts[*subnetworkID]
}

// totalNumAddresses returns the number of addresses known to the address manager.
func (am *AddressManager) totalNumAddresses() int {
	total := am.fullNodeNewAddressCount + am.fullNodeTriedAddressCount
	for _, numAddresses := range am.subnetworkTriedAddressCounts {
		total += numAddresses
	}
	for _, numAddresses := range am.subnetworkNewAddressCounts {
		total += numAddresses
	}
	return total
}

// TotalNumAddresses returns the number of addresses known to the address manager.
func (am *AddressManager) TotalNumAddresses() int {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.totalNumAddresses()
}

// NeedMoreAddresses returns whether or not the address manager needs more
// addresses.
func (am *AddressManager) NeedMoreAddresses() bool {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	allAddresses := am.numAddresses(am.localSubnetworkID)
	if am.localSubnetworkID != nil {
		allAddresses += am.numAddresses(nil)
	}
	return allAddresses < needAddressThreshold
}

// AddressCache returns the current address cache. It must be treated as
// read-only (but since it is a copy now, this is not as dangerous).
func (am *AddressManager) AddressCache(includeAllSubnetworks bool, subnetworkID *externalapi.DomainSubnetworkID) []*appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if len(am.addressIndex) == 0 {
		return nil
	}

	allAddresses := []*appmessage.NetAddress{}
	// Iteration order is undefined here, but we randomise it anyway.
	for _, v := range am.addressIndex {
		if includeAllSubnetworks || subnetworks.IsEqual(v.SubnetworkID(), subnetworkID) {
			allAddresses = append(allAddresses, v.netAddress)
		}
	}

	numAddresses := len(allAddresses) * getAddrPercent / 100
	if numAddresses > GetAddressesMax {
		numAddresses = GetAddressesMax
	}
	if len(allAddresses) < getAddrMin {
		numAddresses = len(allAddresses)
	}
	if len(allAddresses) > getAddrMin && numAddresses < getAddrMin {
		numAddresses = getAddrMin
	}

	// Fisher-Yates shuffle the array. We only need to do the first
	// `numAddresses' since we are throwing the rest.
	for i := 0; i < numAddresses; i++ {
		// pick a number between current index and the end
		j := rand.Intn(len(allAddresses)-i) + i
		allAddresses[i], allAddresses[j] = allAddresses[j], allAddresses[i]
	}

	// slice off the limit we are willing to share.
	return allAddresses[0:numAddresses]
}

// reset resets the address manager by reinitialising the random source
// and allocating fresh empty bucket storage.
func (am *AddressManager) reset() {
	am.addressIndex = make(map[AddressKey]*KnownAddress)

	// fill key with bytes from a good random source.
	io.ReadFull(crand.Reader, am.key[:])
	am.subnetworkNewAddressBucketArrays = make(map[externalapi.DomainSubnetworkID]*newAddressBucketArray)
	am.subnetworkTriedAddresBucketArrays = make(map[externalapi.DomainSubnetworkID]*triedAddressBucketArray)

	am.subnetworkNewAddressCounts = make(map[externalapi.DomainSubnetworkID]int)
	am.subnetworkTriedAddressCounts = make(map[externalapi.DomainSubnetworkID]int)

	for i := range am.fullNodeNewAddressBucketArray {
		am.fullNodeNewAddressBucketArray[i] = make(map[AddressKey]*KnownAddress)
	}
	for i := range am.fullNodeTriedAddressBucketArray {
		am.fullNodeTriedAddressBucketArray[i] = nil
	}
	am.fullNodeNewAddressCount = 0
	am.fullNodeTriedAddressCount = 0
}

// HostToNetAddress returns a netaddress given a host address. If
// the host is not an IP address it will be resolved.
func (am *AddressManager) HostToNetAddress(host string, port uint16, services appmessage.ServiceFlag) (*appmessage.NetAddress, error) {
	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := am.lookupFunc(host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, errors.Errorf("no addresses found for %s", host)
		}
		ip = ips[0]
	}

	return appmessage.NewNetAddressIPPort(ip, port, services), nil
}

// NetAddressKey returns a key in the form of ip:port for IPv4 addresses
// or [ip]:port for IPv6 addresses for use as keys in maps.
func NetAddressKey(netAddress *appmessage.NetAddress) AddressKey {
	port := strconv.FormatUint(uint64(netAddress.Port), 10)

	return AddressKey(net.JoinHostPort(netAddress.IP.String(), port))
}

// GetAddress returns a single address that should be routable. It picks a
// random one from the possible addresses with preference given to ones that
// have not been used recently and should not pick 'close' addresses
// consecutively.
func (am *AddressManager) GetAddress() *KnownAddress {
	// Protect concurrent access.
	am.mutex.Lock()
	defer am.mutex.Unlock()

	triedAddressBucketArray := am.triedAddressBucketArray(am.localSubnetworkID)
	triedAddressCount := am.triedAddressCount(am.localSubnetworkID)
	newAddressBucketArray := am.newAddressBucketArray(am.localSubnetworkID)
	newAddressCount := am.newAddressCount(am.localSubnetworkID)
	knownAddress := am.getAddress(triedAddressBucketArray, triedAddressCount, newAddressBucketArray, newAddressCount)

	return knownAddress

}

// getAddress returns a single address that should be routable.
// See GetAddress for further details.
func (am *AddressManager) getAddress(triedAddressBucketArray *triedAddressBucketArray, triedAddressCount int,
	newAddressBucketArray *newAddressBucketArray, newAddressCount int) *KnownAddress {

	// Use a 50% chance for choosing between tried and new addresses.
	var bucketArray addressBucketArray
	if triedAddressCount > 0 && (newAddressCount == 0 || am.random.Intn(2) == 0) {
		bucketArray = triedAddressBucketArray
	} else if newAddressCount > 0 {
		bucketArray = newAddressBucketArray
	} else {
		// There aren't any addresses in any of the buckets
		return nil
	}

	// Pick a random bucket
	randomBucket := bucketArray.randomBucket(am.random)

	// Get the sum of all chances
	totalChance := float64(0)
	for _, knownAddress := range randomBucket {
		totalChance += knownAddress.chance()
	}

	// Pick a random address weighted by chance
	randomValue := am.random.Float64()
	accumulatedChance := float64(0)
	for _, knownAddress := range randomBucket {
		normalizedChance := knownAddress.chance() / totalChance
		accumulatedChance += normalizedChance
		if randomValue < accumulatedChance {
			return knownAddress
		}
	}

	panic("randomValue is equal to or greater than 1, which cannot happen")
}

type addressBucketArray interface {
	name() string
	randomBucket(random *rand.Rand) []*KnownAddress
}

func (nb *newAddressBucketArray) randomBucket(random *rand.Rand) []*KnownAddress {
	nonEmptyBuckets := make([]map[AddressKey]*KnownAddress, 0, NewBucketCount)
	for _, bucket := range nb {
		if len(bucket) > 0 {
			nonEmptyBuckets = append(nonEmptyBuckets, bucket)
		}
	}
	randomIndex := random.Intn(len(nonEmptyBuckets))
	randomBucket := nonEmptyBuckets[randomIndex]

	// Collect the known addresses into a slice
	randomBucketSlice := make([]*KnownAddress, 0, len(randomBucket))
	for _, knownAddress := range randomBucket {
		randomBucketSlice = append(randomBucketSlice, knownAddress)
	}
	return randomBucketSlice
}

func (nb *newAddressBucketArray) name() string {
	return "new"
}

func (tb *triedAddressBucketArray) randomBucket(random *rand.Rand) []*KnownAddress {
	nonEmptyBuckets := make([][]*KnownAddress, 0, TriedBucketCount)
	for _, bucket := range tb {
		if len(bucket) > 0 {
			nonEmptyBuckets = append(nonEmptyBuckets, bucket)
		}
	}
	randomIndex := random.Intn(len(nonEmptyBuckets))
	return nonEmptyBuckets[randomIndex]
}

func (tb *triedAddressBucketArray) name() string {
	return "tried"
}

func (am *AddressManager) knownAddress(address *appmessage.NetAddress) *KnownAddress {
	return am.addressIndex[NetAddressKey(address)]
}

// Attempt increases the given address' attempt counter and updates
// the last attempt time.
func (am *AddressManager) Attempt(address *appmessage.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// find address.
	// Surely address will be in tried by now?
	knownAddress := am.knownAddress(address)
	if knownAddress == nil {
		return
	}
	// set last tried time to now
	knownAddress.attempts++
	knownAddress.lastAttempt = mstime.Now()
}

// Connected Marks the given address as currently connected and working at the
// current time. The address must already be known to AddressManager else it will
// be ignored.
func (am *AddressManager) Connected(address *appmessage.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	knownAddress := am.knownAddress(address)
	if knownAddress == nil {
		return
	}

	// Update the time as long as it has been 20 minutes since last we did
	// so.
	now := mstime.Now()
	if now.After(knownAddress.netAddress.Timestamp.Add(time.Minute * 20)) {
		// knownAddress.netAddress is immutable, so replace it.
		netAddressCopy := *knownAddress.netAddress
		netAddressCopy.Timestamp = mstime.Now()
		knownAddress.netAddress = &netAddressCopy
	}
}

// Good marks the given address as good. To be called after a successful
// connection and version exchange. If the address is unknown to the address
// manager it will be ignored.
func (am *AddressManager) Good(address *appmessage.NetAddress, subnetworkID *externalapi.DomainSubnetworkID) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	knownAddress := am.knownAddress(address)
	if knownAddress == nil {
		return
	}
	oldSubnetworkID := knownAddress.subnetworkID

	// knownAddress.Timestamp is not updated here to avoid leaking information
	// about currently connected peers.
	now := mstime.Now()
	knownAddress.lastSuccess = now
	knownAddress.lastAttempt = now
	knownAddress.attempts = 0
	knownAddress.subnetworkID = subnetworkID

	addressKey := NetAddressKey(address)
	triedAddressBucketIndex := am.triedAddressBucketIndex(knownAddress.netAddress)

	if knownAddress.tried {
		// If this address was already tried, and subnetworkID didn't change - don't do anything
		if *subnetworkID == *oldSubnetworkID {
			return
		}

		// If this address was already tried, but subnetworkID was changed -
		// update subnetworkID, than continue as though this is a new address
		bucket := am.subnetworkTriedAddresBucketArrays[*oldSubnetworkID][triedAddressBucketIndex]
		var toRemoveIndex int
		toRemoveIndexFound := false
		for i, knownAddress := range bucket {
			if NetAddressKey(knownAddress.NetAddress()) == addressKey {
				toRemoveIndex = i
				toRemoveIndexFound = true
				break
			}
		}
		if toRemoveIndexFound {
			am.subnetworkTriedAddresBucketArrays[*oldSubnetworkID][triedAddressBucketIndex] =
				append(bucket[:toRemoveIndex], bucket[toRemoveIndex+1:]...)
		}
	}

	// Ok, need to move it to tried.

	// Remove from all new buckets.
	// Record one of the buckets in question and call it the `oldBucketIndex'
	var oldBucketIndex int
	oldBucketIndexFound := false
	if !knownAddress.tried {
		newAddressBucketArray := am.newAddressBucketArray(oldSubnetworkID)
		for i := range newAddressBucketArray {
			// we check for existence so we can record the first one
			if _, ok := newAddressBucketArray[i][addressKey]; ok {
				if !oldBucketIndexFound {
					oldBucketIndex = i
					oldBucketIndexFound = true
				}

				delete(newAddressBucketArray[i], addressKey)
				knownAddress.referenceCount--
			}
		}

		am.decrementNewAddressCount(oldSubnetworkID)
	}

	// Room in this tried bucket?
	triedAddressBucketArray := am.triedAddressBucketArray(knownAddress.subnetworkID)
	triedAddressCount := am.triedAddressCount(knownAddress.subnetworkID)
	if triedAddressCount == 0 || len(triedAddressBucketArray[triedAddressBucketIndex]) < triedBucketSize {
		knownAddress.tried = true
		am.updateAddrTried(triedAddressBucketIndex, knownAddress)
		am.incrementTriedAddressCount(knownAddress.subnetworkID)
		return
	}

	// No room, we have to evict something else.
	knownAddressToRemove, knownAddressToRemoveIndex := am.pickTried(knownAddress.subnetworkID, triedAddressBucketIndex)

	// First bucket index it would have been put in.
	newAddressBucketIndex := am.newAddressBucketIndex(knownAddressToRemove.netAddress, knownAddressToRemove.sourceAddress)

	// If no room in the original bucket, we put it in a bucket we just
	// freed up a space in.
	newAddressBucketArray := am.newAddressBucketArray(knownAddress.subnetworkID)
	if len(newAddressBucketArray[newAddressBucketIndex]) >= newBucketSize {
		if !oldBucketIndexFound {
			// If address was a tried bucket with updated subnetworkID - oldBucketIndex will be equal to -1.
			// In that case - find some non-full bucket.
			// If no such bucket exists - throw knownAddressToRemove away
			for newBucket := range newAddressBucketArray {
				if len(newAddressBucketArray[newBucket]) < newBucketSize {
					break
				}
			}
		} else {
			newAddressBucketIndex = oldBucketIndex
		}
	}

	// Replace with knownAddress in the slice
	knownAddress.tried = true
	triedAddressBucketArray[triedAddressBucketIndex][knownAddressToRemoveIndex] = knownAddress

	knownAddressToRemove.tried = false
	knownAddressToRemove.referenceCount++

	// We don't touch a.subnetworkTriedAddressCounts here since the number of tried stays the same
	// but we decremented new above, raise it again since we're putting
	// something back.
	am.incrementNewAddressCount(knownAddress.subnetworkID)

	knownAddressToRemoveKey := NetAddressKey(knownAddressToRemove.netAddress)
	log.Tracef("Replacing %s with %s in tried", knownAddressToRemoveKey, addressKey)

	// We made sure there is space here just above.
	newAddressBucketArray[newAddressBucketIndex][knownAddressToRemoveKey] = knownAddressToRemove
}

func (am *AddressManager) newAddressBucketArray(subnetworkID *externalapi.DomainSubnetworkID) *newAddressBucketArray {
	if subnetworkID == nil {
		return &am.fullNodeNewAddressBucketArray
	}
	return am.subnetworkNewAddressBucketArrays[*subnetworkID]
}

func (am *AddressManager) triedAddressBucketArray(subnetworkID *externalapi.DomainSubnetworkID) *triedAddressBucketArray {
	if subnetworkID == nil {
		return &am.fullNodeTriedAddressBucketArray
	}
	return am.subnetworkTriedAddresBucketArrays[*subnetworkID]
}

func (am *AddressManager) incrementNewAddressCount(subnetworkID *externalapi.DomainSubnetworkID) {
	if subnetworkID == nil {
		am.fullNodeNewAddressCount++
		return
	}
	am.subnetworkNewAddressCounts[*subnetworkID]++
}

func (am *AddressManager) decrementNewAddressCount(subnetworkID *externalapi.DomainSubnetworkID) {
	if subnetworkID == nil {
		am.fullNodeNewAddressCount--
		return
	}
	am.subnetworkNewAddressCounts[*subnetworkID]--
}

func (am *AddressManager) triedAddressCount(subnetworkID *externalapi.DomainSubnetworkID) int {
	if subnetworkID == nil {
		return am.fullNodeTriedAddressCount
	}
	return am.subnetworkTriedAddressCounts[*subnetworkID]
}

func (am *AddressManager) newAddressCount(subnetworkID *externalapi.DomainSubnetworkID) int {
	if subnetworkID == nil {
		return am.fullNodeNewAddressCount
	}
	return am.subnetworkNewAddressCounts[*subnetworkID]
}

func (am *AddressManager) incrementTriedAddressCount(subnetworkID *externalapi.DomainSubnetworkID) {
	if subnetworkID == nil {
		am.fullNodeTriedAddressCount++
		return
	}
	am.subnetworkTriedAddressCounts[*subnetworkID]++
}

// AddLocalAddress adds netAddress to the list of known local addresses to advertise
// with the given priority.
func (am *AddressManager) AddLocalAddress(netAddress *appmessage.NetAddress, priority AddressPriority) error {
	if !am.IsRoutable(netAddress) {
		return errors.Errorf("address %s is not routable", netAddress.IP)
	}

	am.localAddressesLock.Lock()
	defer am.localAddressesLock.Unlock()

	addressKey := NetAddressKey(netAddress)
	address, ok := am.localAddresses[addressKey]
	if !ok || address.score < priority {
		if ok {
			address.score = priority + 1
		} else {
			am.localAddresses[addressKey] = &localAddress{
				netAddress: netAddress,
				score:      priority,
			}
		}
	}
	return nil
}

// getReachabilityFrom returns the relative reachability of the provided local
// address to the provided remote address.
func (am *AddressManager) getReachabilityFrom(localAddress, remoteAddress *appmessage.NetAddress) int {
	const (
		Unreachable = 0
		Default     = iota
		Teredo
		Ipv6Weak
		Ipv4
		Ipv6Strong
		Private
	)

	if !am.IsRoutable(remoteAddress) {
		return Unreachable
	}

	if IsRFC4380(remoteAddress) {
		if !am.IsRoutable(localAddress) {
			return Default
		}

		if IsRFC4380(localAddress) {
			return Teredo
		}

		if IsIPv4(localAddress) {
			return Ipv4
		}

		return Ipv6Weak
	}

	if IsIPv4(remoteAddress) {
		if am.IsRoutable(localAddress) && IsIPv4(localAddress) {
			return Ipv4
		}
		return Unreachable
	}

	/* ipv6 */
	var tunnelled bool
	// Is our v6 is tunnelled?
	if IsRFC3964(localAddress) || IsRFC6052(localAddress) || IsRFC6145(localAddress) {
		tunnelled = true
	}

	if !am.IsRoutable(localAddress) {
		return Default
	}

	if IsRFC4380(localAddress) {
		return Teredo
	}

	if IsIPv4(localAddress) {
		return Ipv4
	}

	if tunnelled {
		// only prioritise ipv6 if we aren't tunnelling it.
		return Ipv6Weak
	}

	return Ipv6Strong
}

// GetBestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (am *AddressManager) GetBestLocalAddress(remoteAddress *appmessage.NetAddress) *appmessage.NetAddress {
	am.localAddressesLock.Lock()
	defer am.localAddressesLock.Unlock()

	bestReach := 0
	var bestScore AddressPriority
	var bestAddress *appmessage.NetAddress
	for _, localAddress := range am.localAddresses {
		reach := am.getReachabilityFrom(localAddress.netAddress, remoteAddress)
		if reach > bestReach ||
			(reach == bestReach && localAddress.score > bestScore) {
			bestReach = reach
			bestScore = localAddress.score
			bestAddress = localAddress.netAddress
		}
	}
	if bestAddress != nil {
		log.Debugf("Suggesting address %s:%d for %s:%d", bestAddress.IP,
			bestAddress.Port, remoteAddress.IP, remoteAddress.Port)
	} else {
		log.Debugf("No worthy address for %s:%d", remoteAddress.IP,
			remoteAddress.Port)

		// Send something unroutable if nothing suitable.
		var ip net.IP
		if !IsIPv4(remoteAddress) {
			ip = net.IPv6zero
		} else {
			ip = net.IPv4zero
		}
		services := appmessage.SFNodeNetwork | appmessage.SFNodeBloom
		bestAddress = appmessage.NewNetAddressIPPort(ip, 0, services)
	}

	return bestAddress
}

// Ban marks the given address as banned
func (am *AddressManager) Ban(address *appmessage.NetAddress) error {
	return am.setBanned(address, true, mstime.Now())
}

// Unban marks the given address as not banned
func (am *AddressManager) Unban(address *appmessage.NetAddress) error {
	return am.setBanned(address, false, mstime.Time{})
}

func (am *AddressManager) setBanned(address *appmessage.NetAddress, isBanned bool, bannedTime mstime.Time) error {
	am.localAddressesLock.Lock()
	defer am.localAddressesLock.Unlock()

	knownAddress := am.knownAddress(address)
	if knownAddress == nil {
		return errors.Wrapf(ErrAddressNotFound, "address %s "+
			"is not registered with the address manager", address.TCPAddress())
	}
	knownAddress.isBanned = isBanned
	knownAddress.bannedTime = bannedTime
	return nil
}

// IsBanned returns whether the given address is banned
func (am *AddressManager) IsBanned(address *appmessage.NetAddress) (bool, error) {
	am.localAddressesLock.Lock()
	defer am.localAddressesLock.Unlock()

	knownAddress := am.knownAddress(address)
	if knownAddress == nil {
		return false, errors.Wrapf(ErrAddressNotFound, "address %s "+
			"is not registered with the address manager", address.TCPAddress())
	}
	return knownAddress.isBanned, nil
}

// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager
func (am *AddressManager) initListeners() error {
	if len(am.cfg.ExternalIPs) != 0 {
		defaultPort, err := strconv.ParseUint(am.cfg.NetParams().DefaultPort, 10, 16)
		if err != nil {
			log.Errorf("Can not parse default port %s for active DAG: %s",
				am.cfg.NetParams().DefaultPort, err)
			return err
		}

		for _, sip := range am.cfg.ExternalIPs {
			eport := uint16(defaultPort)
			host, portstr, err := net.SplitHostPort(sip)
			if err != nil {
				// no port, use default.
				host = sip
			} else {
				port, err := strconv.ParseUint(portstr, 10, 16)
				if err != nil {
					log.Warnf("Can not parse port from %s for "+
						"externalip: %s", sip, err)
					continue
				}
				eport = uint16(port)
			}
			na, err := am.HostToNetAddress(host, eport, appmessage.DefaultServices)
			if err != nil {
				log.Warnf("Not adding %s as externalip: %s", sip, err)
				continue
			}

			err = am.AddLocalAddress(na, ManualPrio)
			if err != nil {
				log.Warnf("Skipping specified external IP: %s", err)
			}
		}
	} else {
		// Listen for TCP connections at the configured addresses
		netAddrs, err := parseListeners(am.cfg.Listeners)
		if err != nil {
			return err
		}

		// Add bound addresses to address manager to be advertised to peers.
		for _, addr := range netAddrs {
			listener, err := net.Listen(addr.Network(), addr.String())
			if err != nil {
				log.Warnf("Can't listen on %s: %s", addr, err)
				continue
			}
			addr := listener.Addr().String()
			err = listener.Close()
			if err != nil {
				return err
			}
			err = am.addLocalAddress(addr)
			if err != nil {
				log.Warnf("Skipping bound address %s: %s", addr, err)
			}
		}
	}

	return nil
}

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			hostAddrs, err := net.LookupHost(host)
			if err != nil {
				return nil, err
			}
			ip = net.ParseIP(hostAddrs[0])
			if ip == nil {
				return nil, errors.Errorf("Cannot resolve IP address for host '%s'", host)
			}
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
		} else {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
		}
	}
	return netAddrs, nil
}

// addLocalAddress adds an address that this node is listening on to the
// address manager so that it may be relayed to peers.
func (am *AddressManager) addLocalAddress(addr string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return err
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		// If bound to unspecified address, advertise all local interfaces
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			ifaceIP, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			// If bound to 0.0.0.0, do not add IPv6 interfaces and if bound to
			// ::, do not add IPv4 interfaces.
			if (ip.To4() == nil) != (ifaceIP.To4() == nil) {
				continue
			}

			netAddr := appmessage.NewNetAddressIPPort(ifaceIP, uint16(port), appmessage.DefaultServices)
			am.AddLocalAddress(netAddr, BoundPrio)
		}
	} else {
		netAddr, err := am.HostToNetAddress(host, uint16(port), appmessage.DefaultServices)
		if err != nil {
			return err
		}

		am.AddLocalAddress(netAddr, BoundPrio)
	}

	return nil
}

// simpleAddr implements the net.Addr interface with two struct fields
type simpleAddr struct {
	net, addr string
}

// String returns the address.
//
// This is part of the net.Addr interface.
func (a simpleAddr) String() string {
	return a.addr
}

// Network returns the network.
//
// This is part of the net.Addr interface.
func (a simpleAddr) Network() string {
	return a.net
}

// Ensure simpleAddr implements the net.Addr interface.
var _ net.Addr = simpleAddr{}
