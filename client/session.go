package client

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"time"

	"github.com/mythologyli/zju-connect/log"
)

// Session represents the persistent session data for EasyConnect
type Session struct {
	TwfID        string    `json:"twfid"`
	TokenHex     string    `json:"token"`      // hex encoded token
	IPStr            string    `json:"ip"`         // IP as string
	IPReverseHex     string    `json:"ip_reverse"` // hex encoded reverse IP
	Timestamp        time.Time `json:"timestamp"`  // AuthTimestamp (legacy/primary)
	SessionTimestamp time.Time `json:"session_timestamp,omitempty"`
	Server           string    `json:"server"`
}

// SaveSession saves the current session to a file
func (c *EasyConnectClient) SaveSession(path string) error {
	if c.twfID == "" || c.token == nil || c.ip == nil {
		return errors.New("incomplete session data")
	}

	session := Session{
		TwfID:        c.twfID,
		TokenHex:         hex.EncodeToString(c.token[:]),
		IPStr:            c.ip.String(),
		IPReverseHex:     hex.EncodeToString(c.ipReverse),
		Timestamp:        c.authTimestamp,
		SessionTimestamp: c.sessionTimestamp,
		Server:           c.server,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return err
	}

	log.Printf("Session saved to %s", path)
	return nil
}

// LoadSession loads a session from a file
func (c *EasyConnectClient) LoadSession(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var session Session
	err = json.Unmarshal(data, &session)
	if err != nil {
		return err
	}

	// Validate server matches
	if session.Server != c.server {
		log.Printf("Warning: Session server (%s) differs from current config (%s). Reusing session server.", session.Server, c.server)
		c.server = session.Server
	}

	// Check if session is too old (e.g., > 24 hours)
	if time.Since(session.Timestamp) > 24*time.Hour {
		log.Printf("Warning: Session might be expired (age: %v), but will attempt to restore it anyway.", time.Since(session.Timestamp))
	}

	// Decode token
	tokenBytes, err := hex.DecodeString(session.TokenHex)
	if err != nil {
		return err
	}
	if len(tokenBytes) != 48 {
		return errors.New("invalid token length")
	}

	// Decode IP reverse
	ipReverse, err := hex.DecodeString(session.IPReverseHex)
	if err != nil {
		return err
	}

	// Parse IP
	ip := net.ParseIP(session.IPStr)
	if ip == nil {
		return errors.New("invalid IP in session")
	}

	// Apply session data
	c.twfID = session.TwfID
	c.token = (*[48]byte)(tokenBytes)
	c.ip = ip
	c.ipReverse = ipReverse
	c.authTimestamp = session.Timestamp
	if !session.SessionTimestamp.IsZero() {
		c.sessionTimestamp = session.SessionTimestamp
	} else {
		// Fallback for old session files
		c.sessionTimestamp = session.Timestamp
	}

	log.Printf("Session loaded from %s (age: %v)", path, time.Since(session.Timestamp))
	return nil
}

// TryRestoreSession attempts to restore a session from file and validate it
// Returns true if session was successfully restored and validated
func (c *EasyConnectClient) TryRestoreSession(path string) bool {
	if path == "" {
		return false
	}

	originalServer := c.server
	err := c.LoadSession(path)
	if err != nil {
		c.server = originalServer
		log.Printf("Failed to load session: %v", err)
		return false
	}

	// Try to verify the session is still valid by requesting IP
	log.Printf("Validating restored session...")
	log.Printf("  TWFID: %s", c.twfID)
	log.Printf("  Token: %x", c.token[:])
	log.Printf("  IP: %s", c.ip.String())

	err = c.requestIP(true)
	if err != nil {
		log.Printf("Session validation failed: %v", err)
		log.Printf("Attempting to refresh token using existing TWFID...")

		// The TLS session ID might have expired, but the TWFID might still be valid.
		// Try to get a new token using the existing TWFID.
		if refreshErr := c.RefreshToken(); refreshErr == nil {
			log.Printf("Token refreshed successfully. New token: %x", c.token[:])
			log.Printf("Re-validating session with new token...")

			err = c.requestIP(false) // Token was refreshed, so this is no longer a pure restore, it's a new session
			if err != nil {
				log.Printf("Session re-validation failed after token refresh: %v", err)
				log.Printf("The server rejected the new token.")
			}
		} else {
			log.Printf("Token refresh failed: %v", refreshErr)
			log.Printf("This likely means the TWFID has expired and full re-authentication is required.")
		}
	}

	// If we still have an error after recovery attempts, fail
	if err != nil {
		log.Printf("Clearing invalid session data and falling back to full authentication...")
		c.server = originalServer
		c.twfID = ""
		c.token = nil
		c.ip = nil
		c.ipReverse = nil
		return false
	}

	log.Printf("Session restored successfully, client IP: %s", c.ip.String())

	// Fetch resources if enabled
	if c.parseResource {
		log.Print("Fetching resources for restored session...")
		resources, err := c.requestResources()
		if err != nil {
			log.Printf("Warning: Failed to fetch resources: %v. Continuing using cached session.", err)
		} else {
			err = c.parseResources(resources)
			if err != nil {
				log.Printf("Warning: Failed to parse resources: %v. Continuing using cached session.", err)
			}
		}
	}

	return true
}
