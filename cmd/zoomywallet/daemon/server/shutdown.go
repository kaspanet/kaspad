package server

import (
	"context"
<<<<<<< Updated upstream:cmd/kaspawallet/daemon/server/shutdown.go
=======

>>>>>>> Stashed changes:cmd/zoomywallet/daemon/server/shutdown.go
	"github.com/zoomy-network/zoomyd/cmd/zoomywallet/daemon/pb"
)

func (s *server) Shutdown(ctx context.Context, request *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	close(s.shutdown)
	return &pb.ShutdownResponse{}, nil
}
