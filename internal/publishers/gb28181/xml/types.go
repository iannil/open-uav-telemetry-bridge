package xml

import (
	"encoding/xml"
	"fmt"
	"math"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

// XMLDeclaration is the standard XML declaration for GB28181 messages
const XMLDeclaration = `<?xml version="1.0" encoding="GB2312"?>`

// CmdType represents GB28181 command types
type CmdType string

const (
	CmdTypeCatalog        CmdType = "Catalog"
	CmdTypeDeviceInfo     CmdType = "DeviceInfo"
	CmdTypeDeviceStatus   CmdType = "DeviceStatus"
	CmdTypeMobilePosition CmdType = "MobilePosition"
	CmdTypeKeepalive      CmdType = "Keepalive"
	CmdTypeRecordInfo     CmdType = "RecordInfo"
)

// Query represents a GB28181 query message
type Query struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

// Response represents a GB28181 response message
type Response struct {
	XMLName  xml.Name `xml:"Response"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Result   string   `xml:"Result,omitempty"`
}

// Notify represents a GB28181 notification message
type Notify struct {
	XMLName  xml.Name `xml:"Notify"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

// KeepaliveNotify represents a keepalive notification
type KeepaliveNotify struct {
	XMLName  xml.Name `xml:"Notify"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Status   string   `xml:"Status"`
}

// MobilePositionNotify represents a mobile position notification
type MobilePositionNotify struct {
	XMLName   xml.Name `xml:"Notify"`
	CmdType   CmdType  `xml:"CmdType"`
	SN        int      `xml:"SN"`
	DeviceID  string   `xml:"DeviceID"`
	Time      string   `xml:"Time"`
	Longitude float64  `xml:"Longitude"`
	Latitude  float64  `xml:"Latitude"`
	Speed     float64  `xml:"Speed"`
	Direction float64  `xml:"Direction"`
	Altitude  float64  `xml:"Altitude"`
}

// NewMobilePositionNotify creates a MobilePosition notification from DroneState
func NewMobilePositionNotify(state *models.DroneState, sn int) *MobilePositionNotify {
	// Convert timestamp (Unix milliseconds) to ISO8601
	t := time.UnixMilli(state.Timestamp)
	timeStr := t.Format("2006-01-02T15:04:05")

	// Calculate ground speed from Vx, Vy
	speed := math.Sqrt(state.Velocity.Vx*state.Velocity.Vx + state.Velocity.Vy*state.Velocity.Vy)

	// Normalize yaw to 0-360 range
	direction := state.Attitude.Yaw
	for direction < 0 {
		direction += 360
	}
	for direction >= 360 {
		direction -= 360
	}

	return &MobilePositionNotify{
		CmdType:   CmdTypeMobilePosition,
		SN:        sn,
		DeviceID:  state.DeviceID,
		Time:      timeStr,
		Longitude: state.Location.Lon,
		Latitude:  state.Location.Lat,
		Speed:     speed,
		Direction: direction,
		Altitude:  state.Location.AltGNSS,
	}
}

// Marshal serializes the notification to XML with declaration
func (n *MobilePositionNotify) Marshal() (string, error) {
	data, err := xml.MarshalIndent(n, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal mobile position: %w", err)
	}
	return XMLDeclaration + "\r\n" + string(data), nil
}

// NewKeepaliveNotify creates a keepalive notification
func NewKeepaliveNotify(deviceID string, sn int) *KeepaliveNotify {
	return &KeepaliveNotify{
		CmdType:  CmdTypeKeepalive,
		SN:       sn,
		DeviceID: deviceID,
		Status:   "OK",
	}
}

// Marshal serializes the keepalive notification to XML with declaration
func (n *KeepaliveNotify) Marshal() (string, error) {
	data, err := xml.MarshalIndent(n, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal keepalive: %w", err)
	}
	return XMLDeclaration + "\r\n" + string(data), nil
}
