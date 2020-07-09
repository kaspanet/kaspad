package grpcserver

import (
	"io"

	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
)

func (c *gRPCConnection) serverConnectionLoop(stream protowire.P2P_MessageStreamServer) error {
	spawn(func() {
		for {
			message, err := stream.Recv()
			if err == io.EOF {
				c.disconnectChan <- struct{}{}
			}
			c.receiveChan <- message
		}
	})

	for {
		select {
		case <-c.disconnectChan:
			c.errChan <- nil
			return nil

		case message := <-c.sendChan:
			err := stream.SendMsg(message)
			c.errChan <- err
			if err == io.EOF {
				return nil
			}
		}
	}
}

func (c *gRPCConnection) clientConnectionLoop(stream protowire.P2P_MessageStreamClient) error {
	spawn(func() {
		for {
			message, err := stream.Recv()
			if err == io.EOF {
				c.disconnectChan <- struct{}{}
			}
			c.receiveChan <- message
		}
	})

	for {
		select {
		case <-c.disconnectChan:
			err := stream.CloseSend()
			if err != nil {
				c.errChan <- err
			}
			err = c.clientConn.Close()
			if err != nil {
				c.errChan <- err
			}
			return nil

		case message := <-c.sendChan:
			err := stream.SendMsg(message)
			c.errChan <- err
			if err == io.EOF {
				return nil
			}
		}
	}
}
