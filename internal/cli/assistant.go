package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/cli/seeder"
)

type assistantEvalMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type assistantEvalPageContext struct {
	Path        string            `json:"path,omitempty"`
	Kind        string            `json:"kind,omitempty"`
	Title       string            `json:"title,omitempty"`
	Summary     string            `json:"summary,omitempty"`
	PromptHints []string          `json:"prompt_hints,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
}

type assistantEvalCase struct {
	Name               string                    `json:"name"`
	Kind               string                    `json:"kind"` // chat_stream | admin_tool
	Auth               string                    `json:"auth,omitempty"`
	SkipIfNoAuth       bool                      `json:"skip_if_no_auth,omitempty"`
	Messages           []assistantEvalMessage    `json:"messages,omitempty"`
	PageContext        *assistantEvalPageContext `json:"page_context,omitempty"`
	Endpoint           string                    `json:"endpoint,omitempty"`
	Payload            map[string]any            `json:"payload,omitempty"`
	ContainsAny        []string                  `json:"contains_any,omitempty"`
	ContainsAll        []string                  `json:"contains_all,omitempty"`
	ExpectCitation     bool                      `json:"expect_citation,omitempty"`
	ExpectInsightKinds []string                  `json:"expect_insight_kinds,omitempty"`
	ExpectDraftLabels  []string                  `json:"expect_draft_labels,omitempty"`
}

type assistantEvalMeta struct {
	Intent      string `json:"intent,omitempty"`
	IntentLabel string `json:"intent_label,omitempty"`
	Fallback    bool   `json:"fallback"`
	Provider    string `json:"provider"`
	Insights    []struct {
		Kind string `json:"kind"`
	} `json:"insights,omitempty"`
}

type assistantEvalAdminToolResult struct {
	Tool     string `json:"tool"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Fallback bool   `json:"fallback"`
	Drafts   []struct {
		Label string `json:"label"`
	} `json:"drafts,omitempty"`
}

type assistantEvalCaseResult struct {
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	Status           string   `json:"status"` // passed | failed | skipped
	DurationMs       int64    `json:"duration_ms"`
	FirstTokenMs     int64    `json:"first_token_ms,omitempty"`
	Provider         string   `json:"provider,omitempty"`
	Fallback         bool     `json:"fallback,omitempty"`
	IntentLabel      string   `json:"intent_label,omitempty"`
	ReplyPreview     string   `json:"reply_preview,omitempty"`
	FailureReason    string   `json:"failure_reason,omitempty"`
	ExpectedInsights []string `json:"expected_insights,omitempty"`
	ObservedInsights []string `json:"observed_insights,omitempty"`
	ObservedDrafts   []string `json:"observed_draft_labels,omitempty"`
	RawSummary       string   `json:"raw_summary,omitempty"`
}

type assistantEvalReport struct {
	GeneratedAt time.Time                 `json:"generated_at"`
	ServerURL   string                    `json:"server_url"`
	Cases       []assistantEvalCaseResult `json:"cases"`
	Passed      int                       `json:"passed"`
	Failed      int                       `json:"failed"`
	Skipped     int                       `json:"skipped"`
}

func newAssistantCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assistant",
		Short: "Assistant evaluation and regression tooling",
	}

	var (
		casesPath  string
		outputPath string
		email      string
		password   string
		failFast   bool
	)

	evalCmd := &cobra.Command{
		Use:   "eval",
		Short: "Run assistant regression cases against a running backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" {
				email = seeder.DefaultDemoAdminEmail
			}
			if password == "" {
				password = seeder.DemoPassword
			}
			return runAssistantEval(cmd.Context(), opts, assistantEvalOptions{
				CasesPath:  casesPath,
				OutputPath: outputPath,
				Email:      email,
				Password:   password,
				FailFast:   failFast,
			})
		},
	}

	evalCmd.Flags().StringVar(&casesPath, "cases", filepath.Join("internal", "cli", "testdata", "assistant_eval_cases.json"), "Path to assistant eval cases JSON")
	evalCmd.Flags().StringVar(&outputPath, "out-json", "", "Write a JSON report to this file")
	evalCmd.Flags().StringVar(&email, "email", "", "Admin login email for authenticated eval cases")
	evalCmd.Flags().StringVar(&password, "password", "", "Admin login password for authenticated eval cases")
	evalCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop on the first failed case")

	cmd.AddCommand(evalCmd)
	return cmd
}

type assistantEvalOptions struct {
	CasesPath  string
	OutputPath string
	Email      string
	Password   string
	FailFast   bool
}

