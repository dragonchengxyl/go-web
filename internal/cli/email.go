package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/studio/platform/configs"
	pkgemail "github.com/studio/platform/internal/pkg/email"
)

func newEmailCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "email",
		Short: "Inspect and validate outbound email delivery",
	}

	var to string

	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Show the effective outbound email configuration summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEmailInspect(opts)
		},
	}

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Send a test email using the configured SMTP settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEmailTest(opts, to)
		},
	}
	testCmd.Flags().StringVar(&to, "to", "", "Recipient email address")
	_ = testCmd.MarkFlagRequired("to")

	cmd.AddCommand(inspectCmd, testCmd)
	return cmd
}

func runEmailInspect(opts *Options) error {
	cfg, err := opts.loadConfig(true)
	if err != nil {
		return err
	}

	printEmailSummary(opts, cfg)
	return nil
}

func runEmailTest(opts *Options, to string) error {
	cfg, err := opts.loadConfig(true)
	if err != nil {
		return err
	}

	trimmedTo := strings.TrimSpace(to)
	if trimmedTo == "" {
		return fmt.Errorf("recipient email is required")
	}

	sender := pkgemail.NewSender(cfg.Email)
	printEmailSummary(opts, cfg)

	if !sender.Enabled() {
		return fmt.Errorf("email sender is not configured: host=%q port=%d from=%q", cfg.Email.Host, cfg.Email.Port, cfg.Email.From)
	}

	if err := sender.SendTest(trimmedTo); err != nil {
		return fmt.Errorf("send test email to %s via %s:%d failed: %w", trimmedTo, cfg.Email.Host, cfg.Email.Port, err)
	}

	fmt.Fprintf(opts.Out, "[OK] test email sent to %s\n", trimmedTo)
	return nil
}

func printEmailSummary(opts *Options, cfg *configs.Config) {
	sender := pkgemail.NewSender(cfg.Email)
	frontendURL := resolveFrontendURL(cfg)
	deliveryEnabled := sender.Enabled() && frontendURL != ""

	writeSection(opts.Out, "Email")
	fmt.Fprintf(opts.Out, "config=%s\n", opts.ConfigPath)
	fmt.Fprintf(opts.Out, "smtp_host=%s\n", emptyDash(cfg.Email.Host))
	fmt.Fprintf(opts.Out, "smtp_port=%d\n", cfg.Email.Port)
	fmt.Fprintf(opts.Out, "smtp_username=%s\n", maskEmailIdentity(cfg.Email.Username))
	fmt.Fprintf(opts.Out, "email_from=%s\n", emptyDash(cfg.Email.From))
	fmt.Fprintf(opts.Out, "frontend_url=%s\n", emptyDash(frontendURL))
	fmt.Fprintf(opts.Out, "sender_enabled=%t\n", sender.Enabled())
	fmt.Fprintf(opts.Out, "delivery_enabled=%t\n", deliveryEnabled)
}

func resolveFrontendURL(cfg *configs.Config) string {
	frontendURL := strings.TrimSpace(cfg.Server.FrontendURL)
	if frontendURL != "" {
		return strings.TrimRight(frontendURL, "/")
	}

	for _, origin := range cfg.Server.AllowOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" && trimmed != "*" {
			return strings.TrimRight(trimmed, "/")
		}
	}

	return ""
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func maskEmailIdentity(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "—"
	}

	at := strings.Index(trimmed, "@")
	if at <= 0 {
		if len(trimmed) <= 2 {
			return trimmed[:1] + "***"
		}
		return trimmed[:2] + "***"
	}

	local := trimmed[:at]
	domain := trimmed[at:]
	switch {
	case len(local) <= 1:
		return local + "***" + domain
	case len(local) == 2:
		return local[:1] + "***" + domain
	default:
		return local[:2] + "***" + domain
	}
}
