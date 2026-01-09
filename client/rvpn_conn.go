package client

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/mythologyli/zju-connect/log"
)

type RvpnConn struct {
	easyConnectClient *EasyConnectClient

	sendConn     io.WriteCloser
	sendLock     sync.Mutex
	sendErrCount int

	recvConn     io.ReadCloser
	recvLock     sync.Mutex
	recvErrCount int
}

// try best to read, if return err!=nil, please panic
func (r *RvpnConn) Read(p []byte) (n int, err error) {
	r.recvLock.Lock()
	defer r.recvLock.Unlock()

	backoff := time.Second
	maxBackoff := 5 * time.Minute

	for {
		n, err = r.recvConn.Read(p)
		if err == nil {
			r.recvErrCount = 0
			return n, nil
		}

		r.recvErrCount++
		log.Printf("Error occurred while receiving (attempt %d): %v", r.recvErrCount, err)

		// Close the old connection
		_ = r.recvConn.Close()

		// Reconnect loop with exponential backoff
		innerRetryCount := 0
		for {
			innerRetryCount++
			// Wait before retry (except first attempt)
			if backoff > time.Second {
				log.Printf("Waiting %v before reconnect...", backoff)
				time.Sleep(backoff)
			}

			r.recvConn, err = r.easyConnectClient.RecvConn()
			if err == nil {
				log.Printf("RecvConn reconnected successfully")
				r.recvErrCount = 0
				backoff = time.Second // Reset backoff on success
				break
			}

			log.Printf("RecvConn failed (attempt %d): %v", innerRetryCount, err)

			// Detailed diagnostic analysis
			errStr := err.Error()
			var diagnose string
			isLikelySessionExpiry := false

			switch {
			case errors.Is(err, io.EOF) || strings.Contains(errStr, "EOF"):
				diagnose = "Server abruptly closed connection during handshake. This strongly suggests the Session/Token has expired on the server side."
				isLikelySessionExpiry = true
			case strings.Contains(errStr, "handshake reply"):
				diagnose = "VPN Handshake rejected by server. The token/session ID is likely no longer valid."
				isLikelySessionExpiry = true
			case strings.Contains(errStr, "connection reset by peer"):
				diagnose = "Connection reset by peer. This can be a transient network issue OR the server killing the session. Retrying to distinguish."
			case strings.Contains(errStr, "connection refused"):
				diagnose = "Target server refused connection. Possibly a temporary service disruption or firewall block."
			default:
				diagnose = errStr
			}
			log.Printf("Diagnostic [Attempt %d]: %s", innerRetryCount, diagnose)

			// Only trigger RefreshSession after 6 failed attempts (cumulative or inner)
			// We strictly follow the 6-retry rule even if we suspect session expiry (EOF/Handshake).
			if r.recvErrCount > 7 || innerRetryCount > 7 {
				reason := "exceeded 7 retries for general connection failures"
				if isLikelySessionExpiry {
					reason = "exceeded 7 retries with high-confidence session expiry indicators (EOF/Handshake)"
				}
				log.Printf("Session refresh triggered (%s). Final error: %v (total: %d, inner: %d). Refreshing session...", reason, err, r.recvErrCount, innerRetryCount)

				if rtokErr := r.easyConnectClient.RefreshSession(); rtokErr != nil {
					log.Printf("Session refresh failed: %v", rtokErr)
					// Return fatal error to caller - session cannot be recovered
					return 0, fmt.Errorf("session refresh failed: %w", rtokErr)
				} else {
					log.Printf("Session refreshed successfully. Resetting counters and retrying connection.")
					r.recvErrCount = 0 // Reset attempt counters on successful refresh
					innerRetryCount = 0
					backoff = time.Second // Reset backoff
					// Immediately try to reconnect with new session in next loop iteration
					continue
				}
			}

			// Increase backoff for next attempt
			backoff = backoff * 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}

		// Reconnected, continue to next Read attempt
	}
}