func runAssistantEval(ctx context.Context, opts *Options, input assistantEvalOptions) error {
	cases, err := loadAssistantEvalCases(input.CasesPath)
	if err != nil {
		return err
	}

	client := newHTTPClient(opts.Timeout)
	baseURL := opts.serverBaseURL()

	var token string
	token, _ = loginForAssistantEval(ctx, client, baseURL, input.Email, input.Password)

	report := assistantEvalReport{
		GeneratedAt: time.Now(),
		ServerURL:   baseURL,
		Cases:       make([]assistantEvalCaseResult, 0, len(cases)),
	}

	for _, tc := range cases {
		result := runAssistantEvalCase(ctx, client, baseURL, token, tc)
		report.Cases = append(report.Cases, result)
		switch result.Status {
		case "passed":
			report.Passed++
		case "failed":
			report.Failed++
		case "skipped":
			report.Skipped++
		}
		fmt.Fprintf(opts.Out, "[%s] %s", strings.ToUpper(result.Status), result.Name)
		if result.FailureReason != "" {
			fmt.Fprintf(opts.Out, " - %s", result.FailureReason)
		}
		fmt.Fprintln(opts.Out)
		if input.FailFast && result.Status == "failed" {
			break
		}
	}

	fmt.Fprintf(opts.Out, "\npassed=%d failed=%d skipped=%d\n", report.Passed, report.Failed, report.Skipped)

	if strings.TrimSpace(input.OutputPath) != "" {
		raw, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal eval report: %w", err)
		}
		if err := os.WriteFile(input.OutputPath, raw, 0o644); err != nil {
			return fmt.Errorf("write eval report: %w", err)
		}
		fmt.Fprintf(opts.Out, "json report written to %s\n", input.OutputPath)
	}

	if report.Failed > 0 {
		return fmt.Errorf("%d assistant eval case(s) failed", report.Failed)
	}
	return nil
}

func loginForAssistantEval(ctx context.Context, client *http.Client, baseURL, email, password string) (string, error) {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return "", fmt.Errorf("missing login credentials")
	}

	var loginEnvelope apiEnvelope[loginResponse]
	status, err := doJSON(ctx, client, http.MethodPost, baseURL+"/api/v1/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	}, &loginEnvelope)
	if err != nil {
		return "", fmt.Errorf("login failed (status=%d): %w", status, err)
	}
	return loginEnvelope.Data.AccessToken, nil
}

