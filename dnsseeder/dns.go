// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/pkg/errors"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/connmgr"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
	"github.com/miekg/dns"
)

// DNSServer struct
type DNSServer struct {
	hostname   string
	listen     string
	nameserver string
}

// Start - starts server
func (d *DNSServer) Start() {
	defer wg.Done()

	rr := fmt.Sprintf("%s 86400 IN NS %s", d.hostname, d.nameserver)
	authority, err := dns.NewRR(rr)
	if err != nil {
		log.Infof("NewRR: %v", err)
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", d.listen)
	if err != nil {
		log.Infof("ResolveUDPAddr: %v", err)
		return
	}

	udpListen, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Infof("ListenUDP: %v", err)
		return
	}
	defer udpListen.Close()

	for {
		b := make([]byte, 512)
	mainLoop:
		err := udpListen.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			log.Infof("SetReadDeadline: %v", err)
			os.Exit(1)
		}
		_, addr, err := udpListen.ReadFromUDP(b)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				if atomic.LoadInt32(&systemShutdown) == 0 {
					// use goto in order to do not re-allocate 'b' buffer
					goto mainLoop
				}
				log.Infof("DNS server shutdown")
				return
			}
			log.Infof("Read: %T", err.(*net.OpError).Err)
			continue
		}

		wg.Add(1)

		spawn(func() { d.handleDNSRequest(addr, authority, udpListen, b) })
	}
}

// NewDNSServer - create DNS server
func NewDNSServer(hostname, nameserver, listen string) *DNSServer {
	if hostname[len(hostname)-1] != '.' {
		hostname = hostname + "."
	}
	if nameserver[len(nameserver)-1] != '.' {
		nameserver = nameserver + "."
	}

	return &DNSServer{
		hostname:   hostname,
		listen:     listen,
		nameserver: nameserver,
	}
}

func (d *DNSServer) extractServicesSubnetworkID(addr *net.UDPAddr, domainName string) (wire.ServiceFlag, *subnetworkid.SubnetworkID, bool, error) {
	// Domain name may be in following format:
	//   [n[subnetwork].][xservice.]hostname
	// where connmgr.SubnetworkIDPrefixChar and connmgr.ServiceFlagPrefixChar are prefexes
	wantedSF := wire.SFNodeNetwork
	var subnetworkID *subnetworkid.SubnetworkID
	includeAllSubnetworks := true
	if d.hostname != domainName {
		idx := 0
		labels := dns.SplitDomainName(domainName)
		if labels[0][0] == connmgr.SubnetworkIDPrefixChar {
			includeAllSubnetworks = false
			if len(labels[0]) > 1 {
				idx = 1
				subnetworkID, err := subnetworkid.NewFromStr(labels[0][1:])
				if err != nil {
					log.Infof("%s: subnetworkid.NewFromStr: %v", addr, err)
					return wantedSF, subnetworkID, includeAllSubnetworks, err
				}
			}
		}
		if labels[idx][0] == connmgr.ServiceFlagPrefixChar && len(labels[idx]) > 1 {
			wantedSFStr := labels[idx][1:]
			u, err := strconv.ParseUint(wantedSFStr, 10, 64)
			if err != nil {
				log.Infof("%s: ParseUint: %v", addr, err)
				return wantedSF, subnetworkID, includeAllSubnetworks, err
			}
			wantedSF = wire.ServiceFlag(u)
		}
	}
	return wantedSF, subnetworkID, includeAllSubnetworks, nil
}

