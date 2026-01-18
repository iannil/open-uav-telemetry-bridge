package gb28181

import (
	"context"
	"encoding/xml"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/emiago/sipgo/sip"

	gbxml "github.com/open-uav/telemetry-bridge/internal/publishers/gb28181/xml"
)

// RequestHandler handles incoming SIP requests for GB28181
type RequestHandler struct {
	deviceMgr *DeviceManager
	subMgr    *SubscriptionManager
	sipClient *SIPClient
}

// NewRequestHandler creates a new request handler
func NewRequestHandler(deviceMgr *DeviceManager, subMgr *SubscriptionManager, sipClient *SIPClient) *RequestHandler {
	return &RequestHandler{
		deviceMgr: deviceMgr,
		subMgr:    subMgr,
		sipClient: sipClient,
	}
}

// HandleRequest processes incoming SIP requests
func (h *RequestHandler) HandleRequest(req *sip.Request) *sip.Response {
	contentType := req.GetHeader("Content-Type")
	if contentType == nil {
		return sip.NewResponseFromRequest(req, 200, "OK", nil)
	}

	// Parse XML body
	body := req.Body()
	if len(body) == 0 {
		return sip.NewResponseFromRequest(req, 200, "OK", nil)
	}

	// Detect message type from XML
	if req.Method == sip.MESSAGE {
		return h.handleMessage(req, body)
	} else if req.Method == sip.SUBSCRIBE {
		return h.handleSubscribe(req)
	}

	return sip.NewResponseFromRequest(req, 200, "OK", nil)
}

// handleMessage handles MESSAGE requests (queries from platform)
func (h *RequestHandler) handleMessage(req *sip.Request, body []byte) *sip.Response {
	// Try to detect query type
	var query gbxml.Query
	if err := xml.Unmarshal(body, &query); err != nil {
		log.Printf("[GB28181] Failed to parse query XML: %v", err)
		return sip.NewResponseFromRequest(req, 400, "Bad Request", nil)
	}

	switch query.CmdType {
	case gbxml.CmdTypeCatalog:
		return h.handleCatalogQuery(req, query.SN, query.DeviceID)
	case gbxml.CmdTypeDeviceInfo:
		return h.handleDeviceInfoQuery(req, query.SN, query.DeviceID)
	case gbxml.CmdTypeDeviceStatus:
		return h.handleDeviceStatusQuery(req, query.SN, query.DeviceID)
	default:
		log.Printf("[GB28181] Unknown query type: %s", query.CmdType)
	}

	return sip.NewResponseFromRequest(req, 200, "OK", nil)
}

// handleCatalogQuery responds to Catalog queries
func (h *RequestHandler) handleCatalogQuery(req *sip.Request, sn int, deviceID string) *sip.Response {
	log.Printf("[GB28181] Received Catalog query (SN=%d, DeviceID=%s)", sn, deviceID)

	// Build catalog items from registered channels
	channels := h.deviceMgr.GetAllChannels()
	items := make([]gbxml.CatalogItem, len(channels))
	for i, ch := range channels {
		items[i] = gbxml.NewCatalogItem(
			ch.DeviceID,
			ch.Name,
			h.deviceMgr.GatewayID(),
			h.deviceMgr.CivilCode(),
			ch.Online,
		)
	}

	// Create response
	catalogResp := gbxml.NewCatalogResponse(h.deviceMgr.GatewayID(), sn, items)
	respBody, err := catalogResp.Marshal()
	if err != nil {
		log.Printf("[GB28181] Failed to marshal catalog response: %v", err)
		return sip.NewResponseFromRequest(req, 500, "Internal Server Error", nil)
	}

	// Send response via MESSAGE (async)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.sipClient.SendMessage(ctx, "Application/MANSCDP+xml", respBody); err != nil {
			log.Printf("[GB28181] Failed to send catalog response: %v", err)
		}
	}()

	return sip.NewResponseFromRequest(req, 200, "OK", nil)
}

