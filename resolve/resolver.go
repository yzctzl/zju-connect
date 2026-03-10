package resolve

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/mythologyli/zju-connect/client"
	"github.com/mythologyli/zju-connect/log"
	"github.com/mythologyli/zju-connect/stack"
	"github.com/patrickmn/go-cache"
)

type resolveResult struct {
	done chan struct{}
	once sync.Once
}

type Resolver struct {
	remoteDNSServers  []string
	stack             stack.Stack
	secondaryResolver *net.Resolver
	ttl               uint64
	domainResources   map[string]client.DomainResource
	dnsResource       map[string]net.IP
	useRemoteDNS      bool
	upstreamDNSMode   string

	dnsCache *cache.Cache

	timer  *time.Timer
	useTCP bool
	tcpGen int64 // generation counter for tcp mode timer
	// check to use tcp resolver or udp resolver
	tcpLock sync.RWMutex
	// singleflight pattern: ensures only one goroutine resolves a given host at a time
	// maps hostname to a channel that signals when resolution is complete
	concurResolveLock sync.Map // map[string]*resolveResult
}

type contextKey string

var (
	ContextKeyResolveHost    = contextKey("RESOLVE_HOST")
	ContextKeyDomainResource = contextKey("DOMAIN_RESOURCE")
)

// Resolve ip address. If the host could be visited via VPN, this function set a DOMAIN_RESOURCE value in context. If resolve success, this function set a RESOLVE_HOST value in context.
func (r *Resolver) Resolve(ctx context.Context, host string) (resCtx context.Context, resIP net.IP, resErr error) {
	resCtx = ctx
	defer func() {
		if resErr == nil {
			resCtx = context.WithValue(resCtx, ContextKeyResolveHost, host)
		}
	}()

	// Determine VPN domain before touching cache/static DNS so cached answers
	// still restore the correct routing context.
	var domainRes *client.DomainResource
	isVPNDomain := r.matchVPNDomain(host, &domainRes)
	if isVPNDomain && domainRes != nil {
		resCtx = context.WithValue(resCtx, ContextKeyDomainResource, *domainRes)
		log.DebugPrintf("Domain %s matched VPN resource", host)
	}

	// 1. Check Cache FIRST (fastest path)
	if entry, found := r.getDNSCache(host); found {
		log.DebugPrintf("%s -> %s (Cache)", host, entry.IP.String())
		if entry.DomainResource != nil {
			resCtx = context.WithValue(resCtx, ContextKeyDomainResource, *entry.DomainResource)
		} else if domainRes != nil {
			resCtx = context.WithValue(resCtx, ContextKeyDomainResource, *domainRes)
		}
		return resCtx, entry.IP, nil
	}

	// 2. Static IP Resource Check (Hosts)
	if r.dnsResource != nil {
		if ip, found := r.dnsResource[host]; found {
			log.DebugPrintf("%s -> %s (Static)", host, ip.String())
			if domainRes != nil {
				resCtx = context.WithValue(resCtx, ContextKeyDomainResource, *domainRes)
			}
			return resCtx, ip, nil
		}
	}

	switch r.upstreamDNSMode {
	case "remote-only":
		return r.resolveWithRemoteDNS(resCtx, host, domainRes, false)
	case "remote-first":
		return r.resolveWithRemoteDNS(resCtx, host, domainRes, true)
	}

	// 3. Decide DNS resolver based on split DNS policy
	// VPN domains MUST use VPN DNS; others use local DNS unless forced
	shouldUseVPNResolver := isVPNDomain || r.useRemoteDNS

	if !shouldUseVPNResolver {
		log.DebugPrintf("%s -> using local DNS (not VPN domain)", host)
		return r.ResolveWithSecondaryDNS(resCtx, host)
	}

	return r.resolveWithRemoteDNS(resCtx, host, domainRes, true)
}

