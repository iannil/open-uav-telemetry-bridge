package core

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/open-uav/telemetry-bridge/internal/core/coordinator"
	"github.com/open-uav/telemetry-bridge/internal/core/statestore"
	"github.com/open-uav/telemetry-bridge/internal/core/throttler"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// Engine is the core message routing engine
type Engine struct {
	adapters    []Adapter
	publishers  []Publisher
	stateStore  *statestore.StateStore
	throttler   *throttler.Throttler
	coordinator *coordinator.Converter
	events      chan *models.DroneState
	wg          sync.WaitGroup
}

// EngineConfig holds configuration for the engine
type EngineConfig struct {
	RateHz       float64
	ConvertGCJ02 bool
	ConvertBD09  bool
}

// NewEngine creates a new core engine
func NewEngine(cfg EngineConfig) *Engine {
	return &Engine{
		adapters:    make([]Adapter, 0),
		publishers:  make([]Publisher, 0),
		stateStore:  statestore.New(),
		throttler:   throttler.New(cfg.RateHz),
		coordinator: coordinator.New(cfg.ConvertGCJ02, cfg.ConvertBD09),
		events:      make(chan *models.DroneState, 100),
	}
}

// RegisterAdapter adds an adapter to the engine
func (e *Engine) RegisterAdapter(adapter Adapter) {
	e.adapters = append(e.adapters, adapter)
}

// RegisterPublisher adds a publisher to the engine
func (e *Engine) RegisterPublisher(publisher Publisher) {
	e.publishers = append(e.publishers, publisher)
}

// Start begins the engine processing
func (e *Engine) Start(ctx context.Context) error {
	// Start all publishers first
	for _, pub := range e.publishers {
		if err := pub.Start(ctx); err != nil {
			return fmt.Errorf("starting publisher %s: %w", pub.Name(), err)
		}
		log.Printf("[Engine] Publisher started: %s", pub.Name())
	}

	// Start all adapters
	for _, adapter := range e.adapters {
		if err := adapter.Start(ctx, e.events); err != nil {
			return fmt.Errorf("starting adapter %s: %w", adapter.Name(), err)
		}
		log.Printf("[Engine] Adapter started: %s", adapter.Name())
	}

	// Start the routing goroutine
	e.wg.Add(1)
	go e.routeMessages(ctx)

	log.Printf("[Engine] Started with %d adapters and %d publishers",
		len(e.adapters), len(e.publishers))

	return nil
}

// routeMessages processes incoming events and routes them to publishers
func (e *Engine) routeMessages(ctx context.Context) {
	defer e.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case state := <-e.events:
			e.processState(state)
		}
	}
}

// processState handles a single state update
func (e *Engine) processState(state *models.DroneState) {
	// Apply coordinate conversion
	e.applyCoordinateConversion(state)

	// Update state store
	e.stateStore.Update(state)

	// Check throttle
	if !e.throttler.ShouldPublish(state) {
		return
	}

	// Publish to all publishers
	for _, pub := range e.publishers {
		if err := pub.Publish(state); err != nil {
			log.Printf("[Engine] Publish error (%s): %v", pub.Name(), err)
		}
	}
}

// applyCoordinateConversion converts WGS84 coordinates to GCJ02/BD09 if configured
func (e *Engine) applyCoordinateConversion(state *models.DroneState) {
	if e.coordinator == nil {
		return
	}

	result := e.coordinator.Convert(state.Location.Lat, state.Location.Lon)

	state.Location.LatGCJ02 = result.LatGCJ02
	state.Location.LonGCJ02 = result.LonGCJ02
	state.Location.LatBD09 = result.LatBD09
	state.Location.LonBD09 = result.LonBD09
}

// Stop gracefully stops the engine
func (e *Engine) Stop() error {
	// Stop adapters first
	for _, adapter := range e.adapters {
		if err := adapter.Stop(); err != nil {
			log.Printf("[Engine] Error stopping adapter %s: %v", adapter.Name(), err)
		}
	}

	// Wait for routing to complete
	e.wg.Wait()

	// Stop publishers
	for _, pub := range e.publishers {
		if err := pub.Stop(); err != nil {
			log.Printf("[Engine] Error stopping publisher %s: %v", pub.Name(), err)
		}
	}

	log.Printf("[Engine] Stopped")
	return nil
}

// GetState returns the current state for a device
func (e *Engine) GetState(deviceID string) *models.DroneState {
	return e.stateStore.Get(deviceID)
}

// GetAllStates returns all current device states
func (e *Engine) GetAllStates() []*models.DroneState {
	return e.stateStore.GetAll()
}

// GetDeviceCount returns the number of tracked devices
func (e *Engine) GetDeviceCount() int {
	return e.stateStore.Count()
}

// SetThrottleRate updates the publish rate
func (e *Engine) SetThrottleRate(rateHz float64) {
	e.throttler.SetRate(rateHz)
}
