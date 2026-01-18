package gb28181

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

// DigestAuth handles SIP digest authentication
type DigestAuth struct {
	username string
	password string
	realm    string
	nonce    string
	qop      string
	nc       int
	cnonce   string
}

// NewDigestAuth creates a new digest authentication handler
func NewDigestAuth(username, password string) *DigestAuth {
	return &DigestAuth{
		username: username,
		password: password,
		nc:       0,
	}
}

// ParseChallenge parses WWW-Authenticate header from 401 response
func (d *DigestAuth) ParseChallenge(wwwAuth string) error {
	// Extract realm
	realmRe := regexp.MustCompile(`realm="([^"]+)"`)
	if matches := realmRe.FindStringSubmatch(wwwAuth); len(matches) > 1 {
		d.realm = matches[1]
	}

	// Extract nonce
	nonceRe := regexp.MustCompile(`nonce="([^"]+)"`)
	if matches := nonceRe.FindStringSubmatch(wwwAuth); len(matches) > 1 {
		d.nonce = matches[1]
	}

	// Extract qop (optional)
	qopRe := regexp.MustCompile(`qop="([^"]+)"`)
	if matches := qopRe.FindStringSubmatch(wwwAuth); len(matches) > 1 {
		d.qop = matches[1]
	}

	if d.realm == "" || d.nonce == "" {
		return fmt.Errorf("invalid WWW-Authenticate header: missing realm or nonce")
	}

	return nil
}

// GenerateResponse generates the Authorization header value
func (d *DigestAuth) GenerateResponse(method, uri string) string {
	d.nc++
	d.cnonce = generateCNonce()

	// Calculate HA1 = MD5(username:realm:password)
	ha1 := md5Hash(fmt.Sprintf("%s:%s:%s", d.username, d.realm, d.password))

	// Calculate HA2 = MD5(method:uri)
	ha2 := md5Hash(fmt.Sprintf("%s:%s", method, uri))

	// Calculate response
	var response string
	if d.qop != "" {
		// qop=auth: MD5(HA1:nonce:nc:cnonce:qop:HA2)
		response = md5Hash(fmt.Sprintf("%s:%s:%08x:%s:%s:%s",
			ha1, d.nonce, d.nc, d.cnonce, d.qop, ha2))
	} else {
		// No qop: MD5(HA1:nonce:HA2)
		response = md5Hash(fmt.Sprintf("%s:%s:%s", ha1, d.nonce, ha2))
	}

	// Build Authorization header
	var authHeader strings.Builder
	authHeader.WriteString(fmt.Sprintf(`Digest username="%s"`, d.username))
	authHeader.WriteString(fmt.Sprintf(`, realm="%s"`, d.realm))
	authHeader.WriteString(fmt.Sprintf(`, nonce="%s"`, d.nonce))
	authHeader.WriteString(fmt.Sprintf(`, uri="%s"`, uri))
	authHeader.WriteString(fmt.Sprintf(`, response="%s"`, response))

	if d.qop != "" {
		authHeader.WriteString(fmt.Sprintf(`, qop=%s`, d.qop))
		authHeader.WriteString(fmt.Sprintf(`, nc=%08x`, d.nc))
		authHeader.WriteString(fmt.Sprintf(`, cnonce="%s"`, d.cnonce))
	}

	return authHeader.String()
}

// md5Hash computes MD5 hash and returns hex string
func md5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", hash)
}

// generateCNonce generates a random client nonce
func generateCNonce() string {
	const charset = "abcdef0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
