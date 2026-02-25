package adapters

import (
	"context"
	"sl651-platform/internal/model"
)

// DestinationAdapter is the interface for all forwarding destinations
type DestinationAdapter interface {
	// Send forwards the device data to the destination
	Send(ctx context.Context, data *model.DeviceData, dest model.Destination) error
	// Close closes the adapter
	Close() error
}
