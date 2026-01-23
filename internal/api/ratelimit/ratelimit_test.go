package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(10.0, 20)

	if limiter == nil {
		t.Fatal("NewIPRateLimiter should return non-nil limiter")
	}
	if limiter.r != 10.0 {
		t.Errorf("Rate = %v, want 10.0", limiter.r)
	}
	if limiter.b != 20 {
		t.Errorf("Burst = %d, want 20", limiter.b)
	}
	if limiter.ips == nil {
		t.Error("IPs map should be initialized")
	}
}

func TestIPRateLimiter_Allow(t *testing.T) {
	// Create limiter with 2 requests/sec, burst of 2
	limiter := NewIPRateLimiter(2.0, 2)

	ip := "192.168.1.1"

	// First two requests should be allowed (burst)
	if !limiter.Allow(ip) {
		t.Error("First request should be allowed")
	}
	if !limiter.Allow(ip) {
		t.Error("Second request should be allowed (within burst)")
	}

	// Third request should be denied (exceeded burst)
	if limiter.Allow(ip) {
		t.Error("Third request should be denied (burst exceeded)")
	}
}

func TestIPRateLimiter_DifferentIPs(t *testing.T) {
	// Create limiter with 1 request/sec, burst of 1
	limiter := NewIPRateLimiter(1.0, 1)

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// First IP uses its quota
	if !limiter.Allow(ip1) {
		t.Error("First IP first request should be allowed")
	}
	if limiter.Allow(ip1) {
		t.Error("First IP second request should be denied")
	}

	// Second IP should have its own quota
	if !limiter.Allow(ip2) {
		t.Error("Second IP first request should be allowed")
	}
}

func TestIPRateLimiter_Recovery(t *testing.T) {
	// Create limiter with 10 requests/sec, burst of 1
	limiter := NewIPRateLimiter(10.0, 1)

	ip := "192.168.1.1"

	// Use the quota
	if !limiter.Allow(ip) {
		t.Error("First request should be allowed")
	}
	if limiter.Allow(ip) {
		t.Error("Second request should be denied")
	}

	// Wait for token recovery
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !limiter.Allow(ip) {
		t.Error("Request after recovery should be allowed")
	}
}

func TestMiddleware_AllowedRequest(t *testing.T) {
	limiter := NewIPRateLimiter(100.0, 100)
	middleware := Middleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "success" {
		t.Errorf("Body = %s, want 'success'", rr.Body.String())
	}
}

func TestMiddleware_RateLimited(t *testing.T) {
	limiter := NewIPRateLimiter(1.0, 1)
	middleware := Middleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("First request status = %d, want %d", rr1.Code, http.StatusOK)
	}

	// Second request (should be rate limited)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request status = %d, want %d", rr2.Code, http.StatusTooManyRequests)
	}

	// Check Retry-After header
	if rr2.Header().Get("Retry-After") != "1" {
		t.Errorf("Retry-After = %s, want '1'", rr2.Header().Get("Retry-After"))
	}

	// Check Content-Type header
	if rr2.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %s, want 'application/json'", rr2.Header().Get("Content-Type"))
	}
}

func TestGetIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	ip := getIP(req)

	if ip != "192.168.1.1:12345" {
		t.Errorf("IP = %s, want '192.168.1.1:12345'", ip)
	}
}

func TestGetIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.195")

	ip := getIP(req)

	if ip != "203.0.113.195" {
		t.Errorf("IP = %s, want '203.0.113.195'", ip)
	}
}

func TestGetIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Real-IP", "198.51.100.178")

	ip := getIP(req)

	if ip != "198.51.100.178" {
		t.Errorf("IP = %s, want '198.51.100.178'", ip)
	}
}

func TestGetIP_XForwardedForPriority(t *testing.T) {
	// X-Forwarded-For should take priority over X-Real-IP
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.195")
	req.Header.Set("X-Real-IP", "198.51.100.178")

	ip := getIP(req)

	if ip != "203.0.113.195" {
		t.Errorf("IP = %s, want '203.0.113.195' (X-Forwarded-For takes priority)", ip)
	}
}

func TestIPRateLimiter_Concurrent(t *testing.T) {
	limiter := NewIPRateLimiter(1000.0, 1000)

	// Launch concurrent requests from same IP
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			limiter.Allow("concurrent-test-ip")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should not panic with concurrent access
}
