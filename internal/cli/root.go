package cli

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

// Options holds global CLI options.
type Options struct {
	ConfigPath string
	ServerURL  string
	Timeout    time.Duration
	Out        *os.File
	ErrOut     *os.File
}

// NewRootCmd constructs the studio CLI root command.
func NewRootCmd() *cobra.Command {
	opts := &Options{
		ConfigPath: "configs/config.local.yaml",
		ServerURL:  "",
		Timeout:    5 * time.Second,
		Out:        os.Stdout,
		ErrOut:     os.Stderr,
	}

	cmd := &cobra.Command{
		Use:           "studio-cli",
		Short:         "Operational and QA tooling for the studio platform",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", opts.ConfigPath, "Path to the YAML config file")
	cmd.PersistentFlags().StringVar(&opts.ServerURL, "server", opts.ServerURL, "Backend base URL, e.g. http://localhost:8080")
	cmd.PersistentFlags().DurationVar(&opts.Timeout, "timeout", opts.Timeout, "Default network timeout")

	cmd.AddCommand(
		newHealthCmd(opts),
		newPerfCmd(opts),
		newPprofCmd(opts),
		newSeedCmd(opts),
		newSmokeCmd(opts),
	)

	return cmd
}
