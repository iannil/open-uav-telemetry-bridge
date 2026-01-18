// +build ignore

// WebSocket test client for OUTB gateway
// Usage: go run scripts/test_ws_client.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type     string          `json:"type"`
	DeviceID string          `json:"device_id,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

func main() {
	fmt.Println("WebSocket Test Client")
	fmt.Println("=====================")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/api/v1/ws"}
	fmt.Printf("Connecting to %s...\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer c.Close()

	fmt.Println("Connected! Waiting for messages...")
	fmt.Println()

	done := make(chan struct{})

	// Read messages
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			var msg WSMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Invalid message: %s", string(message))
				continue
			}

			switch msg.Type {
			case "state_update":
				var state map[string]interface{}
				json.Unmarshal(msg.Data, &state)
				loc := state["location"].(map[string]interface{})
				status := state["status"].(map[string]interface{})
				fmt.Printf("[%s] STATE: device=%s lat=%.4f lon=%.4f alt=%.1f mode=%s battery=%.0f%%\n",
					time.Now().Format("15:04:05"),
					msg.DeviceID,
					loc["lat"], loc["lon"], loc["alt_gnss"],
					status["flight_mode"], status["battery_percent"])

			case "drone_online":
				fmt.Printf("[%s] ONLINE: %s\n", time.Now().Format("15:04:05"), msg.DeviceID)

			case "drone_offline":
				fmt.Printf("[%s] OFFLINE: %s\n", time.Now().Format("15:04:05"), msg.DeviceID)

			default:
				fmt.Printf("[%s] %s: %s\n", time.Now().Format("15:04:05"), msg.Type, string(message))
			}
		}
	}()

	// Wait for interrupt or done
	select {
	case <-done:
	case <-interrupt:
		fmt.Println("\nInterrupted, closing connection...")

		// Cleanly close the connection
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Write close error:", err)
			return
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}

	fmt.Println("Done")
}
