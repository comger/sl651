package notifier

import (
	"context"
	"log"
)

type AlertType string

const (
	AlertTypeOnline  AlertType = "online"
	AlertTypeOffline AlertType = "offline"
	AlertTypeError   AlertType = "error"
)

type Alert struct {
	Type    AlertType
	Device  string
	Message string
}

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) SendAlert(ctx context.Context, alert *Alert) error {
	log.Printf("[ALERT] [%s] Device: %s, Message: %s", alert.Type, alert.Device, alert.Message)
	return nil
}
