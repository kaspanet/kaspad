package addressmanager

import (
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

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

type localAddress struct {
	netAddress *appmessage.NetAddress
	score      AddressPriority
}

type localAddressManager struct {
	localAddresses map[addressKey]*localAddress
	lookupFunc     func(string) ([]net.IP, error)
	cfg            *Config
	mutex          sync.Mutex
}

func newLocalAddressManager(cfg *Config) (*localAddressManager, error) {
	localAddressManager := localAddressManager{
		localAddresses: map[addressKey]*localAddress{},
		cfg:            cfg,
		lookupFunc:     cfg.Lookup,
	}

	err := localAddressManager.initListeners()
	if err != nil {
		return nil, err
	}

	return &localAddressManager, nil
}

// addLocalNetAddress adds netAddress to the list of known local addresses to advertise
// with the given priority.
func (lam *localAddressManager) addLocalNetAddress(netAddress *appmessage.NetAddress, priority AddressPriority) error {
	if !IsRoutable(netAddress, lam.cfg.AcceptUnroutable) {
		return errors.Errorf("address %s is not routable", netAddress.IP)
	}

	lam.mutex.Lock()
	defer lam.mutex.Unlock()

	addressKey := netAddressKey(netAddress)
	address, ok := lam.localAddresses[addressKey]
	if !ok || address.score < priority {
		if ok {
			address.score = priority + 1
		} else {
			lam.localAddresses[addressKey] = &localAddress{
				netAddress: netAddress,
				score:      priority,
			}
		}
	}
	return nil
}

// bestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (lam *localAddressManager) bestLocalAddress(remoteAddress *appmessage.NetAddress) *appmessage.NetAddress {
	lam.mutex.Lock()
	defer lam.mutex.Unlock()

	bestReach := 0
	var bestScore AddressPriority
	var bestAddress *appmessage.NetAddress
	for _, localAddress := range lam.localAddresses {
		reach := reachabilityFrom(localAddress.netAddress, remoteAddress, lam.cfg.AcceptUnroutable)
		if reach > bestReach ||
			(reach == bestReach && localAddress.score > bestScore) {
			bestReach = reach
			bestScore = localAddress.score
			bestAddress = localAddress.netAddress
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
		bestAddress = appmessage.NewNetAddressIPPort(ip, 0)
	}

	return bestAddress
}

// addLocalAddress adds an address that this node is listening on to the
// address manager so that it may be relayed to peers.
func (lam *localAddressManager) addLocalAddress(addr string) error {
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

			netAddr := appmessage.NewNetAddressIPPort(ifaceIP, uint16(port))
			lam.addLocalNetAddress(netAddr, BoundPrio)
		}
	} else {
		netAddr, err := lam.hostToNetAddress(host, uint16(port))
		if err != nil {
			return err
		}

		lam.addLocalNetAddress(netAddr, BoundPrio)
	}

	return nil
}

// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager
func (lam *localAddressManager) initListeners() error {
	if len(lam.cfg.ExternalIPs) != 0 {
		defaultPort, err := strconv.ParseUint(lam.cfg.DefaultPort, 10, 16)
		if err != nil {
			log.Errorf("Can not parse default port %s for active DAG: %s",
				lam.cfg.DefaultPort, err)
			return err
		}

		for _, sip := range lam.cfg.ExternalIPs {
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
			na, err := lam.hostToNetAddress(host, eport)
			if err != nil {
				log.Warnf("Not adding %s as externalip: %s", sip, err)
				continue
			}

			err = lam.addLocalNetAddress(na, ManualPrio)
			if err != nil {
				log.Warnf("Skipping specified external IP: %s", err)
			}
		}
	} else {
		// Listen for TCP connections at the configured addresses
		netAddrs, err := parseListeners(lam.cfg.Listeners)
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
			err = lam.addLocalAddress(addr)
			if err != nil {
				log.Warnf("Skipping bound address %s: %s", addr, err)
			}
		}
	}

	return nil
}

// hostToNetAddress returns a netaddress given a host address. If
// the host is not an IP address it will be resolved.
func (lam *localAddressManager) hostToNetAddress(host string, port uint16) (*appmessage.NetAddress, error) {
	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := lam.lookupFunc(host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, errors.Errorf("no addresses found for %s", host)
		}
		ip = ips[0]
	}

	return appmessage.NewNetAddressIPPort(ip, port), nil
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
