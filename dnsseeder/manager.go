// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
	"github.com/miekg/dns"
)

// Node repesents a node in the Kaspa network
type Node struct {
	Addr         *wire.NetAddress
	Services     wire.ServiceFlag
	LastAttempt  time.Time
	LastSuccess  time.Time
	LastSeen     time.Time
	SubnetworkID *subnetworkid.SubnetworkID
}

// Manager is dnsseeder's main worker-type, storing all information required
// for operation
type Manager struct {
	mtx sync.RWMutex

	nodes     map[string]*Node
	wg        sync.WaitGroup
	quit      chan struct{}
	peersFile string
}

const (
	// defaultMaxAddresses is the maximum number of addresses to return.
	defaultMaxAddresses = 16

	// defaultStaleTimeout is the time in which a host is considered
	// stale.
	defaultStaleTimeout = time.Hour

	// dumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	dumpAddressInterval = time.Second * 30

	// peersFilename is the name of the file.
	peersFilename = "nodes.json"

	// pruneAddressInterval is the interval used to run the address
	// pruner.
	pruneAddressInterval = time.Minute * 1

	// pruneExpireTimeout is the expire time in which a node is
	// considered dead.
	pruneExpireTimeout = time.Hour * 8
)

var (
	// rfc1918Nets specifies the IPv4 private address blocks as defined by
	// by RFC1918 (10.0.0.0/8, 172.16.0.0/12, and 192.168.0.0/16).
	rfc1918Nets = []net.IPNet{
		ipNet("10.0.0.0", 8, 32),
		ipNet("172.16.0.0", 12, 32),
		ipNet("192.168.0.0", 16, 32),
	}

	// rfc3964Net specifies the IPv6 to IPv4 encapsulation address block as
	// defined by RFC3964 (2002::/16).
	rfc3964Net = ipNet("2002::", 16, 128)

	// rfc4380Net specifies the IPv6 teredo tunneling over UDP address block
	// as defined by RFC4380 (2001::/32).
	rfc4380Net = ipNet("2001::", 32, 128)

	// rfc4843Net specifies the IPv6 ORCHID address block as defined by
	// RFC4843 (2001:10::/28).
	rfc4843Net = ipNet("2001:10::", 28, 128)

	// rfc4862Net specifies the IPv6 stateless address autoconfiguration
	// address block as defined by RFC4862 (FE80::/64).
	rfc4862Net = ipNet("FE80::", 64, 128)

	// rfc4193Net specifies the IPv6 unique local address block as defined
	// by RFC4193 (FC00::/7).
	rfc4193Net = ipNet("FC00::", 7, 128)
)

// ipNet returns a net.IPNet struct given the passed IP address string, number
// of one bits to include at the start of the mask, and the total number of bits
// for the mask.
func ipNet(ip string, ones, bits int) net.IPNet {
	return net.IPNet{IP: net.ParseIP(ip), Mask: net.CIDRMask(ones, bits)}
}

func isRoutable(addr net.IP) bool {
	if ActiveConfig().NetParams().AcceptUnroutable {
		return true
	}

	for _, n := range rfc1918Nets {
		if n.Contains(addr) {
			return false
		}
	}
	if rfc3964Net.Contains(addr) ||
		rfc4380Net.Contains(addr) ||
		rfc4843Net.Contains(addr) ||
		rfc4862Net.Contains(addr) ||
		rfc4193Net.Contains(addr) {
		return false
	}

	return true
}

// NewManager constructs and returns a new dnsseeder manager, with the provided dataDir
func NewManager(dataDir string) (*Manager, error) {
	amgr := Manager{
		nodes:     make(map[string]*Node),
		peersFile: filepath.Join(dataDir, peersFilename),
		quit:      make(chan struct{}),
	}

	err := amgr.deserializePeers()
	if err != nil {
		log.Warnf("Failed to parse file %s: %v", amgr.peersFile, err)
		// if it is invalid we nuke the old one unconditionally.
		err = os.Remove(amgr.peersFile)
		if err != nil {
			log.Warnf("Failed to remove corrupt peers file %s: %v",
				amgr.peersFile, err)
		}
	}

	amgr.wg.Add(1)
	spawn(amgr.addressHandler)

	return &amgr, nil
}

// AddAddresses adds an address to this dnsseeder manager, and returns the number of
// address currently held
func (m *Manager) AddAddresses(addrs []*wire.NetAddress) int {
	var count int

	m.mtx.Lock()
	for _, addr := range addrs {
		if !isRoutable(addr.IP) {
			continue
		}
		addrStr := addr.IP.String()

		_, exists := m.nodes[addrStr]
		if exists {
			m.nodes[addrStr].LastSeen = time.Now()
			continue
		}
		node := Node{
			Addr:     addr,
			LastSeen: time.Now(),
		}
		m.nodes[addrStr] = &node
		count++
	}
	m.mtx.Unlock()

	return count
}

