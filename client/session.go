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
	IPStr        string    `json:"ip"`         // IP as string
	IPReverseHex string    `json:"ip_reverse"` // hex encoded reverse IP
	Timestamp    time.Time `json:"timestamp"`
	Server       string    `json:"server"`
}

// SaveSession saves the current session to a file
func (c *EasyConnectClient) SaveSession(path string) error {
	if c.twfID == "" || c.token == nil || c.ip == nil {
		return errors.New("incomplete session data")
	}

	session := Session{
		TwfID:        c.twfID,
		TokenHex:     hex.EncodeToString(c.token[:]),
		IPStr:        c.ip.String(),
		IPReverseHex: hex.EncodeToString(c.ipReverse),
		Timestamp:    time.Now(),
		Server:       c.server,
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
		return errors.New("session server mismatch")
	}

	// Check if session is too old (e.g., > 24 hours)
	if time.Since(session.Timestamp) > 24*time.Hour {
		log.Printf("Session expired (age: %v)", time.Since(session.Timestamp))
		return errors.New("session expired")
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
	c.ip = ip.To4()
	c.ipReverse = ipReverse

	log.Printf("Session loaded from %s (age: %v)", path, time.Since(session.Timestamp))
	return nil
}

// TryRestoreSession attempts to restore a session from file and validate it
// Returns true if session was successfully restored and validated
func (c *EasyConnectClient) TryRestoreSession(path string) bool {
	if path == "" {
		return false
	}

	err := c.LoadSession(path)
	if err != nil {
		log.Printf("Failed to load session: %v", err)
		return false
	}

	// Try to verify the session is still valid by requesting IP
	// This also ensures the connection to server is working
	log.Printf("Validating restored session...")
	err = c.requestIP()
	if err != nil {
		log.Printf("Session validation failed: %v", err)
		// Clear the loaded session data
		c.twfID = ""
		c.token = nil
		c.ip = nil
		c.ipReverse = nil
		return false
	}

	log.Printf("Session restored successfully, client IP: %s", c.ip.String())
	return true
}