// handleDeviceInfoQuery responds to DeviceInfo queries
func (h *RequestHandler) handleDeviceInfoQuery(req *sip.Request, sn int, deviceID string) *sip.Response {
	log.Printf("[GB28181] Received DeviceInfo query (SN=%d, DeviceID=%s)", sn, deviceID)

	// Build device info response
	respBody := buildDeviceInfoResponse(h.deviceMgr.GatewayID(), sn)

	// Send response via MESSAGE (async)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.sipClient.SendMessage(ctx, "Application/MANSCDP+xml", respBody); err != nil {
			log.Printf("[GB28181] Failed to send device info response: %v", err)
		}
	}()

	return sip.NewResponseFromRequest(req, 200, "OK", nil)
}

// handleDeviceStatusQuery responds to DeviceStatus queries
func (h *RequestHandler) handleDeviceStatusQuery(req *sip.Request, sn int, deviceID string) *sip.Response {
	log.Printf("[GB28181] Received DeviceStatus query (SN=%d, DeviceID=%s)", sn, deviceID)

	// Build device status response
	respBody := buildDeviceStatusResponse(h.deviceMgr.GatewayID(), sn)

	// Send response via MESSAGE (async)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.sipClient.SendMessage(ctx, "Application/MANSCDP+xml", respBody); err != nil {
			log.Printf("[GB28181] Failed to send device status response: %v", err)
		}
	}()

	return sip.NewResponseFromRequest(req, 200, "OK", nil)
}

// handleSubscribe handles SUBSCRIBE requests for position updates
func (h *RequestHandler) handleSubscribe(req *sip.Request) *sip.Response {
	log.Printf("[GB28181] Received SUBSCRIBE request")

	// Extract subscription parameters
	expiresHeader := req.GetHeader("Expires")
	expires := 3600 // Default 1 hour
	if expiresHeader != nil {
		if val, err := strconv.Atoi(expiresHeader.Value()); err == nil {
			expires = val
		}
	}

	// Extract event type
	eventHeader := req.GetHeader("Event")
	eventType := "presence"
	if eventHeader != nil {
		eventType = eventHeader.Value()
	}

	// Extract interval from body if present
	interval := 5 // Default 5 seconds
	body := req.Body()
	if len(body) > 0 {
		// Try to extract Interval from XML
		intervalRe := regexp.MustCompile(`<Interval>(\d+)</Interval>`)
		if matches := intervalRe.FindSubmatch(body); len(matches) > 1 {
			if val, err := strconv.Atoi(string(matches[1])); err == nil {
				interval = val
			}
		}
	}

	// Create subscription
	callID := req.GetHeader("Call-ID")
	subID := ""
	if callID != nil {
		subID = callID.Value()
	} else {
		subID = time.Now().Format("20060102150405")
	}

	sub := &Subscription{
		ID:        subID,
		DeviceID:  "*", // Subscribe to all devices
		Interval:  interval,
		Expires:   time.Now().Add(time.Duration(expires) * time.Second),
		EventType: eventType,
	}
	h.subMgr.Add(sub)

	log.Printf("[GB28181] Subscription created: ID=%s, Interval=%ds, Expires=%v", sub.ID, sub.Interval, sub.Expires)

	// Create 200 OK response with Expires header
	resp := sip.NewResponseFromRequest(req, 200, "OK", nil)
	resp.AppendHeader(sip.NewHeader("Expires", strconv.Itoa(expires)))

	return resp
}

// buildDeviceInfoResponse builds a DeviceInfo response XML
func buildDeviceInfoResponse(deviceID string, sn int) string {
	return gbxml.XMLDeclaration + "\r\n" +
		`<Response>
  <CmdType>DeviceInfo</CmdType>
  <SN>` + strconv.Itoa(sn) + `</SN>
  <DeviceID>` + deviceID + `</DeviceID>
  <Result>OK</Result>
  <DeviceName>UAV-Gateway</DeviceName>
  <Manufacturer>OUTB</Manufacturer>
  <Model>v1.0</Model>
  <Firmware>1.0.0</Firmware>
  <Channel>0</Channel>
</Response>`
}

// buildDeviceStatusResponse builds a DeviceStatus response XML
func buildDeviceStatusResponse(deviceID string, sn int) string {
	return gbxml.XMLDeclaration + "\r\n" +
		`<Response>
  <CmdType>DeviceStatus</CmdType>
  <SN>` + strconv.Itoa(sn) + `</SN>
  <DeviceID>` + deviceID + `</DeviceID>
  <Result>OK</Result>
  <Online>ONLINE</Online>
  <Status>OK</Status>
</Response>`
}
