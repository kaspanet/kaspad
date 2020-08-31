package grpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
	"time"
)

type clientRouter struct {
	rpcClient     *RPCClient
	router        *routerpkg.Router
	incomingRoute *routerpkg.Route
}

func newClientRouter(rpcClient *RPCClient) (*clientRouter, error) {
	router := routerpkg.NewRouter()
	clientRouter := &clientRouter{
		rpcClient: rpcClient,
		router:    router,
	}

	incomingRoute, err := clientRouter.registerIncomingRoute()
	if err != nil {
		return nil, err
	}
	clientRouter.incomingRoute = incomingRoute

	return clientRouter, nil
}

func (cr *clientRouter) registerIncomingRoute() (*routerpkg.Route, error) {
	messageTypes := make([]appmessage.MessageCommand, 0, len(appmessage.RPCMessageCommandToString))
	for messageType := range appmessage.RPCMessageCommandToString {
		messageTypes = append(messageTypes, messageType)
	}
	return cr.router.AddIncomingRoute(messageTypes)
}

func (cr *clientRouter) start() {
	spawn("clientRouter.start-sendLoop", func() {
		for {
			message, err := cr.router.OutgoingRoute().Dequeue()
			if err != nil {
				cr.handleError(err)
				return
			}
			err = cr.send(message)
			if err != nil {
				cr.handleError(err)
				return
			}
		}
	})
	spawn("clientRouter.start-receiveLoop", func() {
		for {
			message, err := cr.receive()
			if err != nil {
				cr.handleError(err)
				return
			}
			err = cr.router.EnqueueIncomingMessage(message)
			if err != nil {
				cr.handleError(err)
				return
			}
		}
	})
}

func (cr *clientRouter) send(requestAppMessage appmessage.Message) error {
	request, err := protowire.FromAppMessage(requestAppMessage)
	if err != nil {
		return errors.Wrapf(err, "error converting the request")
	}
	return cr.rpcClient.stream.Send(request)
}

func (cr *clientRouter) receive() (appmessage.Message, error) {
	response, err := cr.rpcClient.stream.Recv()
	if err != nil {
		return nil, err
	}
	return response.ToAppMessage()
}

func (cr *clientRouter) handleError(err error) {
	panic(err)
}

func (cr *clientRouter) close() {
	cr.router.Close()
}

func (cr *clientRouter) enqueue(message appmessage.Message) error {
	return cr.router.OutgoingRoute().Enqueue(message)
}

func (cr *clientRouter) dequeue() (appmessage.Message, error) {
	return cr.incomingRoute.Dequeue()
}

func (cr *clientRouter) dequeueWithTimeout(timeout time.Duration) (appmessage.Message, error) {
	return cr.incomingRoute.DequeueWithTimeout(timeout)
}
