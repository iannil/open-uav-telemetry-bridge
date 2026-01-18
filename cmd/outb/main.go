package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/open-uav/telemetry-bridge/internal/adapters/dji"
	"github.com/open-uav/telemetry-bridge/internal/adapters/mavlink"
	"github.com/open-uav/telemetry-bridge/internal/api"
	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/core"
	"github.com/open-uav/telemetry-bridge/internal/publishers/mqtt"
)

const version = "0.3.1-dev"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	fmt.Printf("Open-UAV-Telemetry-Bridge v%s\n", version)
	fmt.Println("Protocol-agnostic UAV telemetry gateway")
	fmt.Println()

	// Determine config path
	configPath := "configs/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}
	log.Printf("Configuration loaded from %s", configPath)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create core engine with coordinate conversion and track storage
	engineCfg := core.EngineConfig{
		RateHz:                cfg.Throttle.DefaultRateHz,
		ConvertGCJ02:          cfg.Coordinate.ConvertGCJ02,
		ConvertBD09:           cfg.Coordinate.ConvertBD09,
		TrackEnabled:          cfg.Track.Enabled,
		TrackMaxPoints:        cfg.Track.MaxPointsPerDrone,
		TrackSampleIntervalMs: cfg.Track.SampleIntervalMs,
	}
	engine := core.NewEngine(engineCfg)
	log.Printf("Core engine created (throttle: %.1f Hz, GCJ02: %v, BD09: %v, track: %v)",
		cfg.Throttle.DefaultRateHz, cfg.Coordinate.ConvertGCJ02, cfg.Coordinate.ConvertBD09, cfg.Track.Enabled)

	// Register adapters
	if cfg.MAVLink.Enabled {
		mavlinkAdapter := mavlink.New(cfg.MAVLink)
		engine.RegisterAdapter(mavlinkAdapter)
		log.Printf("MAVLink adapter registered (%s: %s)",
			cfg.MAVLink.ConnectionType, cfg.MAVLink.Address)
	}

	if cfg.DJI.Enabled {
		djiAdapter := dji.New(cfg.DJI)
		engine.RegisterAdapter(djiAdapter)
		log.Printf("DJI adapter registered (listen: %s, max clients: %d)",
			cfg.DJI.ListenAddress, cfg.DJI.MaxClients)
	}

	// Register publishers
	if cfg.MQTT.Enabled {
		mqttPublisher := mqtt.New(cfg.MQTT)
		engine.RegisterPublisher(mqttPublisher)
		log.Printf("MQTT publisher registered (broker: %s)", cfg.MQTT.Broker)
	}

	// Start engine
	if err := engine.Start(ctx); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}

	// Start HTTP API server
	var httpServer *api.Server
	if cfg.HTTP.Enabled {
		httpServer = api.New(cfg.HTTP, engine, version)
		if err := httpServer.Start(ctx); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
		// Connect WebSocket broadcast to engine state updates
		engine.SetStateCallback(httpServer.BroadcastState)
		log.Printf("HTTP API server started (address: %s, WebSocket: /api/v1/ws)", cfg.HTTP.Address)
	}

	log.Println("Gateway is running. Press Ctrl+C to stop.")
	fmt.Println()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Println()
	log.Printf("Received signal %v, shutting down...", sig)

	// Cancel context to stop all goroutines
	cancel()

	// Stop HTTP server first
	if httpServer != nil {
		if err := httpServer.Stop(); err != nil {
			log.Printf("Error stopping HTTP server: %v", err)
		}
	}

	// Stop engine gracefully
	if err := engine.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Shutdown complete")
}
