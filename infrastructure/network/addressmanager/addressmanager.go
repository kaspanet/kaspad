// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addressmanager

import (
	"net"
	"sync"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/randomaddress"
	"github.com/pkg/errors"
)

// Configer is the interface for the config needed for the AddressManager.
type Configer interface {
	AcceptUnroutable() bool
}

// AddressRandomizer is the interface for the randomizer needed for the AddressManager.
type AddressRandomizer interface {
	RandomAddress(addresses []*appmessage.NetAddress) *appmessage.NetAddress
	RandomAddresses(addresses []*appmessage.NetAddress, count int) []*appmessage.NetAddress
}

// Config implement addressManagerConfig interface and represent the wrrapper for the config.Config needed for the AddressManager.
type Config struct {
	cfg *config.Config
}

// NewConfig returns a new address manager Config.
func NewConfig(cfg *config.Config) *Config {
	return &Config{
		cfg: cfg,
	}
}

// AcceptUnroutable specifies whether this network accepts unroutable
// IP addresses, such as 10.0.0.0/8
func (amc *Config) AcceptUnroutable() bool {
	return amc.cfg.NetParams().AcceptUnroutable
}

// AddressManager provides a concurrency safe address manager for caching potential
// peers on the Kaspa network.
type AddressManager struct {
	addresses map[AddressKey]*netAddressWrapper
	mutex     sync.Mutex
	cfg       Configer
	random    AddressRandomizer
}

type netAddressWrapper struct {
	netAddress *appmessage.NetAddress
	isBanned   bool
	isLocal    bool
}

// AddressKey represents a "string" key of the ip addresses
// for use as keys in maps.
type AddressKey string

// ErrAddressNotFound is an error returned from some functions when a
// given address is not found in the address manager
var ErrAddressNotFound = errors.New("address not found")

// NetAddressKey returns a key of the ip address to use it in maps.
func netAddressKey(netAddress *appmessage.NetAddress) AddressKey {
	return AddressKey(append(netAddress.IP, byte(netAddress.Port), byte(netAddress.Port>>8)))
}

// netAddressKeys returns a key of the ip address to use it in maps.
func netAddressesKeys(netAddresses []*appmessage.NetAddress) map[AddressKey]bool {
	result := make(map[AddressKey]bool, len(netAddresses))
	for _, netAddress := range netAddresses {
		key := netAddressKey(netAddress)
		result[key] = true
	}

	return result
}

// New returns a new Kaspa address manager.
func New(cfg Configer) (*AddressManager, error) {
	return &AddressManager{
		addresses: map[AddressKey]*netAddressWrapper{},
		random:    randomaddress.NewAddressRandomize(),
		cfg:       cfg,
	}, nil
}

// AddAddresses adds addresses to the address manager
func (am *AddressManager) AddAddresses(addresses ...*appmessage.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, address := range addresses {
		if !am.IsRoutable(address) {
			continue
		}

		key := netAddressKey(address)
		if _, ok := am.addresses[key]; !ok {
			am.addresses[key] = &netAddressWrapper{
				netAddress: address,
			}
		}
	}
}

// AddLocalAddresses adds local netAddresses to the address manager
func (am *AddressManager) AddLocalAddresses(addresses ...*appmessage.NetAddress) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, address := range addresses {
		if !am.IsRoutable(address) {
			continue
		}

		key := netAddressKey(address)
		if _, ok := am.addresses[key]; !ok {
			am.addresses[key] = &netAddressWrapper{
				netAddress: address,
				isLocal:    true,
			}
		}

	}
}

// Addresses returns all addresses
func (am *AddressManager) Addresses() []*appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	result := make([]*appmessage.NetAddress, 0, len(am.addresses))
	for _, address := range am.addresses {
		result = append(result, address.netAddress)
	}
	return result
}

// NotBannedAddressesWithException returns all not banned addresses with excpetion
func (am *AddressManager) NotBannedAddressesWithException(exceptions []*appmessage.NetAddress) []*appmessage.NetAddress {
	exceptionsKeys := netAddressesKeys(exceptions)
	am.mutex.Lock()
	defer am.mutex.Unlock()

	result := make([]*appmessage.NetAddress, 0, len(am.addresses))
	for key, address := range am.addresses {
		if !address.isBanned && !exceptionsKeys[key] {
			result = append(result, address.netAddress)
		}
	}

	return result
}

// RandomAddress returns a random address that isn't banned and isn't in exceptions
func (am *AddressManager) RandomAddress(exceptions []*appmessage.NetAddress) *appmessage.NetAddress {
	validAddresses := am.NotBannedAddressesWithException(exceptions)
	return am.random.RandomAddress(validAddresses)
}

// RandomAddresses returns count addresses at random that aren't banned and aren't in exceptions
func (am *AddressManager) RandomAddresses(count int, exceptions []*appmessage.NetAddress) []*appmessage.NetAddress {
	validAddresses := am.NotBannedAddressesWithException(exceptions)
	return am.random.RandomAddresses(validAddresses, count)
}

// BestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (am *AddressManager) BestLocalAddress(remoteAddress *appmessage.NetAddress) *appmessage.NetAddress {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	bestReach := 0
	var bestAddress *appmessage.NetAddress
	for _, address := range am.addresses {
		if address.isLocal {
			reach := reachabilityFrom(address.netAddress, remoteAddress, am.cfg.AcceptUnroutable())
			if reach > bestReach {
				bestReach = reach
				bestAddress = address.netAddress
			}
		}
	}

	if bestAddress == nil {
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
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	if address, ok := am.addresses[key]; ok {
		address.isBanned = true
		return nil
	}

	return errors.Wrapf(ErrAddressNotFound, "address %s "+
		"is not registered with the address manager", address.TCPAddress())
}

// Unban unmarks the given address as banned
func (am *AddressManager) Unban(address *appmessage.NetAddress) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	if address, ok := am.addresses[key]; ok {
		address.isBanned = false
		return nil
	}

	return errors.Wrapf(ErrAddressNotFound, "address %s "+
		"is not registered with the address manager", address.TCPAddress())
}

// IsBanned returns true if the given address is marked as banned
func (am *AddressManager) IsBanned(address *appmessage.NetAddress) (bool, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := netAddressKey(address)
	if address, ok := am.addresses[key]; ok {
		return address.isBanned, nil
	}

	return false, errors.Wrapf(ErrAddressNotFound, "address %s "+
		"is not registered with the address manager", address.TCPAddress())
}

// reachabilityFrom returns the relative reachability of the provided local
// address to the provided remote address.
func reachabilityFrom(localAddress, remoteAddress *appmessage.NetAddress, acceptUnroutable bool) int {
	const (
		Unreachable = 0
		Default     = iota
		Teredo
		Ipv6Weak
		Ipv4
		Ipv6Strong
		Private
	)

	IsRoutable := func(na *appmessage.NetAddress) bool {
		if acceptUnroutable {
			return !IsLocal(na)
		}

		return IsValid(na) && !(IsRFC1918(na) || IsRFC2544(na) ||
			IsRFC3927(na) || IsRFC4862(na) || IsRFC3849(na) ||
			IsRFC4843(na) || IsRFC5737(na) || IsRFC6598(na) ||
			IsLocal(na) || (IsRFC4193(na)))
	}

	if !IsRoutable(remoteAddress) {
		return Unreachable
	}

	if IsRFC4380(remoteAddress) {
		if !IsRoutable(localAddress) {
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
		if IsRoutable(localAddress) && IsIPv4(localAddress) {
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

	if !IsRoutable(localAddress) {
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
