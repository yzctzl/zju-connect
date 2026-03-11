package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/mythologyli/zju-connect/log"
	"github.com/pquerna/otp/totp"
	utls "github.com/refraction-networking/utls"
)

var ErrSMSRequired = errors.New("SMS code required")
var errTOTPRequired = errors.New("TOTP required")
var errCertRequired = errors.New("cert required")

func (c *EasyConnectClient) requestTwfID(isAuto bool) error {
	err := c.loginAuthAndPsw(isAuto)
	if err != nil {
		if errors.Is(err, ErrSMSRequired) {
			if isAuto {
				return ErrSMSRequired
			}
			err = c.loginSMS()
			if err != nil {
				return err
			}
		} else if errors.Is(err, errTOTPRequired) {
			err = c.loginTOTP()
			if err != nil {
				return err
			}
		} else if errors.Is(err, errCertRequired) {
			err = c.loginCert()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (c *EasyConnectClient) loginAuthAndPsw(isAuto bool) error {
	// First we request the TwfID from server
	addr := "https://" + c.server + "/por/login_auth.csp?apiversion=1"
	log.Printf("Request: %s", addr)

	resp, err := c.httpClient.Get(addr)
	if err != nil {
		debug.PrintStack()
		return err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}

	bodyStr := buf.String()
	log.DebugPrintln("Response:", bodyStr)

	// Pre-emptive SMS protection: Stop if SMS is detected and we are in auto mode
	// Detection of <Type>2</Type> or <NextAuth>2</NextAuth> or <NextService>auth/sms</NextService>
	if isAuto && (strings.Contains(bodyStr, "<Type>2</Type>") ||
		strings.Contains(bodyStr, "<NextAuth>2</NextAuth>") ||
		strings.Contains(bodyStr, "<NextService>auth/sms</NextService>")) {

		// If the user has a safer alternative (TOTP or Cert), we might still be okay,
		// but if they ONLY have password+TOTP/nothing, SMS is a high risk.
		// However, to be absolutely safe as requested: "Once it might need SMS, must stop before sending"
		// We'll stop here if it looks like SMS is mandatory or a priority.
		if !strings.Contains(bodyStr, "auth/token") && !strings.Contains(bodyStr, "<Type>7</Type>") {
			log.Printf("PRE-EMPTIVE ABORT: Server indicates SMS authentication may be required. Aborting automatic login to avoid triggering SMS.")
			return ErrSMSRequired
		}
	}

	vpnMatch := regexp.MustCompile(`<VPNVERSION>(.*)</VPNVERSION>`).FindSubmatch(buf.Bytes())
	if vpnMatch != nil {
		log.Printf("VPN server version: %s", string(vpnMatch[1]))
	}

	twfidMatch := regexp.MustCompile(`<TwfID>(.*)</TwfID>`).FindSubmatch(buf.Bytes())
	if len(twfidMatch) < 2 {
		return errors.New("missing TwfID in server response (might be an unexpected page)")
	}
	c.twfID = string(twfidMatch[1])
	log.Printf("TWFID: %s", c.twfID)

	rsaKeyMatch := regexp.MustCompile(`<RSA_ENCRYPT_KEY>(.*)</RSA_ENCRYPT_KEY>`).FindSubmatch(buf.Bytes())
	if len(rsaKeyMatch) < 2 {
		return errors.New("missing RSA_ENCRYPT_KEY in server response (might be an unexpected page)")
	}
	rsaKey := string(rsaKeyMatch[1])
	log.Printf("RSA key: %s", rsaKey)

	rsaExpMatch := regexp.MustCompile(`<RSA_ENCRYPT_EXP>(.*)</RSA_ENCRYPT_EXP>`).FindSubmatch(buf.Bytes())
	rsaExp := ""
	if rsaExpMatch != nil {
		rsaExp = string(rsaExpMatch[1])
	} else {
		log.Printf("Warning: No RSA_ENCRYPT_EXP, using default")
		rsaExp = "65537"
	}
	log.Printf("RSA exp: %s", rsaExp)

	csrfMatch := regexp.MustCompile(`<CSRF_RAND_CODE>(.*)</CSRF_RAND_CODE>`).FindSubmatch(buf.Bytes())
	csrfCode := ""
	password := c.password
	if csrfMatch != nil {
		csrfCode = string(csrfMatch[1])
		log.Printf("CSRF Code: %s", csrfCode)
		password += "_" + csrfCode
	} else {
		log.Printf("Warning: No CSRF rand code")
	}

	pubKey := rsa.PublicKey{}
	pubKey.E, _ = strconv.Atoi(rsaExp)
	modulus := big.Int{}
	modulus.SetString(rsaKey, 16)
	pubKey.N = &modulus

	encryptedPassword, err := rsa.EncryptPKCS1v15(rand.Reader, &pubKey, []byte(password))
	if err != nil {
		return err
	}
	encryptedPasswordHex := hex.EncodeToString(encryptedPassword)

	addr = "https://" + c.server + "/por/login_psw.csp?anti_replay=1&encrypt=1&type=cs"
	log.Printf("Request: %s", addr)

	form := url.Values{
		"svpn_rand_code":    {""},
		"mitm":              {""},
		"svpn_req_randcode": {csrfCode},
		"svpn_name":         {c.username},
		"svpn_password":     {encryptedPasswordHex},
	}

	req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
	req.Header.Set("Cookie", "TWFID="+c.twfID)
	req.Header.Set("User-Agent", "EasyConnect_windows")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	buf.Reset()
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(buf.String(), "<NextService>auth/sms</NextService>") || strings.Contains(buf.String(), "<NextAuth>2</NextAuth>") {
		log.Print("SMS code required")

		return ErrSMSRequired
	}

	if strings.Contains(buf.String(), "<NextService>auth/token</NextService>") || strings.Contains(buf.String(), "<NextAuth>7</NextAuth>") {
		log.Print("TOTP required")

		return errTOTPRequired
	}

	if strings.Contains(buf.String(), "<NextAuth>0</NextAuth>") {
		log.Print("Cert required")

		return errCertRequired
	}

	if strings.Contains(buf.String(), "<NextAuth>-1</NextAuth>") || !strings.Contains(buf.String(), "<NextAuth>") {
		log.Print("No NextAuth found")
	} else {
		return errors.New("Not implemented auth: " + buf.String())
	}

	if !strings.Contains(buf.String(), "<Result>1</Result>") {
		return errors.New("Login failed: " + buf.String())
	}

	twfIDMatch := regexp.MustCompile(`<TwfID>(.*)</TwfID>`).FindSubmatch(buf.Bytes())
	if twfIDMatch != nil {
		c.twfID = string(twfIDMatch[1])
		c.authTimestamp = time.Now()
		log.Printf("Update TWFID: %s", c.twfID)
	}

	log.Printf("TWFID has been authorized")

	return nil
}

func (c *EasyConnectClient) loginSMS() error {
	addr := "https://" + c.server + "/por/login_sms.csp?apiversion=1"
	log.Printf("%s", "SMS request: "+addr)
	req, err := http.NewRequest("POST", addr, nil)
	req.Header.Set("Cookie", "TWFID="+c.twfID)
	req.Header.Set("User-Agent", "EasyConnect_windows")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(buf.String(), "验证码已发送到您的手机") && !strings.Contains(buf.String(), "<USER_PHONE>") {
		return errors.New("unexpected SMS response: " + buf.String())
	}

	log.Printf("SMS code is sent or still valid")

	fmt.Print("Please enter your SMS code:")
	smsCode := ""
	_, err = fmt.Scan(&smsCode)
	if err != nil {
		return err
	}

	addr = "https://" + c.server + "/por/login_sms1.csp?apiversion=1"
	log.Printf("%s", "SMS Request: "+addr)
	form := url.Values{
		"svpn_inputsms": {smsCode},
	}

	req, err = http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
	req.Header.Set("Cookie", "TWFID="+c.twfID)
	req.Header.Set("User-Agent", "EasyConnect_windows")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	buf.Reset()
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(buf.String(), "Auth sms suc") {
		debug.PrintStack()
		return errors.New("SMS code verification failed: " + buf.String())
	}

	twfidMatch := regexp.MustCompile(`<TwfID>(.*)</TwfID>`).FindSubmatch(buf.Bytes())
	if len(twfidMatch) < 2 {
		return errors.New("missing TwfID in SMS verification response")
	}
	c.twfID = string(twfidMatch[1])
	c.authTimestamp = time.Now()
	log.Print("SMS code verification success")

	return nil
}

func (c *EasyConnectClient) loginTOTP() error {
	var totpCode string
	var err error
	if c.totpSecret == "" {
		fmt.Print("Please enter your TOTP code:")
		_, err = fmt.Scan(&totpCode)
	} else {
		totpCode, err = totp.GenerateCode(c.totpSecret, time.Now())
		if err == nil {
			log.Println("Generate TOTP code:", totpCode)
		}
	}
	if err != nil {
		return err
	}

	addr := "https://" + c.server + "/por/login_token.csp"
	log.Printf("TOTP Request: %s", addr)
	form := url.Values{
		"svpn_inputtoken": {totpCode},
	}
	req, err := http.NewRequest("POST", addr, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", "TWFID="+c.twfID)
	req.Header.Set("User-Agent", "EasyConnect_windows")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(buf.String(), "Totp auth succ") {
		debug.PrintStack()
		return errors.New("TOTP verification failed: " + buf.String())
	}

	twfidMatch := regexp.MustCompile(`<TwfID>(.*)</TwfID>`).FindSubmatch(buf.Bytes())
	if len(twfidMatch) < 2 {
		return errors.New("missing TwfID in TOTP verification response")
	}
	c.twfID = string(twfidMatch[1])
	c.authTimestamp = time.Now()
	log.Print("TOTP verification success")

	return nil
}

func (c *EasyConnectClient) loginCert() error {
	addr := "https://" + c.server + "/com/server.crt"
	log.Printf("Get server cert: %s", addr)
	req, err := http.NewRequest("POST", addr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", "TWFID="+c.twfID)
	req.Header.Set("User-Agent", "EasyConnect_windows")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(buf.Bytes())
	if !ok {
		return errors.New("failed to parse server certificate")
	}

	c.httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Renegotiation:      tls.RenegotiateOnceAsClient,
			Certificates:       []tls.Certificate{c.tlsCert},
			RootCAs:            caCertPool,
		},
	}

	addr = "https://" + c.server + "/por/login_cert.csp?anti_replay=1&encrypt=1&type=cs"
	log.Printf("Cert Request: %s", addr)
	req, err = http.NewRequest("POST", addr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", "TWFID="+c.twfID)
	req.Header.Set("User-Agent", "EasyConnect_windows")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	buf.Reset()
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(buf.String(), "<Result>1</Result>") {
		debug.PrintStack()
		return errors.New("Cert verification failed: " + buf.String())
	}

	log.Print("Cert verification success")

	return nil
}

func (c *EasyConnectClient) requestConfig() (string, error) {
	addr := "https://" + c.server + "/por/conf.csp"
	log.Printf("Request: %s", addr)

	req, err := http.NewRequest("GET", addr, nil)
	req.Header.Set("Cookie", "TWFID="+c.twfID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *EasyConnectClient) requestResources() (string, error) {
	addr := "https://" + c.server + "/por/rclist.csp"
	log.Printf("Request: %s", addr)

	req, err := http.NewRequest("GET", addr, nil)
	req.Header.Set("Cookie", "TWFID="+c.twfID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *EasyConnectClient) requestToken() error {
	if c.twfID == "" {
		return errors.New("cannot request token with empty TWFID")
	}

	log.Printf("Requesting token with TWFID: %s", c.twfID)

	dialConn, err := net.Dial("tcp", c.server)
	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %w", c.server, err)
	}
	defer func(dialConn net.Conn) {
		_ = dialConn.Close()
	}(dialConn)

	conn := utls.UClient(dialConn, &utls.Config{InsecureSkipVerify: true}, utls.HelloGolang)
	defer func(conn *utls.UConn) {
		_ = conn.Close()
	}(conn)

	// When establish an HTTPS connection to server and send a valid request with TWFID to it
	// The **TLS ServerHello SessionId** is the first part of token
	log.Printf("ECAgent request: /por/conf.csp & /por/rclist.csp")
	_, err = io.WriteString(
		conn,
		"GET /por/conf.csp HTTP/1.1\r\nHost: "+c.server+
			"\r\nCookie: TWFID="+c.twfID+
			"\r\n\r\nGET /por/rclist.csp HTTP/1.1\r\nHost: "+c.server+
			"\r\nCookie: TWFID="+c.twfID+"\r\n\r\n",
	)
	if err != nil {
		return fmt.Errorf("failed to send ECAgent request: %w", err)
	}

	if conn.HandshakeState.ServerHello == nil {
		return errors.New("TLS handshake failed: no ServerHello received (TWFID may be invalid)")
	}

	sessionID := hex.EncodeToString(conn.HandshakeState.ServerHello.SessionId)
	log.Printf("Server session ID: %s", sessionID)

	if len(sessionID) < 31 {
		return fmt.Errorf("invalid session ID length: %d (expected >= 31)", len(sessionID))
	}

	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("ECAgent request failed: %w", err)
	}
	if n == 0 {
		return errors.New("ECAgent request invalid: no response from server (TWFID may have expired)")
	}

	tokenBytes := []byte(sessionID[:31] + "\x00" + c.twfID)
	if len(tokenBytes) < 48 {
		return fmt.Errorf("incorrect token length: %d (expected at least 48, twfid might be too short)", len(tokenBytes))
	}
	c.token = (*[48]byte)(tokenBytes)

	log.Printf("Token: %s", hex.EncodeToString(c.token[:]))

	return nil
}

func (c *EasyConnectClient) requestIP() error {
	conn, err := c.tlsConn()
	if err != nil {
		return err
	}

	// Request IP Packet
	message := []byte{0x00, 0x00, 0x00, 0x00}
	message = append(message, c.token[:]...)
	message = append(message, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff}...)

	n, err := conn.Write(message)
	if err != nil {
		_ = conn.Close()
		return err
	}

	log.DebugPrintf("Request IP: wrote %d bytes", n)
	log.DebugDumpHex(message[:n])

	reply := make([]byte, 0x80)
	n, err = conn.Read(reply)
	if err != nil {
		_ = conn.Close()
		return err
	}

	log.DebugPrintf("Request IP: read %d bytes", n)
	log.DebugDumpHex(reply[:n])

	if n < 8 {
		_ = conn.Close()
		log.Printf("Request IP reply too short: got %d bytes, need at least 8", n)
		if n > 0 {
			log.DumpHex(reply[:n])
		}
		return errors.New("request IP reply too short")
	}

	if reply[0] != 0x00 {
		_ = conn.Close()
		log.Printf("Unexpected request IP reply (first byte: 0x%02x, expected: 0x00)", reply[0])
		log.Printf("Full reply (%d bytes):", n)
		log.DumpHex(reply[:n])

		// Provide diagnostic hints based on the reply
		switch reply[0] {
		case 0x01, 0x08:
			log.Printf("Hint: Reply 0x%02x often indicates token expiration or invalid token", reply[0])
		case 0xff:
			log.Printf("Hint: Reply 0xff often indicates authentication failure or session timeout")
		}

		return errors.New("unexpected request IP reply")
	}

	c.ip = reply[4:8]
	c.ipReverse = []byte{c.ip[3], c.ip[2], c.ip[1], c.ip[0]}

	log.Printf("Client IP: %s", c.ip.String())

	// Close old connection and goroutine
	c.CloseIPConn()

	c.ipConn = conn
	ctx, cancel := context.WithCancel(context.Background())
	c.ipConnCancel = cancel

	// Request IP conn CAN NOT be closed, otherwise tx/rx handshake will fail
	go func(ctx context.Context, conn io.Closer) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 10):
				runtime.KeepAlive(conn)
			}
		}
	}(ctx, conn)

	// Auto-save session if sessionFile is configured
	c.sessionTimestamp = time.Now()
	if c.sessionFile != "" {
		if err := c.SaveSession(c.sessionFile); err != nil {
			log.Printf("Warning: failed to save session: %v", err)
		}
	}

	return nil
}

