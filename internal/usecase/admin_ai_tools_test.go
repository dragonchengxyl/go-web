package usecase

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/domain/report"
)

func TestAssessReportRisk(t *testing.T) {
	rep := &report.Report{
		ID:          uuid.New(),
		TargetType:  report.TargetTypeComment,
		Reason:      "骚扰辱骂",
		Description: "这条评论里有明显的人身攻击和威胁措辞",
	}

	level, bullets, action := assessReportRisk(rep, "目标是评论内容。评论摘要：你这种人别出现了。")
	if level != "高" {
		t.Fatalf("risk level = %q, want 高", level)
	}
	if !strings.Contains(action, "删除评论") {
		t.Fatalf("action = %q, want delete comment suggestion", action)
	}
	if len(bullets) == 0 {
		t.Fatalf("expected non-empty risk bullets")
	}
}

func TestBuildEventPublishChecklist(t *testing.T) {
	item := &event.Event{
		Title:         "线上分享会",
		AttendeeCount: 12,
		MaxCapacity:   50,
	}

	bullets := buildEventPublishChecklist(item, "线上活动")
	if len(bullets) < 2 {
		t.Fatalf("expected multiple checklist bullets, got %v", bullets)
	}
	foundPlatformHint := false
	for _, bullet := range bullets {
		if strings.Contains(bullet, "平台") || strings.Contains(bullet, "进群") {
			foundPlatformHint = true
			break
		}
	}
	if !foundPlatformHint {
		t.Fatalf("expected online-event checklist to mention platform or group entry, got %v", bullets)
	}
}
