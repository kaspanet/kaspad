package grpcserver

import (
	"io"

	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
)

type grpcStream interface {
	Send(*protowire.KaspadMessage) error
	Recv() (*protowire.KaspadMessage, error)
}

func (c *gRPCConnection) connectionLoops() error {
	errChan := make(chan error, 1) // buffered channel because one of the loops might try write after disconnect

	spawn(func() { errChan <- c.receiveLoop() })
	spawn(func() { errChan <- c.sendLoop() })

	err := <-errChan

	disconnectErr := c.Disconnect()
	if disconnectErr != nil {
		log.Errorf("Error from disconnect: %s", disconnectErr)
	}
	return err
}

func (c *gRPCConnection) sendLoop() error {
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
		err = func() error {
			c.writeToErrChanDuringDisconnectLock.Lock()
			defer c.writeToErrChanDuringDisconnectLock.Unlock()
			err := c.stream.Send(messageProto)
			if c.IsConnected() {
				c.errChan <- err
				if err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}

	}
	return nil
}

func (c *gRPCConnection) receiveLoop() error {
	for c.IsConnected() {
		protoMessage, err := c.stream.Recv()
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
