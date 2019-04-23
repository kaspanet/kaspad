// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addrmgr

import (
	"container/list"
	crand "crypto/rand" // for seeding
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daglabs/btcd/util/subnetworkid"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

type newBucket [newBucketCount]map[string]*KnownAddress
type triedBucket [triedBucketCount]*list.List

// AddrManager provides a concurrency safe address manager for caching potential
// peers on the bitcoin network.
type AddrManager struct {
	mtx                sync.Mutex
	peersFile          string
	lookupFunc         func(string) ([]net.IP, error)
	rand               *rand.Rand
	key                [32]byte
	addrIndex          map[string]*KnownAddress // address key to ka for all addrs.
	addrNew            map[subnetworkid.SubnetworkID]*newBucket
	addrNewFullNodes   newBucket
	addrTried          map[subnetworkid.SubnetworkID]*triedBucket
	addrTriedFullNodes triedBucket
	addrTrying         map[*KnownAddress]bool
	started            int32
	shutdown           int32
	wg                 sync.WaitGroup
	quit               chan struct{}
	nTried             map[subnetworkid.SubnetworkID]int
	nNew               map[subnetworkid.SubnetworkID]int
	nTriedFullNodes    int
	nNewFullNodes      int
	lamtx              sync.Mutex
	localAddresses     map[string]*localAddress
	localSubnetworkID  *subnetworkid.SubnetworkID
}

type serializedKnownAddress struct {
	Addr         string
	Src          string
	SubnetworkID string
	Attempts     int
	TimeStamp    int64
	LastAttempt  int64
	LastSuccess  int64
	// no refcount or tried, that is available from context.
}

type serializedNewBucket [newBucketCount][]string
type serializedTriedBucket [triedBucketCount][]string

type serializedAddrManager struct {
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

	// triedBucketCount is the number of buckets we split tried
	// addresses over.
	triedBucketCount = 64

	// newBucketSize is the maximum number of addresses in each new address
	// bucket.
	newBucketSize = 64

	// newBucketCount is the number of buckets that we spread new addresses
	// over.
	newBucketCount = 1024

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

	// getAddrMax is the most addresses that we will send in response
	// to a getAddr (in practise the most addresses we will return from a
	// call to AddressCache()).
	getAddrMax = 2500

	// getAddrPercent is the percentage of total addresses known that we
	// will share with a call to AddressCache.
	getAddrPercent = 23

	// serialisationVersion is the current version of the on-disk format.
	serialisationVersion = 1
)

// updateAddress is a helper function to either update an address already known
// to the address manager, or to add the address if not already known.
func (a *AddrManager) updateAddress(netAddr, srcAddr *wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	// Filter out non-routable addresses. Note that non-routable
	// also includes invalid and local addresses.
	if !IsRoutable(netAddr) {
		return
	}

	addr := NetAddressKey(netAddr)
	ka := a.find(netAddr)
	if ka != nil {
		// TODO: only update addresses periodically.
		// Update the last seen time and services.
		// note that to prevent causing excess garbage on getaddr
		// messages the netaddresses in addrmaanger are *immutable*,
		// if we need to change them then we replace the pointer with a
		// new copy so that we don't have to copy every na for getaddr.
		if netAddr.Timestamp.After(ka.na.Timestamp) ||
			(ka.na.Services&netAddr.Services) !=
				netAddr.Services {

			naCopy := *ka.na
			naCopy.Timestamp = netAddr.Timestamp
			naCopy.AddService(netAddr.Services)
			ka.na = &naCopy
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
		if a.rand.Int31n(factor) != 0 {
			return
		}
	} else {
		// Make a copy of the net address to avoid races since it is
		// updated elsewhere in the addrmanager code and would otherwise
		// change the actual netaddress on the peer.
		netAddrCopy := *netAddr
		ka = &KnownAddress{na: &netAddrCopy, srcAddr: srcAddr, subnetworkID: subnetworkID}
		a.addrIndex[addr] = ka
		if subnetworkID == nil {
			a.nNewFullNodes++
		} else {
			a.nNew[*subnetworkID]++
		}
		// XXX time penalty?
	}

	bucket := a.getNewBucket(netAddr, srcAddr)

	// Already exists?
	if ka.subnetworkID == nil {
		if _, ok := a.addrNewFullNodes[bucket][addr]; ok {
			return
		}
	} else if a.addrNew[*ka.subnetworkID] != nil {
		if _, ok := a.addrNew[*ka.subnetworkID][bucket][addr]; ok {
			return
		}
	}

	// Enforce max addresses.
	if ka.subnetworkID == nil {
		if len(a.addrNewFullNodes[bucket]) > newBucketSize {
			log.Tracef("new bucket of full nodes is full, expiring old")
			a.expireNewFullNodes(bucket)
		}
	} else if a.addrNew[*ka.subnetworkID] != nil && len(a.addrNew[*ka.subnetworkID][bucket]) > newBucketSize {
		log.Tracef("new bucket is full, expiring old")
		a.expireNewBySubnetworkID(ka.subnetworkID, bucket)
	}

	// Add to new bucket.
	ka.refs++
	a.updateAddrNew(bucket, addr, ka)

	if ka.subnetworkID == nil {
		log.Tracef("Added new full node address %s for a total of %d addresses", addr,
			a.nTriedFullNodes+a.nNewFullNodes)
	} else {
		log.Tracef("Added new address %s for a total of %d addresses", addr,
			a.nTried[*ka.subnetworkID]+a.nNew[*ka.subnetworkID])
	}
}

func (a *AddrManager) updateAddrNew(bucket int, addr string, ka *KnownAddress) {
	if ka.subnetworkID == nil {
		a.addrNewFullNodes[bucket][addr] = ka
		return
	}

	if _, ok := a.addrNew[*ka.subnetworkID]; !ok {
		a.addrNew[*ka.subnetworkID] = &newBucket{}
		for i := range a.addrNew[*ka.subnetworkID] {
			a.addrNew[*ka.subnetworkID][i] = make(map[string]*KnownAddress)
		}
	}
	a.addrNew[*ka.subnetworkID][bucket][addr] = ka
}

func (a *AddrManager) updateAddrTried(bucket int, ka *KnownAddress) {
	if ka.subnetworkID == nil {
		a.addrTriedFullNodes[bucket].PushBack(ka)
		return
	}

	if _, ok := a.addrTried[*ka.subnetworkID]; !ok {
		a.addrTried[*ka.subnetworkID] = &triedBucket{}
		for i := range a.addrTried[*ka.subnetworkID] {
			a.addrTried[*ka.subnetworkID][i] = list.New()
		}
	}
	a.addrTried[*ka.subnetworkID][bucket].PushBack(ka)
}

// expireNew makes space in the new buckets by expiring the really bad entries.
// If no bad entries are available we look at a few and remove the oldest.
func (a *AddrManager) expireNew(bucket *newBucket, idx int, decrNewCounter func()) {
	// First see if there are any entries that are so bad we can just throw
	// them away. otherwise we throw away the oldest entry in the cache.
	// Bitcoind here chooses four random and just throws the oldest of
	// those away, but we keep track of oldest in the initial traversal and
	// use that information instead.
	var oldest *KnownAddress
	for k, v := range bucket[idx] {
		if v.isBad() {
			log.Tracef("expiring bad address %s", k)
			delete(bucket[idx], k)
			v.refs--
			if v.refs == 0 {
				decrNewCounter()
				delete(a.addrIndex, k)
			}
			continue
		}
		if oldest == nil {
			oldest = v
		} else if !v.na.Timestamp.After(oldest.na.Timestamp) {
			oldest = v
		}
	}

	if oldest != nil {
		key := NetAddressKey(oldest.na)
		log.Tracef("expiring oldest address %s", key)

		delete(bucket[idx], key)
		oldest.refs--
		if oldest.refs == 0 {
			decrNewCounter()
			delete(a.addrIndex, key)
		}
	}
}

// expireNewBySubnetworkID makes space in the new buckets by expiring the really bad entries.
// If no bad entries are available we look at a few and remove the oldest.
func (a *AddrManager) expireNewBySubnetworkID(subnetworkID *subnetworkid.SubnetworkID, bucket int) {
	a.expireNew(a.addrNew[*subnetworkID], bucket, func() { a.nNew[*subnetworkID]-- })
}

// expireNewFullNodes makes space in the new buckets by expiring the really bad entries.
// If no bad entries are available we look at a few and remove the oldest.
func (a *AddrManager) expireNewFullNodes(bucket int) {
	a.expireNew(&a.addrNewFullNodes, bucket, func() { a.nNewFullNodes-- })
}

// pickTried selects an address from the tried bucket to be evicted.
// We just choose the eldest. Bitcoind selects 4 random entries and throws away
// the older of them.
func (a *AddrManager) pickTried(subnetworkID *subnetworkid.SubnetworkID, bucket int) *list.Element {
	var oldest *KnownAddress
	var oldestElem *list.Element
	var lst *list.List
	if subnetworkID == nil {
		lst = a.addrTriedFullNodes[bucket]
	} else {
		lst = a.addrTried[*subnetworkID][bucket]
	}
	for e := lst.Front(); e != nil; e = e.Next() {
		ka := e.Value.(*KnownAddress)
		if oldest == nil || oldest.na.Timestamp.After(ka.na.Timestamp) {
			oldestElem = e
			oldest = ka
		}

	}
	return oldestElem
}

func (a *AddrManager) getNewBucket(netAddr, srcAddr *wire.NetAddress) int {
	// bitcoind:
	// doublesha256(key + sourcegroup + int64(doublesha256(key + group + sourcegroup))%bucket_per_source_group) % num_new_buckets

	data1 := []byte{}
	data1 = append(data1, a.key[:]...)
	data1 = append(data1, []byte(GroupKey(netAddr))...)
	data1 = append(data1, []byte(GroupKey(srcAddr))...)
	hash1 := daghash.DoubleHashB(data1)
	hash64 := binary.LittleEndian.Uint64(hash1)
	hash64 %= newBucketsPerGroup
	var hashbuf [8]byte
	binary.LittleEndian.PutUint64(hashbuf[:], hash64)
	data2 := []byte{}
	data2 = append(data2, a.key[:]...)
	data2 = append(data2, GroupKey(srcAddr)...)
	data2 = append(data2, hashbuf[:]...)

	hash2 := daghash.DoubleHashB(data2)
	return int(binary.LittleEndian.Uint64(hash2) % newBucketCount)
}

func (a *AddrManager) getTriedBucket(netAddr *wire.NetAddress) int {
	// bitcoind hashes this as:
	// doublesha256(key + group + truncate_to_64bits(doublesha256(key)) % buckets_per_group) % num_buckets
	data1 := []byte{}
	data1 = append(data1, a.key[:]...)
	data1 = append(data1, []byte(NetAddressKey(netAddr))...)
	hash1 := daghash.DoubleHashB(data1)
	hash64 := binary.LittleEndian.Uint64(hash1)
	hash64 %= triedBucketsPerGroup
	var hashbuf [8]byte
	binary.LittleEndian.PutUint64(hashbuf[:], hash64)
	data2 := []byte{}
	data2 = append(data2, a.key[:]...)
	data2 = append(data2, GroupKey(netAddr)...)
	data2 = append(data2, hashbuf[:]...)

	hash2 := daghash.DoubleHashB(data2)
	return int(binary.LittleEndian.Uint64(hash2) % triedBucketCount)
}

// addressHandler is the main handler for the address manager.  It must be run
// as a goroutine.
func (a *AddrManager) addressHandler() {
	dumpAddressTicker := time.NewTicker(dumpAddressInterval)
	defer dumpAddressTicker.Stop()
out:
	for {
		select {
		case <-dumpAddressTicker.C:
			a.savePeers()

		case <-a.quit:
			break out
		}
	}
	a.savePeers()
	a.wg.Done()
	log.Trace("Address handler done")
}

// savePeers saves all the known addresses to a file so they can be read back
// in at next run.
func (a *AddrManager) savePeers() {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	// First we make a serialisable datastructure so we can encode it to
	// json.
	sam := new(serializedAddrManager)
	sam.Version = serialisationVersion
	copy(sam.Key[:], a.key[:])

	sam.Addresses = make([]*serializedKnownAddress, len(a.addrIndex))
	i := 0
	for k, v := range a.addrIndex {
		ska := new(serializedKnownAddress)
		ska.Addr = k
		if v.subnetworkID == nil {
			ska.SubnetworkID = ""
		} else {
			ska.SubnetworkID = v.subnetworkID.String()
		}
		ska.TimeStamp = v.na.Timestamp.Unix()
		ska.Src = NetAddressKey(v.srcAddr)
		ska.Attempts = v.attempts
		ska.LastAttempt = v.lastattempt.Unix()
		ska.LastSuccess = v.lastsuccess.Unix()
		// Tried and refs are implicit in the rest of the structure
		// and will be worked out from context on unserialisation.
		sam.Addresses[i] = ska
		i++
	}

	sam.NewBuckets = make(map[string]*serializedNewBucket)
	for subnetworkID := range a.addrNew {
		subnetworkIDStr := subnetworkID.String()
		sam.NewBuckets[subnetworkIDStr] = &serializedNewBucket{}

		for i := range a.addrNew[subnetworkID] {
			sam.NewBuckets[subnetworkIDStr][i] = make([]string, len(a.addrNew[subnetworkID][i]))
			j := 0
			for k := range a.addrNew[subnetworkID][i] {
				sam.NewBuckets[subnetworkIDStr][i][j] = k
				j++
			}
		}
	}

	for i := range a.addrNewFullNodes {
		sam.NewBucketFullNodes[i] = make([]string, len(a.addrNewFullNodes[i]))
		j := 0
		for k := range a.addrNewFullNodes[i] {
			sam.NewBucketFullNodes[i][j] = k
			j++
		}
	}

	sam.TriedBuckets = make(map[string]*serializedTriedBucket)
	for subnetworkID := range a.addrTried {
		subnetworkIDStr := subnetworkID.String()
		sam.TriedBuckets[subnetworkIDStr] = &serializedTriedBucket{}

		for i := range a.addrTried[subnetworkID] {
			sam.TriedBuckets[subnetworkIDStr][i] = make([]string, a.addrTried[subnetworkID][i].Len())
			j := 0
			for e := a.addrTried[subnetworkID][i].Front(); e != nil; e = e.Next() {
				ka := e.Value.(*KnownAddress)
				sam.TriedBuckets[subnetworkIDStr][i][j] = NetAddressKey(ka.na)
				j++
			}
		}
	}

	for i := range a.addrTriedFullNodes {
		sam.TriedBucketFullNodes[i] = make([]string, a.addrTriedFullNodes[i].Len())
		j := 0
		for e := a.addrTriedFullNodes[i].Front(); e != nil; e = e.Next() {
			ka := e.Value.(*KnownAddress)
			sam.TriedBucketFullNodes[i][j] = NetAddressKey(ka.na)
			j++
		}
	}

	w, err := os.Create(a.peersFile)
	if err != nil {
		log.Errorf("Error opening file %s: %s", a.peersFile, err)
		return
	}
	enc := json.NewEncoder(w)
	defer w.Close()
	if err := enc.Encode(&sam); err != nil {
		log.Errorf("Failed to encode file %s: %s", a.peersFile, err)
		return
	}
}

// loadPeers loads the known address from the saved file.  If empty, missing, or
// malformed file, just don't load anything and start fresh
func (a *AddrManager) loadPeers() {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	err := a.deserializePeers(a.peersFile)
	if err != nil {
		log.Errorf("Failed to parse file %s: %s", a.peersFile, err)
		// if it is invalid we nuke the old one unconditionally.
		err = os.Remove(a.peersFile)
		if err != nil {
			log.Warnf("Failed to remove corrupt peers file %s: %s",
				a.peersFile, err)
		}
		a.reset()
		return
	}
	log.Infof("Loaded %d addresses from file '%s'", a.totalNumAddresses(), a.peersFile)
}

func (a *AddrManager) deserializePeers(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	r, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("%s error opening file: %s", filePath, err)
	}
	defer r.Close()

	var sam serializedAddrManager
	dec := json.NewDecoder(r)
	err = dec.Decode(&sam)
	if err != nil {
		return fmt.Errorf("error reading %s: %s", filePath, err)
	}

	if sam.Version != serialisationVersion {
		return fmt.Errorf("unknown version %d in serialized "+
			"addrmanager", sam.Version)
	}
	copy(a.key[:], sam.Key[:])

	for _, v := range sam.Addresses {
		ka := new(KnownAddress)
		ka.na, err = a.DeserializeNetAddress(v.Addr)
		if err != nil {
			return fmt.Errorf("failed to deserialize netaddress "+
				"%s: %s", v.Addr, err)
		}
		ka.srcAddr, err = a.DeserializeNetAddress(v.Src)
		if err != nil {
			return fmt.Errorf("failed to deserialize netaddress "+
				"%s: %s", v.Src, err)
		}
		if v.SubnetworkID != "" {
			ka.subnetworkID, err = subnetworkid.NewFromStr(v.SubnetworkID)
			if err != nil {
				return fmt.Errorf("failed to deserialize subnetwork id "+
					"%s: %s", v.SubnetworkID, err)
			}
		}
		ka.attempts = v.Attempts
		ka.lastattempt = time.Unix(v.LastAttempt, 0)
		ka.lastsuccess = time.Unix(v.LastSuccess, 0)
		a.addrIndex[NetAddressKey(ka.na)] = ka
	}

	for subnetworkIDStr := range sam.NewBuckets {
		subnetworkID, err := subnetworkid.NewFromStr(subnetworkIDStr)
		if err != nil {
			return err
		}
		for i, subnetworkNewBucket := range sam.NewBuckets[subnetworkIDStr] {
			for _, val := range subnetworkNewBucket {
				ka, ok := a.addrIndex[val]
				if !ok {
					return fmt.Errorf("newbucket contains %s but "+
						"none in address list", val)
				}

				if ka.refs == 0 {
					a.nNew[*subnetworkID]++
				}
				ka.refs++
				a.updateAddrNew(i, val, ka)
			}
		}
	}

	for i, newBucket := range sam.NewBucketFullNodes {
		for _, val := range newBucket {
			ka, ok := a.addrIndex[val]
			if !ok {
				return fmt.Errorf("full nodes newbucket contains %s but "+
					"none in address list", val)
			}

			if ka.refs == 0 {
				a.nNewFullNodes++
			}
			ka.refs++
			a.updateAddrNew(i, val, ka)
		}
	}

	for subnetworkIDStr := range sam.TriedBuckets {
		subnetworkID, err := subnetworkid.NewFromStr(subnetworkIDStr)
		if err != nil {
			return err
		}
		for i, subnetworkTriedBucket := range sam.TriedBuckets[subnetworkIDStr] {
			for _, val := range subnetworkTriedBucket {
				ka, ok := a.addrIndex[val]
				if !ok {
					return fmt.Errorf("Tried bucket contains %s but "+
						"none in address list", val)
				}

				ka.tried = true
				a.nTried[*subnetworkID]++
				a.addrTried[*subnetworkID][i].PushBack(ka)
			}
		}
	}

	for i, triedBucket := range sam.TriedBucketFullNodes {
		for _, val := range triedBucket {
			ka, ok := a.addrIndex[val]
			if !ok {
				return fmt.Errorf("Full nodes tried bucket contains %s but "+
					"none in address list", val)
			}

			ka.tried = true
			a.nTriedFullNodes++
			a.addrTriedFullNodes[i].PushBack(ka)
		}
	}

	// Sanity checking.
	for k, v := range a.addrIndex {
		if v.refs == 0 && !v.tried {
			return fmt.Errorf("address %s after serialisation "+
				"with no references", k)
		}

		if v.refs > 0 && v.tried {
			return fmt.Errorf("address %s after serialisation "+
				"which is both new and tried!", k)
		}
	}

	return nil
}

// DeserializeNetAddress converts a given address string to a *wire.NetAddress
func (a *AddrManager) DeserializeNetAddress(addr string) (*wire.NetAddress, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}

	return a.HostToNetAddress(host, uint16(port), wire.SFNodeNetwork)
}

