package grpcserver

import (
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/pkg/errors"
	"io"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/logger"

	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
)

type grpcStream interface {
	Send(*protowire.KaspadMessage) error
	Recv() (*protowire.KaspadMessage, error)
}

func (c *gRPCConnection) connectionLoops() error {
	errChan := make(chan error, 1) // buffered channel because one of the loops might try write after disconnect

	spawn("gRPCConnection.receiveLoop", func() { errChan <- c.receiveLoop() })
	spawn("gRPCConnection.sendLoop", func() { errChan <- c.sendLoop() })

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
		message, err := outgoingRoute.Dequeue()
		if err != nil {
			return err
		}

		log.Tracef("outgoing '%s' message to %s: %s", message.Command(), c, logger.NewLogClosure(func() string {
			return spew.Sdump(message)
		}))

		messageProto, err := protowire.FromWireMessage(message)
		if err != nil {
			return err
		}

		err = c.stream.Send(messageProto)
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

		log.Tracef("incoming '%s' message from %s: %s", message.Command(), c, logger.NewLogClosure(func() string {
			return spew.Sdump(message)
		}))

		err = c.router.EnqueueIncomingMessage(message)
		if err != nil {
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				log.Debugf("Router for %s is closed. Exiting the receive loop", c)
				return nil
			}
			return err
		}
	}
	return nil
}
