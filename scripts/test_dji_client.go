// +build ignore

// Test client that simulates an Android DJI Forwarder
// Usage: go run scripts/test_dji_client.go [options]
//
// Options:
//   -addr string     Gateway address (default "127.0.0.1:14560")
//   -device string   Device ID (default "dji-test-001")
//   -count int       Number of telemetry messages to send (default 5)
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"time"
)

var (
	serverAddr = flag.String("addr", "127.0.0.1:14560", "Gateway address")
	deviceID   = flag.String("device", "dji-test-001", "Device ID")
	msgCount   = flag.Int("count", 5, "Number of telemetry messages to send")
)

type Message struct {
	Type      string          `json:"type"`
	DeviceID  string          `json:"device_id,omitempty"`
	Version   string          `json:"sdk_version,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type Location struct {
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	AltGNSS     float64 `json:"alt_gnss"`
	AltRelative float64 `json:"alt_relative"`
}

type Attitude struct {
	Pitch float64 `json:"pitch"`
	Roll  float64 `json:"roll"`
	Yaw   float64 `json:"yaw"`
}

type Status struct {
	Armed         bool    `json:"armed"`
	FlightMode    string  `json:"flight_mode"`
	BatteryPct    float64 `json:"battery_percent"`
	SignalQuality int     `json:"signal_quality"`
}

type Velocity struct {
	VX float64 `json:"vx"`
	VY float64 `json:"vy"`
	VZ float64 `json:"vz"`
}

type DroneState struct {
	DeviceID       string   `json:"device_id"`
	Timestamp      int64    `json:"timestamp"`
	ProtocolSource string   `json:"protocol_source"`
	Location       Location `json:"location"`
	Attitude       Attitude `json:"attitude"`
	Status         Status   `json:"status"`
	Velocity       Velocity `json:"velocity"`
}

func sendMessage(conn net.Conn, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Send length prefix (4 bytes, big-endian)
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	if _, err := conn.Write(lengthBuf); err != nil {
		return err
	}

	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

func readMessage(conn net.Conn) (*Message, error) {
	// Read length prefix
	lengthBuf := make([]byte, 4)
	if _, err := conn.Read(lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// Read message data
	data := make([]byte, length)
	if _, err := conn.Read(data); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func main() {
	flag.Parse()

	fmt.Println("DJI Forwarder Test Client")
	fmt.Println("========================")
	fmt.Printf("Connecting to OUTB gateway at %s...\n", *serverAddr)

	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected!")

	// Send HELLO message
	helloMsg := Message{
		Type:     "hello",
		DeviceID: *deviceID,
		Version:  "5.0.0",
	}

	fmt.Println("Sending HELLO message...")
	if err := sendMessage(conn, helloMsg); err != nil {
		fmt.Printf("Failed to send HELLO: %v\n", err)
		return
	}

	// Wait for ACK
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	ackMsg, err := readMessage(conn)
	if err != nil {
		fmt.Printf("Failed to read ACK: %v\n", err)
		return
	}
	fmt.Printf("Received: %s\n", ackMsg.Type)

	// Send some simulated telemetry data
	fmt.Printf("\nSending simulated telemetry (%d messages)...\n", *msgCount)

	for i := 0; i < *msgCount; i++ {
		state := DroneState{
			DeviceID:       *deviceID,
			Timestamp:      time.Now().UnixMilli(),
			ProtocolSource: "dji_mobile_sdk",
			Location: Location{
				Lat:         39.9042 + float64(i)*0.0001, // Beijing coordinates
				Lon:         116.4074 + float64(i)*0.0001,
				AltGNSS:     100.0 + float64(i)*5,
				AltRelative: 50.0 + float64(i)*5,
			},
			Attitude: Attitude{
				Pitch: 2.5,
				Roll:  -1.2,
				Yaw:   float64(45 + i*10),
			},
			Status: Status{
				Armed:         true,
				FlightMode:    "GPS_NORMAL",
				BatteryPct:    85.0 - float64(i),
				SignalQuality: 95,
			},
			Velocity: Velocity{
				VX: 5.0,
				VY: 3.0,
				VZ: -1.0,
			},
		}

		stateData, _ := json.Marshal(state)
		stateMsg := Message{
			Type: "state",
			Data: stateData,
		}

		if err := sendMessage(conn, stateMsg); err != nil {
			fmt.Printf("Failed to send state: %v\n", err)
			return
		}

		fmt.Printf("  [%d] Sent: lat=%.4f, lon=%.4f, alt=%.1f, battery=%.0f%%\n",
			i+1, state.Location.Lat, state.Location.Lon,
			state.Location.AltGNSS, state.Status.BatteryPct)

		time.Sleep(500 * time.Millisecond)
	}

	// Send heartbeat
	fmt.Println("\nSending HEARTBEAT...")
	heartbeatMsg := Message{
		Type:      "heartbeat",
		Timestamp: time.Now().UnixMilli(),
	}
	if err := sendMessage(conn, heartbeatMsg); err != nil {
		fmt.Printf("Failed to send heartbeat: %v\n", err)
		return
	}

	// Wait for ACK
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	ackMsg, err = readMessage(conn)
	if err != nil {
		fmt.Printf("Failed to read heartbeat ACK: %v\n", err)
		return
	}
	fmt.Printf("Received: %s\n", ackMsg.Type)

	fmt.Println("\n========================")
	fmt.Println("Test completed successfully!")
}
