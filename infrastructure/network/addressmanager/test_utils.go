package addressmanager

import (
	"net"
	"strconv"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// AddAddressByIP adds an address where we are given an ip:port and not a
// appmessage.NetAddress.
func AddAddressByIP(am *AddressManager, addressIP string, subnetworkID *externalapi.DomainSubnetworkID) error {
	// Split IP and port
	ipString, portString, err := net.SplitHostPort(addressIP)
	if err != nil {
		return err
	}
	// Put it in appmessage.Netaddress
	ip := net.ParseIP(ipString)
	if ip == nil {
		return errors.Errorf("invalid ip %s", ipString)
	}
	port, err := strconv.ParseUint(portString, 10, 0)
	if err != nil {
		return errors.Errorf("invalid port %s: %s", portString, err)
	}
	netAddress := appmessage.NewNetAddressIPPort(ip, uint16(port))
	return am.AddAddresses(netAddress)
}
