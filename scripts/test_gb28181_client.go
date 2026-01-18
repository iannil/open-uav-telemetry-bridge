// Test client for GB28181 protocol
// This script simulates a GB28181 SIP server (platform) for testing the GB28181 publisher
//
// Usage: go run scripts/test_gb28181_client.go [options]
//
// Options:
//   -listen string   Listen address (default ":5060")
//   -user   string   Expected device ID (default "34020000001320000001")
//   -pass   string   Device password (default "password123")
//   -realm  string   SIP realm (default "3402000000")
//
// The test client will:
// 1. Accept REGISTER requests and authenticate
// 2. Send Catalog queries
// 3. Accept and log MobilePosition notifications
// 4. Accept and log Keepalive messages

package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
)

var (
	listenAddr = flag.String("listen", ":5060", "Listen address")
	username   = flag.String("user", "34020000001320000001", "Expected device ID")
	password   = flag.String("pass", "password123", "Device password")
	realm      = flag.String("realm", "3402000000", "SIP realm")
)

func main() {
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("GB28181 Test Server starting on %s", *listenAddr)
	log.Printf("Expected device: %s, realm: %s", *username, *realm)

	// Create SIP user agent
	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent("GB28181-Test-Server/1.0"),
	)
	if err != nil {
		log.Fatalf("Failed to create UA: %v", err)
	}

	// Create SIP server
	server, err := sipgo.NewServer(ua)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Register handlers
	server.OnRegister(handleRegister)
	server.OnMessage(handleMessage)
	server.OnNotify(handleNotify)
	server.OnSubscribe(handleSubscribe)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start listening
	go func() {
		log.Printf("Listening for SIP messages on %s (UDP)", *listenAddr)
		if err := server.ListenAndServe(ctx, "udp", *listenAddr); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Press Ctrl+C to stop")
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}

// handleRegister handles REGISTER requests
func handleRegister(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("📥 REGISTER from %s", req.Source())

	// Check for Authorization header
	authHeader := req.GetHeader("Authorization")
	if authHeader == nil {
		// Send 401 Unauthorized with challenge
		log.Println("  ⚠️  No Authorization header, sending 401 challenge")
		resp := sip.NewResponseFromRequest(req, 401, "Unauthorized", nil)
		challenge := fmt.Sprintf(`Digest realm="%s", nonce="%s", algorithm=MD5`, *realm, generateNonce())
		resp.AppendHeader(sip.NewHeader("WWW-Authenticate", challenge))
		tx.Respond(resp)
		return
	}

	// Simple auth validation (in production, verify the digest properly)
	authValue := authHeader.Value()
	if !strings.Contains(authValue, *username) {
		log.Printf("  ❌ Authorization failed: wrong username")
		resp := sip.NewResponseFromRequest(req, 403, "Forbidden", nil)
		tx.Respond(resp)
		return
	}

	log.Printf("  ✅ Registration successful")

	// Send 200 OK
	resp := sip.NewResponseFromRequest(req, 200, "OK", nil)
	resp.AppendHeader(sip.NewHeader("Expires", "3600"))
	tx.Respond(resp)
}

// handleMessage handles MESSAGE requests (keepalive, catalog response, etc.)
func handleMessage(req *sip.Request, tx sip.ServerTransaction) {
	body := string(req.Body())
	log.Printf("📥 MESSAGE from %s", req.Source())

	// Parse XML content
	if strings.Contains(body, "<CmdType>Keepalive</CmdType>") {
		log.Println("  💓 Keepalive received")
		parseKeepalive(body)
	} else if strings.Contains(body, "<CmdType>Catalog</CmdType>") {
		log.Println("  📋 Catalog response received")
		parseCatalog(body)
	} else if strings.Contains(body, "<CmdType>DeviceInfo</CmdType>") {
		log.Println("  ℹ️  DeviceInfo response received")
	} else if strings.Contains(body, "<CmdType>DeviceStatus</CmdType>") {
		log.Println("  📊 DeviceStatus response received")
	} else {
		log.Printf("  📄 Unknown MESSAGE:\n%s", body)
	}

	// Send 200 OK
	resp := sip.NewResponseFromRequest(req, 200, "OK", nil)
	tx.Respond(resp)
}

