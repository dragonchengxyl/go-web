package cli

import "testing"

func TestParseCLIEventBlock(t *testing.T) {
	event, data := parseCLIEventBlock("event: token\ndata: {\"content\":\"你好\"}\n\n")
	if event != "token" {
		t.Fatalf("event = %q, want token", event)
	}
	if data != "{\"content\":\"你好\"}" {
		t.Fatalf("data = %q", data)
	}
}

func TestValidateChatCase(t *testing.T) {
	tc := assistantEvalCase{
		Name:               "chat-onboarding",
		ContainsAny:        []string{"发现页", "活动"},
		ExpectCitation:     true,
		ExpectInsightKinds: []string{"title_options"},
	}
	meta := assistantEvalMeta{
		Insights: []struct {
			Kind string `json:"kind"`
		}{
			{Kind: "title_options"},
		},
	}
	reply := "建议先看发现页 [R1]"
	if err := validateChatCase(tc, meta, reply); err != nil {
		t.Fatalf("validateChatCase error: %v", err)
	}
}