// try best to write, if return err!=nil, please panic
func (r *RvpnConn) Write(p []byte) (n int, err error) {
	r.sendLock.Lock()
	defer r.sendLock.Unlock()

	backoff := time.Second
	maxBackoff := 5 * time.Minute

	for {
		n, err = r.sendConn.Write(p)
		if err == nil {
			r.sendErrCount = 0
			return n, nil
		}

		r.sendErrCount++
		log.Printf("Error occurred while sending (attempt %d): %v", r.sendErrCount, err)

		// Close the old connection
		_ = r.sendConn.Close()

		// Reconnect loop with exponential backoff
		innerRetryCount := 0
		for {
			innerRetryCount++
			// Wait before retry (except first attempt)
			if backoff > time.Second {
				log.Printf("Waiting %v before reconnect...", backoff)
				time.Sleep(backoff)
			}

			r.sendConn, err = r.easyConnectClient.SendConn()
			if err == nil {
				log.Printf("SendConn reconnected successfully")
				r.sendErrCount = 0
				backoff = time.Second // Reset backoff on success
				break
			}

			log.Printf("SendConn failed (attempt %d): %v", innerRetryCount, err)

			// Detailed diagnostic analysis
			errStr := err.Error()
			var diagnose string
			isLikelySessionExpiry := false

			switch {
			case errors.Is(err, io.EOF) || strings.Contains(errStr, "EOF"):
				diagnose = "Server abruptly closed connection during handshake. This strongly suggests the Session/Token has expired on the server side."
				isLikelySessionExpiry = true
			case strings.Contains(errStr, "handshake reply"):
				diagnose = "VPN Handshake rejected by server. The token/session ID is likely no longer valid."
				isLikelySessionExpiry = true
			case strings.Contains(errStr, "connection reset by peer"):
				diagnose = "Connection reset by peer. This can be a transient network issue OR the server killing the session. Retrying to distinguish."
			case strings.Contains(errStr, "connection refused"):
				diagnose = "Target server refused connection. Possibly a temporary service disruption or firewall block."
			default:
				diagnose = errStr
			}
			log.Printf("Diagnostic [Attempt %d]: %s", innerRetryCount, diagnose)

			// Only trigger RefreshSession after 6 failed attempts (cumulative or inner)
			if r.sendErrCount > 7 || innerRetryCount > 7 {
				reason := "exceeded 7 retries for general connection failures"
				if isLikelySessionExpiry {
					reason = "exceeded 7 retries with high-confidence session expiry indicators (EOF/Handshake)"
				}
				log.Printf("Session refresh triggered (%s). Final error: %v (total: %d, inner: %d). Refreshing session...", reason, err, r.sendErrCount, innerRetryCount)

				if rtokErr := r.easyConnectClient.RefreshSession(); rtokErr != nil {
					log.Printf("Session refresh failed: %v", rtokErr)
					// Return fatal error to caller - session cannot be recovered
					return 0, fmt.Errorf("session refresh failed: %w", rtokErr)
				} else {
					log.Printf("Session refreshed successfully. Resetting counters and retrying connection.")
					r.sendErrCount = 0 // Reset attempt counters on successful refresh
					innerRetryCount = 0
					backoff = time.Second
					// Immediately try to reconnect with new session
					continue
				}
			}

			// Increase backoff for next attempt
			backoff = backoff * 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}

		// Reconnected, continue to next Write attempt
	}
}

func (r *RvpnConn) Close() error {
	if r.sendConn != nil {
		_ = r.sendConn.Close()
	}
	if r.recvConn != nil {
		_ = r.recvConn.Close()
	}
	return nil
}

func NewRvpnConn(ec *EasyConnectClient) (*RvpnConn, error) {
	c := &RvpnConn{
		easyConnectClient: ec,
		sendErrCount:      0,
		recvErrCount:      0,
	}

	var err error
	c.sendConn, err = ec.SendConn()
	if err != nil {
		log.Printf("Error occurred while creating sendConn: %v", err)
		return nil, err
	}

	c.recvConn, err = ec.RecvConn()
	if err != nil {
		log.Printf("Error occurred while creating recvConn: %v", err)
		return nil, err
	}
	return c, nil
}
