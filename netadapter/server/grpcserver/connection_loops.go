package grpcserver

import (
	"io"

	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
)

type grpcStream interface {
	Send(*protowire.KaspadMessage) error
	Recv() (*protowire.KaspadMessage, error)
}

func (c *gRPCConnection) connectionLoops(stream grpcStream) error {
	errChan := make(chan error, 1) // buffered channel because one of the loops might try write after disconnect

	spawn(func() { c.receiveLoop(stream, errChan) })

	spawn(func() { c.sendLoop(stream, errChan) })

	err := <-errChan

	disconnectErr := c.Disconnect()
	if disconnectErr != nil {
		log.Errorf("Error from disconnect: %s", disconnectErr)
	}
	return err
}

func (c *gRPCConnection) sendLoop(stream grpcStream, errChan chan error) {
	for c.IsConnected() {
		message, ok := <-c.sendChan
		if !ok {
			errChan <- nil
			return
		}
		err := stream.Send(message)
		c.errChan <- err
		if err != nil {
			errChan <- err
			return
		}
	}
}

func (c *gRPCConnection) receiveLoop(stream grpcStream, errChan chan error) {
	for c.IsConnected() {
		message, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			errChan <- err
			return
		}

		c.receiveChan <- message
	}
}

func (c *gRPCConnection) serverConnectionLoop(stream protowire.P2P_MessageStreamServer) error {
	return c.connectionLoops(stream)
}

func (c *gRPCConnection) clientConnectionLoop(stream protowire.P2P_MessageStreamClient) error {
	err := c.connectionLoops(stream)

	_ = stream.CloseSend() // ignore error because we don't really know what's the status of the connection

	return err
}
