/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// defaultTimeout caps every outgoing request. Proxmox usually responds in a
// few hundred ms; 30s protects against network hangs without being painful
// for manual CLI use.
const defaultTimeout = 30 * time.Second

// Client issues authenticated requests against a single Proxmox VE endpoint.
// It is intentionally small in Phase 1 — enough to validate credentials via
// /version and return JSON payloads. Phase 2+ will add VM/CT operations.
type Client struct {
	profile *Profile
	http    *http.Client
	baseURL string

	// ticket + csrfToken are populated for password-based auth by calling
	// Login(). Token-based profiles leave these empty.
	ticket    string
	csrfToken string
}

// NewClient builds a Client for the given profile. It configures TLS
// verification according to Profile.VerifyTLS but does not perform any network
// I/O — call Version() or Login() to validate.
func NewClient(p *Profile) (*Client, error) {
	if p == nil {
		return nil, fmt.Errorf("nil profile")
	}
	if p.Host == "" {
		return nil, fmt.Errorf("profile has empty host")
	}
	port := p.Port
	if port == 0 {
		port = 8006
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !p.VerifyTLS}, //nolint:gosec // controlled by user flag
	}
	return &Client{
		profile: p,
		http:    &http.Client{Timeout: defaultTimeout, Transport: transport},
		baseURL: fmt.Sprintf("https://%s:%d/api2/json", p.Host, port),
	}, nil
}

// BaseURL returns the computed API base (useful for diagnostics and tests).
func (c *Client) BaseURL() string { return c.baseURL }

// Login acquires a session ticket using username+password. It is only needed
// for AuthTypePassword profiles; token-based profiles can skip it. The ticket
// and CSRF token are stored on the Client for subsequent authenticated calls.
func (c *Client) Login(password string) error {
	if c.profile.AuthType != AuthTypePassword {
		return fmt.Errorf("login() only applies to password-based profiles")
	}
	if c.profile.User == "" {
		return fmt.Errorf("profile has empty user — cannot log in")
	}
	form := url.Values{}
	form.Set("username", c.profile.User)
	form.Set("password", password)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/access/ticket", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call /access/ticket: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var env struct {
		Data struct {
			Ticket              string `json:"ticket"`
			CSRFPreventionToken string `json:"CSRFPreventionToken"`
			Username            string `json:"username"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return fmt.Errorf("decode login response: %w", err)
	}
	if env.Data.Ticket == "" {
		return fmt.Errorf("login response missing ticket")
	}
	c.ticket = env.Data.Ticket
	c.csrfToken = env.Data.CSRFPreventionToken
	return nil
}

// Version calls GET /version to retrieve server version and simultaneously
// validate that the configured credentials work. This is the primary check
// used by `proxmox auth status`.
func (c *Client) Version() (*VersionInfo, error) {
	var v VersionInfo
	if err := c.get("/version", &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// get issues an authenticated GET and unmarshals the `data` envelope into out.
func (c *Client) get(path string, out interface{}) error {
	return c.do(http.MethodGet, path, nil, out)
}

// postForm issues a POST with application/x-www-form-urlencoded body. Proxmox
// API uses form encoding for write operations (create, snapshot, config).
// Returns the decoded `data` value in out. For endpoints that return only a
// UPID (task ID) string, pass `*string` as out.
func (c *Client) postForm(path string, form url.Values, out interface{}) error {
	return c.do(http.MethodPost, path, form, out)
}

// putForm issues a PUT with form-encoded body (used by cloud-init / config
// updates).
func (c *Client) putForm(path string, form url.Values, out interface{}) error {
	return c.do(http.MethodPut, path, form, out)
}

// delete removes a resource. Proxmox returns a UPID for async operations like
// VM destruction.
func (c *Client) delete(path string, out interface{}) error {
	return c.do(http.MethodDelete, path, nil, out)
}

// do is the single HTTP primitive used by every verb. Passing a non-nil form
// automatically sets Content-Type. out may be nil to ignore the response body.
func (c *Client) do(method, path string, form url.Values, out interface{}) error {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if err := c.attachAuth(req); err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", path, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed (HTTP 401) — check token or re-run `proxmox auth login`")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s: HTTP %d: %s", method, path, resp.StatusCode,
			strings.TrimSpace(string(respBody)))
	}
	if out == nil {
		return nil
	}
	env := apiResponse{Data: out}
	if err := json.Unmarshal(respBody, &env); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

// WaitTask polls task status until finished or the context is cancelled. UPID
// values are returned by most write endpoints (e.g., status/start). Returns
// the final task status string ("OK" on success, error message otherwise).
func (c *Client) WaitTask(ctx context.Context, node, upid string) error {
	const pollInterval = 1 * time.Second
	path := fmt.Sprintf("/nodes/%s/tasks/%s/status", url.PathEscape(node), url.PathEscape(upid))
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait cancelled: %w", ctx.Err())
		default:
		}
		var st struct {
			Status     string `json:"status"`
			ExitStatus string `json:"exitstatus"`
		}
		if err := c.get(path, &st); err != nil {
			return err
		}
		if st.Status == "stopped" {
			if st.ExitStatus != "" && st.ExitStatus != "OK" {
				return fmt.Errorf("task %s failed: %s", upid, st.ExitStatus)
			}
			return nil
		}
		time.Sleep(pollInterval)
	}
}

// attachAuth adds the proper auth headers/cookies to req based on profile
// AuthType. Token auth requires no prior call; password auth needs Login() to
// have run first.
func (c *Client) attachAuth(req *http.Request) error {
	switch c.profile.AuthType {
	case AuthTypeToken:
		if c.profile.TokenID == "" || c.profile.TokenSecret == "" {
			return fmt.Errorf("profile is missing token_id or token_secret")
		}
		req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s",
			c.profile.TokenID, c.profile.TokenSecret))
	case AuthTypePassword:
		if c.ticket == "" {
			return fmt.Errorf("no active session — call Login() first")
		}
		req.AddCookie(&http.Cookie{Name: "PVEAuthCookie", Value: c.ticket})
		if req.Method != http.MethodGet && c.csrfToken != "" {
			req.Header.Set("CSRFPreventionToken", c.csrfToken)
		}
	default:
		return fmt.Errorf("unknown auth type %q", c.profile.AuthType)
	}
	return nil
}
