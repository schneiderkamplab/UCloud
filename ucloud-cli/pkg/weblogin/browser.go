package weblogin

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Login opens a fresh profile, leaving the user's existing browser sessions alone.
// Only cookies applicable to the selected UCloud origin are inspected. Credentials
// entered into the institutional login pages are never read by this package.
func Login(ctx context.Context, rawServer, executable string) (Session, error) {
	server, err := serverURL(rawServer)
	if err != nil {
		return Session{}, err
	}
	profile, err := os.MkdirTemp("", "ucloud-web-login-*")
	if err != nil {
		return Session{}, err
	}
	defer os.RemoveAll(profile)
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	opts = append(opts, chromedp.Flag("headless", false), chromedp.UserDataDir(profile))
	if executable != "" {
		opts = append(opts, chromedp.ExecPath(executable))
	}
	allocator, closeAllocator := chromedp.NewExecAllocator(ctx, opts...)
	defer closeAllocator()
	browser, closeBrowser := chromedp.NewContext(allocator)
	defer closeBrowser()
	if err := chromedp.Do(browser, chromedp.Navigate(server.String()+"/app/login")); err != nil {
		return Session{}, fmt.Errorf("could not open login browser; install Chrome/Chromium or set --browser")
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-browser.Done():
			return Session{}, fmt.Errorf("web login interrupted: %w", browser.Err())
		case <-ticker.C:
			result, err := chromedp.Run(browser, func(ctx context.Context, target *chromedp.Target) (network.GetCookiesResult, error) {
				return cdp.Call(ctx, target, network.GetCookies, network.GetCookiesParams{URLs: []string{server.String() + "/"}})
			})
			if err != nil {
				return Session{}, fmt.Errorf("login browser closed or disconnected")
			}
			for _, cookie := range result.Cookies {
				if cookie.Name == "refreshToken" && cookie.Value != "" && cookie.HTTPOnly {
					return validateSession(ctx, server, cookie.Value)
				}
			}
		}
	}
}
