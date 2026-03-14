package cli

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/cli/seeder"
)

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func newSmokeCmd(opts *Options) *cobra.Command {
	var email string
	var password string

	cmd := &cobra.Command{
		Use:   "smoke",
		Short: "Run API smoke checks against a running backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" {
				email = seeder.DefaultDemoAdminEmail
			}
			if password == "" {
				password = seeder.DemoPassword
			}
			return runSmoke(cmd.Context(), opts, email, password)
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Login email used for authenticated checks")
	cmd.Flags().StringVar(&password, "password", "", "Login password used for authenticated checks")
	return cmd
}

func runSmoke(ctx context.Context, opts *Options, email, password string) error {
	client := newHTTPClient(opts.Timeout)
	baseURL := opts.serverBaseURL()

	runRaw := func(name, path string) error {
		status, _, err := doRequest(ctx, client, http.MethodGet, baseURL+path, "")
		if err != nil {
			return fmt.Errorf("%s failed (status=%d): %w", name, status, err)
		}
		fmt.Fprintf(opts.Out, "[OK] %s\n", name)
		return nil
	}
	runAPI := func(name, method, path, token string, payload any) error {
		var envelope apiEnvelope[map[string]any]
		status, err := doJSON(ctx, client, method, baseURL+path, token, payload, &envelope)
		if err != nil {
			return fmt.Errorf("%s failed (status=%d): %w", name, status, err)
		}
		fmt.Fprintf(opts.Out, "[OK] %s\n", name)
		return nil
	}

	if err := runRaw("health endpoint", "/health"); err != nil {
		return err
	}
	if err := runRaw("ready endpoint", "/ready"); err != nil {
		return err
	}

	var loginEnvelope apiEnvelope[loginResponse]
	status, err := doJSON(ctx, client, http.MethodPost, baseURL+"/api/v1/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	}, &loginEnvelope)
	if err != nil {
		return fmt.Errorf("login failed (status=%d): %w", status, err)
	}
	token := loginEnvelope.Data.AccessToken
	fmt.Fprintf(opts.Out, "[OK] login as %s\n", email)

	for _, check := range []struct {
		name string
		path string
	}{
		{"profile endpoint", "/api/v1/users/me"},
		{"feed endpoint", "/api/v1/feed?page=1&page_size=5"},
		{"notifications endpoint", "/api/v1/notifications?page=1&page_size=5"},
		{"explore endpoint", "/api/v1/explore?page=1&page_size=5"},
		{"groups endpoint", "/api/v1/groups?page=1&page_size=5"},
		{"events endpoint", "/api/v1/events?page=1&page_size=5"},
		{"search endpoint", "/api/v1/search?q=AI"},
	} {
		if err := runAPI(check.name, http.MethodGet, check.path, token, nil); err != nil {
			return err
		}
	}

	fmt.Fprintln(opts.Out, "smoke test passed")
	return nil
}
