package usecase

import "testing"

func TestDetectAssistantIntent(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  assistantIntent
	}{
		{
			name:  "onboarding question",
			query: "我第一次来，先逛哪里？",
			want:  assistantIntentOnboarding,
		},
		{
			name:  "group discovery question",
			query: "推荐几个有意思的圈子",
			want:  assistantIntentGroups,
		},
		{
			name:  "event question",
			query: "最近有什么活动值得看？",
			want:  assistantIntentEvents,
		},
		{
			name:  "posting help question",
			query: "怎么发布我的第一条动态？",
			want:  assistantIntentPosting,
		},
		{
			name:  "user recommendation question",
			query: "推荐几个值得关注的用户",
			want:  assistantIntentUsers,
		},
		{
			name:  "content discovery question",
			query: "有什么热门帖子和标签",
			want:  assistantIntentContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectAssistantIntent(tt.query, nil); got != tt.want {
				t.Fatalf("detectAssistantIntent(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}

func TestScoreAssistantCardIntentBoost(t *testing.T) {
	postingIntent := detectAssistantIntent("怎么发布我的第一条动态？", nil)
	createPage := AssistantCard{
		Kind:  "page",
		Title: "发布动态",
		Href:  "/posts/create",
	}
	explorePage := AssistantCard{
		Kind:  "page",
		Title: "发现页",
		Href:  "/explore",
	}
	if scoreAssistantCard(createPage, "怎么发布我的第一条动态？", postingIntent) <= scoreAssistantCard(explorePage, "怎么发布我的第一条动态？", postingIntent) {
		t.Fatalf("posting intent should prioritize /posts/create over /explore")
	}

	groupIntent := detectAssistantIntent("推荐几个有意思的圈子", nil)
	groupCard := AssistantCard{
		Kind:    "group",
		Title:   "兽设灵感交流局",
		Summary: "分享角色设定、配色灵感和参考图",
		Href:    "/groups/1",
	}
	postCard := AssistantCard{
		Kind:    "post",
		Title:   "本周热门作品",
		Summary: "整理近期高互动动态",
		Href:    "/posts/1",
	}
	if scoreAssistantCard(groupCard, "推荐几个有意思的圈子", groupIntent) <= scoreAssistantCard(postCard, "推荐几个有意思的圈子", groupIntent) {
		t.Fatalf("group intent should prioritize groups over generic posts")
	}
}

func TestDetectAssistantIntentWithPageContext(t *testing.T) {
	pageContext := &AssistantPageContext{
		Kind:  "post_create",
		Title: "发布动态",
	}

	if got := detectAssistantIntent("帮我润色一下", pageContext); got != assistantIntentPosting {
		t.Fatalf("detectAssistantIntent with post_create context = %q, want %q", got, assistantIntentPosting)
	}
}
