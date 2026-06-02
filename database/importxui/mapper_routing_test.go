package importxui

import (
	"testing"
)

// firstMappedRule runs MapXrayRouting over a single-rule config and returns the
// produced sing-box rule (or nil) plus counts.
func firstMappedRule(t *testing.T, raw string) (map[string]any, int, int, []string) {
	t.Helper()
	mapped, warnings, mappedCount, manualCount := MapXrayRouting(raw, map[string]string{"out": "direct"})
	route := mapped["route"].(map[string]any)
	rules := route["rules"].([]any)
	var rule map[string]any
	if len(rules) > 0 {
		rule = rules[0].(map[string]any)
	}
	return rule, mappedCount, manualCount, warnings
}

func TestRoutingMatchers_PortsNetworkProtocol(t *testing.T) {
	raw := `{"routing":{"rules":[{"outboundTag":"out","port":"443,1000-2000","network":"tcp,udp","protocol":["tls","http"]}]}}`
	rule, mapped, manual, _ := firstMappedRule(t, raw)
	if mapped != 1 || manual != 0 {
		t.Fatalf("mapped=%d manual=%d, want 1/0", mapped, manual)
	}
	ports, _ := rule["port"].([]int)
	if len(ports) != 1 || ports[0] != 443 {
		t.Errorf("port = %v, want [443]", rule["port"])
	}
	pr := rule["port_range"].([]string)
	if len(pr) != 1 || pr[0] != "1000:2000" {
		t.Errorf("port_range = %v, want [1000:2000]", rule["port_range"])
	}
	nets := rule["network"].([]string)
	if len(nets) != 2 || nets[0] != "tcp" || nets[1] != "udp" {
		t.Errorf("network = %v, want [tcp udp]", rule["network"])
	}
	protos := rule["protocol"].([]string)
	if len(protos) != 2 || protos[0] != "tls" {
		t.Errorf("protocol = %v", rule["protocol"])
	}
	if _, ok := rule["outbound"]; !ok {
		t.Errorf("rule missing outbound: %#v", rule)
	}
}

func TestRoutingMatchers_SourceInboundUser(t *testing.T) {
	raw := `{"routing":{"rules":[{"outboundTag":"out","source":["geoip:private","10.0.0.0/8"],"sourcePort":"50000-60000","inboundTag":["in-1","in-2"],"user":["alice@example.com"]}]}}`
	rule, mapped, manual, _ := firstMappedRule(t, raw)
	if mapped != 1 || manual != 0 {
		t.Fatalf("mapped=%d manual=%d, want 1/0", mapped, manual)
	}
	if sg := rule["source_geoip"].([]string); len(sg) != 1 || sg[0] != "private" {
		t.Errorf("source_geoip = %v", rule["source_geoip"])
	}
	if sc := rule["source_ip_cidr"].([]string); len(sc) != 1 || sc[0] != "10.0.0.0/8" {
		t.Errorf("source_ip_cidr = %v", rule["source_ip_cidr"])
	}
	if spr := rule["source_port_range"].([]string); len(spr) != 1 || spr[0] != "50000:60000" {
		t.Errorf("source_port_range = %v", rule["source_port_range"])
	}
	if inb := rule["inbound"].([]string); len(inb) != 2 || inb[0] != "in-1" {
		t.Errorf("inbound = %v", rule["inbound"])
	}
	if au := rule["auth_user"].([]string); len(au) != 1 || au[0] != "alice@example.com" {
		t.Errorf("auth_user = %v", rule["auth_user"])
	}
}

func TestRoutingMatchers_DomainPrefixes(t *testing.T) {
	raw := `{"routing":{"rules":[{"outboundTag":"out","domain":["full:exact.com","domain:sub.com","keyword:ads","regexp:.*\\.evil\\.com","bare.com"]}]}}`
	rule, mapped, manual, _ := firstMappedRule(t, raw)
	if mapped != 1 || manual != 0 {
		t.Fatalf("mapped=%d manual=%d, want 1/0", mapped, manual)
	}
	if d := rule["domain"].([]string); len(d) != 1 || d[0] != "exact.com" {
		t.Errorf("domain = %v, want [exact.com]", rule["domain"])
	}
	ds := rule["domain_suffix"].([]string)
	if len(ds) != 2 {
		t.Errorf("domain_suffix = %v, want [sub.com bare.com]", ds)
	}
	if dk := rule["domain_keyword"].([]string); len(dk) != 1 || dk[0] != "ads" {
		t.Errorf("domain_keyword = %v", rule["domain_keyword"])
	}
	if dr := rule["domain_regex"].([]string); len(dr) != 1 {
		t.Errorf("domain_regex = %v", rule["domain_regex"])
	}
}

func TestRoutingMatchers_AttrsAndNoMatcherAreManual(t *testing.T) {
	// attrs -> whole rule manual (cannot represent in sing-box).
	rawAttrs := `{"routing":{"rules":[{"outboundTag":"out","attrs":{":method":"GET"},"domain":["full:x.com"]}]}}`
	_, mapped, manual, warnings := firstMappedRule(t, rawAttrs)
	if mapped != 0 || manual != 1 {
		t.Fatalf("attrs rule: mapped=%d manual=%d, want 0/1", mapped, manual)
	}
	if len(warnings) == 0 {
		t.Error("attrs rule should warn")
	}

	// outboundTag resolvable but no matchers -> manual (cannot be a sing-box rule).
	rawEmpty := `{"routing":{"rules":[{"outboundTag":"out"}]}}`
	_, mapped2, manual2, _ := firstMappedRule(t, rawEmpty)
	if mapped2 != 0 || manual2 != 1 {
		t.Fatalf("empty-matcher rule: mapped=%d manual=%d, want 0/1", mapped2, manual2)
	}
}