// CloseIPConn closes the IP holding connection and stops its maintenance goroutine
func (c *EasyConnectClient) CloseIPConn() {
	if c.ipConnCancel != nil {
		c.ipConnCancel()
		c.ipConnCancel = nil
	}
	if c.ipConn != nil {
		_ = c.ipConn.Close()
		c.ipConn = nil
	}
}

// KeepWebSessionAlive sends a dummy request to refresh the web session timeout
func (c *EasyConnectClient) KeepWebSessionAlive() error {
	addr := "https://" + c.server + "/por/conf.csp"
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", "TWFID="+c.twfID)
	req.Header.Set("User-Agent", "EasyConnect_windows")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("KeepWebSessionAlive failed with status: %d", resp.StatusCode)
	}

	log.DebugPrintf("KeepWebSessionAlive: OK")
	return nil
}

// RefreshToken attempts to refresh the token using the current twfID
// Returns nil on success, error otherwise
func (c *EasyConnectClient) RefreshToken() error {
	return c.requestToken()
}

// RefreshSession attempts to refresh both token and IP
// If forceFull is true, it will clear TWFID and perform a full re-login even if session is not old.
func (c *EasyConnectClient) RefreshSession(forceFull bool) error {
	c.refreshMutex.Lock()
	defer c.refreshMutex.Unlock()

	// Singleflight: If multiple goroutines trigger RefreshSession simultaneously,
	// only the first one executes it. Others will exit gracefully upon acquiring the lock.
	if time.Since(c.lastRefreshTime) < 15*time.Second {
		log.Printf("Session was completely refreshed within the last 15s. Skipping duplicate refresh to prevent overlapping handshake errors.")
		return nil
	}

	// Proactive full re-authentication if the session is old (e.g., > 24 hours)
	// or if we are explicitly told to force it.
	isProactive := forceFull || (!c.authTimestamp.IsZero() && time.Since(c.authTimestamp) > 24*time.Hour)
	var oldTwfID string

	if isProactive {
		if forceFull {
			log.Printf("Forced proactive session refresh triggered. Clearing session for fresh login...")
		} else {
			log.Printf("Master session (TWFID) is over 24h old (%v). Attempting proactive re-authentication...", time.Since(c.authTimestamp))
		}
		oldTwfID = c.twfID
		oldToken := c.token
		oldIP := c.ip
		oldIPReverse := c.ipReverse
		oldAuthTimestamp := c.authTimestamp

		// c.twfID = "" // Clear TWFID to force a fresh login in SetupAuto

		// Directly attempt a full setup for proactive refresh
		log.Printf("Attempting proactive re-setup...")
		if setupErr := c.SetupAuto(); setupErr != nil {
			log.Printf("Proactive re-authentication failed: %v. Reverting to old session.", setupErr)
			c.twfID = oldTwfID
			c.token = oldToken
			c.ip = oldIP
			c.ipReverse = oldIPReverse
			c.authTimestamp = oldAuthTimestamp

			// Re-save old session to restore it completely in the session file
			if c.sessionFile != "" {
				_ = c.SaveSession(c.sessionFile)
			}
			return setupErr
		}
		c.lastRefreshTime = time.Now()
		return nil
	}

	err := c.requestToken()
	if err != nil {
		log.Printf("Token refresh failed: %v. Attempting automatic re-setup...", err)
		// Try auto setup
		if setupErr := c.SetupAuto(); setupErr != nil {
			if errors.Is(setupErr, ErrSMSRequired) {
				log.Printf("CRITICAL: SMS verification required. Please restart the program and enter SMS code.")
				os.Exit(1)
			}
			return fmt.Errorf("auto setup failed: %w (original: %v)", setupErr, err)
		}
		c.lastRefreshTime = time.Now()
		return nil // SetupAuto already calls requestIP and saves session
	}

	err = c.requestIP()
	if err != nil {
		if isProactive {
			log.Printf("requestIP failed during proactive rotation: %v. Reverting to old session.", err)
			c.twfID = oldTwfID
			return nil
		}
		return err
	}

	c.lastRefreshTime = time.Now()
	return nil
}
