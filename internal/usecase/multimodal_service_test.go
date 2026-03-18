package usecase

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseMediaAnalysisResponseEvalSet(t *testing.T) {
	type evalCase struct {
		Name              string `json:"name"`
		Raw               string `json:"raw"`
		ExpectAltContains string `json:"expect_alt_contains"`
		ExpectRiskLevel   string `json:"expect_risk_level"`
	}

	path := filepath.Join("testdata", "multimodal_eval_cases.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read eval cases: %v", err)
	}

	var cases []evalCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("unmarshal eval cases: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			item, err := parseMediaAnalysisResponse(tc.Raw, "https://cdn.example.com/demo.png", "post_create", 24*time.Hour)
			if err != nil {
				t.Fatalf("parseMediaAnalysisResponse error: %v", err)
			}
			if !strings.Contains(item.AltText, tc.ExpectAltContains) {
				t.Fatalf("alt_text = %q, want substring %q", item.AltText, tc.ExpectAltContains)
			}
			if item.RiskLevel != tc.ExpectRiskLevel {
				t.Fatalf("risk_level = %q, want %q", item.RiskLevel, tc.ExpectRiskLevel)
			}
		})
	}
}

func TestBuildFallbackMediaAnalysis(t *testing.T) {
	item := buildFallbackMediaAnalysis("https://cdn.example.com/uploads/fox-poster.webp", "moderation", time.Hour)
	if item == nil {
		t.Fatalf("expected fallback analysis")
	}
	if !item.Fallback {
		t.Fatalf("expected fallback flag to be true")
	}
	if item.RiskLevel == "" {
		t.Fatalf("expected fallback risk level")
	}
	if len(item.SafetyNotes) == 0 {
		t.Fatalf("expected fallback safety notes")
	}
}