func (r *Resolver) resolveWithRemoteDNS(ctx context.Context, host string, domainRes *client.DomainResource, allowLocalFallback bool) (context.Context, net.IP, error) {
	// Remote DNS Resolution with singleflight pattern
	resultItem, loaded := r.concurResolveLock.LoadOrStore(host, &resolveResult{
		done: make(chan struct{}),
	})
	result := resultItem.(*resolveResult)

	if !loaded {
		// We're the first goroutine for this host - do the actual resolution
		defer func() {
			// Remove the entry from the map so that after TTL expires,
			// a new resolution can be triggered.
			r.concurResolveLock.Delete(host)
			// Signal completion to all waiting goroutines
			result.once.Do(func() {
				close(result.done)
			})
		}()

		ip, err := r.resolveViaRemoteDNS(ctx, host)
		if err == nil {
			r.setDNSCache(host, ip, domainRes)
			log.DebugPrintf("%s -> %s (VPN DNS)", host, ip.String())
			return ctx, ip, nil
		}

		if !allowLocalFallback {
			return ctx, nil, err
		}

		// VPN DNS failed, fallback to secondary
		log.Printf("VPN DNS failed for %s: %v, trying local DNS", host, err)
		fallbackCtx, ip, fallbackErr := r.ResolveWithSecondaryDNS(ctx, host)
		if fallbackErr == nil {
			r.setDNSCache(host, ip, domainRes)
		}
		return fallbackCtx, ip, fallbackErr
	}

	// Another goroutine is resolving this host, wait for it to complete
	select {
	case <-result.done:
	case <-ctx.Done():
		return ctx, nil, ctx.Err()
	}

	// Check cache after the resolving goroutine has completed
	if entry, found := r.getDNSCache(host); found {
		log.DebugPrintf("%s -> %s (VPN DNS, from concurrent resolution)", host, entry.IP.String())
		if entry.DomainResource != nil {
			ctx = context.WithValue(ctx, ContextKeyDomainResource, *entry.DomainResource)
		} else if domainRes != nil {
			ctx = context.WithValue(ctx, ContextKeyDomainResource, *domainRes)
		}
		return ctx, entry.IP, nil
	}

	if !allowLocalFallback {
		return ctx, nil, fmt.Errorf("VPN DNS resolution failed for %s", host)
	}

	// Concurrent resolution failed, try secondary as fallback
	return r.ResolveWithSecondaryDNS(ctx, host)
}

// matchVPNDomain checks if host matches any VPN domain resource
// Returns true if matched, and sets domainRes to the matched resource
func (r *Resolver) matchVPNDomain(host string, domainRes **client.DomainResource) bool {
	if r.domainResources == nil {
		return false
	}

	var longestMatch string
	for domain, resource := range r.domainResources {
		// Strict matching: exact match or subdomain (ends with .domain)
		if host == domain || strings.HasSuffix(host, "."+domain) {
			if len(domain) > len(longestMatch) {
				longestMatch = domain
				copyRes := resource
				*domainRes = &copyRes
			}
		}
	}
	return longestMatch != ""
}

// resolveViaRemoteDNS performs DNS resolution using VPN DNS (UDP with TCP fallback)
func (r *Resolver) resolveViaRemoteDNS(ctx context.Context, host string) (net.IP, error) {
	r.tcpLock.RLock()
	useTCP := r.useTCP
	r.tcpLock.RUnlock()

	var lastErr error
	for _, server := range r.remoteDNSServers {
		var ips []net.IP
		var err error

		if !useTCP {
			// Try UDP
			ips, err = r.lookupViaServer(ctx, server, "udp", host)
			if err != nil {
				// UDP failed, try TCP on same server
				log.DebugPrintf("UDP DNS failed for %s on %s, trying TCP", host, server)
				ips, err = r.lookupViaServer(ctx, server, "tcp", host)
				if err == nil {
					r.switchToTCPMode()
				}
			}
		} else {
			// Already in TCP mode
			ips, err = r.lookupViaServer(ctx, server, "tcp", host)
		}

		if err == nil && len(ips) > 0 {
			return ips[0], nil
		}
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no IP addresses found for %s", host)
}

func (r *Resolver) lookupViaServer(ctx context.Context, server, network, host string) ([]net.IP, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, n, addr string) (net.Conn, error) {
			if network == "tcp" {
				return r.stack.DialTCP(&net.TCPAddr{
					IP:   net.ParseIP(server),
					Port: 53,
				})
			}
			return r.stack.DialUDP(&net.UDPAddr{
				IP:   net.ParseIP(server),
				Port: 53,
			})
		},
	}
	return resolver.LookupIP(ctx, "ip4", host)
}

