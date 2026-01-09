package service

import (
	"context"
	"net"
	"time"

	"github.com/mythologyli/zju-connect/client"
	"github.com/mythologyli/zju-connect/log"
	"github.com/mythologyli/zju-connect/resolve"
)

func KeepAlive(vpnClient *client.EasyConnectClient, resolver *resolve.Resolver) {
	var remoteUDPResolver *net.Resolver
	var err error

	// Retry loop for creating resolver
	backoff := time.Second
	maxBackoff := 5 * time.Minute
	for {
		remoteUDPResolver, err = resolver.RemoteUDPResolver()
		if err == nil {
			break
		}
		log.Printf("KeepAlive: failed to create resolver: %v. Retrying in %v...", err, backoff)
		time.Sleep(backoff)
		backoff = backoff * 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	consecutiveFailures := 0
	lastWebKeepAlive := time.Now()

	for {
		_, err := remoteUDPResolver.LookupIP(context.Background(), "ip4", "www.henu.edu.cn")
		if err != nil {
			consecutiveFailures++
			log.Printf("KeepAlive: %s (consecutive failures: %d)", err, consecutiveFailures)

			// If too many consecutive failures, try to recreate the resolver
			if consecutiveFailures >= 5 {
				log.Printf("KeepAlive: too many consecutive failures, attempting to recreate resolver...")
				newResolver, newErr := resolver.RemoteUDPResolver()
				if newErr == nil {
					remoteUDPResolver = newResolver
					consecutiveFailures = 0
					log.Printf("KeepAlive: resolver recreated successfully")
				}
			}
		} else {
			if consecutiveFailures > 0 {
				log.Printf("KeepAlive: recovered after %d failures", consecutiveFailures)
			}
			consecutiveFailures = 0
			log.Printf("KeepAlive: OK")
		}

		// Refresh web session every 30 minutes
		if time.Since(lastWebKeepAlive) >= 30*time.Minute {
			if kwErr := vpnClient.KeepWebSessionAlive(); kwErr != nil {
				log.Printf("KeepWebSession: Failed: %v", kwErr)
			} else {
				lastWebKeepAlive = time.Now()
			}
		}

		time.Sleep(60 * time.Second)
	}
}
