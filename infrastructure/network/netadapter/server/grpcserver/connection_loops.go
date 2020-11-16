package grpcserver

import (
	"io"
	"time"

	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/infrastructure/logger"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
)

func (c *gRPCConnection) connectionLoops() error {
	errChan := make(chan error, 1) // buffered channel because one of the loops might try write after disconnect

	spawn("gRPCConnection.receiveLoop", func() { errChan <- c.receiveLoop() })
	spawn("gRPCConnection.sendLoop", func() { errChan <- c.sendLoop() })

	err := <-errChan

	c.Disconnect()

	return err
}

func (c *gRPCConnection) sendLoop() error {
	outgoingRoute := c.router.OutgoingRoute()
	for c.IsConnected() {
		message, err := outgoingRoute.Dequeue()
		if err != nil {
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				return nil
			}
			return err
		}

		log.Debugf("outgoing '%s' message to %s", message.Command(), c)
		log.Tracef("outgoing '%s' message to %s: %s", message.Command(), c, logger.NewLogClosure(func() string {
			return spew.Sdump(message)
		}))

		messageProto, err := protowire.FromAppMessage(message)
		if err != nil {
			return err
		}

		err = c.send(messageProto)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *gRPCConnection) receiveLoop() error {
	messageNumber := uint64(0)
	for c.IsConnected() {
		protoMessage, err := c.receive()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		message, err := protoMessage.ToAppMessage()
		if err != nil {
			if c.onInvalidMessageHandler != nil {
				c.onInvalidMessageHandler(err)
			}
			return err
		}

		messageNumber++
		message.SetMessageNumber(messageNumber)
		message.SetReceivedAt(time.Now())

		log.Debugf("incoming '%s' message from %s (message number %d)", message.Command(), c,
			message.MessageNumber())

		log.Tracef("incoming '%s' message from %s  (message number %d): %s", message.Command(),
			c, message.MessageNumber(), logger.NewLogClosure(func() string {
				return spew.Sdump(message)
			}))

		err = c.router.EnqueueIncomingMessage(message)
		if err != nil {
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				return nil
			}
			if c.onInvalidMessageHandler != nil {
				c.onInvalidMessageHandler(err)
			}
			return err
		}
	}
	return nil
}
