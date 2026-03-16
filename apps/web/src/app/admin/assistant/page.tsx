"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bot, Loader2, Save, Sparkles } from "lucide-react";
import { apiClient, AssistantSettings } from "@/lib/api-client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { AdminMetricCard } from "@/components/admin/admin-metric-card";
import { AdminPageHeader } from "@/components/admin/admin-page-header";
import { showAdminToast } from "@/components/admin/admin-toast";

const defaultSettings: AssistantSettings = {
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
    <label className="flex cursor-pointer items-start justify-between gap-4 rounded-2xl border border-slate-200 p-4 transition-colors hover:bg-slate-50">
      <div className="min-w-0">
        <p className="text-sm font-medium text-slate-900">{label}</p>
        <p className="mt-1 text-xs leading-5 text-slate-500">{desc}</p>
      </div>
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="mt-1 h-4 w-4 rounded accent-slate-950"
      />
    </label>
  );
}

export default function AdminAssistantPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<AssistantSettings>(defaultSettings);
  const [message, setMessage] = useState("");

  const { data, isLoading } = useQuery<AssistantSettings>({
    queryKey: ["admin-assistant-settings"],
    queryFn: () => apiClient.getAssistantSettings(),
  });

  useEffect(() => {
    if (data) {
      setForm({
        ...defaultSettings,
        ...data,
      });
    }
  }, [data]);

  const saveMutation = useMutation({
    mutationFn: (payload: AssistantSettings) =>
      apiClient.updateAssistantSettings(payload),
    onSuccess: (next) => {
      setForm({
        ...defaultSettings,
        ...next,
      });
      setMessage("设置已保存，新的对话请求会立即生效。");
      showAdminToast("AI 助手设置已保存", "success");
      queryClient.invalidateQueries({ queryKey: ["admin-assistant-settings"] });
    },
    onError: (err: unknown) => {
      const nextMessage = err instanceof Error ? err.message : "保存失败，请重试";
      setMessage(nextMessage);
      showAdminToast(nextMessage, "error");
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
      <AdminPageHeader
        eyebrow="AI Operations"
        title="AI 助手设置"
        description="在这里管理站内 AI 助手的人设、检索来源和系统提示词，让助手输出更贴合业务。"
        actions={
          <Button
            type="submit"
            form="assistant-settings-form"
            disabled={saveMutation.isPending}
            className="bg-slate-950 text-white hover:bg-slate-800"
          >
            {saveMutation.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Save className="mr-2 h-4 w-4" />
            )}
            保存设置
          </Button>
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        <AdminMetricCard
          label="助手状态"
          value={form.enabled ? "启用中" : "已关闭"}
          hint="前端聊天入口是否接受请求"
          icon={Bot}
          tone={form.enabled ? "success" : "warning"}
        />
        <AdminMetricCard
          label="人格名称"
          value={form.persona_name || "未命名"}
          hint="影响前端和回复上下文中的角色展示"
          icon={Sparkles}
          tone="brand"
        />
        <AdminMetricCard
          label="上下文卡片"
          value={String(form.max_context_items)}
          hint="每次对话最多注入的站内推荐卡片数"
          icon={Save}
          tone="default"
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
        </div>
      ) : (
        <form id="assistant-settings-form" onSubmit={handleSubmit} className="space-y-6">
          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardHeader>
              <CardTitle className="text-lg">基础设置</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <ToggleRow
                label="启用 AI 助手"
                desc="关闭后，前端会收到“AI 助手当前已关闭”的提示。"
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
                  <label className="text-sm font-medium">最大上下文卡片数</label>
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
                  placeholder="补充额外规则，例如回答更克制、优先推荐活动、减少跑题等。"
                />
                <p className="text-xs text-slate-500">
                  会追加到默认系统提示词之后，对新会话立即生效。
                </p>
              </div>
            </CardContent>
          </Card>

          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardHeader>
              <CardTitle className="text-lg">检索来源</CardTitle>
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

          <div className="rounded-3xl border border-slate-200 bg-white px-5 py-4 text-sm text-slate-500 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            {message || "保存后，新的对话请求会立即使用最新设置。"}
          </div>
        </form>
      )}
    </div>
  );
}
