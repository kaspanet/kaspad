package server

import (
	"context"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) Send(_ context.Context, request *pb.SendRequest) (*pb.SendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
