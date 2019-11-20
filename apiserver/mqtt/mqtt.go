package mqtt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/daglabs/btcd/apiserver/apimodels"
	"github.com/daglabs/btcd/apiserver/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// client is an instance of the MQTT client, in case we have an active connection
var client mqtt.Client

// GetClient returns an instance of the MQTT client, in case we have an active connection
func GetClient() (mqtt.Client, error) {
	if client == nil {
		return nil, errors.New("MQTT is not connected")
	}
	return client, nil
}

// IsConnected returns true is MQTT client was connected, false otherwise.
func IsConnected() bool {
	return client != nil
}

// PublishTransactionForAddress publishes a transaction message for the topic to transactions/address.
func PublishTransactionNotification(transaction *apimodels.TransactionResponse, address string) error {
	payload, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	token := client.Publish(transactionsTopic(address), 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

func transactionsTopic(address string) string {
	return fmt.Sprintf("transactions/%s", address)
}

// Connect initiates a connection to the MQTT server, if defined
func Connect() error {
	cfg := config.ActiveConfig()
	if cfg.MQTTBrokerAddress == "" {
		// MQTT broker not defined -- nothing to do
		return nil
	}

	options := mqtt.NewClientOptions()
	options.AddBroker(cfg.MQTTBrokerAddress)
	options.SetUsername(cfg.MQTTUser)
	options.SetPassword(cfg.MQTTPassword)
	options.SetAutoReconnect(true)

	newClient := mqtt.NewClient(options)
	if token := newClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	client = newClient

	return nil
}

// Close closes the connection to the MQTT server, if previously connected
func Close() {
	if client == nil {
		return
	}
	client.Disconnect(250)
	client = nil
}