func (d *DNSServer) validateDNSRequest(addr *net.UDPAddr, b []byte) (dnsMsg *dns.Msg, domainName string, atype string, err error) {
	dnsMsg = new(dns.Msg)
	err = dnsMsg.Unpack(b[:])
	if err != nil {
		log.Infof("%s: invalid dns message: %v", addr, err)
		return nil, "", "", err
	}
	if len(dnsMsg.Question) != 1 {
		str := fmt.Sprintf("%s sent more than 1 question: %d", addr, len(dnsMsg.Question))
		log.Infof("%s", str)
		return nil, "", "", errors.Errorf("%s", str)
	}
	domainName = strings.ToLower(dnsMsg.Question[0].Name)
	ff := strings.LastIndex(domainName, d.hostname)
	if ff < 0 {
		str := fmt.Sprintf("invalid name: %s", dnsMsg.Question[0].Name)
		log.Infof("%s", str)
		return nil, "", "", errors.Errorf("%s", str)
	}
	atype, err = translateDNSQuestion(addr, dnsMsg)
	return dnsMsg, domainName, atype, err
}

func translateDNSQuestion(addr *net.UDPAddr, dnsMsg *dns.Msg) (string, error) {
	var atype string
	qtype := dnsMsg.Question[0].Qtype
	switch qtype {
	case dns.TypeA:
		atype = "A"
	case dns.TypeAAAA:
		atype = "AAAA"
	case dns.TypeNS:
		atype = "NS"
	default:
		str := fmt.Sprintf("%s: invalid qtype: %d", addr, dnsMsg.Question[0].Qtype)
		log.Infof("%s", str)
		return "", errors.Errorf("%s", str)
	}
	return atype, nil
}

func (d *DNSServer) buildDNSResponse(addr *net.UDPAddr, authority dns.RR, dnsMsg *dns.Msg,
	wantedSF wire.ServiceFlag, includeAllSubnetworks bool, subnetworkID *subnetworkid.SubnetworkID, atype string) ([]byte, error) {
	respMsg := dnsMsg.Copy()
	respMsg.Authoritative = true
	respMsg.Response = true

	qtype := dnsMsg.Question[0].Qtype
	if qtype != dns.TypeNS {
		respMsg.Ns = append(respMsg.Ns, authority)
		addrs := amgr.GoodAddresses(qtype, wantedSF, includeAllSubnetworks, subnetworkID)
		for _, a := range addrs {
			rr := fmt.Sprintf("%s 30 IN %s %s", dnsMsg.Question[0].Name, atype, a.IP.String())
			newRR, err := dns.NewRR(rr)
			if err != nil {
				log.Infof("%s: NewRR: %v", addr, err)
				return nil, err
			}

			respMsg.Answer = append(respMsg.Answer, newRR)
		}
	} else {
		rr := fmt.Sprintf("%s 86400 IN NS %s", dnsMsg.Question[0].Name, d.nameserver)
		newRR, err := dns.NewRR(rr)
		if err != nil {
			log.Infof("%s: NewRR: %v", addr, err)
			return nil, err
		}

		respMsg.Answer = append(respMsg.Answer, newRR)
	}

	sendBytes, err := respMsg.Pack()
	if err != nil {
		log.Infof("%s: failed to pack response: %v", addr, err)
		return nil, err
	}
	return sendBytes, nil
}

func (d *DNSServer) handleDNSRequest(addr *net.UDPAddr, authority dns.RR, udpListen *net.UDPConn, b []byte) {
	defer wg.Done()

	dnsMsg, domainName, atype, err := d.validateDNSRequest(addr, b)
	if err != nil {
		return
	}

	wantedSF, subnetworkID, includeAllSubnetworks, err := d.extractServicesSubnetworkID(addr, domainName)
	if err != nil {
		return
	}

	log.Infof("%s: query %d for services %v, subnetwork ID %v",
		addr, dnsMsg.Question[0].Qtype, wantedSF, subnetworkID)

	sendBytes, err := d.buildDNSResponse(addr, authority, dnsMsg, wantedSF, includeAllSubnetworks, subnetworkID, atype)
	if err != nil {
		return
	}

	_, err = udpListen.WriteToUDP(sendBytes, addr)
	if err != nil {
		log.Infof("%s: failed to write response: %v", addr, err)
		return
	}
}
