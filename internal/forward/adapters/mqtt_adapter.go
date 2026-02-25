package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"sl651-platform/internal/model"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTAdapter struct {
	clients map[string]mqtt.Client
}

func NewMQTTAdapter() *MQTTAdapter {
	return &MQTTAdapter{
		clients: make(map[string]mqtt.Client),
	}
}

func (a *MQTTAdapter) Send(ctx context.Context, data *model.DeviceData, dest model.Destination) error {
	client, err := a.getClient(dest)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Default topic if not specified in headers
	topic := "sl651/data"
	for _, h := range dest.Headers {
		if t, ok := h["topic"]; ok {
			topic = t
			break
		}
	}

	token := client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to MQTT: %w", token.Error())
	}

	return nil
}

func (a *MQTTAdapter) getClient(dest model.Destination) (mqtt.Client, error) {
	broker := dest.URL
	if client, ok := a.clients[broker]; ok {
		if client.IsConnected() {
			return client, nil
		}
	}

	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetClientID(fmt.Sprintf("sl651-forwarder-%d", time.Now().UnixNano()))

	if dest.Auth != nil {
		if dest.Auth.Username != nil {
			opts.SetUsername(*dest.Auth.Username)
		}
		if dest.Auth.Password != nil {
			opts.SetPassword(*dest.Auth.Password)
		}
	}

	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(10 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker %s: %w", broker, token.Error())
	}

	a.clients[broker] = client
	return client, nil
}

func (a *MQTTAdapter) Close() error {
	for _, client := range a.clients {
		client.Disconnect(250)
	}
	return nil
}
