package xuihttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/deposist/s-ui-x/util/ssrf"
)

// errBlockedRemoteAddress is returned by the guarded dialer when the remote
// x-ui panel resolves to an address disallowed by the active SSRF policy.
var errBlockedRemoteAddress = errors.New("xuihttp: remote address is not allowed")

// addrBlocked applies the SSRF policy to a single resolved address.
// Infrastructure addresses (cloud-metadata / link-local, multicast, unspecified)
// are rejected for everyone; loopback and private/LAN ranges are rejected only
// when restrictPrivate is set (an untrusted, token-scoped caller).
func addrBlocked(addr netip.Addr, restrictPrivate bool) bool {
	if restrictPrivate {
		return ssrf.IsBlockedAddr(addr)
	}
	return ssrf.IsInfrastructureAddr(addr)
}

// validateBaseURL is the pre-flight check run before any request is issued.
// A restricted (untrusted) caller gets the full private/loopback block-list; a
// trusted caller may target loopback/LAN/single-label hosts but never an
// IP-literal infrastructure address. Hostnames resolving to a blocked address
// are caught at dial time (defeating DNS-rebinding).
func validateBaseURL(ctx context.Context, baseURL string, restrictPrivate bool) error {
	if restrictPrivate {
		return ssrf.ValidateOutboundURL(ctx, baseURL, "http", "https")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	if scheme := strings.ToLower(parsed.Scheme); scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported url scheme")
	}
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("missing url host")
	}
	if ip, err := netip.ParseAddr(host); err == nil && ssrf.IsInfrastructureAddr(ip) {
		return fmt.Errorf("url host is not allowed")
	}
	return nil
}

// newGuardedClient builds an HTTP client that re-validates every dialed IP at
// connection time and bounds redirects, defending the importer against SSRF and
// DNS-rebinding. The block-list itself lives in util/ssrf; this only adds the
// dial-time plumbing and the trust policy.
func newGuardedClient(restrictPrivate bool) *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           guardedDialContext(dialer, restrictPrivate),
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   2,
	}
	return &http.Client{
		Timeout:   2 * time.Minute,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("xuihttp: too many redirects")
			}
			// The address policy for redirect targets is enforced by the
			// guarded dialer when the client connects; here we only refuse
			// scheme downgrades to non-http(s) targets.
			if scheme := req.URL.Scheme; scheme != "http" && scheme != "https" {
				return errors.New("xuihttp: refusing non-http redirect")
			}
			return nil
		},
	}
}

func guardedDialContext(dialer *net.Dialer, restrictPrivate bool) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		if ip, err := netip.ParseAddr(host); err == nil {
			if addrBlocked(ip, restrictPrivate) {
				return nil, errBlockedRemoteAddress
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ip.Unmap().String(), port))
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, ipAddr := range ips {
			ip, ok := netip.AddrFromSlice(ipAddr.IP)
			if !ok || addrBlocked(ip, restrictPrivate) {
				lastErr = errBlockedRemoteAddress
				continue
			}
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.Unmap().String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		if lastErr == nil {
			lastErr = errBlockedRemoteAddress
		}
		return nil, lastErr
	}
}
