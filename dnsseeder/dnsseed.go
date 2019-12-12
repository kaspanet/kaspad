// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/panics"

	"github.com/kaspanet/kaspad/connmgr"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/wire"
)

const (
	// nodeTimeout defines the timeout time waiting for
	// a response from a node.
	nodeTimeout = time.Second * 3

	// requiredServices describes the default services that are
	// required to be supported by outbound peers.
	requiredServices = wire.SFNodeNetwork
)

var (
	amgr             *Manager
	wg               sync.WaitGroup
	peersDefaultPort int
	systemShutdown   int32
)

// hostLookup returns the correct DNS lookup function to use depending on the
// passed host and configuration options. For example, .onion addresses will be
// resolved using the onion specific proxy if one was specified, but will
// otherwise treat the normal proxy as tor unless --noonion was specified in
// which case the lookup will fail. Meanwhile, normal IP addresses will be
// resolved using tor if a proxy was specified unless --noonion was also
// specified in which case the normal system DNS resolver will be used.
func hostLookup(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

func creep() {
	defer wg.Done()

	onAddr := make(chan struct{})
	onVersion := make(chan struct{})
	cfg := peer.Config{
		UserAgentName:    "daglabs-sniffer",
		UserAgentVersion: "0.0.1",
		DAGParams:        ActiveConfig().NetParams(),
		DisableRelayTx:   true,
		SelectedTip:      func() *daghash.Hash { return ActiveConfig().NetParams().GenesisBlock.BlockHash() },

		Listeners: peer.MessageListeners{
			OnAddr: func(p *peer.Peer, msg *wire.MsgAddr) {
				added := amgr.AddAddresses(msg.AddrList)
				log.Infof("Peer %v sent %v addresses, %d new",
					p.Addr(), len(msg.AddrList), added)
				onAddr <- struct{}{}
			},
			OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) {
				log.Infof("Adding peer %v with services %v and subnetword ID %v",
					p.NA().IP.String(), msg.Services, msg.SubnetworkID)
				// Mark this peer as a good node.
				amgr.Good(p.NA().IP, msg.Services, msg.SubnetworkID)
				// Ask peer for some addresses.
				p.QueueMessage(wire.NewMsgGetAddr(true, nil), nil)
				// notify that version is received and Peer's subnetwork ID is updated
				onVersion <- struct{}{}
			},
		},
	}

	var wgCreep sync.WaitGroup
	for {
		peers := amgr.Addresses()
		if len(peers) == 0 && amgr.AddressCount() == 0 {
			// Add peers discovered through DNS to the address manager.
			connmgr.SeedFromDNS(ActiveConfig().NetParams(), requiredServices, true, nil, hostLookup, func(addrs []*wire.NetAddress) {
				amgr.AddAddresses(addrs)
			})
			peers = amgr.Addresses()
		}
		if len(peers) == 0 {
			log.Infof("No stale addresses -- sleeping for 10 minutes")
			for i := 0; i < 600; i++ {
				time.Sleep(time.Second)
				if atomic.LoadInt32(&systemShutdown) != 0 {
					log.Infof("Creep thread shutdown")
					return
				}
			}
			continue
		}

		for _, addr := range peers {
			if atomic.LoadInt32(&systemShutdown) != 0 {
				log.Infof("Waiting creep threads to terminate")
				wgCreep.Wait()
				log.Infof("Creep thread shutdown")
				return
			}
			wgCreep.Add(1)
			go func(addr *wire.NetAddress) {
				defer wgCreep.Done()

				host := net.JoinHostPort(addr.IP.String(), strconv.Itoa(int(addr.Port)))
				p, err := peer.NewOutboundPeer(&cfg, host)
				if err != nil {
					log.Warnf("NewOutboundPeer on %v: %v",
						host, err)
					return
				}
				amgr.Attempt(addr.IP)
				conn, err := net.DialTimeout("tcp", p.Addr(), nodeTimeout)
				if err != nil {
					log.Warnf("%v", err)
					return
				}
				p.AssociateConnection(conn)

				// Wait version messsage or timeout in case of failure.
				select {
				case <-onVersion:
				case <-time.After(nodeTimeout):
					log.Warnf("version timeout on peer %v",
						p.Addr())
					p.Disconnect()
					return
				}

				select {
				case <-onAddr:
				case <-time.After(nodeTimeout):
					log.Warnf("getaddr timeout on peer %v",
						p.Addr())
					p.Disconnect()
					return
				}
				p.Disconnect()
			}(addr)
		}
		wgCreep.Wait()
	}
}

func main() {
	defer panics.HandlePanic(log, nil, nil)
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "loadConfig: %v\n", err)
		os.Exit(1)
	}
	amgr, err = NewManager(defaultHomeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewManager: %v\n", err)
		os.Exit(1)
	}

	peersDefaultPort, err = strconv.Atoi(ActiveConfig().NetParams().DefaultPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid peers default port %s: %v\n", ActiveConfig().NetParams().DefaultPort, err)
		os.Exit(1)
	}

	if len(cfg.Seeder) != 0 {
		ip := net.ParseIP(cfg.Seeder)
		if ip == nil {
			hostAddrs, err := net.LookupHost(cfg.Seeder)
			if err != nil {
				log.Warnf("Failed to resolve seed host: %v, %v, ignoring", cfg.Seeder, err)
			} else {
				ip = net.ParseIP(hostAddrs[0])
				if ip == nil {
					log.Warnf("Failed to resolve seed host: %v, ignoring", cfg.Seeder)
				}
			}
		}
		if ip != nil {
			amgr.AddAddresses([]*wire.NetAddress{
				wire.NewNetAddressIPPort(ip, uint16(peersDefaultPort),
					requiredServices)})
		}
	}

	wg.Add(1)
	spawn(creep)

	dnsServer := NewDNSServer(cfg.Host, cfg.Nameserver, cfg.Listen)
	wg.Add(1)
	spawn(dnsServer.Start)

	defer func() {
		log.Infof("Gracefully shutting down the seeder...")
		atomic.StoreInt32(&systemShutdown, 1)
		close(amgr.quit)
		wg.Wait()
		amgr.wg.Wait()
		log.Infof("Seeder shutdown complete")
	}()

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	interrupt := signal.InterruptListener()
	<-interrupt
}
