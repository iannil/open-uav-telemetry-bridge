package gb28181

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"

	"github.com/open-uav/telemetry-bridge/internal/config"
)

// SIPClient handles SIP protocol operations for GB28181
type SIPClient struct {
	cfg       config.GB28181Config
	ua        *sipgo.UserAgent
	client    *sipgo.Client
	server    *sipgo.Server
	auth      *DigestAuth
	cseq      atomic.Int32
	sn        atomic.Int32
	callID    string
	localAddr string
	mu        sync.RWMutex

	registered     bool
	registeredAt   time.Time
	registerCancel context.CancelFunc

	// Handler for incoming requests
	requestHandler func(req *sip.Request) *sip.Response
}

// NewSIPClient creates a new SIP client
func NewSIPClient(cfg config.GB28181Config) *SIPClient {
	return &SIPClient{
		cfg:       cfg,
		auth:      NewDigestAuth(cfg.Username, cfg.Password),
		localAddr: fmt.Sprintf("%s:%d", cfg.LocalIP, cfg.LocalPort),
	}
}

// Start initializes the SIP client and begins listening
func (c *SIPClient) Start(ctx context.Context) error {
	// Create user agent
	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent("OUTB-GB28181/1.0"),
	)
	if err != nil {
		return fmt.Errorf("create SIP user agent: %w", err)
	}
	c.ua = ua

	// Create SIP client
	client, err := sipgo.NewClient(ua)
	if err != nil {
		return fmt.Errorf("create SIP client: %w", err)
	}
	c.client = client

	// Create SIP server for incoming requests
	server, err := sipgo.NewServer(ua)
	if err != nil {
		return fmt.Errorf("create SIP server: %w", err)
	}
	c.server = server

	// Register request handlers
	c.registerHandlers()

	// Start listening
	transport := strings.ToLower(c.cfg.Transport)
	if transport == "" {
		transport = "udp"
	}

	go func() {
		listenErr := c.server.ListenAndServe(ctx, transport, c.localAddr)
		if listenErr != nil && ctx.Err() == nil {
			log.Printf("[GB28181] SIP server error: %v", listenErr)
		}
	}()

	// Generate unique call ID
	c.callID = fmt.Sprintf("%d@%s", time.Now().UnixNano(), c.cfg.LocalIP)

	return nil
}

// Stop gracefully stops the SIP client
func (c *SIPClient) Stop() error {
	if c.registerCancel != nil {
		c.registerCancel()
	}
	return nil
}

// SetRequestHandler sets the handler for incoming SIP requests
func (c *SIPClient) SetRequestHandler(handler func(req *sip.Request) *sip.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestHandler = handler
}

// registerHandlers registers SIP request handlers
func (c *SIPClient) registerHandlers() {
	// Handle MESSAGE requests (keepalive, catalog query, etc.)
	c.server.OnMessage(func(req *sip.Request, tx sip.ServerTransaction) {
		c.handleRequest(req, tx)
	})

	// Handle SUBSCRIBE requests
	c.server.OnSubscribe(func(req *sip.Request, tx sip.ServerTransaction) {
		c.handleRequest(req, tx)
	})

	// Handle INFO requests
	c.server.OnInfo(func(req *sip.Request, tx sip.ServerTransaction) {
		c.handleRequest(req, tx)
	})
}

// handleRequest handles incoming SIP requests
func (c *SIPClient) handleRequest(req *sip.Request, tx sip.ServerTransaction) {
	c.mu.RLock()
	handler := c.requestHandler
	c.mu.RUnlock()

	var resp *sip.Response
	if handler != nil {
		resp = handler(req)
	}

	if resp == nil {
		// Default 200 OK response
		resp = sip.NewResponseFromRequest(req, 200, "OK", nil)
	}

	if err := tx.Respond(resp); err != nil {
		log.Printf("[GB28181] Error sending response: %v", err)
	}
}

