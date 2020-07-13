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

	spawn(func() { errChan <- c.receiveLoop(stream) })
	spawn(func() { errChan <- c.sendLoop(stream) })

	err := <-errChan

	disconnectErr := c.Disconnect()
	if disconnectErr != nil {
		log.Errorf("Error from disconnect: %s", disconnectErr)
	}
	return err
}

func (c *gRPCConnection) sendLoop(stream grpcStream) error {
	outgoingRoute := c.router.OutgoingRoute()
	for c.IsConnected() {
		message, isOpen := outgoingRoute.Dequeue()
		if !isOpen {
			return nil
		}
		messageProto, err := protowire.FromWireMessage(message)
		if err != nil {
			return err
		}
		err = stream.Send(messageProto)
		c.errChan <- err
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *gRPCConnection) receiveLoop(stream grpcStream) error {
	for c.IsConnected() {
		protoMessage, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		message, err := protoMessage.ToWireMessage()
		if err != nil {
			return err
		}
		isOpen, err := c.router.EnqueueIncomingMessage(message)
		if err != nil {
			return err
		}
		if !isOpen {
			return nil
		}
	}
	return nil
}

func (c *gRPCConnection) serverConnectionLoop(stream protowire.P2P_MessageStreamServer) error {
	return c.connectionLoops(stream)
}

func (c *gRPCConnection) clientConnectionLoop(stream protowire.P2P_MessageStreamClient) error {
	err := c.connectionLoops(stream)

	_ = stream.CloseSend() // ignore error because we don't really know what's the status of the connection

	return err
}
