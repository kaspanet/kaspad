package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/network"
	"net"
	"strconv"
)

// handleNode handles node commands.
func handleNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.NodeCmd)

	var addr string
	var nodeID uint64
	var errN, err error
	params := s.cfg.DAGParams
	switch c.SubCmd {
	case "disconnect":
		// If we have a valid uint disconnect by node id. Otherwise,
		// attempt to disconnect by address, returning an error if a
		// valid IP address is not supplied.
		if nodeID, errN = strconv.ParseUint(c.Target, 10, 32); errN == nil {
			err = s.cfg.ConnMgr.DisconnectByID(int32(nodeID))
		} else {
			if _, _, errP := net.SplitHostPort(c.Target); errP == nil || net.ParseIP(c.Target) != nil {
				addr = network.NormalizeAddress(c.Target, params.DefaultPort)
				err = s.cfg.ConnMgr.DisconnectByAddr(addr)
			} else {
				return nil, &rpcmodel.RPCError{
					Code:    rpcmodel.ErrRPCInvalidParameter,
					Message: "invalid address or node ID",
				}
			}
		}
		if err != nil && peerExists(s.cfg.ConnMgr, addr, int32(nodeID)) {

			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCMisc,
				Message: "can't disconnect a permanent peer, use remove",
			}
		}

	case "remove":
		// If we have a valid uint disconnect by node id. Otherwise,
		// attempt to disconnect by address, returning an error if a
		// valid IP address is not supplied.
		if nodeID, errN = strconv.ParseUint(c.Target, 10, 32); errN == nil {
			err = s.cfg.ConnMgr.RemoveByID(int32(nodeID))
		} else {
			if _, _, errP := net.SplitHostPort(c.Target); errP == nil || net.ParseIP(c.Target) != nil {
				addr = network.NormalizeAddress(c.Target, params.DefaultPort)
				err = s.cfg.ConnMgr.RemoveByAddr(addr)
			} else {
				return nil, &rpcmodel.RPCError{
					Code:    rpcmodel.ErrRPCInvalidParameter,
					Message: "invalid address or node ID",
				}
			}
		}
		if err != nil && peerExists(s.cfg.ConnMgr, addr, int32(nodeID)) {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCMisc,
				Message: "can't remove a temporary peer, use disconnect",
			}
		}

	case "connect":
		addr = network.NormalizeAddress(c.Target, params.DefaultPort)

		// Default to temporary connections.
		subCmd := "temp"
		if c.ConnectSubCmd != nil {
			subCmd = *c.ConnectSubCmd
		}

		switch subCmd {
		case "perm", "temp":
			err = s.cfg.ConnMgr.Connect(addr, subCmd == "perm")
		default:
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInvalidParameter,
				Message: "invalid subcommand for node connect",
			}
		}
	default:
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidParameter,
			Message: "invalid subcommand for node",
		}
	}

	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}

// peerExists determines if a certain peer is currently connected given
// information about all currently connected peers. Peer existence is
// determined using either a target address or node id.
func peerExists(connMgr rpcserverConnManager, addr string, nodeID int32) bool {
	for _, p := range connMgr.ConnectedPeers() {
		if p.ToPeer().ID() == nodeID || p.ToPeer().Addr() == addr {
			return true
		}
	}
	return false
}
