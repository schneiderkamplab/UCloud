package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"ucloud.dk/ucloud_cli/pkg/shared"
	"ucloud.dk/ucloud_cli/pkg/weblogin"
)

func performWebConnection(browser string) error {
	cfg, err := shared.ReadConfig()
	if err != nil {
		return err
	}
	environment, ok := cfg.Environments[cfg.DefaultEnvironment]
	if !ok {
		return fmt.Errorf("selected environment does not exist")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	fmt.Printf("Opening %s for login. Complete login in the new browser window.\n", environment.URL)
	fmt.Println("The CLI will save this login session for future commands. Ctrl-C cancels.")
	session, err := weblogin.Login(ctx, environment.URL, browser)
	if err != nil {
		return err
	}
	// Preserve unrelated changes made while the browser was open, but never save a
	// credential against a different server if the active environment changed.
	current, err := shared.ReadConfig()
	if err != nil {
		return err
	}
	if current.DefaultEnvironment != cfg.DefaultEnvironment || current.Environments[current.DefaultEnvironment].URL != environment.URL {
		return fmt.Errorf("environment changed during login; run connect --web again")
	}
	current.Username = session.Username
	current.TokenRef = session.RefreshToken
	current.TokenServer = strings.TrimRight(environment.URL, "/")
	if _, err = shared.SaveConfig(current); err != nil {
		return err
	}
	fmt.Printf("Connected as %s. Session saved to %s.\n", session.Username, shared.GetConfigPath())
	return nil
}
