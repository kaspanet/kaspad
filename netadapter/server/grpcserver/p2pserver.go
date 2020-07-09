package grpcserver

import (
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
)

type p2pServer struct {
	protowire.UnimplementedP2PServer
}

func (*p2pServer) MessageStream(stream protowire.P2P_MessageStreamServer) error {
	defer panics.HandlePanic(log, nil)
	return serverConnectionLoop(stream)
}
