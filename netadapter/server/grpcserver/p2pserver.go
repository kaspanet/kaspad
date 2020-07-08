package grpcserver

import (
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
)

type p2pServer struct {
	protowire.UnimplementedP2PServer
}

func (*p2pServer) MessageStream(stream protowire.P2P_MessageStreamServer) error {
	// TODO(libp2p)
	panic("unimplemented!")
}