// Addresses returns IPs that need to be tested again.
func (m *Manager) Addresses() []*wire.NetAddress {
	addrs := make([]*wire.NetAddress, 0, defaultMaxAddresses*8)
	now := time.Now()
	i := defaultMaxAddresses

	m.mtx.RLock()
	for _, node := range m.nodes {
		if i == 0 {
			break
		}
		if now.Sub(node.LastSuccess) < defaultStaleTimeout ||
			now.Sub(node.LastAttempt) < defaultStaleTimeout {
			continue
		}
		addrs = append(addrs, node.Addr)
		i--
	}
	m.mtx.RUnlock()

	return addrs
}

// AddressCount returns number of known nodes.
func (m *Manager) AddressCount() int {
	return len(m.nodes)
}

// GoodAddresses returns good working IPs that match both the
// passed DNS query type and have the requested services.
func (m *Manager) GoodAddresses(qtype uint16, services wire.ServiceFlag, includeAllSubnetworks bool, subnetworkID *subnetworkid.SubnetworkID) []*wire.NetAddress {
	addrs := make([]*wire.NetAddress, 0, defaultMaxAddresses)
	i := defaultMaxAddresses

	if qtype != dns.TypeA && qtype != dns.TypeAAAA {
		return addrs
	}

	now := time.Now()
	m.mtx.RLock()
	for _, node := range m.nodes {
		if i == 0 {
			break
		}

		if node.Addr.Port != uint16(peersDefaultPort) {
			continue
		}

		if !includeAllSubnetworks && !node.SubnetworkID.IsEqual(subnetworkID) {
			continue
		}

		if qtype == dns.TypeA && node.Addr.IP.To4() == nil {
			continue
		} else if qtype == dns.TypeAAAA && node.Addr.IP.To4() != nil {
			continue
		}

		if node.LastSuccess.IsZero() ||
			now.Sub(node.LastSuccess) > defaultStaleTimeout {
			continue
		}

		// Does the node have the requested services?
		if node.Services&services != services {
			continue
		}
		addrs = append(addrs, node.Addr)
		i--
	}
	m.mtx.RUnlock()

	return addrs
}

// Attempt updates the last connection attempt for the specified ip address to now
func (m *Manager) Attempt(ip net.IP) {
	m.mtx.Lock()
	node, exists := m.nodes[ip.String()]
	if exists {
		node.LastAttempt = time.Now()
	}
	m.mtx.Unlock()
}

// Good updates the last successful connection attempt for the specified ip address to now
func (m *Manager) Good(ip net.IP, services wire.ServiceFlag, subnetworkid *subnetworkid.SubnetworkID) {
	m.mtx.Lock()
	node, exists := m.nodes[ip.String()]
	if exists {
		node.Services = services
		node.LastSuccess = time.Now()
		node.SubnetworkID = subnetworkid
	}
	m.mtx.Unlock()
}

// addressHandler is the main handler for the address manager. It must be run
// as a goroutine.
func (m *Manager) addressHandler() {
	defer m.wg.Done()
	pruneAddressTicker := time.NewTicker(pruneAddressInterval)
	defer pruneAddressTicker.Stop()
	dumpAddressTicker := time.NewTicker(dumpAddressInterval)
	defer dumpAddressTicker.Stop()
out:
	for {
		select {
		case <-dumpAddressTicker.C:
			m.savePeers()
		case <-pruneAddressTicker.C:
			m.prunePeers()
		case <-m.quit:
			break out
		}
	}
	log.Infof("Address manager: saving peers")
	m.savePeers()
	log.Infof("Address manager shoutdown")
}

func (m *Manager) prunePeers() {
	var count int
	now := time.Now()
	m.mtx.Lock()
	for k, node := range m.nodes {
		if now.Sub(node.LastSeen) > pruneExpireTimeout {
			delete(m.nodes, k)
			count++
			continue
		}
		if !node.LastSuccess.IsZero() &&
			now.Sub(node.LastSuccess) > pruneExpireTimeout {
			delete(m.nodes, k)
			count++
			continue
		}
	}
	l := len(m.nodes)
	m.mtx.Unlock()

	log.Infof("Pruned %d addresses: %d remaining", count, l)
}

func (m *Manager) deserializePeers() error {
	filePath := m.peersFile
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	r, err := os.Open(filePath)
	if err != nil {
		return errors.Errorf("%s error opening file: %v", filePath, err)
	}
	defer r.Close()

	var nodes map[string]*Node
	dec := json.NewDecoder(r)
	err = dec.Decode(&nodes)
	if err != nil {
		return errors.Errorf("error reading %s: %v", filePath, err)
	}

	l := len(nodes)

	m.mtx.Lock()
	m.nodes = nodes
	m.mtx.Unlock()

	log.Infof("%d nodes loaded", l)
	return nil
}

func (m *Manager) savePeers() {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	// Write temporary peers file and then move it into place.
	tmpfile := m.peersFile + ".new"
	w, err := os.Create(tmpfile)
	if err != nil {
		log.Errorf("Error opening file %s: %v", tmpfile, err)
		return
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(&m.nodes); err != nil {
		log.Errorf("Failed to encode file %s: %v", tmpfile, err)
		return
	}
	if err := w.Close(); err != nil {
		log.Errorf("Error closing file %s: %v", tmpfile, err)
		return
	}
	if err := os.Rename(tmpfile, m.peersFile); err != nil {
		log.Errorf("Error writing file %s: %v", m.peersFile, err)
		return
	}
}
