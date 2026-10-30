package shared

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ucloud.dk/shared/pkg/rpc"
)

func TestCredentialRequestsDoNotFollowRedirects(t *testing.T) {
	previous := rpc.DefaultClient
	defer func() { rpc.DefaultClient = previous }()
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("followed redirect with session credentials")
	}))
	defer destination.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	cfg := &Config{DefaultEnvironment: "test", Environments: map[string]Environment{"test": {URL: server.URL}}}
	cfg.InitUCloudClient()
	req, err := http.NewRequest(http.MethodPost, server.URL+"/auth/refresh", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-secret")
	resp, err := rpc.DefaultClient.Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestWebSessionServerBinding(t *testing.T) {
	previous := rpc.DefaultClient
	defer func() { rpc.DefaultClient = previous }()
	cfg := &Config{TokenRef: "secret", TokenServer: "https://cloud.sdu.dk", DefaultEnvironment: "test", Environments: map[string]Environment{"test": {URL: "https://other.example"}}}
	cfg.InitUCloudClient()
	if rpc.DefaultClient.RefreshToken != "" {
		t.Fatal("credential would be sent to a different server")
	}
	cfg.Environments["test"] = Environment{URL: "https://cloud.sdu.dk/"}
	cfg.InitUCloudClient()
	if rpc.DefaultClient.RefreshToken != "secret" {
		t.Fatal("credential missing for its own server")
	}
}

func TestSessionConfigPermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.TokenRef = "secret"
	cfg.TokenServer = "https://cloud.sdu.dk"
	if _, err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(GetConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("credential permissions: %v", info.Mode())
	}
	loaded, err := ReadConfig()
	if err != nil || loaded.TokenServer != cfg.TokenServer || loaded.TokenRef != cfg.TokenRef {
		t.Fatal("session did not round trip")
	}
}
