// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addrmgr

import (
	"bytes"
	crand "crypto/rand" // for seeding
	"encoding/binary"
	"encoding/gob"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// AddressKey represents a "string" key in the form of ip:port for IPv4 addresses
// or [ip]:port for IPv6 addresses for use as keys in maps.
type AddressKey string
type newAddressBucketArray [NewBucketCount]map[AddressKey]*KnownAddress
type triedAddressBucketArray [TriedBucketCount][]*KnownAddress

// AddrManager provides a concurrency safe address manager for caching potential
// peers on the Kaspa network.
type AddrManager struct {
	mutex             sync.Mutex
	lookupFunc        func(string) ([]net.IP, error)
	rand              *rand.Rand
	key               [32]byte
	addressIndex      map[AddressKey]*KnownAddress // address keys to known addresses for all addresses.
	started           int32
	shutdown          int32
	wg                sync.WaitGroup
	quit              chan struct{}
	lamtx             sync.Mutex
	localAddresses    map[AddressKey]*localAddress
	localSubnetworkID *subnetworkid.SubnetworkID

	fullNodeNewAddressBucketArray     newAddressBucketArray
	fullNodeNewAddressCount           int
	fullNodeTriedAddressBucketArray   triedAddressBucketArray
	fullNodeTriedAddressCount         int
	subnetworkNewAddressBucketArrays  map[subnetworkid.SubnetworkID]*newAddressBucketArray
	subnetworkNewAddressCounts        map[subnetworkid.SubnetworkID]int
	subnetworkTriedAddresBucketArrays map[subnetworkid.SubnetworkID]*triedAddressBucketArray
	subnetworkTriedAddressCounts      map[subnetworkid.SubnetworkID]int
}

type serializedKnownAddress struct {
	Addr         AddressKey
	Src          AddressKey
	SubnetworkID string
	Attempts     int
	TimeStamp    int64
	LastAttempt  int64
	LastSuccess  int64
	// no refcount or tried, that is available from context.
}

type serializedNewBucket [NewBucketCount][]AddressKey
type serializedTriedBucket [TriedBucketCount][]AddressKey

// PeersStateForSerialization is the data model that is used to
// serialize the peers state to any encoding.
type PeersStateForSerialization struct {
	Version              int
	Key                  [32]byte
	Addresses            []*serializedKnownAddress
	NewBuckets           map[string]*serializedNewBucket // string is Subnetwork ID
	NewBucketFullNodes   serializedNewBucket
	TriedBuckets         map[string]*serializedTriedBucket // string is Subnetwork ID
	TriedBucketFullNodes serializedTriedBucket
}

type localAddress struct {
	na    *wire.NetAddress
	score AddressPriority
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
	// to a getAddr. If we have less than this amount, we send everything.
	getAddrMin = 50

	// GetAddrMax is the most addresses that we will send in response
	// to a getAddr (in practise the most addresses we will return from a
	// call to AddressCache()).
	GetAddrMax = 2500

	// getAddrPercent is the percentage of total addresses known that we
	// will share with a call to AddressCache.
	getAddrPercent = 23

	// serializationVersion is the current version of the on-disk format.
	serializationVersion = 1
)

// updateAddress is a helper function to either update an address already known
// to the address manager, or to add the address if not already known.
func (am *AddrManager) updateAddress(netAddr, srcAddr *wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	// Filter out non-routable addresses. Note that non-routable
	// also includes invalid and local addresses.
	if !IsRoutable(netAddr) {
		return
	}

	addressKey := NetAddressKey(netAddr)
	ka := am.find(netAddr)
	if ka != nil {
		// TODO: only update addresses periodically.
		// Update the last seen time and services.
		// note that to prevent causing excess garbage on getaddr
		// messages the netaddresses in addrmaanger are *immutable*,
		// if we need to change them then we replace the pointer with a
		// new copy so that we don't have to copy every netAddress for getaddr.
		if netAddr.Timestamp.After(ka.netAddress.Timestamp) ||
			(ka.netAddress.Services&netAddr.Services) !=
				netAddr.Services {

			naCopy := *ka.netAddress
			naCopy.Timestamp = netAddr.Timestamp
			naCopy.AddService(netAddr.Services)
			ka.netAddress = &naCopy
		}

		// If already in tried, we have nothing to do here.
		if ka.tried {
			return
		}

		// Already at our max?
		if ka.refs == newBucketsPerAddress {
			return
		}

		// The more entries we have, the less likely we are to add more.
		// likelihood is 2N.
		factor := int32(2 * ka.refs)
		if am.rand.Int31n(factor) != 0 {
			return
		}
	} else {
		// Make a copy of the net address to avoid races since it is
		// updated elsewhere in the addrmanager code and would otherwise
		// change the actual netaddress on the peer.
		netAddrCopy := *netAddr
		ka = &KnownAddress{netAddress: &netAddrCopy, srcAddr: srcAddr, subnetworkID: subnetworkID}
		am.addressIndex[addressKey] = ka
		am.incrementNewAddressCount(subnetworkID)
	}

	newAddressBucketIndex := am.getNewAddressBucketIndex(netAddr, srcAddr)

	// Enforce max addresses.
	newAddressBucketArray := am.newAddressBucketArray(ka.subnetworkID)
	if newAddressBucketArray != nil && len(newAddressBucketArray[newAddressBucketIndex]) > newBucketSize {
		log.Tracef("new bucket is full, expiring old")
		am.expireNew(ka.subnetworkID, newAddressBucketIndex)
	}

	// Add to new bucket.
	ka.refs++
	am.updateAddrNew(newAddressBucketIndex, addressKey, ka)

	totalAddressCount := am.newAddressCount(ka.subnetworkID) + am.triedAddressCount(ka.subnetworkID)
	log.Tracef("Added new address %s for a total of %d addresses", addressKey, totalAddressCount)

}

func (am *AddrManager) updateAddrNew(bucket int, addressKey AddressKey, ka *KnownAddress) {
	if ka.subnetworkID == nil {
		am.fullNodeNewAddressBucketArray[bucket][addressKey] = ka
		return
	}

	if _, ok := am.subnetworkNewAddressBucketArrays[*ka.subnetworkID]; !ok {
		am.subnetworkNewAddressBucketArrays[*ka.subnetworkID] = &newAddressBucketArray{}
		for i := range am.subnetworkNewAddressBucketArrays[*ka.subnetworkID] {
			am.subnetworkNewAddressBucketArrays[*ka.subnetworkID][i] = make(map[AddressKey]*KnownAddress)
		}
	}
	am.subnetworkNewAddressBucketArrays[*ka.subnetworkID][bucket][addressKey] = ka
}

func (am *AddrManager) updateAddrTried(bucket int, ka *KnownAddress) {
	if ka.subnetworkID == nil {
		am.fullNodeTriedAddressBucketArray[bucket] = append(am.fullNodeTriedAddressBucketArray[bucket], ka)
		return
	}

	if _, ok := am.subnetworkTriedAddresBucketArrays[*ka.subnetworkID]; !ok {
		am.subnetworkTriedAddresBucketArrays[*ka.subnetworkID] = &triedAddressBucketArray{}
		for i := range am.subnetworkTriedAddresBucketArrays[*ka.subnetworkID] {
			am.subnetworkTriedAddresBucketArrays[*ka.subnetworkID][i] = nil
		}
	}
	am.subnetworkTriedAddresBucketArrays[*ka.subnetworkID][bucket] = append(am.subnetworkTriedAddresBucketArrays[*ka.subnetworkID][bucket], ka)
}

// expireNew makes space in the new buckets by expiring the really bad entries.
// If no bad entries are available we look at a few and remove the oldest.
func (am *AddrManager) expireNew(subnetworkID *subnetworkid.SubnetworkID, bucketIndex int) {
	// First see if there are any entries that are so bad we can just throw
	// them away. otherwise we throw away the oldest entry in the cache.
	// We keep track of oldest in the initial traversal and use that
	// information instead.
	var oldest *KnownAddress
	newAddressBucketArray := am.newAddressBucketArray(subnetworkID)
	for k, v := range newAddressBucketArray[bucketIndex] {
		if v.isBad() {
			log.Tracef("expiring bad address %s", k)
			delete(newAddressBucketArray[bucketIndex], k)
			v.refs--
			if v.refs == 0 {
				am.decrementNewAddressCount(subnetworkID)
				delete(am.addressIndex, k)
			}
			continue
		}
		if oldest == nil {
			oldest = v
		} else if !v.netAddress.Timestamp.After(oldest.netAddress.Timestamp) {
			oldest = v
		}
	}

	if oldest != nil {
		key := NetAddressKey(oldest.netAddress)
		log.Tracef("expiring oldest address %s", key)

		delete(newAddressBucketArray[bucketIndex], key)
		oldest.refs--
		if oldest.refs == 0 {
			am.decrementNewAddressCount(subnetworkID)
			delete(am.addressIndex, key)
		}
	}
}

// pickTried selects an address from the tried bucket to be evicted.
// We just choose the eldest.
func (am *AddrManager) pickTried(subnetworkID *subnetworkid.SubnetworkID, bucketIndex int) (ka *KnownAddress, index int) {
	var oldest *KnownAddress
	oldestIndex := -1
	triedAddressBucketArray := am.triedAddressBucketArray(subnetworkID)
	for i, ka := range triedAddressBucketArray[bucketIndex] {
		if oldest == nil || oldest.netAddress.Timestamp.After(ka.netAddress.Timestamp) {
			oldestIndex = i
			oldest = ka
		}
	}
	return oldest, oldestIndex
}

func (am *AddrManager) getNewAddressBucketIndex(netAddr, srcAddr *wire.NetAddress) int {
	// doublesha256(key + sourcegroup + int64(doublesha256(key + group + sourcegroup))%bucket_per_source_group) % num_new_buckets

	data1 := []byte{}
	data1 = append(data1, am.key[:]...)
	data1 = append(data1, []byte(GroupKey(netAddr))...)
	data1 = append(data1, []byte(GroupKey(srcAddr))...)
	hash1 := daghash.DoubleHashB(data1)
	hash64 := binary.LittleEndian.Uint64(hash1)
	hash64 %= newBucketsPerGroup
	var hashbuf [8]byte
	binary.LittleEndian.PutUint64(hashbuf[:], hash64)
	data2 := []byte{}
	data2 = append(data2, am.key[:]...)
	data2 = append(data2, GroupKey(srcAddr)...)
	data2 = append(data2, hashbuf[:]...)

	hash2 := daghash.DoubleHashB(data2)
	return int(binary.LittleEndian.Uint64(hash2) % NewBucketCount)
}

func (am *AddrManager) getTriedAddressBucketIndex(netAddr *wire.NetAddress) int {
	// doublesha256(key + group + truncate_to_64bits(doublesha256(key)) % buckets_per_group) % num_buckets
	data1 := []byte{}
	data1 = append(data1, am.key[:]...)
	data1 = append(data1, []byte(NetAddressKey(netAddr))...)
	hash1 := daghash.DoubleHashB(data1)
	hash64 := binary.LittleEndian.Uint64(hash1)
	hash64 %= triedBucketsPerGroup
	var hashbuf [8]byte
	binary.LittleEndian.PutUint64(hashbuf[:], hash64)
	data2 := []byte{}
	data2 = append(data2, am.key[:]...)
	data2 = append(data2, GroupKey(netAddr)...)
	data2 = append(data2, hashbuf[:]...)

	hash2 := daghash.DoubleHashB(data2)
	return int(binary.LittleEndian.Uint64(hash2) % TriedBucketCount)
}

// addressHandler is the main handler for the address manager. It must be run
// as a goroutine.
func (am *AddrManager) addressHandler() {
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
func (am *AddrManager) savePeers() error {
	serializedPeersState, err := am.serializePeersState()
	if err != nil {
		return err
	}

	return dbaccess.StorePeersState(dbaccess.NoTx(), serializedPeersState)
}

func (am *AddrManager) serializePeersState() ([]byte, error) {
	peersState, err := am.PeersStateForSerialization()
	if err != nil {
		return nil, err
	}

	w := &bytes.Buffer{}
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(&peersState)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode peers state")
	}

	return w.Bytes(), nil
}

// PeersStateForSerialization returns the data model that is used to serialize the peers state to any encoding.
func (am *AddrManager) PeersStateForSerialization() (*PeersStateForSerialization, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// First we make a serializable data structure so we can encode it to
	// gob.
	peersState := new(PeersStateForSerialization)
	peersState.Version = serializationVersion
	copy(peersState.Key[:], am.key[:])

	peersState.Addresses = make([]*serializedKnownAddress, len(am.addressIndex))
	i := 0
	for k, v := range am.addressIndex {
		ska := new(serializedKnownAddress)
		ska.Addr = k
		if v.subnetworkID == nil {
			ska.SubnetworkID = ""
		} else {
			ska.SubnetworkID = v.subnetworkID.String()
		}
		ska.TimeStamp = v.netAddress.Timestamp.UnixMilliseconds()
		ska.Src = NetAddressKey(v.srcAddr)
		ska.Attempts = v.attempts
		ska.LastAttempt = v.lastattempt.UnixMilliseconds()
		ska.LastSuccess = v.lastsuccess.UnixMilliseconds()
		// Tried and refs are implicit in the rest of the structure
		// and will be worked out from context on unserialisation.
		peersState.Addresses[i] = ska
		i++
	}

	peersState.NewBuckets = make(map[string]*serializedNewBucket)
	for subnetworkID := range am.subnetworkNewAddressBucketArrays {
		subnetworkIDStr := subnetworkID.String()
		peersState.NewBuckets[subnetworkIDStr] = &serializedNewBucket{}

		for i := range am.subnetworkNewAddressBucketArrays[subnetworkID] {
			peersState.NewBuckets[subnetworkIDStr][i] = make([]AddressKey, len(am.subnetworkNewAddressBucketArrays[subnetworkID][i]))
			j := 0
			for k := range am.subnetworkNewAddressBucketArrays[subnetworkID][i] {
				peersState.NewBuckets[subnetworkIDStr][i][j] = k
				j++
			}
		}
	}

	for i := range am.fullNodeNewAddressBucketArray {
		peersState.NewBucketFullNodes[i] = make([]AddressKey, len(am.fullNodeNewAddressBucketArray[i]))
		j := 0
		for k := range am.fullNodeNewAddressBucketArray[i] {
			peersState.NewBucketFullNodes[i][j] = k
			j++
		}
	}

	peersState.TriedBuckets = make(map[string]*serializedTriedBucket)
	for subnetworkID := range am.subnetworkTriedAddresBucketArrays {
		subnetworkIDStr := subnetworkID.String()
		peersState.TriedBuckets[subnetworkIDStr] = &serializedTriedBucket{}

		for i := range am.subnetworkTriedAddresBucketArrays[subnetworkID] {
			peersState.TriedBuckets[subnetworkIDStr][i] = make([]AddressKey, len(am.subnetworkTriedAddresBucketArrays[subnetworkID][i]))
			j := 0
			for _, ka := range am.subnetworkTriedAddresBucketArrays[subnetworkID][i] {
				peersState.TriedBuckets[subnetworkIDStr][i][j] = NetAddressKey(ka.netAddress)
				j++
			}
		}
	}

	for i := range am.fullNodeTriedAddressBucketArray {
		peersState.TriedBucketFullNodes[i] = make([]AddressKey, len(am.fullNodeTriedAddressBucketArray[i]))
		j := 0
		for _, ka := range am.fullNodeTriedAddressBucketArray[i] {
			peersState.TriedBucketFullNodes[i][j] = NetAddressKey(ka.netAddress)
			j++
		}
	}

	return peersState, nil
}

// loadPeers loads the known address from the database. If missing,
// just don't load anything and start fresh.
func (am *AddrManager) loadPeers() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	serializedPeerState, err := dbaccess.FetchPeersState(dbaccess.NoTx())
	if dbaccess.IsNotFoundError(err) {
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

func (am *AddrManager) deserializePeersState(serializedPeerState []byte) error {
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

	for _, v := range peersState.Addresses {
		ka := new(KnownAddress)
		ka.netAddress, err = am.DeserializeNetAddress(v.Addr)
		if err != nil {
			return errors.Errorf("failed to deserialize netaddress "+
				"%s: %s", v.Addr, err)
		}
		ka.srcAddr, err = am.DeserializeNetAddress(v.Src)
		if err != nil {
			return errors.Errorf("failed to deserialize netaddress "+
				"%s: %s", v.Src, err)
		}
		if v.SubnetworkID != "" {
			ka.subnetworkID, err = subnetworkid.NewFromStr(v.SubnetworkID)
			if err != nil {
				return errors.Errorf("failed to deserialize subnetwork id "+
					"%s: %s", v.SubnetworkID, err)
			}
		}
		ka.attempts = v.Attempts
		ka.lastattempt = mstime.UnixMilliseconds(v.LastAttempt)
		ka.lastsuccess = mstime.UnixMilliseconds(v.LastSuccess)
		am.addressIndex[NetAddressKey(ka.netAddress)] = ka
	}

	for subnetworkIDStr := range peersState.NewBuckets {
		subnetworkID, err := subnetworkid.NewFromStr(subnetworkIDStr)
		if err != nil {
			return err
		}
		for i, subnetworkNewBucket := range peersState.NewBuckets[subnetworkIDStr] {
			for _, val := range subnetworkNewBucket {
				ka, ok := am.addressIndex[val]
				if !ok {
					return errors.Errorf("newbucket contains %s but "+
						"none in address list", val)
				}

				if ka.refs == 0 {
					am.subnetworkNewAddressCounts[*subnetworkID]++
				}
				ka.refs++
				am.updateAddrNew(i, val, ka)
			}
		}
	}

	for i, newBucket := range peersState.NewBucketFullNodes {
		for _, val := range newBucket {
			ka, ok := am.addressIndex[val]
			if !ok {
				return errors.Errorf("full nodes newbucket contains %s but "+
					"none in address list", val)
			}

			if ka.refs == 0 {
				am.fullNodeNewAddressCount++
			}
			ka.refs++
			am.updateAddrNew(i, val, ka)
		}
	}

	for subnetworkIDStr := range peersState.TriedBuckets {
		subnetworkID, err := subnetworkid.NewFromStr(subnetworkIDStr)
		if err != nil {
			return err
		}
		for i, subnetworkTriedBucket := range peersState.TriedBuckets[subnetworkIDStr] {
			for _, val := range subnetworkTriedBucket {
				ka, ok := am.addressIndex[val]
				if !ok {
					return errors.Errorf("Tried bucket contains %s but "+
						"none in address list", val)
				}

				ka.tried = true
				am.subnetworkTriedAddressCounts[*subnetworkID]++
				am.subnetworkTriedAddresBucketArrays[*subnetworkID][i] = append(am.subnetworkTriedAddresBucketArrays[*subnetworkID][i], ka)
			}
		}
	}

	for i, triedBucket := range peersState.TriedBucketFullNodes {
		for _, val := range triedBucket {
			ka, ok := am.addressIndex[val]
			if !ok {
				return errors.Errorf("Full nodes tried bucket contains %s but "+
					"none in address list", val)
			}

			ka.tried = true
			am.fullNodeTriedAddressCount++
			am.fullNodeTriedAddressBucketArray[i] = append(am.fullNodeTriedAddressBucketArray[i], ka)
		}
	}

	// Sanity checking.
	for k, v := range am.addressIndex {
		if v.refs == 0 && !v.tried {
			return errors.Errorf("address %s after serialisation "+
				"with no references", k)
		}

		if v.refs > 0 && v.tried {
			return errors.Errorf("address %s after serialisation "+
				"which is both new and tried!", k)
		}
	}

	return nil
}

// DeserializeNetAddress converts a given address string to a *wire.NetAddress
func (am *AddrManager) DeserializeNetAddress(addr AddressKey) (*wire.NetAddress, error) {
	host, portStr, err := net.SplitHostPort(string(addr))
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}

	return am.HostToNetAddress(host, uint16(port), wire.SFNodeNetwork)
}

// Start begins the core address handler which manages a pool of known
// addresses, timeouts, and interval based writes.
func (am *AddrManager) Start() error {
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
	spawn(am.addressHandler)
	return nil
}

// Stop gracefully shuts down the address manager by stopping the main handler.
func (am *AddrManager) Stop() error {
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
func (am *AddrManager) AddAddresses(addrs []*wire.NetAddress, srcAddr *wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, na := range addrs {
		am.updateAddress(na, srcAddr, subnetworkID)
	}
}

// AddAddress adds a new address to the address manager. It enforces a max
// number of addresses and silently ignores duplicate addresses. It is
// safe for concurrent access.
func (am *AddrManager) AddAddress(addr, srcAddr *wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.updateAddress(addr, srcAddr, subnetworkID)
}

// AddAddressByIP adds an address where we are given an ip:port and not a
// wire.NetAddress.
func (am *AddrManager) AddAddressByIP(addrIP string, subnetworkID *subnetworkid.SubnetworkID) error {
	// Split IP and port
	addr, portStr, err := net.SplitHostPort(addrIP)
	if err != nil {
		return err
	}
	// Put it in wire.Netaddress
	ip := net.ParseIP(addr)
	if ip == nil {
		return errors.Errorf("invalid ip address %s", addr)
	}
	port, err := strconv.ParseUint(portStr, 10, 0)
	if err != nil {
		return errors.Errorf("invalid port %s: %s", portStr, err)
	}
	na := wire.NewNetAddressIPPort(ip, uint16(port), 0)
	am.AddAddress(na, na, subnetworkID) // XXX use correct src address
	return nil
}

// numAddresses returns the number of addresses that belongs to a specific subnetwork id
// which are known to the address manager.
func (am *AddrManager) numAddresses(subnetworkID *subnetworkid.SubnetworkID) int {
	if subnetworkID == nil {
		return am.fullNodeNewAddressCount + am.fullNodeTriedAddressCount
	}
	return am.subnetworkTriedAddressCounts[*subnetworkID] + am.subnetworkNewAddressCounts[*subnetworkID]
}

// totalNumAddresses returns the number of addresses known to the address manager.
func (am *AddrManager) totalNumAddresses() int {
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
func (am *AddrManager) TotalNumAddresses() int {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.totalNumAddresses()
}

// NeedMoreAddresses returns whether or not the address manager needs more
// addresses.
func (am *AddrManager) NeedMoreAddresses() bool {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	allAddrs := am.numAddresses(am.localSubnetworkID)
	if am.localSubnetworkID != nil {
		allAddrs += am.numAddresses(nil)
	}
	return allAddrs < needAddressThreshold
}

// AddressCache returns the current address cache. It must be treated as
// read-only (but since it is a copy now, this is not as dangerous).
func (am *AddrManager) AddressCache(includeAllSubnetworks bool, subnetworkID *subnetworkid.SubnetworkID) []*wire.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if len(am.addressIndex) == 0 {
		return nil
	}

	allAddr := []*wire.NetAddress{}
	// Iteration order is undefined here, but we randomise it anyway.
	for _, v := range am.addressIndex {
		if includeAllSubnetworks || v.SubnetworkID().IsEqual(subnetworkID) {
			allAddr = append(allAddr, v.netAddress)
		}
	}

	numAddresses := len(allAddr) * getAddrPercent / 100
	if numAddresses > GetAddrMax {
		numAddresses = GetAddrMax
	}
	if len(allAddr) < getAddrMin {
		numAddresses = len(allAddr)
	}
	if len(allAddr) > getAddrMin && numAddresses < getAddrMin {
		numAddresses = getAddrMin
	}

	// Fisher-Yates shuffle the array. We only need to do the first
	// `numAddresses' since we are throwing the rest.
	for i := 0; i < numAddresses; i++ {
		// pick a number between current index and the end
		j := rand.Intn(len(allAddr)-i) + i
		allAddr[i], allAddr[j] = allAddr[j], allAddr[i]
	}

	// slice off the limit we are willing to share.
	return allAddr[0:numAddresses]
}

// reset resets the address manager by reinitialising the random source
// and allocating fresh empty bucket storage.
func (am *AddrManager) reset() {
	am.addressIndex = make(map[AddressKey]*KnownAddress)

	// fill key with bytes from a good random source.
	io.ReadFull(crand.Reader, am.key[:])
	am.subnetworkNewAddressBucketArrays = make(map[subnetworkid.SubnetworkID]*newAddressBucketArray)
	am.subnetworkTriedAddresBucketArrays = make(map[subnetworkid.SubnetworkID]*triedAddressBucketArray)

	am.subnetworkNewAddressCounts = make(map[subnetworkid.SubnetworkID]int)
	am.subnetworkTriedAddressCounts = make(map[subnetworkid.SubnetworkID]int)

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
func (am *AddrManager) HostToNetAddress(host string, port uint16, services wire.ServiceFlag) (*wire.NetAddress, error) {
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

	return wire.NewNetAddressIPPort(ip, port, services), nil
}

// NetAddressKey returns a "string" key in the form of ip:port for IPv4 addresses
// or [ip]:port for IPv6 addresses for use as keys in maps.
func NetAddressKey(na *wire.NetAddress) AddressKey {
	port := strconv.FormatUint(uint64(na.Port), 10)

	return AddressKey(net.JoinHostPort(na.IP.String(), port))
}

// GetAddress returns a single address that should be routable. It picks a
// random one from the possible addresses with preference given to ones that
// have not been used recently and should not pick 'close' addresses
// consecutively.
func (am *AddrManager) GetAddress() *KnownAddress {
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
func (am *AddrManager) getAddress(addrTried *triedAddressBucketArray, nTried int, addrNew *newAddressBucketArray, nNew int) *KnownAddress {
	// Use a 50% chance for choosing between tried and new table entries.
	var addrBucket bucket
	if nTried > 0 && (nNew == 0 || am.rand.Intn(2) == 0) {
		addrBucket = addrTried
	} else if nNew > 0 {
		addrBucket = addrNew
	} else {
		// There aren't any addresses in any of the buckets
		return nil
	}

	// Pick a random bucket
	randomBucket := addrBucket.randomBucket(am.rand)

	// Get the sum of all chances
	totalChance := float64(0)
	for _, ka := range randomBucket {
		totalChance += ka.chance()
	}

	// Pick a random address weighted by chance
	randomValue := am.rand.Float64()
	accumulatedChance := float64(0)
	for _, ka := range randomBucket {
		normalizedChance := ka.chance() / totalChance
		accumulatedChance += normalizedChance
		if randomValue < accumulatedChance {
			return ka
		}
	}

	panic("randomValue is exactly 1, which cannot happen")
}

type bucket interface {
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
	for _, ka := range randomBucket {
		randomBucketSlice = append(randomBucketSlice, ka)
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

func (am *AddrManager) find(address *wire.NetAddress) *KnownAddress {
	return am.addressIndex[NetAddressKey(address)]
}

// Attempt increases the given address' attempt counter and updates
// the last attempt time.
func (am *AddrManager) Attempt(address *wire.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// find address.
	// Surely address will be in tried by now?
	knownAddress := am.find(address)
	if knownAddress == nil {
		return
	}
	// set last tried time to now
	knownAddress.attempts++
	knownAddress.lastattempt = mstime.Now()
}

// Connected Marks the given address as currently connected and working at the
// current time. The address must already be known to AddrManager else it will
// be ignored.
func (am *AddrManager) Connected(address *wire.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	knownAddress := am.find(address)
	if knownAddress == nil {
		return
	}

	// Update the time as long as it has been 20 minutes since last we did
	// so.
	now := mstime.Now()
	if now.After(knownAddress.netAddress.Timestamp.Add(time.Minute * 20)) {
		// knownAddress.netAddress is immutable, so replace it.
		naCopy := *knownAddress.netAddress
		naCopy.Timestamp = mstime.Now()
		knownAddress.netAddress = &naCopy
	}
}

// Good marks the given address as good. To be called after a successful
// connection and version exchange. If the address is unknown to the address
// manager it will be ignored.
func (am *AddrManager) Good(address *wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	knownAddress := am.find(address)
	if knownAddress == nil {
		return
	}
	oldSubnetworkID := knownAddress.subnetworkID

	// knownAddress.Timestamp is not updated here to avoid leaking information
	// about currently connected peers.
	now := mstime.Now()
	knownAddress.lastsuccess = now
	knownAddress.lastattempt = now
	knownAddress.attempts = 0
	knownAddress.subnetworkID = subnetworkID

	addressKey := NetAddressKey(address)
	triedAddressBucketIndex := am.getTriedAddressBucketIndex(knownAddress.netAddress)

	if knownAddress.tried {
		// If this address was already tried, and subnetworkID didn't change - don't do anything
		if subnetworkID.IsEqual(oldSubnetworkID) {
			return
		}

		// If this address was already tried, but subnetworkID was changed -
		// update subnetworkID, than continue as though this is a new address
		bucket := am.subnetworkTriedAddresBucketArrays[*oldSubnetworkID][triedAddressBucketIndex]
		toRemoveIndex := -1
		for i, ka := range bucket {
			if NetAddressKey(ka.NetAddress()) == addressKey {
				toRemoveIndex = i
			}
		}
		if toRemoveIndex != -1 {
			am.subnetworkTriedAddresBucketArrays[*oldSubnetworkID][triedAddressBucketIndex] =
				append(bucket[:toRemoveIndex], bucket[toRemoveIndex+1:]...)
		}
	}

	// Ok, need to move it to tried.

	// Remove from all new buckets.
	// Record one of the buckets in question and call it the `first'
	oldBucket := -1
	if !knownAddress.tried {
		newAddressBucketArray := am.newAddressBucketArray(oldSubnetworkID)
		for i := range newAddressBucketArray {
			// we check for existence so we can record the first one
			if _, ok := newAddressBucketArray[i][addressKey]; ok {
				delete(newAddressBucketArray[i], addressKey)
				knownAddress.refs--
				if oldBucket == -1 {
					oldBucket = i
				}
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
	kaToRemove, kaToRemoveIndex := am.pickTried(knownAddress.subnetworkID, triedAddressBucketIndex)

	// First bucket index it would have been put in.
	newAddressBucketIndex := am.getNewAddressBucketIndex(kaToRemove.netAddress, kaToRemove.srcAddr)

	// If no room in the original bucket, we put it in a bucket we just
	// freed up a space in.
	newAddressBucketArray := am.newAddressBucketArray(knownAddress.subnetworkID)
	if len(newAddressBucketArray[newAddressBucketIndex]) >= newBucketSize {
		if oldBucket == -1 {
			// If address was a tried bucket with updated subnetworkID - oldBucket will be equal to -1.
			// In that case - find some non-full bucket.
			// If no such bucket exists - throw kaToRemove away
			for newBucket := range newAddressBucketArray {
				if len(newAddressBucketArray[newBucket]) < newBucketSize {
					break
				}
			}
		} else {
			newAddressBucketIndex = oldBucket
		}
	}

	// Replace with knownAddress in the slice
	knownAddress.tried = true
	triedAddressBucketArray[triedAddressBucketIndex][kaToRemoveIndex] = knownAddress

	kaToRemove.tried = false
	kaToRemove.refs++

	// We don't touch a.subnetworkTriedAddressCounts here since the number of tried stays the same
	// but we decremented new above, raise it again since we're putting
	// something back.
	am.incrementNewAddressCount(knownAddress.subnetworkID)

	kaToRemoveKey := NetAddressKey(kaToRemove.netAddress)
	log.Tracef("Replacing %s with %s in tried", kaToRemoveKey, addressKey)

	// We made sure there is space here just above.
	newAddressBucketArray[newAddressBucketIndex][kaToRemoveKey] = kaToRemove
}

func (am *AddrManager) newAddressBucketArray(subnetworkID *subnetworkid.SubnetworkID) *newAddressBucketArray {
	if subnetworkID == nil {
		return &am.fullNodeNewAddressBucketArray
	}
	return am.subnetworkNewAddressBucketArrays[*subnetworkID]
}

func (am *AddrManager) triedAddressBucketArray(subnetworkID *subnetworkid.SubnetworkID) *triedAddressBucketArray {
	if subnetworkID == nil {
		return &am.fullNodeTriedAddressBucketArray
	}
	return am.subnetworkTriedAddresBucketArrays[*subnetworkID]
}

func (am *AddrManager) incrementNewAddressCount(subnetworkID *subnetworkid.SubnetworkID) {
	if subnetworkID == nil {
		am.fullNodeNewAddressCount++
		return
	}
	am.subnetworkNewAddressCounts[*subnetworkID]++
}

func (am *AddrManager) decrementNewAddressCount(subnetworkID *subnetworkid.SubnetworkID) {
	if subnetworkID == nil {
		am.fullNodeNewAddressCount--
		return
	}
	am.subnetworkNewAddressCounts[*subnetworkID]--
}

func (am *AddrManager) triedAddressCount(subnetworkID *subnetworkid.SubnetworkID) int {
	if subnetworkID == nil {
		return am.fullNodeTriedAddressCount
	}
	return am.subnetworkTriedAddressCounts[*subnetworkID]
}

func (am *AddrManager) newAddressCount(subnetworkID *subnetworkid.SubnetworkID) int {
	if subnetworkID == nil {
		return am.fullNodeNewAddressCount
	}
	return am.subnetworkNewAddressCounts[*subnetworkID]
}

func (am *AddrManager) incrementTriedAddressCount(subnetworkID *subnetworkid.SubnetworkID) {
	if subnetworkID == nil {
		am.fullNodeTriedAddressCount++
		return
	}
	am.subnetworkTriedAddressCounts[*subnetworkID]++
}

// AddLocalAddress adds netAddress to the list of known local addresses to advertise
// with the given priority.
func (am *AddrManager) AddLocalAddress(na *wire.NetAddress, priority AddressPriority) error {
	if !IsRoutable(na) {
		return errors.Errorf("address %s is not routable", na.IP)
	}

	am.lamtx.Lock()
	defer am.lamtx.Unlock()

	key := NetAddressKey(na)
	la, ok := am.localAddresses[key]
	if !ok || la.score < priority {
		if ok {
			la.score = priority + 1
		} else {
			am.localAddresses[key] = &localAddress{
				na:    na,
				score: priority,
			}
		}
	}
	return nil
}

// getReachabilityFrom returns the relative reachability of the provided local
// address to the provided remote address.
func getReachabilityFrom(localAddr, remoteAddr *wire.NetAddress) int {
	const (
		Unreachable = 0
		Default     = iota
		Teredo
		Ipv6Weak
		Ipv4
		Ipv6Strong
		Private
	)

	if !IsRoutable(remoteAddr) {
		return Unreachable
	}

	if IsRFC4380(remoteAddr) {
		if !IsRoutable(localAddr) {
			return Default
		}

		if IsRFC4380(localAddr) {
			return Teredo
		}

		if IsIPv4(localAddr) {
			return Ipv4
		}

		return Ipv6Weak
	}

	if IsIPv4(remoteAddr) {
		if IsRoutable(localAddr) && IsIPv4(localAddr) {
			return Ipv4
		}
		return Unreachable
	}

	/* ipv6 */
	var tunnelled bool
	// Is our v6 is tunnelled?
	if IsRFC3964(localAddr) || IsRFC6052(localAddr) || IsRFC6145(localAddr) {
		tunnelled = true
	}

	if !IsRoutable(localAddr) {
		return Default
	}

	if IsRFC4380(localAddr) {
		return Teredo
	}

	if IsIPv4(localAddr) {
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
func (am *AddrManager) GetBestLocalAddress(remoteAddr *wire.NetAddress) *wire.NetAddress {
	am.lamtx.Lock()
	defer am.lamtx.Unlock()

	bestreach := 0
	var bestscore AddressPriority
	var bestAddress *wire.NetAddress
	for _, la := range am.localAddresses {
		reach := getReachabilityFrom(la.na, remoteAddr)
		if reach > bestreach ||
			(reach == bestreach && la.score > bestscore) {
			bestreach = reach
			bestscore = la.score
			bestAddress = la.na
		}
	}
	if bestAddress != nil {
		log.Debugf("Suggesting address %s:%d for %s:%d", bestAddress.IP,
			bestAddress.Port, remoteAddr.IP, remoteAddr.Port)
	} else {
		log.Debugf("No worthy address for %s:%d", remoteAddr.IP,
			remoteAddr.Port)

		// Send something unroutable if nothing suitable.
		var ip net.IP
		if !IsIPv4(remoteAddr) {
			ip = net.IPv6zero
		} else {
			ip = net.IPv4zero
		}
		services := wire.SFNodeNetwork | wire.SFNodeBloom
		bestAddress = wire.NewNetAddressIPPort(ip, 0, services)
	}

	return bestAddress
}

// New returns a new Kaspa address manager.
// Use Start to begin processing asynchronous address updates.
func New(lookupFunc func(string) ([]net.IP, error), subnetworkID *subnetworkid.SubnetworkID) *AddrManager {
	am := AddrManager{
		lookupFunc:        lookupFunc,
		rand:              rand.New(rand.NewSource(time.Now().UnixNano())),
		quit:              make(chan struct{}),
		localAddresses:    make(map[AddressKey]*localAddress),
		localSubnetworkID: subnetworkID,
	}
	am.reset()
	return &am
}
