package client

import (
	"context"
	"time"

	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/daemon/server"

	"github.com/pkg/errors"

	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/daemon/pb"
	"google.golang.org/grpc"
)

// Connect connects to the c4exwalletd server, and returns the client instance
func Connect(address string) (pb.C4exwalletdClient, func(), error) {
	// Connection is local, so 1 second timeout is sufficient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(server.MaxDaemonSendMsgSize)))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, errors.New("c4exwallet daemon is not running, start it with `c4exwallet start-daemon`")
		}
		return nil, nil, err
	}

	return pb.NewC4exwalletdClient(conn), func() {
		conn.Close()
	}, nil
}
