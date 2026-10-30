// Package weblogin establishes a CLI session through UCloud's normal web login.
package weblogin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Session struct {
	Username     string
	RefreshToken string
}

func serverURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, fmt.Errorf("server must be an HTTPS origin without credentials, path, query, or fragment")
	}
	loopback := u.Hostname() == "localhost"
	if ip := net.ParseIP(u.Hostname()); ip != nil {
		loopback = ip.IsLoopback()
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && loopback) {
		return nil, fmt.Errorf("server must use HTTPS (HTTP is allowed only on loopback)")
	}
	u.Path = ""
	return u, nil
}

func validateSession(ctx context.Context, server *url.URL, token string) (Session, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, server.String()+"/auth/refresh", nil)
	if err != nil {
		return Session{}, fmt.Errorf("cannot construct refresh request")
	}
	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{
		Timeout:       15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	response, err := client.Do(request)
	if err != nil {
		return Session{}, fmt.Errorf("could not validate session with UCloud")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Session{}, fmt.Errorf("UCloud rejected CLI token refresh (HTTP %d); session not saved", response.StatusCode)
	}
	var body struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&body); err != nil {
		return Session{}, fmt.Errorf("invalid token refresh response")
	}
	// This JWT came directly from the trusted server over TLS. Decode only its
	// subject for display/configuration; this is not local signature verification.
	parts := strings.Split(body.AccessToken, ".")
	if len(parts) != 3 {
		return Session{}, fmt.Errorf("refresh response did not contain a JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	var claims struct {
		Subject string `json:"sub"`
	}
	if err != nil || json.Unmarshal(payload, &claims) != nil || claims.Subject == "" {
		return Session{}, fmt.Errorf("refresh response did not identify the user")
	}
	return Session{Username: claims.Subject, RefreshToken: token}, nil
}