// handleNotify handles NOTIFY requests (position updates)
func handleNotify(req *sip.Request, tx sip.ServerTransaction) {
	body := string(req.Body())
	log.Printf("📥 NOTIFY from %s", req.Source())

	if strings.Contains(body, "<CmdType>MobilePosition</CmdType>") {
		log.Println("  📍 MobilePosition received")
		parsePosition(body)
	} else {
		log.Printf("  📄 Unknown NOTIFY:\n%s", body)
	}

	// Send 200 OK
	resp := sip.NewResponseFromRequest(req, 200, "OK", nil)
	tx.Respond(resp)
}

// handleSubscribe handles SUBSCRIBE requests
func handleSubscribe(req *sip.Request, tx sip.ServerTransaction) {
	log.Printf("📥 SUBSCRIBE from %s", req.Source())
	log.Println("  📝 Subscription request received")

	// Send 200 OK with Expires header
	resp := sip.NewResponseFromRequest(req, 200, "OK", nil)
	resp.AppendHeader(sip.NewHeader("Expires", "3600"))
	tx.Respond(resp)
}

// MobilePosition represents a position notification
type MobilePosition struct {
	XMLName   xml.Name `xml:"Notify"`
	CmdType   string   `xml:"CmdType"`
	SN        int      `xml:"SN"`
	DeviceID  string   `xml:"DeviceID"`
	Time      string   `xml:"Time"`
	Longitude float64  `xml:"Longitude"`
	Latitude  float64  `xml:"Latitude"`
	Speed     float64  `xml:"Speed"`
	Direction float64  `xml:"Direction"`
	Altitude  float64  `xml:"Altitude"`
}

// parsePosition parses MobilePosition XML
func parsePosition(body string) {
	var pos MobilePosition
	if err := xml.Unmarshal([]byte(body), &pos); err != nil {
		log.Printf("    Failed to parse position: %v", err)
		return
	}

	log.Printf("    DeviceID: %s", pos.DeviceID)
	log.Printf("    Time: %s", pos.Time)
	log.Printf("    Position: %.6f, %.6f", pos.Latitude, pos.Longitude)
	log.Printf("    Altitude: %.1f m", pos.Altitude)
	log.Printf("    Speed: %.1f m/s, Direction: %.1f°", pos.Speed, pos.Direction)
}

// Keepalive represents a keepalive notification
type Keepalive struct {
	XMLName  xml.Name `xml:"Notify"`
	CmdType  string   `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Status   string   `xml:"Status"`
}

// parseKeepalive parses Keepalive XML
func parseKeepalive(body string) {
	var ka Keepalive
	if err := xml.Unmarshal([]byte(body), &ka); err != nil {
		log.Printf("    Failed to parse keepalive: %v", err)
		return
	}

	log.Printf("    DeviceID: %s, Status: %s, SN: %d", ka.DeviceID, ka.Status, ka.SN)
}

// CatalogResponse represents a catalog response
type CatalogResponse struct {
	XMLName    xml.Name `xml:"Response"`
	CmdType    string   `xml:"CmdType"`
	SN         int      `xml:"SN"`
	DeviceID   string   `xml:"DeviceID"`
	SumNum     int      `xml:"SumNum"`
	DeviceList struct {
		Items []CatalogItem `xml:"Item"`
	} `xml:"DeviceList"`
}

// CatalogItem represents a device in the catalog
type CatalogItem struct {
	DeviceID string `xml:"DeviceID"`
	Name     string `xml:"Name"`
	Status   string `xml:"Status"`
}

// parseCatalog parses Catalog response XML
func parseCatalog(body string) {
	var cat CatalogResponse
	if err := xml.Unmarshal([]byte(body), &cat); err != nil {
		log.Printf("    Failed to parse catalog: %v", err)
		return
	}

	log.Printf("    DeviceID: %s, SN: %d", cat.DeviceID, cat.SN)
	log.Printf("    Total devices: %d", cat.SumNum)
	for _, item := range cat.DeviceList.Items {
		log.Printf("      - %s: %s (%s)", item.DeviceID, item.Name, item.Status)
	}
}

// generateNonce generates a simple nonce for the challenge
func generateNonce() string {
	return fmt.Sprintf("%d", os.Getpid())
}
