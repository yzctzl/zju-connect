//go:build !tun

package main

import (
	"context"
	"crypto"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mythologyli/zju-connect/client"
	"github.com/mythologyli/zju-connect/configs"
	"github.com/mythologyli/zju-connect/dial"
	"github.com/mythologyli/zju-connect/internal/hook_func"
	"github.com/mythologyli/zju-connect/log"
	"github.com/mythologyli/zju-connect/resolve"
	"github.com/mythologyli/zju-connect/service"
	"github.com/mythologyli/zju-connect/stack"
	"github.com/mythologyli/zju-connect/stack/gvisor"
	"github.com/mythologyli/zju-connect/stack/tun"
	"golang.org/x/crypto/pkcs12"
	"inet.af/netaddr"
)

var conf configs.Config

const zjuConnectVersion = "0.9.0"

func main() {
	log.Init()

	log.Println("Start ZJU Connect v" + zjuConnectVersion)
	if conf.DebugDump {
		log.EnableDebug()
	}

	if errs := hook_func.ExecInitialFunc(context.Background(), conf); errs != nil {
		for _, err := range errs {
			log.Printf("Initial ZJU-Connect failed: %s", err)
		}
		os.Exit(1)
	}

	tlsCert := tls.Certificate{}
	if conf.CertFile != "" {
		p12Data, err := os.ReadFile(conf.CertFile)
		if err != nil {
			log.Fatalf("Read certificate file error: %s", err)
		}

		key, cert, err := pkcs12.Decode(p12Data, conf.CertPassword)
		if err != nil {
			log.Fatalf("Decode certificate file error: %s", err)
		}

		tlsCert = tls.Certificate{
			Certificate: [][]byte{cert.Raw},
			PrivateKey:  key.(crypto.PrivateKey),
			Leaf:        cert,
		}
	}

	vpnClient := client.NewEasyConnectClient(
		conf.ServerAddress+":"+fmt.Sprintf("%d", conf.ServerPort),
		conf.Username,
		conf.Password,
		conf.TOTPSecret,
		tlsCert,
		conf.TwfID,
		!conf.DisableMultiLine,
		!conf.DisableServerConfig,
		!conf.SkipDomainResource,
	)

	// Set session file path if configured
	if conf.SessionFile != "" {
		vpnClient.SetSessionFile(conf.SessionFile)
		// Try to restore session from file first
		if vpnClient.TryRestoreSession(conf.SessionFile) {
			log.Printf("Session restored from %s", conf.SessionFile)
		} else {
			// Session restore failed, do full setup
			err := vpnClient.Setup()
			if err != nil {
				log.Fatalf("EasyConnect client setup error: %s", err)
			}
		}
	} else {
		// No session file configured, do normal setup
		err := vpnClient.Setup()
		if err != nil {
			log.Fatalf("EasyConnect client setup error: %s", err)
		}
	}

	log.Printf("EasyConnect client started")

	ipResources, err := vpnClient.IPResources()
	if err != nil && !conf.DisableServerConfig {
		log.Println("No IP resources")
	}

	ipSet, err := vpnClient.IPSet()
	if err != nil && !conf.DisableServerConfig {
		log.Println("No IP set")
	}

	domainResources, err := vpnClient.DomainResources()
	if err != nil && !conf.DisableServerConfig {
		log.Println("No domain resources")
	}

	dnsResource, err := vpnClient.DNSResource()
	if err != nil && !conf.DisableServerConfig {
		log.Println("No DNS resource")
	}

	if !conf.DisableZJUConfig {
		if domainResources != nil {
			domainResources["zju.edu.cn"] = client.DomainResource{
				PortMin:  1,
				PortMax:  65535,
				Protocol: "all",
			}
		} else {
			domainResources = map[string]client.DomainResource{
				"zju.edu.cn": {
					PortMin:  1,
					PortMax:  65535,
					Protocol: "all",
				},
			}
		}

		if ipResources != nil {
			ipResources = append([]client.IPResource{{
				IPMin:    net.ParseIP("10.0.0.0"),
				IPMax:    net.ParseIP("10.255.255.255"),
				PortMin:  1,
				PortMax:  65535,
				Protocol: "all",
			}}, ipResources...)
		} else {
			ipResources = []client.IPResource{{
				IPMin:    net.ParseIP("10.0.0.0"),
				IPMax:    net.ParseIP("10.255.255.255"),
				PortMin:  1,
				PortMax:  65535,
				Protocol: "all",
			}}
		}

		ipSetBuilder := netaddr.IPSetBuilder{}
		if ipSet != nil {
			ipSetBuilder.AddSet(ipSet)
		}
		ipSetBuilder.AddPrefix(netaddr.MustParseIPPrefix("10.0.0.0/8"))
		ipSet, _ = ipSetBuilder.IPSet()
	}

	for _, customProxyDomain := range conf.CustomProxyDomain {
		if domainResources != nil {
			domainResources[customProxyDomain] = client.DomainResource{
				PortMin:  1,
				PortMax:  65535,
				Protocol: "all",
			}
		} else {
			domainResources = map[string]client.DomainResource{
				customProxyDomain: {
					PortMin:  1,
					PortMax:  65535,
					Protocol: "all",
				},
			}
		}
	}

	var vpnStack stack.Stack
	if conf.TUNMode {
		vpnTUNStack, err := tun.NewStack(vpnClient, conf.DNSHijack, ipResources)
		if err != nil {
			log.Fatalf("Tun stack setup error, make sure you are root user : %s", err)
		}

		if conf.AddRoute && ipSet != nil {
			for _, prefix := range ipSet.Prefixes() {
				log.Printf("Add route to %s", prefix.String())
				_ = vpnTUNStack.AddRoute(prefix.String())
			}
		} else if !conf.AddRoute && !conf.DisableZJUConfig {
			log.Println("Add route to 10.0.0.0/8")
			_ = vpnTUNStack.AddRoute("10.0.0.0/8")
		}

		vpnStack = vpnTUNStack
	} else {
		vpnStack, err = gvisor.NewStack(vpnClient)
		if err != nil {
			log.Fatalf("gVisor stack setup error: %s", err)
		}
	}

	useZJUDNS := !conf.DisableZJUDNS
	var zjuDNSServers []string
	if useZJUDNS {
		if conf.ZJUDNSServer == "auto" {
			servers, err := vpnClient.DNSServers()
			if err != nil {
				useZJUDNS = false
				zjuDNSServers = []string{"10.10.0.21"}
				log.Println("No DNS server provided by server. Disable ZJU DNS")
			} else {
				zjuDNSServers = servers
				log.Printf("Use DNS servers %v provided by server", zjuDNSServers)
			}
		} else {
			zjuDNSServers = []string{conf.ZJUDNSServer}
		}
	}

	vpnResolver := resolve.NewResolver(
		vpnStack,
		zjuDNSServers,
		conf.SecondaryDNSServer,
		conf.DNSTTL,
		domainResources,
		dnsResource,
		useZJUDNS,
	)

	for _, customDns := range conf.CustomDNSList {
		ipAddr := net.ParseIP(customDns.IP)
		if ipAddr == nil {
			log.Printf("Custom DNS for host name %s is invalid, SKIP", customDns.HostName)
		}
		vpnResolver.SetPermanentDNS(customDns.HostName, ipAddr)
		log.Printf("Add custom DNS: %s -> %s\n", customDns.HostName, customDns.IP)
	}
	allDNSServers := append([]string{}, zjuDNSServers...)
	allDNSServers = append(allDNSServers, conf.SecondaryDNSServer)
	localResolver := service.NewDnsServer(vpnResolver, allDNSServers)
	vpnStack.SetupResolve(localResolver)

	go vpnStack.Run()

	vpnDialer := dial.NewDialer(vpnStack, vpnResolver, ipResources, conf.ProxyAll, conf.DialDirectProxy)

	if conf.DNSServerBind != "" {
		go service.ServeDNS(conf.DNSServerBind, localResolver)
	}
	if conf.TUNMode {
		clientIP, _ := vpnClient.IP()
		go service.ServeDNS(clientIP.String()+":53", localResolver)
	}

	if conf.SocksBind != "" {
		go service.ServeSocks5(conf.SocksBind, vpnDialer, vpnResolver, conf.SocksUser, conf.SocksPasswd)
	}

	if conf.HTTPBind != "" {
		go service.ServeHTTP(conf.HTTPBind, vpnDialer)
	}

	if conf.ShadowsocksURL != "" {
		go service.ServeShadowsocks(vpnDialer, conf.ShadowsocksURL)
	}

	for _, portForwarding := range conf.PortForwardingList {
		switch portForwarding.NetworkType {
		case "tcp":
			go service.ServeTCPForwarding(vpnStack, portForwarding.BindAddress, portForwarding.RemoteAddress)
		case "udp":
			go service.ServeUDPForwarding(vpnStack, portForwarding.BindAddress, portForwarding.RemoteAddress)
		default:
			log.Printf("Port forwarding: unknown network type %s. Aborting", portForwarding.NetworkType)
		}
	}

	if !conf.DisableKeepAlive {
		if !useZJUDNS {
			log.Println("Keep alive is disabled because ZJU DNS is disabled")
		} else {
			go service.KeepAlive(vpnClient, vpnResolver, conf.KeepAliveDomain)
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	<-quit
	log.Println("Shutdown ZJU-Connect ......")
	if errs := hook_func.ExecTerminalFunc(context.Background()); errs != nil {
		for _, err := range errs {
			log.Printf("Shutdown ZJU-Connect failed: %s", err)
		}
	} else {
		log.Println("Shutdown ZJU-Connect success, Bye~")
	}
}
