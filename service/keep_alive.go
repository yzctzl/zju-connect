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
	remoteResolverRetryAt := time.Time{}

	tryRemoteResolver := func() {
		if !remoteResolverRetryAt.IsZero() && time.Now().Before(remoteResolverRetryAt) {
			return
		}

		var err error
		remoteUDPResolver, err = resolver.RemoteUDPResolver()
		if err != nil {
			remoteUDPResolver = nil
			remoteResolverRetryAt = time.Now().Add(5 * time.Minute)
			log.Printf("KeepAlive: remote DNS probe disabled for now: %v", err)
			return
		}

		remoteResolverRetryAt = time.Time{}
	}

	tryRemoteResolver()

	consecutiveFailures := 0
	lastWebKeepAlive := time.Now()
	lastAttemptFullRefresh := time.Time{} // Track the last time we tried a full refresh

	for {
		if remoteUDPResolver == nil {
			tryRemoteResolver()
		}

		if remoteUDPResolver != nil {
			_, err := remoteUDPResolver.LookupIP(context.Background(), "ip4", keepAliveDomain)
			if err != nil {
				consecutiveFailures++
				log.Printf("KeepAlive: %s (consecutive failures: %d)", err, consecutiveFailures)

				// If too many consecutive failures, try to recreate the resolver
				if consecutiveFailures >= 5 {
					log.Printf("KeepAlive: too many consecutive failures, attempting to recreate resolver...")
					remoteUDPResolver = nil
					consecutiveFailures = 0
					tryRemoteResolver()
					if remoteUDPResolver != nil {
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
		}

		// Refresh web session every 30 minutes
		if time.Since(lastWebKeepAlive) >= 30*time.Minute {
			if kwErr := vpnClient.KeepWebSessionAlive(); kwErr != nil {
				log.Printf("KeepWebSession: Failed: %v", kwErr)
			} else {
				lastWebKeepAlive = time.Now()
			}
		}

		// Proactively refresh full session every 12 hours base on ACTUAL session age
		// Or if the session is old and we haven't tried recently
		sessionAge := time.Since(vpnClient.AuthTimestamp())
		if sessionAge >= 12*time.Hour && time.Since(lastAttemptFullRefresh) >= 1*time.Hour {
			log.Printf("KeepAlive: Proactive session refresh triggered (Session age: %v)", sessionAge)
			lastAttemptFullRefresh = time.Now() // Mark attempt before calling to prevent overlaps
			if rfErr := vpnClient.RefreshSession(true); rfErr != nil {
				log.Printf("KeepAlive: Proactive session refresh failed: %v. Will retry in 1 hour.", rfErr)
			} else {
				log.Printf("KeepAlive: Proactive session refresh successful. New session age: %v", time.Since(vpnClient.AuthTimestamp()))
				// sessionAge will now be small, so it won't trigger until 12h later
			}
		}

		time.Sleep(60 * time.Second)
	}
}
