package weblogin

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerURL(t *testing.T) {
	for _, raw := range []string{"https://cloud.sdu.dk", "https://cloud.sdu.dk/", "http://localhost:1234", "http://127.0.0.1:1234", "http://[::1]:1234"} {
		if _, err := serverURL(raw); err != nil {
			t.Errorf("%s: %v", raw, err)
		}
	}
	for _, raw := range []string{"http://cloud.sdu.dk", "https://user:secret@cloud.sdu.dk", "https://cloud.sdu.dk/path", "https://cloud.sdu.dk?token=secret", "https://cloud.sdu.dk#fragment", "file:///tmp/test", ""} {
		if _, err := serverURL(raw); err == nil {
			t.Errorf("accepted unsafe server %q", raw)
		}
	}
}

func TestValidateSession(t *testing.T) {
	jwt := "header." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"test-user"}`)) + ".signature"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/auth/refresh" || r.Header.Get("Authorization") != "Bearer test-refresh-token" {
			t.Error("incorrect refresh request")
		}
		fmt.Fprintf(w, `{"accessToken":%q}`, jwt)
	}))
	defer server.Close()
	u, _ := serverURL(server.URL)
	session, err := validateSession(context.Background(), u, "test-refresh-token")
	if err != nil || session.Username != "test-user" || session.RefreshToken != "test-refresh-token" {
		t.Fatalf("unexpected session result: %v", err)
	}
}

func TestRejectInvalidSession(t *testing.T) {
	for _, response := range []struct {
		status int
		body   string
	}{
		{401, "secret-refresh-token"}, {403, "forbidden"}, {200, "invalid json"},
		{200, `{"accessToken":"invalid"}`}, {200, `{"accessToken":"a.e30.c"}`},
	} {
		t.Run(fmt.Sprint(response), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(response.status)
				fmt.Fprint(w, response.body)
			}))
			defer server.Close()
			u, _ := serverURL(server.URL)
			_, err := validateSession(context.Background(), u, "secret-refresh-token")
			if err == nil {
				t.Fatal("accepted invalid session")
			}
			if strings.Contains(err.Error(), "secret-refresh-token") {
				t.Fatal("error disclosed credential")
			}
		})
	}
}

func TestRefreshDoesNotFollowRedirect(t *testing.T) {
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("followed credential-bearing redirect") }))
	defer destination.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	u, _ := serverURL(server.URL)
	if _, err := validateSession(context.Background(), u, "secret"); err == nil {
		t.Fatal("accepted redirect")
	}
}

func TestRefreshCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	u, _ := serverURL("https://cloud.sdu.dk")
	if _, err := validateSession(ctx, u, "secret"); err == nil {
		t.Fatal("ignored cancellation")
	}
}
