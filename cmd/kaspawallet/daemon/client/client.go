package client

import (
	"context"
	"github.com/fabbez/topiad/cmd/topiawallet/daemon/server"
	"time"

	"github.com/pkg/errors"

	"github.com/fabbez/topiad/cmd/topiawallet/daemon/pb"
	"google.golang.org/grpc"
)

// Connect connects to the topiawalletd server, and returns the client instance
func Connect(address string) (pb.topiawalletdClient, func(), error) {
	// Connection is local, so 1 second timeout is sufficient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(server.MaxDaemonSendMsgSize)))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, errors.New("topiawallet daemon is not running, start it with `topiawallet start-daemon`")
		}
		return nil, nil, err
	}

	return pb.NewtopiawalletdClient(conn), func() {
		conn.Close()
	}, nil
}