// Start begins the core address handler which manages a pool of known
// addresses, timeouts, and interval based writes.
func (a *AddrManager) Start() {
	// Already started?
	if atomic.AddInt32(&a.started, 1) != 1 {
		return
	}

	log.Trace("Starting address manager")

	// Load peers we already know about from file.
	a.loadPeers()

	// Start the address ticker to save addresses periodically.
	a.wg.Add(1)
	go a.addressHandler()
}

// Stop gracefully shuts down the address manager by stopping the main handler.
func (a *AddrManager) Stop() error {
	if atomic.AddInt32(&a.shutdown, 1) != 1 {
		log.Warnf("Address manager is already in the process of " +
			"shutting down")
		return nil
	}

	log.Infof("Address manager shutting down")
	close(a.quit)
	a.wg.Wait()
	return nil
}

// AddAddresses adds new addresses to the address manager.  It enforces a max
// number of addresses and silently ignores duplicate addresses.  It is
// safe for concurrent access.
func (a *AddrManager) AddAddresses(addrs []*wire.NetAddress, srcAddr *wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	for _, na := range addrs {
		a.updateAddress(na, srcAddr, subnetworkID)
	}
}

// AddAddress adds a new address to the address manager.  It enforces a max
// number of addresses and silently ignores duplicate addresses.  It is
// safe for concurrent access.
func (a *AddrManager) AddAddress(addr, srcAddr *wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.updateAddress(addr, srcAddr, subnetworkID)
}