// buildRequest creates a SIP request with common headers
func (c *SIPClient) buildRequest(method sip.RequestMethod, requestURI, fromURI, toURI sip.Uri) *sip.Request {
	req := sip.NewRequest(method, requestURI)

	// Via header
	transport := strings.ToUpper(c.cfg.Transport)
	if transport == "" {
		transport = "UDP"
	}
	req.AppendHeader(sip.NewHeader("Via", fmt.Sprintf("SIP/2.0/%s %s;branch=z9hG4bK%d;rport",
		transport, c.localAddr, time.Now().UnixNano())))

	// From header
	req.AppendHeader(&sip.FromHeader{Address: fromURI, Params: sip.NewParams()})

	// To header
	req.AppendHeader(&sip.ToHeader{Address: toURI})

	// Call-ID header
	callID := sip.CallIDHeader(c.callID)
	req.AppendHeader(&callID)

	// CSeq header
	req.AppendHeader(&sip.CSeqHeader{SeqNo: uint32(c.cseq.Add(1)), MethodName: method})

	// Max-Forwards header
	req.AppendHeader(sip.NewHeader("Max-Forwards", "70"))

	return req
}

// Register sends REGISTER request to the SIP server
func (c *SIPClient) Register(ctx context.Context) error {
	requestURI := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.ServerID,
		Host:   c.cfg.ServerDomain,
	}

	fromAddr := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.DeviceID,
		Host:   c.cfg.ServerDomain,
	}

	toAddr := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.DeviceID,
		Host:   c.cfg.ServerDomain,
	}

	// Create REGISTER request
	req := c.buildRequest(sip.REGISTER, requestURI, fromAddr, toAddr)
	req.AppendHeader(&sip.ContactHeader{Address: sip.Uri{
		Scheme: "sip",
		User:   c.cfg.DeviceID,
		Host:   c.cfg.LocalIP,
		Port:   c.cfg.LocalPort,
	}})
	req.AppendHeader(sip.NewHeader("Expires", fmt.Sprintf("%d", c.cfg.RegisterExpires)))

	// Send first REGISTER (will likely get 401 Unauthorized)
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("send REGISTER: %w", err)
	}

	// Handle 401 Unauthorized - need authentication
	if resp.StatusCode == 401 {
		wwwAuth := resp.GetHeader("WWW-Authenticate")
		if wwwAuth == nil {
			return fmt.Errorf("401 response without WWW-Authenticate header")
		}

		// Parse challenge and generate response
		if err := c.auth.ParseChallenge(wwwAuth.Value()); err != nil {
			return fmt.Errorf("parse auth challenge: %w", err)
		}

		// Create new REGISTER with Authorization
		authReq := c.buildRequest(sip.REGISTER, requestURI, fromAddr, toAddr)
		authReq.AppendHeader(&sip.ContactHeader{Address: sip.Uri{
			Scheme: "sip",
			User:   c.cfg.DeviceID,
			Host:   c.cfg.LocalIP,
			Port:   c.cfg.LocalPort,
		}})
		authReq.AppendHeader(sip.NewHeader("Expires", fmt.Sprintf("%d", c.cfg.RegisterExpires)))

		// Add Authorization header
		authHeader := c.auth.GenerateResponse("REGISTER", requestURI.String())
		authReq.AppendHeader(sip.NewHeader("Authorization", authHeader))

		resp, err = c.client.Do(ctx, authReq)
		if err != nil {
			return fmt.Errorf("send authenticated REGISTER: %w", err)
		}
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("REGISTER failed with status %d: %s", resp.StatusCode, resp.Reason)
	}

	c.mu.Lock()
	c.registered = true
	c.registeredAt = time.Now()
	c.mu.Unlock()

	log.Printf("[GB28181] Registered successfully with server %s:%d", c.cfg.ServerIP, c.cfg.ServerPort)
	return nil
}

// StartRegistrationLoop starts the periodic registration refresh
func (c *SIPClient) StartRegistrationLoop(ctx context.Context) {
	regCtx, cancel := context.WithCancel(ctx)
	c.registerCancel = cancel

	// Refresh at 80% of expiry time
	refreshInterval := time.Duration(c.cfg.RegisterExpires*80/100) * time.Second
	if refreshInterval < time.Minute {
		refreshInterval = time.Minute
	}

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-regCtx.Done():
			return
		case <-ticker.C:
			if err := c.Register(regCtx); err != nil {
				log.Printf("[GB28181] Registration refresh failed: %v", err)
			}
		}
	}
}

