package grpcserver

import (
	"context"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/peer"
	"net"
	"time"
)

type p2pServer struct {
	protowire.UnimplementedP2PServer
	gRPCServer
}

// NewP2PServer creates a new P2PServer
func NewP2PServer(listeningAddresses []string) (server.P2PServer, error) {
	gRPCServer := newGRPCServer(listeningAddresses)
	p2pServer := &p2pServer{gRPCServer: *gRPCServer}
	protowire.RegisterP2PServer(gRPCServer.server, p2pServer)
	return p2pServer, nil
}

func (p *p2pServer) MessageStream(stream protowire.P2P_MessageStreamServer) error {
	defer panics.HandlePanic(log, "p2pServer.MessageStream", nil)

	return p.handleInboundConnection(stream.Context(), stream)
}

// Connect connects to the given address
// This is part of the P2PServer interface
func (p *p2pServer) Connect(address string) (server.Connection, error) {
	log.Infof("Dialing to %s", address)

	const dialTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	gRPCClientConnection, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", address)
	}

	client := protowire.NewP2PClient(gRPCClientConnection)
	stream, err := client.MessageStream(context.Background(), grpc.UseCompressor(gzip.Name),
		grpc.MaxCallRecvMsgSize(MaxMessageSize), grpc.MaxCallSendMsgSize(MaxMessageSize))
	if err != nil {
		return nil, errors.Wrapf(err, "error getting client stream for %s", address)
	}

	peerInfo, ok := peer.FromContext(stream.Context())
	if !ok {
		return nil, errors.Errorf("error getting stream peer info from context for %s", address)
	}
	tcpAddress, ok := peerInfo.Addr.(*net.TCPAddr)
	if !ok {
		return nil, errors.Errorf("non-tcp addresses are not supported")
	}

	connection := newConnection(&p.gRPCServer, tcpAddress, stream, gRPCClientConnection)

	err = p.onConnectedHandler(connection)
	if err != nil {
		return nil, err
	}

	log.Infof("Connected to %s", address)

	return connection, nil
}