// switchToTCPMode switches DNS resolver to TCP mode for 10 minutes
func (r *Resolver) switchToTCPMode() {
	r.tcpLock.Lock()
	defer r.tcpLock.Unlock()

	r.useTCP = true
	r.tcpGen++
	currentGen := r.tcpGen

	// Always stop old timer and create new one to capture new currentGen
	if r.timer != nil {
		r.timer.Stop()
	}
	r.timer = time.AfterFunc(10*time.Minute, func() {
		r.tcpLock.Lock()
		defer r.tcpLock.Unlock()

		// Only switch back if we are still in the same generation
		if r.tcpGen == currentGen {
			r.useTCP = false
			r.timer = nil
		}
	})
}

func (r *Resolver) RemoteUDPResolver() (*net.Resolver, error) {
	if len(r.remoteDNSServers) == 0 {
		return nil, errors.New("remote DNS servers are empty")
	}
	// Return a resolver that uses the first available server
	// KeepAlive uses this. We might want to optimize it to try multiple servers too.
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return r.stack.DialUDP(&net.UDPAddr{
				IP:   net.ParseIP(r.remoteDNSServers[0]),
				Port: 53,
			})
		},
	}, nil
}

func (r *Resolver) ResolveWithSecondaryDNS(ctx context.Context, host string) (context.Context, net.IP, error) {
	if targets, err := r.secondaryResolver.LookupIP(ctx, "ip4", host); err != nil {
		log.Printf("%s", "Resolve IPv4 addr failed using secondary DNS: "+host+". Try IPv6 addr")

		if targets, err = r.secondaryResolver.LookupIP(ctx, "ip6", host); err != nil {
			log.Printf("%s", "Resolve IPv6 addr failed using secondary DNS: "+host)
			return ctx, nil, err
		} else {
			log.Printf("%s -> %s", host, targets[0].String())
			return ctx, targets[0], nil
		}
	} else {
		log.Printf("%s -> %s", host, targets[0].String())
		return ctx, targets[0], nil
	}
}

func (r *Resolver) CleanCache(duration time.Duration) {
	// go-cache handles its own cleanup via the background janitor
	// We don't need to manually clean concurResolveLock as it handles itself
	// leaving this function for future use or removing entirely
	select {}
}

func NewResolver(stack stack.Stack, remoteDNSServers []string, secondaryDNSServer string, ttl uint64, domainResources map[string]client.DomainResource, dnsResource map[string]net.IP, useRemoteDNS bool, upstreamDNSMode string) *Resolver {
	resolver := &Resolver{
		remoteDNSServers: remoteDNSServers,
		stack:            stack,
		ttl:              ttl,
		domainResources:  domainResources,
		dnsResource:      dnsResource,
		dnsCache:         cache.New(time.Duration(ttl)*time.Second, time.Duration(ttl)*2*time.Second),
		useRemoteDNS:     useRemoteDNS,
		upstreamDNSMode:  upstreamDNSMode,
	}

	if secondaryDNSServer != "" {
		resolver.secondaryResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.DialUDP(network, nil, &net.UDPAddr{
					IP:   net.ParseIP(secondaryDNSServer),
					Port: 53,
				})
			},
		}
	} else {
		resolver.secondaryResolver = &net.Resolver{
			PreferGo: true,
		}
	}
	return resolver
}
