package client

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mythologyli/zju-connect/log"
	"inet.af/netaddr"
)

type IPResource struct {
	IPMin    net.IP
	IPMax    net.IP
	PortMin  int
	PortMax  int
	Protocol string
}

type DomainResource struct {
	PortMin  int
	PortMax  int
	Protocol string
}

type EasyConnectClient struct {
	server            string // Example: rvpn.zju.edu.cn:443. No protocol prefix
	username          string
	password          string
	totpSecret        string
	tlsCert           tls.Certificate
	testMultiLine     bool
	parseResource     bool
	useDomainResource bool

	httpClient *http.Client

	twfID string
	token *[48]byte

	lineList []string

	ipResources     []IPResource
	domainResources map[string]DomainResource
	ipSet           *netaddr.IPSet
	dnsResource     map[string]net.IP
	dnsServers      []string

	ip        net.IP // Client IP
	ipReverse []byte

	authTimestamp    time.Time // When the TWFID was initially obtained
	sessionTimestamp time.Time // When the current IP/Token was secured
	sessionFile      string    // Path to session file for persistence

	ipConn       io.ReadWriteCloser
	ipConnCancel context.CancelFunc

	refreshMutex    sync.Mutex
	lastRefreshTime time.Time
}

// AuthTimestamp returns the time when the current session was authenticated
func (c *EasyConnectClient) AuthTimestamp() time.Time {
	return c.authTimestamp
}

func (c *EasyConnectClient) SessionTimestamp() time.Time {
	return c.sessionTimestamp
}

// SetSessionFile sets the path for session persistence
func (c *EasyConnectClient) SetSessionFile(path string) {
	c.sessionFile = path
}

func NewEasyConnectClient(server, username, password, totpSecret string, tlsCert tls.Certificate, twfID string, testMultiLine, parseResource, useDomainResource bool) *EasyConnectClient {
	c := &EasyConnectClient{
		server:            server,
		username:          username,
		password:          password,
		totpSecret:        totpSecret,
		tlsCert:           tlsCert,
		testMultiLine:     testMultiLine,
		parseResource:     parseResource,
		useDomainResource: useDomainResource,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		twfID: twfID,
	}
	if twfID != "" {
		// init authTimestamp, overwrite when using session.json
		c.authTimestamp = time.Now()
	}
	return c
}

func (c *EasyConnectClient) IP() (net.IP, error) {
	if c.ip == nil {
		return nil, errors.New("IP not available")
	}

	return c.ip, nil
}

func (c *EasyConnectClient) IPSet() (*netaddr.IPSet, error) {
	if c.ipSet == nil {
		return nil, errors.New("IP set not available")
	}

	return c.ipSet, nil
}

func (c *EasyConnectClient) IPResources() ([]IPResource, error) {
	if c.ipResources == nil {
		return nil, errors.New("IP resources not available")
	}

	return c.ipResources, nil
}

func (c *EasyConnectClient) DomainResources() (map[string]DomainResource, error) {
	if c.domainResources == nil {
		return nil, errors.New("domain resources not available")
	}

	return c.domainResources, nil
}

func (c *EasyConnectClient) DNSResource() (map[string]net.IP, error) {
	if c.dnsResource == nil {
		return nil, errors.New("DNS resource not available")
	}

	return c.dnsResource, nil
}

func (c *EasyConnectClient) DNSServers() ([]string, error) {
	if len(c.dnsServers) == 0 {
		return nil, errors.New("DNS server not available")
	}

	return c.dnsServers, nil
}

func (c *EasyConnectClient) Setup() error {
	return c.setup(false)
}

func (c *EasyConnectClient) SetupAuto() error {
	return c.setup(true)
}

func (c *EasyConnectClient) setup(isAuto bool) error {
	// Use username/password/(SMS code) to get the TwfID
	if c.twfID == "" {
		err := c.requestTwfID(isAuto)
		if err != nil {
			return err
		}
	} // else we use the TwfID provided by user

	// Then we can get config from server and find the best line
	if c.testMultiLine {
		configStr, err := c.requestConfig()
		if err != nil {
			log.Printf("Error occurred while requesting config: %v", err)
		} else {
			err := c.parseLineListFromConfig(configStr)
			if err != nil {
				log.Printf("Error occurred while parsing config: %v", err)
			} else {
				log.Printf("Line list: %v", c.lineList)

				bestLine, err := findBestLine(c.lineList)
				if err != nil {
					log.Printf("Error occurred while finding best line: %v", err)
				} else {
					log.Printf("Best line: %v", bestLine)

					// Now we use the bestLine as new server
					if c.server != bestLine {
						c.server = bestLine
						c.testMultiLine = false
						c.twfID = ""

						return c.setup(isAuto)
					}
				}
			}
		}
	}

	// Then, use the TwfID to get token
	err := c.requestToken()
	if err != nil {
		return err
	}

	startTime := time.Now()

	// Then we get the resources from server
	if c.parseResource {
		resources, err := c.requestResources()
		if err != nil {
			log.Printf("Error occurred while requesting resources: %v", err)
		} else {
			// Parse the resources
			err = c.parseResources(resources)
			if err != nil {
				log.Printf("Error occurred while parsing resources: %v", err)
			}
		}
	}

	// Error may occur if we request too fast
	if time.Since(startTime) < time.Second {
		time.Sleep(time.Second - time.Since(startTime))
	}

	// Finally, use the token to get client IP
	err = c.requestIP(false)
	if err != nil {
		return err
	}

	c.sessionTimestamp = time.Now()

	// If sessionFile is configured, we save the very first successful setup
	if c.sessionFile != "" {
		if err := c.SaveSession(c.sessionFile); err != nil {
			log.Printf("Warning: failed to save initial session: %v", err)
		}
	}

	return nil
}
