package importxui

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// mapXrayDNS translates an Xray `dns` object into a sing-box DNS config. Xray
// uses a flat server list (plain addresses + domain-scoped objects) while
// sing-box (1.13) uses typed servers with tags, domain-scoped DNS rules, a
// final server and a query strategy. Anything without a clean equivalent
// (hosts, fakedns) is reported rather than emitted as an invalid block.
//
// ruleSets/seen are the shared route rule_set accumulators so a geosite used in
// a DNS rule registers a remote rule set exactly once.
func mapXrayDNS(rawDNS map[string]any, ruleSets *[]any, seen map[string]struct{}) (map[string]any, []string) {
	out := map[string]any{}
	var warnings []string
	var servers []any
	var rules []any
	tagIndex := 0
	nextTag := func() string {
		t := fmt.Sprintf("dns-%d", tagIndex)
		tagIndex++
		return t
	}
	finalTag := ""

	serversRaw, _ := rawDNS["servers"].([]any)
	for _, s := range serversRaw {
		switch v := s.(type) {
		case string:
			srv, ok, w := dnsServerFromAddress(v, nextTag)
			warnings = append(warnings, w...)
			if ok {
				servers = append(servers, srv)
				if finalTag == "" {
					finalTag = srv["tag"].(string)
				}
			}
		case map[string]any:
			addr, _ := v["address"].(string)
			srv, ok, w := dnsServerFromAddress(addr, nextTag)
			warnings = append(warnings, w...)
			if !ok {
				continue
			}
			servers = append(servers, srv)
			tag := srv["tag"].(string)
			rule := map[string]any{"action": "route", "server": tag}
			matched, _ := mapDomainMatchers(stringList(v["domains"]), rule, ruleSets, seen)
			if matched {
				rules = append(rules, rule)
			} else if finalTag == "" {
				// A server with no domain scope acts as a default resolver.
				finalTag = tag
			}
		}
	}

	if hosts, ok := rawDNS["hosts"].(map[string]any); ok && len(hosts) > 0 {
		warnings = append(warnings, fmt.Sprintf("dns: %d host override(s) were not migrated; add them as a sing-box \"hosts\" DNS server after import", len(hosts)))
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(rawDNS["queryStrategy"]))) {
	case "useipv4":
		out["strategy"] = "ipv4_only"
	case "useipv6":
		out["strategy"] = "ipv6_only"
	}
	if clientIP, ok := rawDNS["clientIp"].(string); ok && strings.TrimSpace(clientIP) != "" {
		if ip := strings.TrimSpace(clientIP); net.ParseIP(ip) != nil {
			out["client_subnet"] = ip
		} else {
			warnings = append(warnings, fmt.Sprintf("dns: clientIp %q is not a valid IP address; not migrated", ip))
		}
	}

	// Ensure a deterministic default resolver: if every server was domain-scoped
	// (so no bare/default server set finalTag), fall back to the first server so
	// unmapped-domain queries do not resolve to an implementation-defined server.
	if finalTag == "" && len(servers) > 0 {
		finalTag = servers[0].(map[string]any)["tag"].(string)
	}

	if len(servers) > 0 {
		out["servers"] = servers
	}
	if len(rules) > 0 {
		out["rules"] = rules
	}
	if finalTag != "" {
		out["final"] = finalTag
	}
	return out, warnings
}

// dnsServerFromAddress turns one Xray DNS server address into a typed sing-box
// DNS server object. It recognises the scheme prefixes Xray accepts; a bare
// address becomes a plain UDP server.
func dnsServerFromAddress(addr string, nextTag func() string) (map[string]any, bool, []string) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, false, nil
	}
	lower := strings.ToLower(addr)
	switch {
	case lower == "localhost" || lower == "local":
		return map[string]any{"type": "local", "tag": nextTag()}, true, nil
	case lower == "fakedns" || lower == "fakeip":
		return nil, false, []string{"dns: fakedns server not migrated; configure a sing-box fakeip server manually if needed"}
	case strings.HasPrefix(lower, "https://"), strings.HasPrefix(lower, "h3://"):
		typ := "https"
		raw := addr
		if strings.HasPrefix(lower, "h3://") {
			typ = "h3"
			raw = "https://" + addr[len("h3://"):]
		}
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" {
			return nil, false, []string{fmt.Sprintf("dns: invalid server %q; skipped", addr)}
		}
		srv := map[string]any{"type": typ, "tag": nextTag(), "server": u.Hostname()}
		var warnings []string
		if p := u.Port(); p != "" {
			if n, err := strconv.Atoi(p); err == nil && validDNSPort(n) {
				srv["server_port"] = n
			} else if err == nil {
				warnings = append(warnings, fmt.Sprintf("dns: server %q has out-of-range port %d; using the default port", addr, n))
			}
		}
		if u.Path != "" && u.Path != "/" {
			srv["path"] = u.Path
		}
		return srv, true, warnings
	case strings.HasPrefix(lower, "tls://"):
		srv, w := schemeServer("tls", addr[len("tls://"):], nextTag)
		return srv, true, w
	case strings.HasPrefix(lower, "quic://"):
		srv, w := schemeServer("quic", addr[len("quic://"):], nextTag)
		return srv, true, w
	case strings.HasPrefix(lower, "tcp://"):
		srv, w := schemeServer("tcp", addr[len("tcp://"):], nextTag)
		return srv, true, w
	case strings.HasPrefix(lower, "udp://"):
		srv, w := schemeServer("udp", addr[len("udp://"):], nextTag)
		return srv, true, w
	default:
		srv, w := schemeServer("udp", addr, nextTag)
		return srv, true, w
	}
}

// schemeServer builds a typed remote DNS server, splitting an optional :port. An
// out-of-range port is dropped with a warning rather than emitted — sing-box's
// server_port is a uint16 and would reject the whole config on overflow.
func schemeServer(typ, hostPort string, nextTag func() string) (map[string]any, []string) {
	host, port := splitHostPortLoose(strings.TrimSpace(hostPort))
	srv := map[string]any{"type": typ, "tag": nextTag(), "server": host}
	var warnings []string
	if port != 0 {
		if validDNSPort(port) {
			srv["server_port"] = port
		} else {
			warnings = append(warnings, fmt.Sprintf("dns: server %q has out-of-range port %d; using the default port", hostPort, port))
		}
	}
	return srv, warnings
}

func validDNSPort(n int) bool {
	return n > 0 && n <= 65535
}

// splitHostPortLoose splits "host:port" (or "[ipv6]:port"); a value with no
// parseable port yields the whole string as host and port 0.
func splitHostPortLoose(value string) (string, int) {
	value = strings.TrimSpace(strings.TrimSuffix(value, "/"))
	if value == "" {
		return "", 0
	}
	if host, portStr, err := net.SplitHostPort(value); err == nil {
		port, _ := strconv.Atoi(portStr)
		return strings.Trim(host, "[]"), port
	}
	return strings.Trim(value, "[]"), 0
}
