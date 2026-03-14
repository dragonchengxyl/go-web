package cli

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	postgresinfra "github.com/studio/platform/internal/infra/postgres"
	"go.uber.org/zap"
)

func newPerfCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perf",
		Short: "Performance analysis helpers",
	}

	var limit int
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Inspect PostgreSQL performance statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBPerf(cmd.Context(), opts, limit)
		},
	}
	dbCmd.Flags().IntVar(&limit, "limit", 10, "Number of rows to print for each section")

	cmd.AddCommand(dbCmd)
	return cmd
}

func runDBPerf(ctx context.Context, opts *Options, limit int) error {
	cfg, err := opts.loadConfig(true)
	if err != nil {
		return err
	}

	pool, err := postgresinfra.NewPool(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	analyzer := postgresinfra.NewQueryAnalyzer(pool, zap.NewNop())

	writeSection(opts.Out, "DB Connection Stats")
	connectionStats, err := analyzer.GetConnectionStats(ctx)
	if err != nil {
		return err
	}
	for _, key := range []string{"total", "active", "idle", "idle_in_transaction", "waiting"} {
		fmt.Fprintf(opts.Out, "%s: %v\n", key, connectionStats[key])
	}

	if slowQueries, err := analyzer.GetSlowQueries(ctx, limit); err != nil {
		fmt.Fprintf(opts.ErrOut, "warning: slow query analysis unavailable: %v\n", err)
	} else {
		writeSection(opts.Out, "Slow Queries")
		tw := tabwriter.NewWriter(opts.Out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "mean_ms\tcalls\trows\tquery")
		for _, item := range slowQueries {
			fmt.Fprintf(tw, "%.2f\t%d\t%d\t%s\n", item.MeanTime, item.Calls, item.Rows, preview(item.Query, 100))
		}
		_ = tw.Flush()
	}

	if tableStats, err := analyzer.GetTableStats(ctx); err != nil {
		fmt.Fprintf(opts.ErrOut, "warning: table stats unavailable: %v\n", err)
	} else {
		writeSection(opts.Out, "Largest Tables")
		tw := tabwriter.NewWriter(opts.Out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "table\trows\ttotal_size\tindex_size\tseq_scans\tidx_scans")
		for i, item := range tableStats {
			if i >= limit {
				break
			}
			fmt.Fprintf(tw, "%s.%s\t%d\t%s\t%s\t%d\t%d\n",
				item.SchemaName,
				item.TableName,
				item.RowCount,
				postgresinfra.FormatBytes(item.TotalSize),
				postgresinfra.FormatBytes(item.IndexSize),
				item.SeqScans,
				item.IdxScans,
			)
		}
		_ = tw.Flush()
	}

	if missingIndexes, err := analyzer.GetMissingIndexes(ctx); err != nil {
		fmt.Fprintf(opts.ErrOut, "warning: missing index analysis unavailable: %v\n", err)
	} else {
		writeSection(opts.Out, "Missing Index Candidates")
		tw := tabwriter.NewWriter(opts.Out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "table\tseq_scans\tseq_reads\tidx_scans\trecommendation")
		for i, item := range missingIndexes {
			if i >= limit {
				break
			}
			fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%s\n", item.TableName, item.SeqScans, item.SeqTupleReads, item.IdxScans, item.Recommendation)
		}
		_ = tw.Flush()
	}

	if unusedIndexes, err := analyzer.GetUnusedIndexes(ctx); err != nil {
		fmt.Fprintf(opts.ErrOut, "warning: unused index analysis unavailable: %v\n", err)
	} else {
		writeSection(opts.Out, "Unused Index Candidates")
		tw := tabwriter.NewWriter(opts.Out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "table\tindex\tindex_size\tscans")
		for i, item := range unusedIndexes {
			if i >= limit {
				break
			}
			fmt.Fprintf(tw, "%s.%s\t%s\t%s\t%d\n",
				item.SchemaName,
				item.TableName,
				item.IndexName,
				postgresinfra.FormatBytes(item.IndexSize),
				item.IndexScans,
			)
		}
		_ = tw.Flush()
	}

	return nil
}

func preview(input string, limit int) string {
	normalized := strings.Join(strings.Fields(input), " ")
	if len(normalized) <= limit {
		return normalized
	}
	return normalized[:limit] + "..."
}
