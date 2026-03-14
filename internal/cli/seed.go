package cli

import (
	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/cli/seeder"
)

func newSeedCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed development or demo data",
	}

	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Seed demo data used for local smoke tests and interviews",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := opts.loadConfig(true)
			if err != nil {
				return err
			}
			return seeder.SeedDemo(cmd.Context(), cfg, opts.Out)
		},
	}

	cmd.AddCommand(demoCmd)
	return cmd
}