// AddAddressByIP adds an address where we are given an ip:port and not a
// wire.NetAddress.
func (a *AddrManager) AddAddressByIP(addrIP string, subnetworkID *subnetworkid.SubnetworkID) error {
	// Split IP and port
	addr, portStr, err := net.SplitHostPort(addrIP)
	if err != nil {
		return err
	}
	// Put it in wire.Netaddress
	ip := net.ParseIP(addr)
	if ip == nil {
		return fmt.Errorf("invalid ip address %s", addr)
	}
	port, err := strconv.ParseUint(portStr, 10, 0)
	if err != nil {
		return fmt.Errorf("invalid port %s: %s", portStr, err)
	}
	na := wire.NewNetAddressIPPort(ip, uint16(port), 0)
	a.AddAddress(na, na, subnetworkID) // XXX use correct src address
	return nil
}

// numAddresses returns the number of addresses that belongs to a specific subnetwork id
// which are known to the address manager.
func (a *AddrManager) numAddresses(subnetworkID *subnetworkid.SubnetworkID) int {
	if subnetworkID == nil {
		return a.nNewFullNodes + a.nTriedFullNodes
	}
	return a.nTried[*subnetworkID] + a.nNew[*subnetworkID]
}

// totalNumAddresses returns the number of addresses known to the address manager.
func (a *AddrManager) totalNumAddresses() int {
	total := a.nNewFullNodes + a.nTriedFullNodes
	for _, numAddresses := range a.nTried {
		total += numAddresses
	}
	for _, numAddresses := range a.nNew {
		total += numAddresses
	}
	return total
}

