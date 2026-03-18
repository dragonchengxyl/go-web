package usecase

import (
	"context"
	"testing"
)

func TestBuildPostCreateInsights(t *testing.T) {
	svc := &AssistantService{}
	ctx := &AssistantPageContext{
		Kind:  "post_create",
		Title: "发布动态",
		Fields: map[string]string{
			"draft_content": "最近把角色设定又调整了一版，想分享一下新的配色和情绪氛围。",
			"group_name":    "原创设定研究所",
			"visibility":    "公开",
		},
	}

	insights := svc.buildPostCreateInsights(context.Background(), "帮我润色一下这条动态", ctx)
	assertInsightKinds(t, insights, "draft_polish", "title_options", "tag_suggestions", "visibility_suggestion")
}

func TestBuildGroupDetailInsights(t *testing.T) {
	svc := &AssistantService{}
	ctx := &AssistantPageContext{
		Kind:  "group_detail",
		Title: "圈子详情：兽设交流",
		Fields: map[string]string{
			"group_description":  "这里主要交流原创角色设定、配色和世界观。",
			"group_rules":        "发帖请注明原创或二创；禁止直接搬运；讨论保持友好。",
			"group_tags":         "原创设定、角色设计",
			"member_count":       "128",
			"post_count":         "456",
			"is_member":          "未加入",
			"featured_post":      "最近的精选内容都偏设定拆解和创作过程分享",
			"group_announcement": "本周欢迎大家发设定过程贴。",
		},
	}

	insights := svc.buildGroupDetailInsights("这个圈子适合我加入吗？", ctx)
	assertInsightKinds(t, insights, "group_atmosphere", "rules_summary", "join_suggestion", "posting_ideas")
}

func TestBuildEventDetailInsights(t *testing.T) {
	svc := &AssistantService{}
	ctx := &AssistantPageContext{
		Kind:  "event_detail",
		Title: "活动详情：线上设定分享会",
		Fields: map[string]string{
			"event_status":      "报名中",
			"event_time":        "2026-03-20 20:00 - 22:00",
			"event_location":    "线上活动",
			"event_attendees":   "12 / 50",
			"event_tags":        "创作交流、线上分享",
			"event_description": "会聊设定、配色、参考资料整理和角色世界观。",
			"has_attended":      "未报名",
		},
	}

	insights := svc.buildEventDetailInsights("参加前我需要准备什么？", ctx)
	assertInsightKinds(t, insights, "event_summary", "fit_assessment", "signup_notes", "preparation_checklist")
}

func assertInsightKinds(t *testing.T, insights []AssistantInsight, expectedKinds ...string) {
	t.Helper()
	set := make(map[string]struct{}, len(insights))
	for _, insight := range insights {
		set[insight.Kind] = struct{}{}
	}
	for _, kind := range expectedKinds {
		if _, ok := set[kind]; !ok {
			t.Fatalf("expected insight kind %q, got %+v", kind, insights)
		}
	}
}
