package service

import (
	"context"
	"net"
	"time"

	"github.com/mythologyli/zju-connect/client"
	"github.com/mythologyli/zju-connect/log"
	"github.com/mythologyli/zju-connect/resolve"
)

func KeepAlive(vpnClient *client.EasyConnectClient, resolver *resolve.Resolver, keepAliveDomain string) {
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
	lastFullRefresh := time.Now()

	for {
		_, err := remoteUDPResolver.LookupIP(context.Background(), "ip4", keepAliveDomain)
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
			log.DebugPrintf("KeepAlive: OK")
		}

		// Refresh web session every 30 minutes
		if time.Since(lastWebKeepAlive) >= 30*time.Minute {
			if kwErr := vpnClient.KeepWebSessionAlive(); kwErr != nil {
				log.Printf("KeepWebSession: Failed: %v", kwErr)
			} else {
				lastWebKeepAlive = time.Now()
			}
		}

		// Proactively refresh full session every 12 hours to prevent 24h expiration
		if time.Since(lastFullRefresh) >= 12*time.Hour {
			log.Printf("KeepAlive: Proactive session refresh triggered (12h interval)")
			if rfErr := vpnClient.RefreshSession(); rfErr != nil {
				log.Printf("KeepAlive: Proactive session refresh failed: %v. Will retry in next loop.", rfErr)
			} else {
				log.Printf("KeepAlive: Proactive session refresh successful")
				lastFullRefresh = time.Now()
			}
		}

		time.Sleep(60 * time.Second)
	}
}
