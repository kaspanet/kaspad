package server

import (
	"context"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) Sign(_ context.Context, request *pb.SignRequest) (*pb.SignResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sign not implemented")
}