// IsRegistered returns whether the client is registered
func (c *SIPClient) IsRegistered() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.registered
}

// SendMessage sends a SIP MESSAGE request
func (c *SIPClient) SendMessage(ctx context.Context, contentType, body string) error {
	requestURI := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.ServerID,
		Host:   c.cfg.ServerDomain,
	}

	fromAddr := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.DeviceID,
		Host:   c.cfg.ServerDomain,
	}

	toAddr := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.ServerID,
		Host:   c.cfg.ServerDomain,
	}

	// Use a new Call-ID for each MESSAGE
	req := sip.NewRequest(sip.MESSAGE, requestURI)

	transport := strings.ToUpper(c.cfg.Transport)
	if transport == "" {
		transport = "UDP"
	}
	req.AppendHeader(sip.NewHeader("Via", fmt.Sprintf("SIP/2.0/%s %s;branch=z9hG4bK%d;rport",
		transport, c.localAddr, time.Now().UnixNano())))
	req.AppendHeader(&sip.FromHeader{Address: fromAddr, Params: sip.NewParams()})
	req.AppendHeader(&sip.ToHeader{Address: toAddr})

	newCallID := sip.CallIDHeader(fmt.Sprintf("%d@%s", time.Now().UnixNano(), c.cfg.LocalIP))
	req.AppendHeader(&newCallID)

	req.AppendHeader(&sip.CSeqHeader{SeqNo: uint32(c.cseq.Add(1)), MethodName: sip.MESSAGE})
	req.AppendHeader(sip.NewHeader("Max-Forwards", "70"))
	req.AppendHeader(sip.NewHeader("Content-Type", contentType))
	req.SetBody([]byte(body))

	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("send MESSAGE: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("MESSAGE failed with status %d: %s", resp.StatusCode, resp.Reason)
	}

	return nil
}

// SendNotify sends a SIP NOTIFY request
func (c *SIPClient) SendNotify(ctx context.Context, contentType, body string) error {
	requestURI := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.ServerID,
		Host:   c.cfg.ServerDomain,
	}

	fromAddr := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.DeviceID,
		Host:   c.cfg.ServerDomain,
	}

	toAddr := sip.Uri{
		Scheme: "sip",
		User:   c.cfg.ServerID,
		Host:   c.cfg.ServerDomain,
	}

	// Use a new Call-ID for each NOTIFY
	req := sip.NewRequest(sip.NOTIFY, requestURI)

	transport := strings.ToUpper(c.cfg.Transport)
	if transport == "" {
		transport = "UDP"
	}
	req.AppendHeader(sip.NewHeader("Via", fmt.Sprintf("SIP/2.0/%s %s;branch=z9hG4bK%d;rport",
		transport, c.localAddr, time.Now().UnixNano())))
	req.AppendHeader(&sip.FromHeader{Address: fromAddr, Params: sip.NewParams()})
	req.AppendHeader(&sip.ToHeader{Address: toAddr})

	newCallID := sip.CallIDHeader(fmt.Sprintf("%d@%s", time.Now().UnixNano(), c.cfg.LocalIP))
	req.AppendHeader(&newCallID)

	req.AppendHeader(&sip.CSeqHeader{SeqNo: uint32(c.cseq.Add(1)), MethodName: sip.NOTIFY})
	req.AppendHeader(sip.NewHeader("Max-Forwards", "70"))
	req.AppendHeader(sip.NewHeader("Event", "presence"))
	req.AppendHeader(sip.NewHeader("Subscription-State", "active"))
	req.AppendHeader(sip.NewHeader("Content-Type", contentType))
	req.SetBody([]byte(body))

	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("send NOTIFY: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("NOTIFY failed with status %d: %s", resp.StatusCode, resp.Reason)
	}

	return nil
}

// NextSN returns the next sequence number for XML messages
func (c *SIPClient) NextSN() int {
	return int(c.sn.Add(1))
}
