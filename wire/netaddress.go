// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/kaspanet/kaspad/util/mstime"
	"io"
	"net"
)

// maxNetAddressPayload returns the max payload size for a kaspa NetAddress
// based on the protocol version.
func maxNetAddressPayload(pver uint32) uint32 {
	// Services 8 bytes + ip 16 bytes + port 2 bytes + timestamp 8 bytes.
	return uint32(34)
}

// NetAddress defines information about a peer on the network including the time
// it was last seen, the services it supports, its IP address, and port.
type NetAddress struct {
	// Last time the address was seen.
	Timestamp mstime.Time

	// Bitfield which identifies the services supported by the address.
	Services ServiceFlag

	// IP address of the peer.
	IP net.IP

	// Port the peer is using. This is encoded in big endian on the wire
	// which differs from most everything else.
	Port uint16
}

// HasService returns whether the specified service is supported by the address.
func (na *NetAddress) HasService(service ServiceFlag) bool {
	return na.Services&service == service
}

// AddService adds service as a supported service by the peer generating the
// message.
func (na *NetAddress) AddService(service ServiceFlag) {
	na.Services |= service
}

// TCPAddress converts the NetAddress to *net.TCPAddr
func (na *NetAddress) TCPAddress() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   na.IP,
		Port: int(na.Port),
	}
}

// NewNetAddressIPPort returns a new NetAddress using the provided IP, port, and
// supported services with defaults for the remaining fields.
func NewNetAddressIPPort(ip net.IP, port uint16, services ServiceFlag) *NetAddress {
	return NewNetAddressTimestamp(mstime.Now(), services, ip, port)
}

// NewNetAddressTimestamp returns a new NetAddress using the provided
// timestamp, IP, port, and supported services. The timestamp is rounded to
// single millisecond precision.
func NewNetAddressTimestamp(
	timestamp mstime.Time, services ServiceFlag, ip net.IP, port uint16) *NetAddress {
	// Limit the timestamp to one millisecond precision since the protocol
	// doesn't support better.
	na := NetAddress{
		Timestamp: timestamp,
		Services:  services,
		IP:        ip,
		Port:      port,
	}
	return &na
}

// NewNetAddress returns a new NetAddress using the provided TCP address and
// supported services with defaults for the remaining fields.
func NewNetAddress(addr *net.TCPAddr, services ServiceFlag) *NetAddress {
	return NewNetAddressIPPort(addr.IP, uint16(addr.Port), services)
}

// readNetAddress reads an encoded NetAddress from r depending on the protocol
// version and whether or not the timestamp is included per ts. Some messages
// like version do not include the timestamp.
func readNetAddress(r io.Reader, pver uint32, na *NetAddress, ts bool) error {
	var ip [16]byte

	if ts {
		err := ReadElement(r, (*int64Time)(&na.Timestamp))
		if err != nil {
			return err
		}
	}

	err := readElements(r, &na.Services, &ip)
	if err != nil {
		return err
	}
	port, err := binaryserializer.Uint16(r, bigEndian)
	if err != nil {
		return err
	}

	*na = NetAddress{
		Timestamp: na.Timestamp,
		Services:  na.Services,
		IP:        net.IP(ip[:]),
		Port:      port,
	}
	return nil
}

// writeNetAddress serializes a NetAddress to w depending on the protocol
// version and whether or not the timestamp is included per ts. Some messages
// like version do not include the timestamp.
func writeNetAddress(w io.Writer, pver uint32, na *NetAddress, ts bool) error {
	if ts {
		err := WriteElement(w, na.Timestamp.UnixMilliseconds())
		if err != nil {
			return err
		}
	}

	// Ensure to always write 16 bytes even if the ip is nil.
	var ip [16]byte
	if na.IP != nil {
		copy(ip[:], na.IP.To16())
	}
	err := writeElements(w, na.Services, ip)
	if err != nil {
		return err
	}

	return binary.Write(w, bigEndian, na.Port)
}
