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

	var (
		bulkProfile   string
		bulkNamespace string
		bulkSeed      int64
	)

	bulkCmd := &cobra.Command{
		Use:   "bulk",
		Short: "Seed large synthetic data across major modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := opts.loadConfig(true)
			if err != nil {
				return err
			}
			return seeder.SeedBulk(cmd.Context(), cfg, seeder.BulkSeedOptions{
				Profile:   bulkProfile,
				Namespace: bulkNamespace,
				Seed:      bulkSeed,
			}, opts.Out)
		},
	}
	bulkCmd.Flags().StringVar(&bulkProfile, "profile", "medium", "bulk seed profile: small, medium, large")
	bulkCmd.Flags().StringVar(&bulkNamespace, "namespace", "bulk", "namespace prefix used to keep generated data deterministic")
	bulkCmd.Flags().Int64Var(&bulkSeed, "seed", 0, "random seed override; 0 derives a deterministic seed from namespace")

	cmd.AddCommand(demoCmd)
	cmd.AddCommand(bulkCmd)
	return cmd
}