// TotalNumAddresses returns the number of addresses known to the address manager.
func (a *AddrManager) TotalNumAddresses() int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	return a.totalNumAddresses()
}

// NeedMoreAddresses returns whether or not the address manager needs more
// addresses.
func (a *AddrManager) NeedMoreAddresses() bool {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	allAddrs := a.numAddresses(a.localSubnetworkID)
	if a.localSubnetworkID != nil {
		allAddrs += a.numAddresses(nil)
	}
	return allAddrs < needAddressThreshold
}

// AddressCache returns the current address cache.  It must be treated as
// read-only (but since it is a copy now, this is not as dangerous).
func (a *AddrManager) AddressCache(includeAllSubnetworks bool, subnetworkID *subnetworkid.SubnetworkID) []*wire.NetAddress {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	if len(a.addrIndex) == 0 {
		return nil
	}

	allAddr := []*wire.NetAddress{}
	// Iteration order is undefined here, but we randomise it anyway.
	for _, v := range a.addrIndex {
		if includeAllSubnetworks || v.SubnetworkID().IsEqual(subnetworkID) {
			allAddr = append(allAddr, v.na)
		}
	}

	numAddresses := len(allAddr) * getAddrPercent / 100
	if numAddresses > getAddrMax {
		numAddresses = getAddrMax
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
func (a *AddrManager) reset() {

	a.addrIndex = make(map[string]*KnownAddress)

	// fill key with bytes from a good random source.
	io.ReadFull(crand.Reader, a.key[:])
	a.addrNew = make(map[subnetworkid.SubnetworkID]*newBucket)
	a.addrTried = make(map[subnetworkid.SubnetworkID]*triedBucket)

	a.nNew = make(map[subnetworkid.SubnetworkID]int)
	a.nTried = make(map[subnetworkid.SubnetworkID]int)

	for i := range a.addrNewFullNodes {
		a.addrNewFullNodes[i] = make(map[string]*KnownAddress)
	}
	for i := range a.addrTriedFullNodes {
		a.addrTriedFullNodes[i] = list.New()
	}
	a.nNewFullNodes = 0
	a.nTriedFullNodes = 0

	a.addrTrying = make(map[*KnownAddress]bool)
}

// HostToNetAddress returns a netaddress given a host address.  If the address
// is a Tor .onion address this will be taken care of.  Else if the host is
// not an IP address it will be resolved (via Tor if required).
func (a *AddrManager) HostToNetAddress(host string, port uint16, services wire.ServiceFlag) (*wire.NetAddress, error) {
	// Tor address is 16 char base32 + ".onion"
	var ip net.IP
	if len(host) == 22 && host[16:] == ".onion" {
		// go base32 encoding uses capitals (as does the rfc
		// but Tor and bitcoind tend to user lowercase, so we switch
		// case here.
		data, err := base32.StdEncoding.DecodeString(
			strings.ToUpper(host[:16]))
		if err != nil {
			return nil, err
		}
		prefix := []byte{0xfd, 0x87, 0xd8, 0x7e, 0xeb, 0x43}
		ip = net.IP(append(prefix, data...))
	} else if ip = net.ParseIP(host); ip == nil {
		ips, err := a.lookupFunc(host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("no addresses found for %s", host)
		}
		ip = ips[0]
	}

	return wire.NewNetAddressIPPort(ip, port, services), nil
}

// ipString returns a string for the ip from the provided NetAddress. If the
// ip is in the range used for Tor addresses then it will be transformed into
// the relevant .onion address.
func ipString(na *wire.NetAddress) string {
	if IsOnionCatTor(na) {
		// We know now that na.IP is long enough.
		base32 := base32.StdEncoding.EncodeToString(na.IP[6:])
		return strings.ToLower(base32) + ".onion"
	}

	return na.IP.String()
}

// NetAddressKey returns a string key in the form of ip:port for IPv4 addresses
// or [ip]:port for IPv6 addresses.
func NetAddressKey(na *wire.NetAddress) string {
	port := strconv.FormatUint(uint64(na.Port), 10)

	return net.JoinHostPort(ipString(na), port)
}

// GetAddress returns a single address that should be routable.  It picks a
// random one from the possible addresses with preference given to ones that
// have not been used recently and should not pick 'close' addresses
// consecutively.
func (a *AddrManager) GetAddress() *KnownAddress {
	// Protect concurrent access.
	a.mtx.Lock()
	defer a.mtx.Unlock()

	var knownAddress *KnownAddress
	if a.localSubnetworkID == nil {
		knownAddress = a.getAddress(&a.addrTriedFullNodes, a.nTriedFullNodes,
			&a.addrNewFullNodes, a.nNewFullNodes)
	} else {
		subnetworkID := *a.localSubnetworkID
		knownAddress = a.getAddress(a.addrTried[subnetworkID], a.nTried[subnetworkID],
			a.addrNew[subnetworkID], a.nNew[subnetworkID])
	}

	if knownAddress != nil {
		if a.addrTrying[knownAddress] {
			return nil
		}

		a.addrTrying[knownAddress] = true
	}

	return knownAddress

}

// see GetAddress for details
func (a *AddrManager) getAddress(addrTried *triedBucket, nTried int, addrNew *newBucket, nNew int) *KnownAddress {
	// Use a 50% chance for choosing between tried and new table entries.
	if nTried > 0 && (nNew == 0 || a.rand.Intn(2) == 0) {
		// Tried entry.
		large := 1 << 30
		factor := 1.0
		for {
			// pick a random bucket.
			bucket := a.rand.Intn(len(addrTried))
			if addrTried[bucket].Len() == 0 {
				continue
			}

			// Pick a random entry in the list
			e := addrTried[bucket].Front()
			for i :=
				a.rand.Int63n(int64(addrTried[bucket].Len())); i > 0; i-- {
				e = e.Next()
			}
			ka := e.Value.(*KnownAddress)
			randval := a.rand.Intn(large)
			if float64(randval) < (factor * ka.chance() * float64(large)) {
				log.Infof("Selected %s from tried bucket",
					NetAddressKey(ka.na))
				return ka
			}
			factor *= 1.2
		}
	} else if nNew > 0 {
		// new node.
		// XXX use a closure/function to avoid repeating this.
		large := 1 << 30
		factor := 1.0
		for {
			// Pick a random bucket.
			bucket := a.rand.Intn(len(addrNew))
			if len(addrNew[bucket]) == 0 {
				continue
			}
			// Then, a random entry in it.
			var ka *KnownAddress
			nth := a.rand.Intn(len(addrNew[bucket]))
			for _, value := range addrNew[bucket] {
				if nth == 0 {
					ka = value
				}
				nth--
			}
			randval := a.rand.Intn(large)
			if float64(randval) < (factor * ka.chance() * float64(large)) {
				log.Infof("Selected %s from new bucket",
					NetAddressKey(ka.na))
				return ka
			}
			factor *= 1.2
		}
	}
	return nil
}

func (a *AddrManager) find(addr *wire.NetAddress) *KnownAddress {
	return a.addrIndex[NetAddressKey(addr)]
}

// Attempt increases the given address' attempt counter and updates
// the last attempt time.
func (a *AddrManager) Attempt(addr *wire.NetAddress) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	// find address.
	// Surely address will be in tried by now?
	ka := a.find(addr)
	if ka == nil {
		return
	}
	// set last tried time to now
	ka.attempts++
	ka.lastattempt = time.Now()

	delete(a.addrTrying, ka)
}

// Connected Marks the given address as currently connected and working at the
// current time.  The address must already be known to AddrManager else it will
// be ignored.
func (a *AddrManager) Connected(addr *wire.NetAddress) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	ka := a.find(addr)
	if ka == nil {
		return
	}

	// Update the time as long as it has been 20 minutes since last we did
	// so.
	now := time.Now()
	if now.After(ka.na.Timestamp.Add(time.Minute * 20)) {
		// ka.na is immutable, so replace it.
		naCopy := *ka.na
		naCopy.Timestamp = time.Now()
		ka.na = &naCopy
	}
}

