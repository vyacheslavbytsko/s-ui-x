package api

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/deposist/s-ui-x/database/importxui"
	"github.com/deposist/s-ui-x/database/importxui/source/xuihttp"
	"github.com/gin-gonic/gin"
)

// TestRemoteImportIsUntrusted pins the S1 trust boundary: only a non-admin
// token scope is "untrusted"; a full admin session and an admin-scoped token
// stay trusted (and may reach loopback/LAN for legitimate migrations).
func TestRemoteImportIsUntrusted(t *testing.T) {
	a := &ApiService{}
	cases := []struct {
		name     string
		setScope func(c *gin.Context)
		want     bool
	}{
		{"session admin (no token scope)", func(c *gin.Context) {}, false},
		{"admin-scoped token", func(c *gin.Context) { c.Set(apiTokenScopeKey, "admin") }, false},
		{"xui_remote-scoped token", func(c *gin.Context) { c.Set(apiTokenScopeKey, "xui_remote") }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tc.setScope(c)
			if got := a.remoteImportIsUntrusted(c); got != tc.want {
				t.Fatalf("remoteImportIsUntrusted=%v want %v", got, tc.want)
			}
		})
	}
}

// TestApiSourceFromConfigPropagatesRestrictPrivate verifies the trust decision
// reaches the xuihttp source that actually enforces it.
func TestApiSourceFromConfigPropagatesRestrictPrivate(t *testing.T) {
	for _, restrict := range []bool{true, false} {
		src, err := apiSourceFromConfig(importxui.SyncProfileSource{
			Type:    "xuihttp",
			BaseURL: "http://panel.example.com",
		}, restrict)
		if err != nil {
			t.Fatalf("apiSourceFromConfig error: %v", err)
		}
		httpSrc, ok := src.(xuihttp.Source)
		if !ok {
			t.Fatalf("expected xuihttp.Source, got %T", src)
		}
		if httpSrc.RestrictPrivate != restrict {
			t.Fatalf("RestrictPrivate=%v want %v", httpSrc.RestrictPrivate, restrict)
		}
	}
}

// TestValidateRemoteSyncSourceSSRF covers the save-time guard that stops an
// untrusted token from storing a profile the cron job would later fetch.
func TestValidateRemoteSyncSourceSSRF(t *testing.T) {
	cases := []struct {
		name      string
		source    importxui.SyncProfileSource
		wantError bool
	}{
		{"private http rejected", importxui.SyncProfileSource{Type: "xuihttp", BaseURL: "http://10.0.0.5:2053"}, true},
		{"metadata rejected", importxui.SyncProfileSource{Type: "xuihttp", BaseURL: "http://169.254.169.254"}, true},
		{"file source ignored", importxui.SyncProfileSource{Type: "file", URL: "/tmp/x-ui.db"}, false},
		{"ssh source ignored", importxui.SyncProfileSource{Type: "ssh", URL: "ssh://host/x-ui.db"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRemoteSyncSourceSSRF(context.Background(), tc.source)
			if tc.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}