func runAssistantEvalCase(ctx context.Context, client *http.Client, baseURL, token string, tc assistantEvalCase) assistantEvalCaseResult {
	start := time.Now()
	result := assistantEvalCaseResult{
		Name:             tc.Name,
		Kind:             tc.Kind,
		ExpectedInsights: tc.ExpectInsightKinds,
	}

	needsAuth := strings.TrimSpace(tc.Auth) != "" && tc.Auth != "none"
	if needsAuth && token == "" {
		result.Status = "skipped"
		result.FailureReason = "authenticated case skipped because login was unavailable"
		if !tc.SkipIfNoAuth {
			result.Status = "failed"
			result.FailureReason = "authenticated case requires a valid login token"
		}
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	switch tc.Kind {
	case "chat_stream":
		meta, reply, firstTokenMs, err := runAssistantChatEval(ctx, client, baseURL, token, tc)
		result.DurationMs = time.Since(start).Milliseconds()
		result.FirstTokenMs = firstTokenMs
		result.Provider = meta.Provider
		result.Fallback = meta.Fallback
		result.IntentLabel = meta.IntentLabel
		result.ReplyPreview = truncateCLI(reply, 180)
		result.ObservedInsights = insightKinds(meta.Insights)
		if err != nil {
			result.Status = "failed"
			result.FailureReason = err.Error()
			return result
		}
	case "admin_tool":
		summary, drafts, provider, fallback, err := runAssistantAdminToolEval(ctx, client, baseURL, token, tc)
		result.DurationMs = time.Since(start).Milliseconds()
		result.Provider = provider
		result.Fallback = fallback
		result.RawSummary = truncateCLI(summary, 180)
		result.ObservedDrafts = drafts
		if err != nil {
			result.Status = "failed"
			result.FailureReason = err.Error()
			return result
		}
	default:
		result.DurationMs = time.Since(start).Milliseconds()
		result.Status = "failed"
		result.FailureReason = fmt.Sprintf("unsupported case kind %q", tc.Kind)
		return result
	}

	result.Status = "passed"
	return result
}

func runAssistantChatEval(ctx context.Context, client *http.Client, baseURL, token string, tc assistantEvalCase) (assistantEvalMeta, string, int64, error) {
	payload := map[string]any{
		"messages": tc.Messages,
	}
	if tc.PageContext != nil {
		payload["page_context"] = tc.PageContext
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return assistantEvalMeta{}, "", 0, fmt.Errorf("marshal case payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/assistant/chat/stream", strings.NewReader(string(raw)))
	if err != nil {
		return assistantEvalMeta{}, "", 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return assistantEvalMeta{}, "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return assistantEvalMeta{}, "", 0, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	reader := bufio.NewReader(resp.Body)
	var (
		meta         assistantEvalMeta
		reply        strings.Builder
		buffer       strings.Builder
		firstTokenMs int64
		seenFirst    bool
	)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return meta, reply.String(), firstTokenMs, err
		}
		buffer.WriteString(strings.ReplaceAll(line, "\r", ""))
		if strings.HasSuffix(buffer.String(), "\n\n") {
			event, data := parseCLIEventBlock(buffer.String())
			buffer.Reset()
			switch event {
			case "meta":
				if err := json.Unmarshal([]byte(data), &meta); err != nil {
					return meta, reply.String(), firstTokenMs, fmt.Errorf("decode meta: %w", err)
				}
			case "token":
				var payload struct {
					Content string `json:"content"`
				}
				if err := json.Unmarshal([]byte(data), &payload); err != nil {
					return meta, reply.String(), firstTokenMs, fmt.Errorf("decode token: %w", err)
				}
				if !seenFirst {
					firstTokenMs = time.Since(start).Milliseconds()
					seenFirst = true
				}
				reply.WriteString(payload.Content)
			case "error":
				var payload struct {
					Message string `json:"message"`
				}
				_ = json.Unmarshal([]byte(data), &payload)
				return meta, reply.String(), firstTokenMs, errors.New(firstNonEmptyCLI(payload.Message, "assistant stream returned error"))
			}
		}
	}

	replyText := reply.String()
	if err := validateChatCase(tc, meta, replyText); err != nil {
		return meta, replyText, firstTokenMs, err
	}
	return meta, replyText, firstTokenMs, nil
}

func runAssistantAdminToolEval(ctx context.Context, client *http.Client, baseURL, token string, tc assistantEvalCase) (string, []string, string, bool, error) {
	if strings.TrimSpace(tc.Endpoint) == "" {
		return "", nil, "", false, fmt.Errorf("admin tool case missing endpoint")
	}
	var envelope apiEnvelope[assistantEvalAdminToolResult]
	status, err := doJSON(ctx, client, http.MethodPost, baseURL+tc.Endpoint, token, tc.Payload, &envelope)
	if err != nil {
		return "", nil, "", false, fmt.Errorf("admin tool request failed (status=%d): %w", status, err)
	}

	draftLabels := make([]string, 0, len(envelope.Data.Drafts))
	for _, draft := range envelope.Data.Drafts {
		draftLabels = append(draftLabels, draft.Label)
	}
	if err := validateAdminToolCase(tc, envelope.Data, draftLabels); err != nil {
		return envelope.Data.Summary, draftLabels, envelope.Data.Tool, envelope.Data.Fallback, err
	}
	return envelope.Data.Summary, draftLabels, envelope.Data.Tool, envelope.Data.Fallback, nil
}

func validateChatCase(tc assistantEvalCase, meta assistantEvalMeta, reply string) error {
	reply = strings.TrimSpace(reply)
	if reply == "" {
		return fmt.Errorf("empty assistant reply")
	}
	if len(tc.ContainsAll) > 0 {
		for _, expected := range tc.ContainsAll {
			if !strings.Contains(reply, expected) {
				return fmt.Errorf("reply missing required text %q", expected)
			}
		}
	}
	if len(tc.ContainsAny) > 0 {
		matched := false
		for _, expected := range tc.ContainsAny {
			if strings.Contains(reply, expected) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("reply did not match any expected text: %s", strings.Join(tc.ContainsAny, ", "))
		}
	}
	if tc.ExpectCitation {
		if !regexp.MustCompile(`\[R\d+\]`).MatchString(reply) {
			return fmt.Errorf("reply missing source citation")
		}
	}
	if len(tc.ExpectInsightKinds) > 0 {
		observed := insightKinds(meta.Insights)
		for _, expected := range tc.ExpectInsightKinds {
			if !containsString(observed, expected) {
				return fmt.Errorf("missing expected insight kind %q", expected)
			}
		}
	}
	return nil
}

func validateAdminToolCase(tc assistantEvalCase, result assistantEvalAdminToolResult, drafts []string) error {
	joined := result.Title + "\n" + result.Summary
	if len(tc.ContainsAll) > 0 {
		for _, expected := range tc.ContainsAll {
			if !strings.Contains(joined, expected) {
				return fmt.Errorf("tool output missing required text %q", expected)
			}
		}
	}
	if len(tc.ContainsAny) > 0 {
		matched := false
		for _, expected := range tc.ContainsAny {
			if strings.Contains(joined, expected) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("tool output did not match any expected text: %s", strings.Join(tc.ContainsAny, ", "))
		}
	}
	for _, expected := range tc.ExpectDraftLabels {
		if !containsString(drafts, expected) {
			return fmt.Errorf("missing expected draft label %q", expected)
		}
	}
	return nil
}

func loadAssistantEvalCases(path string) ([]assistantEvalCase, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read eval cases: %w", err)
	}
	var cases []assistantEvalCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		return nil, fmt.Errorf("decode eval cases: %w", err)
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("no assistant eval cases found in %s", path)
	}
	return cases, nil
}

func parseCLIEventBlock(block string) (string, string) {
	lines := strings.Split(block, "\n")
	event := "message"
	var dataLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	return event, strings.Join(dataLines, "\n")
}

func insightKinds(items []struct {
	Kind string `json:"kind"`
}) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Kind) != "" {
			out = append(out, item.Kind)
		}
	}
	return out
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func truncateCLI(text string, limit int) string {
	text = strings.TrimSpace(text)
	if limit <= 0 || len([]rune(text)) <= limit {
		return text
	}
	return string([]rune(text)[:limit]) + "..."
}

func firstNonEmptyCLI(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