// Good marks the given address as good.  To be called after a successful
// connection and version exchange.  If the address is unknown to the address
// manager it will be ignored.
func (a *AddrManager) Good(addr *wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	ka := a.find(addr)
	if ka == nil {
		return
	}
	oldSubnetworkID := ka.subnetworkID

	// ka.Timestamp is not updated here to avoid leaking information
	// about currently connected peers.
	now := time.Now()
	ka.lastsuccess = now
	ka.lastattempt = now
	ka.attempts = 0
	ka.subnetworkID = subnetworkID

	addrKey := NetAddressKey(addr)
	triedBucketIndex := a.getTriedBucket(ka.na)

	if ka.tried {
		// If this address was already tried, and subnetworkID didn't change - don't do anything
		if subnetworkID.IsEqual(oldSubnetworkID) {
			return
		}

		// If this address was already tried, but subnetworkID was changed -
		// update subnetworkID, than continue as though this is a new address
		bucketList := a.addrTried[*oldSubnetworkID][triedBucketIndex]
		for e := bucketList.Front(); e != nil; e = e.Next() {
			if NetAddressKey(e.Value.(*KnownAddress).NetAddress()) == addrKey {
				bucketList.Remove(e)
				break
			}
		}
	}

	// Ok, need to move it to tried.

	// Remove from all new buckets.
	// Record one of the buckets in question and call it the `first'
	oldBucket := -1
	if !ka.tried {
		if oldSubnetworkID == nil {
			for i := range a.addrNewFullNodes {
				// we check for existence so we can record the first one
				if _, ok := a.addrNewFullNodes[i][addrKey]; ok {
					delete(a.addrNewFullNodes[i], addrKey)
					ka.refs--
					if oldBucket == -1 {
						oldBucket = i
					}
				}
			}
			a.nNewFullNodes--
		} else {
			for i := range a.addrNew[*oldSubnetworkID] {
				// we check for existence so we can record the first one
				if _, ok := a.addrNew[*oldSubnetworkID][i][addrKey]; ok {
					delete(a.addrNew[*oldSubnetworkID][i], addrKey)
					ka.refs--
					if oldBucket == -1 {
						oldBucket = i
					}
				}
			}
			a.nNew[*oldSubnetworkID]--
		}

		if oldBucket == -1 {
			// What? wasn't in a bucket after all.... Panic?
			return
		}
	}

	// Room in this tried bucket?
	if ka.subnetworkID == nil {
		if a.nTriedFullNodes == 0 || a.addrTriedFullNodes[triedBucketIndex].Len() < triedBucketSize {
			ka.tried = true
			a.updateAddrTried(triedBucketIndex, ka)
			a.nTriedFullNodes++
			return
		}
	} else if a.nTried[*ka.subnetworkID] == 0 || a.addrTried[*ka.subnetworkID][triedBucketIndex].Len() < triedBucketSize {
		ka.tried = true
		a.updateAddrTried(triedBucketIndex, ka)
		a.nTried[*ka.subnetworkID]++
		return
	}

	// No room, we have to evict something else.
	entry := a.pickTried(ka.subnetworkID, triedBucketIndex)
	rmka := entry.Value.(*KnownAddress)

	// First bucket it would have been put in.
	newBucket := a.getNewBucket(rmka.na, rmka.srcAddr)

	// If no room in the original bucket, we put it in a bucket we just
	// freed up a space in.
	if ka.subnetworkID == nil {
		if len(a.addrNewFullNodes[newBucket]) >= newBucketSize {
			if oldBucket == -1 {
				// If addr was a tried bucket with updated subnetworkID - oldBucket will be equal to -1.
				// In that case - find some non-full bucket.
				// If no such bucket exists - throw rmka away
				for newBucket := range a.addrNewFullNodes {
					if len(a.addrNewFullNodes[newBucket]) < newBucketSize {
						break
					}
				}
			} else {
				newBucket = oldBucket
			}
		}
	} else if len(a.addrNew[*ka.subnetworkID][newBucket]) >= newBucketSize {
		if len(a.addrNew[*ka.subnetworkID][newBucket]) >= newBucketSize {
			if oldBucket == -1 {
				// If addr was a tried bucket with updated subnetworkID - oldBucket will be equal to -1.
				// In that case - find some non-full bucket.
				// If no such bucket exists - throw rmka away
				for newBucket := range a.addrNew[*ka.subnetworkID] {
					if len(a.addrNew[*ka.subnetworkID][newBucket]) < newBucketSize {
						break
					}
				}
			} else {
				newBucket = oldBucket
			}
		}
	}

	// Replace with ka in list.
	ka.tried = true
	entry.Value = ka

	rmka.tried = false
	rmka.refs++

	// We don't touch a.nTried here since the number of tried stays the same
	// but we decemented new above, raise it again since we're putting
	// something back.
	if ka.subnetworkID == nil {
		a.nNewFullNodes++
	} else {
		a.nNew[*ka.subnetworkID]++
	}

	rmkey := NetAddressKey(rmka.na)
	log.Tracef("Replacing %s with %s in tried", rmkey, addrKey)

	// We made sure there is space here just above.
	if ka.subnetworkID == nil {
		a.addrNewFullNodes[newBucket][rmkey] = rmka
	} else {
		a.addrNew[*ka.subnetworkID][newBucket][rmkey] = rmka
	}
}

