package core

import (
	"context"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

// Adapter is the interface that all southbound protocol adapters must implement
type Adapter interface {
	// Name returns the adapter name (e.g., "mavlink", "dji", "gb28181")
	Name() string

	// Start begins receiving data and sends DroneState events to the channel
	// The adapter should respect context cancellation for graceful shutdown
	Start(ctx context.Context, events chan<- *models.DroneState) error

	// Stop gracefully stops the adapter
	Stop() error
}

// Publisher is the interface that all northbound publishers must implement
type Publisher interface {
	// Name returns the publisher name (e.g., "mqtt", "websocket", "http")
	Name() string

	// Start initializes the publisher and prepares for publishing
	// The publisher should respect context cancellation for graceful shutdown
	Start(ctx context.Context) error

	// Publish sends a DroneState to the destination
	Publish(state *models.DroneState) error

	// Stop gracefully stops the publisher
	Stop() error
}
