package grpcserver

import (
	"io"
	"time"

	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
)

type grpcStream interface {
	Send(*protowire.KaspadMessage) error
	Recv() (*protowire.KaspadMessage, error)
}

func (c *gRPCConnection) connectionLoop(stream grpcStream) error {
	errChan := make(chan error)

	spawn(func() {
		for c.IsConnected() {
			message, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				c.errChan <- err
				return
			}

			c.receiveChan <- message
		}
	})

	spawn(func() {
		for c.IsConnected() {
			select {
			case message := <-c.sendChan:
				err := stream.Send(message)
				if err != nil {
					if err == io.EOF {
						err = nil
					}
					c.errChan <- err
				}
			case <-time.Tick(1 * time.Second):
			}
		}
		errChan <- nil
	})

	err := c.Disconnect()
	if err != nil {
		log.Errorf("Error from disconnect: %s", err)
	}
	return <-errChan
}

func (c *gRPCConnection) serverConnectionLoop(stream protowire.P2P_MessageStreamServer) error {
	return c.connectionLoop(stream)
}

func (c *gRPCConnection) clientConnectionLoop(stream protowire.P2P_MessageStreamClient) error {
	return c.connectionLoop(stream)
}
