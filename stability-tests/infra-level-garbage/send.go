package main

import (
	"encoding/hex"
	"net"
	"time"

	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/pkg/errors"
)

func sendMessages(address string, messagesChan <-chan []byte) error {
	connection, err := dialToNode(address)
	if err != nil {
		return err
	}
	for message := range messagesChan {
		messageHex := make([]byte, hex.EncodedLen(len(message)))
		hex.Encode(messageHex, message)
		log.Infof("Sending message %s", messageHex)

		err := sendMessage(connection, message)
		if err != nil {
			// if failed once, we might have been disconnected because of a previous message,
			// so re-connect and retry before reporting error
			connection, err = dialToNode(address)
			if err != nil {
				return err
			}
			err = sendMessage(connection, message)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func sendMessage(connection net.Conn, message []byte) error {
	err := connection.SetDeadline(time.Now().Add(common.DefaultTimeout))
	if err != nil {
		return errors.Wrap(err, "Error setting connection deadline")
	}

	_, err = connection.Write(message)
	return err
}

func dialToNode(address string) (net.Conn, error) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return nil, errors.Wrap(err, "Error connecting to node")
	}
	return connection, nil
}
