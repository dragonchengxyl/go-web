package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newPprofCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pprof",
		Short: "Fetch pprof profiles from a running backend",
	}

	var cpuSeconds int
	var cpuOutput string
	var cpuToken string
	cpuCmd := &cobra.Command{
		Use:   "cpu",
		Short: "Fetch a CPU profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout := opts.Timeout
			if minimum := (time.Duration(cpuSeconds) * time.Second) + 10*time.Second; timeout < minimum {
				timeout = minimum
			}
			return fetchProfile(context.Background(), opts.serverBaseURL()+"/debug/pprof/profile?seconds="+strconv.Itoa(cpuSeconds), cpuToken, cpuOutput, timeout, opts.Out)
		},
	}
	cpuCmd.Flags().IntVar(&cpuSeconds, "seconds", 30, "CPU profiling duration")
	cpuCmd.Flags().StringVar(&cpuOutput, "output", "cpu.pprof", "Output file path")
	cpuCmd.Flags().StringVar(&cpuToken, "token", "", "Optional bearer token for protected pprof endpoints")

	var heapOutput string
	var heapToken string
	heapCmd := &cobra.Command{
		Use:   "heap",
		Short: "Fetch a heap profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchProfile(context.Background(), opts.serverBaseURL()+"/debug/pprof/heap", heapToken, heapOutput, opts.Timeout, opts.Out)
		},
	}
	heapCmd.Flags().StringVar(&heapOutput, "output", "heap.pprof", "Output file path")
	heapCmd.Flags().StringVar(&heapToken, "token", "", "Optional bearer token for protected pprof endpoints")

	cmd.AddCommand(cpuCmd, heapCmd)
	return cmd
}

func fetchProfile(ctx context.Context, url, token, output string, timeout time.Duration, out io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := newHTTPClient(timeout)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "saved %d bytes to %s\n", written, output)
	return nil
}
