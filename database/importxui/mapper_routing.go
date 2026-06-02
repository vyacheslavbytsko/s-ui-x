package importxui

import (
	"fmt"
	"strconv"
	"strings"
)

// applyRuleMatchers translates an Xray routing rule's matcher fields into
// sing-box route-rule fields on next. It returns whether at least one matcher
// was added (a rule with none cannot be a sing-box rule) and warnings for
// fields that can only be partially represented.
//
// Caller handles the un-representable matchers (attrs, balancerTag) by marking
// the whole rule manual before calling this — dropping them would silently
// broaden the match.
func applyRuleMatchers(index int, rule, next map[string]any, ruleSets *[]any, seen map[string]struct{}) (bool, []string) {
	added := false
	var warnings []string

	if domains := stringList(rule["domain"]); len(domains) > 0 {
		matched, unknown := mapDomainMatchers(domains, next, ruleSets, seen)
		if matched {
			added = true
		}
		for _, d := range unknown {
			warnings = append(warnings, fmt.Sprintf("routing rule %d domain %q has an unsupported prefix; that entry was dropped", index, d))
		}
	}

	if mapIPMatchers(rule["ip"], next, "geoip", "ip_cidr") {
		added = true
	}
	if mapIPMatchers(rule["source"], next, "source_geoip", "source_ip_cidr") {
		added = true
	}
	if mapPortMatchers(rule["port"], next, "port", "port_range") {
		added = true
	}
	if mapPortMatchers(rule["sourcePort"], next, "source_port", "source_port_range") {
		added = true
	}
	if nets := splitCSVList(rule["network"]); len(nets) > 0 {
		next["network"] = nets
		added = true
	}
	if protos := stringList(rule["protocol"]); len(protos) > 0 {
		next["protocol"] = protos
		added = true
	}
	if inbounds := stringList(rule["inboundTag"]); len(inbounds) > 0 {
		next["inbound"] = inbounds
		added = true
	}
	if users := stringList(rule["user"]); len(users) > 0 {
		next["auth_user"] = users
		added = true
	}
	return added, warnings
}

// mapDomainMatchers translates Xray domain matcher entries (geosite:/domain:/
// full:/regexp:/keyword:/bare) into sing-box domain fields on dst (shared by
// route rules and DNS rules). It returns whether any matcher was added and any
// entries with an unsupported prefix (e.g. ext:) that were dropped.
func mapDomainMatchers(domains []string, dst map[string]any, ruleSets *[]any, seen map[string]struct{}) (bool, []string) {
	added := false
	var unknown []string
	for _, d := range domains {
		switch {
		case strings.HasPrefix(d, "geosite:"):
			name := strings.ReplaceAll(d, ":", "-")
			dst["rule_set"] = appendString(dst["rule_set"], name)
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				*ruleSets = append(*ruleSets, map[string]any{"tag": name, "type": "remote", "format": "binary"})
			}
			added = true
		case strings.HasPrefix(d, "full:"):
			dst["domain"] = appendString(dst["domain"], strings.TrimPrefix(d, "full:"))
			added = true
		case strings.HasPrefix(d, "regexp:"):
			dst["domain_regex"] = appendString(dst["domain_regex"], strings.TrimPrefix(d, "regexp:"))
			added = true
		case strings.HasPrefix(d, "keyword:"):
			dst["domain_keyword"] = appendString(dst["domain_keyword"], strings.TrimPrefix(d, "keyword:"))
			added = true
		case strings.HasPrefix(d, "domain:"):
			dst["domain_suffix"] = appendString(dst["domain_suffix"], strings.TrimPrefix(d, "domain:"))
			added = true
		case strings.Contains(d, ":"):
			unknown = append(unknown, d)
		default:
			// A bare domain is matched by Xray as a domain (and its subdomains).
			dst["domain_suffix"] = appendString(dst["domain_suffix"], d)
			added = true
		}
	}
	return added, unknown
}

// mapIPMatchers splits Xray ip/source entries into a geoip field and a CIDR
// field on next. Returns whether anything was added.
func mapIPMatchers(value any, next map[string]any, geoipKey, cidrKey string) bool {
	added := false
	for _, ip := range stringList(value) {
		if strings.HasPrefix(ip, "geoip:") {
			next[geoipKey] = appendString(next[geoipKey], strings.TrimPrefix(ip, "geoip:"))
		} else {
			next[cidrKey] = appendString(next[cidrKey], ip)
		}
		added = true
	}
	return added
}

// mapPortMatchers splits Xray port specs (a number, an "a-b" range, or a list/
// comma string of either) into a sing-box single-port list and a port-range
// list ("a:b"). Returns whether anything was added.
func mapPortMatchers(value any, next map[string]any, portKey, rangeKey string) bool {
	added := false
	for _, token := range portTokens(value) {
		if lo, hi, ok := splitPortRange(token); ok {
			next[rangeKey] = appendString(next[rangeKey], fmt.Sprintf("%d:%d", lo, hi))
			added = true
			continue
		}
		if p, err := strconv.Atoi(token); err == nil && p >= 0 && p <= 65535 {
			next[portKey] = appendInt(next[portKey], p)
			added = true
		}
	}
	return added
}

// portTokens flattens an Xray port value (number, string, comma list, or array)
// into individual port/range tokens.
func portTokens(value any) []string {
	var out []string
	add := func(s string) {
		for _, part := range strings.Split(s, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	switch v := value.(type) {
	case nil:
		return nil
	case []any:
		for _, item := range v {
			add(strings.TrimSpace(fmt.Sprint(item)))
		}
	case string:
		add(v)
	default:
		add(strings.TrimSpace(fmt.Sprint(v)))
	}
	return out
}

// splitPortRange parses "lo-hi" into its bounds.
func splitPortRange(token string) (int, int, bool) {
	i := strings.IndexByte(token, '-')
	if i <= 0 || i >= len(token)-1 {
		return 0, 0, false
	}
	lo, err1 := strconv.Atoi(strings.TrimSpace(token[:i]))
	hi, err2 := strconv.Atoi(strings.TrimSpace(token[i+1:]))
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return lo, hi, true
}

// splitCSVList flattens a value that may be a comma string ("tcp,udp"), a single
// string, or an array into a trimmed list.
func splitCSVList(value any) []string {
	var out []string
	for _, s := range stringList(value) {
		for _, part := range strings.Split(s, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// appendInt appends an int to a value that is either nil or an existing []int.
func appendInt(value any, item int) []int {
	if existing, ok := value.([]int); ok {
		return append(existing, item)
	}
	return []int{item}
}
