package client

import (
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
		for {
			// Wait before retry (except first attempt)
			if backoff > time.Second {
				log.Printf("Waiting %v before reconnect...", backoff)
				time.Sleep(backoff)
			}

			r.recvConn, err = r.easyConnectClient.RecvConn()
			if err == nil {
				log.Printf("RecvConn reconnected successfully")
				backoff = time.Second // Reset backoff on success
				break
			}

			log.Printf("RecvConn failed: %v. Retrying...", err)

			// If handshake error, try refresh token
			if strings.Contains(err.Error(), "handshake reply") {
				log.Printf("Possible token expiry. Attempting to refresh session...")
				if rtokErr := r.easyConnectClient.RefreshSession(); rtokErr != nil {
					log.Printf("Session refresh failed: %v", rtokErr)
				} else {
					log.Printf("Session refreshed successfully")
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
		for {
			// Wait before retry (except first attempt)
			if backoff > time.Second {
				log.Printf("Waiting %v before reconnect...", backoff)
				time.Sleep(backoff)
			}

			r.sendConn, err = r.easyConnectClient.SendConn()
			if err == nil {
				log.Printf("SendConn reconnected successfully")
				backoff = time.Second // Reset backoff on success
				break
			}

			log.Printf("SendConn failed: %v. Retrying...", err)

			// If handshake error, try refresh session
			if strings.Contains(err.Error(), "handshake reply") {
				log.Printf("Possible token expiry. Attempting to refresh session...")
				if rtokErr := r.easyConnectClient.RefreshSession(); rtokErr != nil {
					log.Printf("Session refresh failed: %v", rtokErr)
				} else {
					log.Printf("Session refreshed successfully")
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
