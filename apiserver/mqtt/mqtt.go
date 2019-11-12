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
	if cfg.MQTTAddress == "" {
		// MQTT not defined -- nothing to do
		return nil
	}

	return nil
}

func Close() {
	if client == nil {
		return
	}
	client.Disconnect(250)
	client = nil
}
