"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bot, Loader2, Save } from "lucide-react";
import { apiClient, AssistantSettings } from "@/lib/api-client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

const DEFAULT_SETTINGS: AssistantSettings = {
  enabled: true,
  persona_name: "霜牙",
  system_prompt: "",
  max_context_items: 6,
  include_pages: true,
  include_posts: true,
  include_users: true,
  include_tags: true,
  include_groups: true,
  include_events: true,
};

function ToggleRow({
  label,
  desc,
  checked,
  onChange,
}: {
  label: string;
  desc: string;
  checked: boolean;
  onChange: (next: boolean) => void;
}) {
  return (
    <label className="flex items-start justify-between gap-4 rounded-xl border p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/60 transition-colors">
      <div className="min-w-0">
        <p className="text-sm font-medium text-gray-900 dark:text-white">
          {label}
        </p>
        <p className="mt-1 text-xs leading-5 text-gray-500">{desc}</p>
      </div>
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="mt-1 h-4 w-4 rounded accent-purple-600"
      />
    </label>
  );
}

export default function AdminAssistantPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<AssistantSettings>(DEFAULT_SETTINGS);
  const [message, setMessage] = useState("");

  const { data, isLoading } = useQuery<AssistantSettings>({
    queryKey: ["admin-assistant-settings"],
    queryFn: () => apiClient.getAssistantSettings(),
  });

  useEffect(() => {
    if (data) {
      setForm({
        ...DEFAULT_SETTINGS,
        ...data,
      });
    }
  }, [data]);

  const saveMutation = useMutation({
    mutationFn: (payload: AssistantSettings) =>
      apiClient.updateAssistantSettings(payload),
    onSuccess: (next) => {
      setForm({
        ...DEFAULT_SETTINGS,
        ...next,
      });
      setMessage("AI 设置已保存");
      queryClient.invalidateQueries({ queryKey: ["admin-assistant-settings"] });
    },
    onError: (err: unknown) => {
      setMessage(err instanceof Error ? err.message : "保存失败，请重试");
    },
  });

  function update<K extends keyof AssistantSettings>(
    key: K,
    value: AssistantSettings[K],
  ) {
    setForm((prev) => ({ ...prev, [key]: value }));
    setMessage("");
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    saveMutation.mutate(form);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <Bot className="h-6 w-6" />
            AI 助手设置
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            修改站内 AI 助手的人设、附加提示词和检索来源。
          </p>
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">基础设置</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <ToggleRow
                label="启用 AI 助手"
                desc="关闭后，前端请求会收到“AI 助手当前已关闭”的提示。"
                checked={form.enabled}
                onChange={(next) => update("enabled", next)}
              />

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium">角色名称</label>
                  <Input
                    value={form.persona_name}
                    onChange={(e) => update("persona_name", e.target.value)}
                    placeholder="例如：霜牙"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">
                    最大上下文卡片数
                  </label>
                  <Input
                    type="number"
                    min={2}
                    max={12}
                    value={form.max_context_items}
                    onChange={(e) =>
                      update("max_context_items", Number(e.target.value) || 6)
                    }
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">额外系统提示词</label>
                <Textarea
                  value={form.system_prompt}
                  onChange={(e) => update("system_prompt", e.target.value)}
                  rows={8}
                  placeholder="补充额外规则，例如回答更克制、避免剧透、优先推荐活动等。"
                />
                <p className="text-xs text-gray-500">
                  这部分会追加到基础系统提示词后面。
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">检索来源</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3 md:grid-cols-2">
              <ToggleRow
                label="页面入口"
                desc="首页、发现页、圈子、活动、创作者面板等固定入口。"
                checked={form.include_pages}
                onChange={(next) => update("include_pages", next)}
              />
              <ToggleRow
                label="帖子"
                desc="公开且审核通过的动态内容。"
                checked={form.include_posts}
                onChange={(next) => update("include_posts", next)}
              />
              <ToggleRow
                label="用户"
                desc="搜索用户、创作者与其主页入口。"
                checked={form.include_users}
                onChange={(next) => update("include_users", next)}
              />
              <ToggleRow
                label="标签"
                desc="热门标签和标签聚合页。"
                checked={form.include_tags}
                onChange={(next) => update("include_tags", next)}
              />
              <ToggleRow
                label="圈子"
                desc="公开圈子、成员数和圈子详情入口。"
                checked={form.include_groups}
                onChange={(next) => update("include_groups", next)}
              />
              <ToggleRow
                label="活动"
                desc="近期公开活动与活动详情入口。"
                checked={form.include_events}
                onChange={(next) => update("include_events", next)}
              />
            </CardContent>
          </Card>

          <div className="flex items-center justify-between gap-4">
            <p
              className={`text-sm ${message.includes("已保存") ? "text-green-600" : "text-red-500"}`}
            >
              {message || "保存后，新的对话请求会立即使用最新设置。"}
            </p>
            <Button
              type="submit"
              disabled={saveMutation.isPending}
              className="bg-purple-600 hover:bg-purple-500 text-white"
            >
              {saveMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Save className="mr-2 h-4 w-4" />
              )}
              保存设置
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}