// AddLocalAddress adds na to the list of known local addresses to advertise
// with the given priority.
func (a *AddrManager) AddLocalAddress(na *wire.NetAddress, priority AddressPriority) error {
	if !IsRoutable(na) {
		return fmt.Errorf("address %s is not routable", na.IP)
	}

	a.lamtx.Lock()
	defer a.lamtx.Unlock()

	key := NetAddressKey(na)
	la, ok := a.localAddresses[key]
	if !ok || la.score < priority {
		if ok {
			la.score = priority + 1
		} else {
			a.localAddresses[key] = &localAddress{
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

	if IsOnionCatTor(remoteAddr) {
		if IsOnionCatTor(localAddr) {
			return Private
		}

		if IsRoutable(localAddr) && IsIPv4(localAddr) {
			return Ipv4
		}

		return Default
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
func (a *AddrManager) GetBestLocalAddress(remoteAddr *wire.NetAddress) *wire.NetAddress {
	a.lamtx.Lock()
	defer a.lamtx.Unlock()

	bestreach := 0
	var bestscore AddressPriority
	var bestAddress *wire.NetAddress
	for _, la := range a.localAddresses {
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
		if !IsIPv4(remoteAddr) && !IsOnionCatTor(remoteAddr) {
			ip = net.IPv6zero
		} else {
			ip = net.IPv4zero
		}
		services := wire.SFNodeNetwork | wire.SFNodeBloom
		bestAddress = wire.NewNetAddressIPPort(ip, 0, services)
	}

	return bestAddress
}

// New returns a new bitcoin address manager.
// Use Start to begin processing asynchronous address updates.
func New(dataDir string, lookupFunc func(string) ([]net.IP, error), subnetworkID *subnetworkid.SubnetworkID) *AddrManager {
	am := AddrManager{
		peersFile:         filepath.Join(dataDir, "peers.json"),
		lookupFunc:        lookupFunc,
		rand:              rand.New(rand.NewSource(time.Now().UnixNano())),
		quit:              make(chan struct{}),
		localAddresses:    make(map[string]*localAddress),
		localSubnetworkID: subnetworkID,
	}
	am.reset()
	return &am
}
