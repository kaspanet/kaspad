package mqtt

import (
	"encoding/json"
	"github.com/daglabs/btcd/apiserver/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
)

// client is an instance of the MQTT client, in case we have an active connection
var client mqtt.Client

const (
	qualityOfService    = 2
	quiesceMilliseconds = 250
)

// GetClient returns an instance of the MQTT client, in case we have an active connection
func GetClient() (mqtt.Client, error) {
	if client == nil {
		return nil, errors.New("MQTT is not connected")
	}
	return client, nil
}

func isConnected() bool {
	return client != nil
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
	client.Disconnect(quiesceMilliseconds)
	client = nil
}

func publish(topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	token := client.Publish(topic, qualityOfService, false, payload)
	token.Wait()
	if token.Error() != nil {
		return errors.WithStack(token.Error())
	}
	return nil
}
