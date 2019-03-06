// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"fmt"
	mrand "math/rand"
	"net"
	"strconv"
	"time"

	"github.com/daglabs/btcd/util/subnetworkid"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/wire"
)

const (
	// These constants are used by the DNS seed code to pick a random last
	// seen time.
	secondsIn3Days int32 = 24 * 60 * 60 * 3
	secondsIn4Days int32 = 24 * 60 * 60 * 4

	// SubnetworkIDPrefixChar is prefix of subnetworkID, when building DNS seed request
	SubnetworkIDPrefixChar byte = 'n'

	// ServiceFlagPrefixChar is prefix of service flag, when building DNS seed request
	ServiceFlagPrefixChar byte = 'x'
)

// OnSeed is the signature of the callback function which is invoked when DNS
// seeding is succesfull.
type OnSeed func(addrs []*wire.NetAddress)

// LookupFunc is the signature of the DNS lookup function.
type LookupFunc func(string) ([]net.IP, error)

// SeedFromDNS uses DNS seeding to populate the address manager with peers.
func SeedFromDNS(dagParams *dagconfig.Params, reqServices wire.ServiceFlag, subnetworkID *subnetworkid.SubnetworkID,
	lookupFn LookupFunc, seedFn OnSeed) {

	for _, dnsseed := range dagParams.DNSSeeds {
		var host string
		if !dnsseed.HasFiltering || reqServices == wire.SFNodeNetwork {
			host = dnsseed.Host
		} else {
			host = fmt.Sprintf("%c%x.%s", ServiceFlagPrefixChar, uint64(reqServices), dnsseed.Host)
		}

		if !subnetworkID.IsEqual(&wire.SubnetworkIDSupportsAll) {
			host = fmt.Sprintf("%c%s.%s", SubnetworkIDPrefixChar, subnetworkID, host)
		}

		go func(host string) {
			randSource := mrand.New(mrand.NewSource(time.Now().UnixNano()))

			seedpeers, err := lookupFn(host)
			if err != nil {
				log.Infof("DNS discovery failed on seed %s: %s", host, err)
				return
			}
			numPeers := len(seedpeers)

			log.Infof("%d addresses found from DNS seed %s", numPeers, host)

			if numPeers == 0 {
				return
			}
			addresses := make([]*wire.NetAddress, len(seedpeers))
			// if this errors then we have *real* problems
			intPort, _ := strconv.Atoi(dagParams.DefaultPort)
			for i, peer := range seedpeers {
				addresses[i] = wire.NewNetAddressTimestamp(
					// bitcoind seeds with addresses from
					// a time randomly selected between 3
					// and 7 days ago.
					time.Now().Add(-1*time.Second*time.Duration(secondsIn3Days+
						randSource.Int31n(secondsIn4Days))),
					0, peer, uint16(intPort))
			}

			seedFn(addresses)
		}(host)
	}
}
