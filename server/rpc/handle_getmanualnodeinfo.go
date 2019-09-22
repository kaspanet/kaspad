package rpc

import (
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/server/serverutils"
	"net"
	"strings"
)

// handleGetManualNodeInfo handles getManualNodeInfo commands.
func handleGetManualNodeInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetManualNodeInfoCmd)
	results, err := getManualNodesInfo(s, c.Details, c.Node)
	if err != nil {
		return nil, err
	}
	if resultsNonDetailed, ok := results.([]string); ok {
		return resultsNonDetailed[0], nil
	}
	resultsDetailed := results.([]*btcjson.GetManualNodeInfoResult)
	return resultsDetailed[0], nil
}

// handleGetAllManualNodesInfo handles getAllManualNodesInfo commands.
func handleGetAllManualNodesInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetAllManualNodesInfoCmd)
	return getManualNodesInfo(s, c.Details, "")
}

// getManualNodesInfo handles getManualNodeInfo and getAllManualNodesInfo commands.
func getManualNodesInfo(s *Server, detailsArg *bool, node string) (interface{}, error) {

	details := detailsArg == nil || *detailsArg

	// Retrieve a list of persistent (manual) peers from the server and
	// filter the list of peers per the specified address (if any).
	peers := s.cfg.ConnMgr.PersistentPeers()
	if node != "" {
		found := false
		for i, peer := range peers {
			if peer.ToPeer().Addr() == node {
				peers = peers[i : i+1]
				found = true
			}
		}
		if !found {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCClientNodeNotAdded,
				Message: "Node has not been added",
			}
		}
	}

	// Without the details flag, the result is just a slice of the addresses as
	// strings.
	if !details {
		results := make([]string, 0, len(peers))
		for _, peer := range peers {
			results = append(results, peer.ToPeer().Addr())
		}
		return results, nil
	}

	// With the details flag, the result is an array of JSON objects which
	// include the result of DNS lookups for each peer.
	results := make([]*btcjson.GetManualNodeInfoResult, 0, len(peers))
	for _, rpcPeer := range peers {
		// Set the "address" of the peer which could be an ip address
		// or a domain name.
		peer := rpcPeer.ToPeer()
		var result btcjson.GetManualNodeInfoResult
		result.ManualNode = peer.Addr()
		result.Connected = btcjson.Bool(peer.Connected())

		// Split the address into host and port portions so we can do
		// a DNS lookup against the host.  When no port is specified in
		// the address, just use the address as the host.
		host, _, err := net.SplitHostPort(peer.Addr())
		if err != nil {
			host = peer.Addr()
		}

		var ipList []string
		switch {
		case net.ParseIP(host) != nil, strings.HasSuffix(host, ".onion"):
			ipList = make([]string, 1)
			ipList[0] = host
		default:
			// Do a DNS lookup for the address.  If the lookup fails, just
			// use the host.
			ips, err := serverutils.BTCDLookup(host)
			if err != nil {
				ipList = make([]string, 1)
				ipList[0] = host
				break
			}
			ipList = make([]string, 0, len(ips))
			for _, ip := range ips {
				ipList = append(ipList, ip.String())
			}
		}

		// Add the addresses and connection info to the result.
		addrs := make([]btcjson.GetManualNodeInfoResultAddr, 0, len(ipList))
		for _, ip := range ipList {
			var addr btcjson.GetManualNodeInfoResultAddr
			addr.Address = ip
			addr.Connected = "false"
			if ip == host && peer.Connected() {
				addr.Connected = logger.DirectionString(peer.Inbound())
			}
			addrs = append(addrs, addr)
		}
		result.Addresses = &addrs
		results = append(results, &result)
	}
	return results, nil
}
