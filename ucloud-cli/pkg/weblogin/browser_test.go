package weblogin

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Opt-in smoke test uses a local fake login, never a real user's credentials.
func TestBrowserLogin(t *testing.T) {
	if os.Getenv("UCLOUD_TEST_BROWSER") == "" {
		t.Skip("set UCLOUD_TEST_BROWSER to a Chrome/Chromium executable")
	}
	jwt := "header." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"browser-test"}`)) + ".signature"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/login":
			http.SetCookie(w, &http.Cookie{Name: "refreshToken", Value: "fake-browser-session", HttpOnly: true, Path: "/"})
			fmt.Fprint(w, "Local CLI login test completed.")
		case "/auth/refresh":
			if r.Header.Get("Authorization") != "Bearer fake-browser-session" {
				w.WriteHeader(401)
				return
			}
			fmt.Fprintf(w, `{"accessToken":%q}`, jwt)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	session, err := Login(ctx, server.URL, os.Getenv("UCLOUD_TEST_BROWSER"))
	if err != nil {
		t.Fatal(err)
	}
	if session.Username != "browser-test" || session.RefreshToken != "fake-browser-session" {
		t.Fatal("wrong browser session")
	}
}
