package xuihttp

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPSourceAcquireSynthesizesSQLite(t *testing.T) {
	// A trusted caller (RestrictPrivate=false) may target loopback, so the
	// loopback httptest server works without any opt-in; only infrastructure
	// (cloud-metadata / link-local) addresses are blocked for trusted callers.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/panel/api/login":
			_, _ = w.Write([]byte(`{"success":true}`))
		case "/panel/api/inbounds/list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"obj":[{"id":1,"remark":"demo","enable":true,"port":443,"protocol":"vless","settings":"{\"clients\":[{\"email\":\"a@example.com\",\"id\":\"11111111-1111-1111-1111-111111111111\"}]}","streamSettings":"{}","sniffing":"{}"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	path, cleanup, err := New(server.URL, "admin", "secret").Acquire(context.Background())
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow("SELECT count(*) FROM client_traffics").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected one synthesized client_traffic row, got %d", count)
	}
}

// TestHTTPSourceAcquireBlocksPrivateForUntrusted asserts that a restricted
// (untrusted, token-scoped) caller is refused private, loopback and
// cloud-metadata targets before any request is issued (S1).
func TestHTTPSourceAcquireBlocksPrivateForUntrusted(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"loopback", "http://127.0.0.1:80/"},
		{"metadata", "http://169.254.169.254/latest/meta-data/"},
		{"private", "http://10.0.0.5:54321/"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path, cleanup, err := New(tc.url, "admin", "secret").WithRestrictPrivate(true).Acquire(context.Background())
			if cleanup != nil {
				defer cleanup()
			}
			if err == nil {
				t.Fatalf("expected Acquire to reject %s target, got path %q", tc.name, path)
			}
		})
	}
}

// TestHTTPSourceAcquireBlocksInfraEvenWhenTrusted asserts that cloud-metadata /
// link-local targets are refused even for a trusted caller (the always-on
// infrastructure block); loopback/LAN remain allowed for trusted callers
// (covered by TestHTTPSourceAcquireSynthesizesSQLite).
func TestHTTPSourceAcquireBlocksInfraEvenWhenTrusted(t *testing.T) {
	for _, target := range []string{
		"http://169.254.169.254/latest/meta-data/",
		"http://[fe80::1]/",
	} {
		path, cleanup, err := New(target, "admin", "secret").Acquire(context.Background())
		if cleanup != nil {
			defer cleanup()
		}
		if err == nil {
			t.Fatalf("expected trusted Acquire to still reject infrastructure target %s, got path %q", target, path)
		}
	}
}

// TestGuardedDialerBlocksLoopbackWhenRestricted exercises the dial-time guard
// (the DNS-rebinding defense) for an untrusted caller: even a loopback dial is
// rejected.
func TestGuardedDialerBlocksLoopbackWhenRestricted(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	transport, ok := newGuardedClient(true).Transport.(*http.Transport)
	if !ok || transport.DialContext == nil {
		t.Fatalf("expected configured *http.Transport with DialContext")
	}
	addr := strings.TrimPrefix(srv.URL, "http://")
	if _, err := transport.DialContext(context.Background(), "tcp", addr); !errors.Is(err, errBlockedRemoteAddress) {
		t.Fatalf("expected errBlockedRemoteAddress, got %v", err)
	}
}

// TestGuardedDialerAllowsLoopbackWhenTrusted verifies a trusted caller's dialer
// connects to loopback (RestrictPrivate off).
func TestGuardedDialerAllowsLoopbackWhenTrusted(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	transport, ok := newGuardedClient(false).Transport.(*http.Transport)
	if !ok || transport.DialContext == nil {
		t.Fatalf("expected configured *http.Transport with DialContext")
	}
	addr := strings.TrimPrefix(srv.URL, "http://")
	conn, err := transport.DialContext(context.Background(), "tcp", addr)
	if err != nil {
		t.Fatalf("expected dialer to allow loopback for trusted caller, got %v", err)
	}
	conn.Close()
}

// TestGuardedDialerBlocksInfraWhenTrusted verifies the always-on infrastructure
// block at dial time: a trusted caller's dialer still refuses a link-local /
// cloud-metadata address (defeats DNS-rebinding to metadata).
func TestGuardedDialerBlocksInfraWhenTrusted(t *testing.T) {
	transport, ok := newGuardedClient(false).Transport.(*http.Transport)
	if !ok || transport.DialContext == nil {
		t.Fatalf("expected configured *http.Transport with DialContext")
	}
	if _, err := transport.DialContext(context.Background(), "tcp", "169.254.169.254:80"); !errors.Is(err, errBlockedRemoteAddress) {
		t.Fatalf("expected errBlockedRemoteAddress for metadata dial, got %v", err)
	}
}
