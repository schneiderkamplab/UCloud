package ucloud_cli

import (
	"testing"
	"ucloud.dk/ucloud_cli/pkg/command"
)

func TestParseWebConnect(t *testing.T) {
	parsed, err := Parse([]string{"connect", "--web", "--browser", "/path/to/chrome"})
	if err != nil {
		t.Fatal(err)
	}
	cmd, ok := parsed.(*command.ConnectCommand)
	if !ok || !cmd.Web || cmd.Browser != "/path/to/chrome" {
		t.Fatal("web connect arguments were not parsed")
	}
}
