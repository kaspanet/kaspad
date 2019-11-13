package mqtt

import (
	"errors"
	"github.com/daglabs/btcd/apiserver/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

func GetClient() (mqtt.Client, error) {
	if client == nil {
		return nil, errors.New("MQTT is not connected")
	}
	return client, nil
}

func Connect(cfg *config.Config) error {
	if cfg.MQTTBrokerAddress == "" {
		// MQTT not defined -- nothing to do
		return nil
	}

	options := mqtt.NewClientOptions()
	options.AddBroker(cfg.MQTTBrokerAddress)
	options.SetUsername(cfg.MQTTUser)
	options.SetPassword(cfg.MQTTPassword)
	options.SetCleanSession(true)
	options.SetAutoReconnect(true)

	newClient := mqtt.NewClient(options)
	if token := newClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	client = newClient

	return nil
}

func Close() {
	if client == nil {
		return
	}
	client.Disconnect(250)
	client = nil
}
