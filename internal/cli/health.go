package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/infra/postgres"
	redisinfra "github.com/studio/platform/internal/infra/redis"
)

type healthResult struct {
	Name   string
	Status string
	Detail string
}

func newHealthCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Run API, dependency, and system health checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHealth(cmd.Context(), opts)
		},
	}
}

func runHealth(ctx context.Context, opts *Options) error {
	client := newHTTPClient(opts.Timeout)
	serverURL := opts.serverBaseURL()
	results := make([]healthResult, 0, 6)
	failures := 0

	add := func(name string, err error, detail string) {
		status := "OK"
		if err != nil {
			status = "FAIL"
			failures++
			if detail == "" {
				detail = err.Error()
			}
		}
		results = append(results, healthResult{Name: name, Status: status, Detail: detail})
	}

	if status, _, err := doRequest(ctx, client, http.MethodGet, serverURL+"/health", ""); err != nil {
		add("API /health", err, fmt.Sprintf("status=%d", status))
	} else {
		add("API /health", nil, "status=200")
	}

	if status, _, err := doRequest(ctx, client, http.MethodGet, serverURL+"/ready", ""); err != nil {
		add("API /ready", err, fmt.Sprintf("status=%d", status))
	} else {
		add("API /ready", nil, "status=200")
	}

	if cfg, err := opts.loadConfig(false); err == nil && cfg != nil {
		dbCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
		pool, err := postgres.NewPool(dbCtx, cfg.Database)
		if err != nil {
			add("PostgreSQL", err, "")
		} else {
			stats := pool.Stat()
			add("PostgreSQL", nil, fmt.Sprintf("max=%d idle=%d total=%d", stats.MaxConns(), stats.IdleConns(), stats.TotalConns()))
			pool.Close()
		}

		redisCtx, redisCancel := context.WithTimeout(ctx, opts.Timeout)
		defer redisCancel()
		client, err := redisinfra.NewClient(redisCtx, cfg.Redis)
		if err != nil {
			add("Redis", err, "")
		} else {
			info, pingErr := client.Info(redisCtx, "stats").Result()
			if pingErr != nil {
				add("Redis", pingErr, "")
			} else {
				add("Redis", nil, compactLine(info))
			}
			_ = client.Close()
		}
	} else {
		results = append(results, healthResult{Name: "PostgreSQL", Status: "SKIP", Detail: "config unavailable"})
		results = append(results, healthResult{Name: "Redis", Status: "SKIP", Detail: "config unavailable"})
	}

	if usage, err := rootDiskUsage("/"); err != nil {
		results = append(results, healthResult{Name: "Disk", Status: "SKIP", Detail: err.Error()})
	} else {
		detail := fmt.Sprintf("root usage=%s", strconv.FormatFloat(usage, 'f', 1, 64)+"%")
		if usage >= 85 {
			failures++
			results = append(results, healthResult{Name: "Disk", Status: "FAIL", Detail: detail})
		} else {
			results = append(results, healthResult{Name: "Disk", Status: "OK", Detail: detail})
		}
	}

	if totalMB, usedPct, err := linuxMemoryUsage(); err != nil {
		results = append(results, healthResult{Name: "Memory", Status: "SKIP", Detail: err.Error()})
	} else {
		detail := fmt.Sprintf("used=%s total=%dMB", strconv.FormatFloat(usedPct, 'f', 1, 64)+"%", totalMB)
		if usedPct >= 90 {
			failures++
			results = append(results, healthResult{Name: "Memory", Status: "FAIL", Detail: detail})
		} else {
			results = append(results, healthResult{Name: "Memory", Status: "OK", Detail: detail})
		}
	}

	writeSection(opts.Out, "Health")
	printHealthResults(opts.Out, results)

	if failures > 0 {
		return fmt.Errorf("%d health check(s) failed", failures)
	}
	return nil
}

func printHealthResults(w io.Writer, results []healthResult) {
	for _, result := range results {
		fmt.Fprintf(w, "[%s] %-14s %s\n", result.Status, result.Name, result.Detail)
	}
}

func compactLine(input string) string {
	parts := strings.FieldsFunc(input, func(r rune) bool { return r == '\n' || r == '\r' })
	if len(parts) == 0 {
		return "ok"
	}
	if len(parts[0]) > 80 {
		return parts[0][:80] + "..."
	}
	return parts[0]
}

func rootDiskUsage(path string) (float64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	if stat.Blocks == 0 {
		return 0, fmt.Errorf("no filesystem blocks reported")
	}
	used := stat.Blocks - stat.Bfree
	return float64(used) * 100 / float64(stat.Blocks), nil
}

func linuxMemoryUsage() (totalMB int64, usedPct float64, err error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	var totalKB int64
	var availableKB int64
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			totalKB, _ = strconv.ParseInt(fields[1], 10, 64)
		case "MemAvailable:":
			availableKB, _ = strconv.ParseInt(fields[1], 10, 64)
		}
	}
	if totalKB == 0 {
		return 0, 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
	}
	usedPct = float64(totalKB-availableKB) * 100 / float64(totalKB)
	return totalKB / 1024, usedPct, nil
}
