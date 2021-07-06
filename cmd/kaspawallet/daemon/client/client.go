package client

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"google.golang.org/grpc"
)

// Connect connects to the kaspawalletd server, and returns the client instance
func Connect(address string) (pb.KaspawalletdClient, func(), error) {
	// Connection is local, so 1 second timeout is sufficient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, errors.New("kaspactl daemon is not running, start it with `kaspactl start-daemon`")
		}
		return nil, nil, err
	}

	return pb.NewKaspawalletdClient(conn), func() {
		conn.Close()
	}, nil
}
