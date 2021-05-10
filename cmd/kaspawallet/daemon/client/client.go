package client

import (
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"google.golang.org/grpc"
)

func Connect(address string) (pb.KaspawalletdClient, func(), error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, err
	}

	return pb.NewKaspawalletdClient(conn), func() {
		conn.Close()
	}, nil
}
