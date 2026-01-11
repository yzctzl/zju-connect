package resolve

import (
	"net"

	"github.com/mythologyli/zju-connect/client"
	"github.com/patrickmn/go-cache"
)

// CachedDNSEntry stores both IP and optional DomainResource for proper context restoration
type CachedDNSEntry struct {
	IP             net.IP
	DomainResource *client.DomainResource // nil if not a VPN domain
}

func (r *Resolver) getDNSCache(host string) (*CachedDNSEntry, bool) {
	if item, found := r.dnsCache.Get(host); found {
		return item.(*CachedDNSEntry), true
	}
	return nil, false
}

func (r *Resolver) setDNSCache(host string, ip net.IP, domainResource *client.DomainResource) {
	entry := &CachedDNSEntry{
		IP:             ip,
		DomainResource: domainResource,
	}
	r.dnsCache.Set(host, entry, cache.DefaultExpiration)
}

func (r *Resolver) SetPermanentDNS(host string, ip net.IP) {
	entry := &CachedDNSEntry{
		IP:             ip,
		DomainResource: nil,
	}
	r.dnsCache.Set(host, entry, cache.NoExpiration)
}
