/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestClient points a Client at a test server so we exercise the HTTP path
// without needing real Proxmox. `fn` is the request handler.
func newTestClient(t *testing.T, p *Profile, fn http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(fn)
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	p.Host = u.Hostname()
	// httptest picks an ephemeral port — feed it into the profile.
	var port int
	if _, err := jsonUnmarshalInt(u.Port(), &port); err != nil {
		t.Fatalf("parse port %q: %v", u.Port(), err)
	}
	p.Port = port
	p.VerifyTLS = false // NewTLSServer uses a self-signed cert

	c, err := NewClient(p)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

// jsonUnmarshalInt parses an integer from a string using the JSON decoder —
// small helper to avoid pulling in strconv just for tests.
func jsonUnmarshalInt(s string, out *int) (int, error) {
	if err := json.Unmarshal([]byte(s), out); err != nil {
		return 0, err
	}
	return *out, nil
}

func TestClient_Version_Token(t *testing.T) {
	profile := &Profile{
		AuthType:    AuthTypeToken,
		TokenID:     "user@pam!ci",
		TokenSecret: "deadbeef",
	}
	c, _ := newTestClient(t, profile, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api2/json/version" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		want := "PVEAPIToken=user@pam!ci=deadbeef"
		if auth != want {
			t.Errorf("Authorization header: got %q, want %q", auth, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"version":"8.2.4","release":"8","repoid":"abcdef"}}`))
	})

	info, err := c.Version()
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if info.Version != "8.2.4" || info.Release != "8" || info.RepoID != "abcdef" {
		t.Errorf("unexpected VersionInfo: %+v", info)
	}
}

func TestClient_Version_Unauthorized(t *testing.T) {
	profile := &Profile{
		AuthType:    AuthTypeToken,
		TokenID:     "u",
		TokenSecret: "bad",
	}
	c, _ := newTestClient(t, profile, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusUnauthorized)
	})
	_, err := c.Version()
	if err == nil || !strings.Contains(err.Error(), "HTTP 401") {
		t.Fatalf("expected 401 error, got %v", err)
	}
}

func TestClient_MissingToken(t *testing.T) {
	profile := &Profile{AuthType: AuthTypeToken}
	// NewClient itself succeeds; the missing token surfaces on the first call.
	c, err := NewClient(profile)
	if err != nil {
		// Some fields are validated later — but Host is required at construction.
		// Our Profile has no Host, so NewClient should reject it.
		if !strings.Contains(err.Error(), "empty host") {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	_, err = c.Version()
	if err == nil || !strings.Contains(err.Error(), "token_id") {
		t.Fatalf("expected missing-token error, got %v", err)
	}
}

func TestClient_PasswordLogin(t *testing.T) {
	profile := &Profile{
		AuthType: AuthTypePassword,
		User:     "root@pam",
	}
	c, _ := newTestClient(t, profile, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api2/json/access/ticket":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm: %v", err)
			}
			if r.PostForm.Get("username") != "root@pam" {
				t.Errorf("unexpected username %q", r.PostForm.Get("username"))
			}
			if r.PostForm.Get("password") != "hunter2" {
				t.Errorf("unexpected password %q", r.PostForm.Get("password"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"ticket":"PVE:root@pam:ABC","CSRFPreventionToken":"CSRF","username":"root@pam"}}`))
		case "/api2/json/version":
			// Ticket must travel in a cookie; CSRF header only required for writes.
			c, err := r.Cookie("PVEAuthCookie")
			if err != nil || c.Value != "PVE:root@pam:ABC" {
				t.Errorf("missing or wrong auth cookie: %+v / err=%v", c, err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"version":"8.2.4","release":"8","repoid":"r"}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})

	if err := c.Login("hunter2"); err != nil {
		t.Fatalf("Login: %v", err)
	}
	if _, err := c.Version(); err != nil {
		t.Fatalf("Version after Login: %v", err)
	}
}

func TestClient_PasswordWithoutLogin(t *testing.T) {
	profile := &Profile{AuthType: AuthTypePassword, User: "u", Host: "example.invalid", Port: 8006}
	c, err := NewClient(profile)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.Version()
	if err == nil || !strings.Contains(err.Error(), "Login") {
		t.Fatalf("expected 'call Login() first' error, got %v", err)
	}
}

func TestNewClient_NilProfile(t *testing.T) {
	if _, err := NewClient(nil); err == nil {
		t.Fatal("expected error for nil profile")
	}
}

func TestNewClient_EmptyHost(t *testing.T) {
	if _, err := NewClient(&Profile{}); err == nil {
		t.Fatal("expected error for empty host")
	}
}
