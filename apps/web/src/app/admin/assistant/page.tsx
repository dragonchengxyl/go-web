"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Bot,
  Copy,
  FileText,
  Loader2,
  Megaphone,
  Save,
  ShieldAlert,
  Sparkles,
  Wand2,
} from "lucide-react";
import {
  AdminAIToolResult,
  AdminEvent,
  apiClient,
  AdminReport,
  AdminUser,
  AssistantMeta,
  AssistantOverviewData,
  AssistantSettings,
} from "@/lib/api-client";
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

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

const SOURCE_KIND_LABELS: Record<string, string> = {
  page: "页面",
  post: "帖子",
  user: "用户",
  tag: "标签",
  group: "圈子",
  event: "活动",
};

function parseSSEBlock(block: string): { event: string; data: string } | null {
  const lines = block.split("\n");
  let event = "message";
  const dataLines: string[] = [];

  for (const line of lines) {
    if (line.startsWith("event:")) {
      event = line.slice(6).trim();
    } else if (line.startsWith("data:")) {
      dataLines.push(line.slice(5).trim());
    }
  }

  if (dataLines.length === 0) return null;
  return { event, data: dataLines.join("\n") };
}

function formatSourceCounts(counts?: Record<string, number>) {
  if (!counts) return "";
  return Object.entries(counts)
    .sort(([kindA], [kindB]) => kindA.localeCompare(kindB))
    .map(([kind, count]) => `${SOURCE_KIND_LABELS[kind] || kind}×${count}`)
    .join(" · ");
}

function formatDateTime(value?: string) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleString("zh-CN");
}

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

function ToolResultCard({
  result,
  loading,
  empty,
  onCopy,
}: {
  result?: AdminAIToolResult;
  loading?: boolean;
  empty: string;
  onCopy: (text: string) => void;
}) {
  if (loading) {
    return (
      <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-6 text-center text-sm text-slate-500">
        <Loader2 className="mx-auto mb-2 h-4 w-4 animate-spin" />
        正在生成内容...
      </div>
    );
  }

  if (!result) {
    return (
      <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
        {empty}
      </div>
    );
  }

  return (
    <div className="space-y-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
      <div className="flex flex-wrap items-center gap-2 text-xs text-slate-500">
        <span className="rounded-full bg-slate-900 px-2.5 py-1 font-medium text-white">
          {result.fallback ? "Fallback" : result.provider || "AI"}
        </span>
        <span>Run ID: {result.run_id}</span>
        <span>{new Date(result.generated_at).toLocaleString("zh-CN")}</span>
      </div>
      <div>
        <p className="text-sm font-semibold text-slate-900">{result.title}</p>
        <p className="mt-1 text-sm leading-6 text-slate-600">{result.summary}</p>
      </div>
      <div className="grid gap-3 lg:grid-cols-2">
        {(result.sections || []).map((section) => (
          <div
            key={section.title}
            className="rounded-2xl border border-white bg-white px-4 py-3"
          >
            <p className="text-sm font-semibold text-slate-900">{section.title}</p>
            <ul className="mt-2 space-y-1.5 pl-4 text-sm leading-6 text-slate-600">
              {(section.bullets || []).map((bullet, index) => (
                <li key={`${section.title}-${index}`} className="list-disc">
                  {bullet}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <div className="space-y-3">
        {(result.drafts || []).map((draft) => (
          <div
            key={draft.label}
            className="rounded-2xl border border-white bg-white px-4 py-3"
          >
            <div className="flex items-center justify-between gap-4">
              <p className="text-sm font-semibold text-slate-900">{draft.label}</p>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => onCopy(draft.content)}
              >
                <Copy className="mr-1 h-4 w-4" />
                复制
              </Button>
            </div>
            <div className="mt-3 whitespace-pre-wrap rounded-2xl bg-slate-50 px-3 py-3 text-sm leading-6 text-slate-700">
              {draft.content}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function AdminAssistantPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<AssistantSettings>(defaultSettings);
  const [message, setMessage] = useState("");
  const [diagnosticPrompt, setDiagnosticPrompt] = useState("我第一次来，先逛哪里？");
  const [diagnosticMeta, setDiagnosticMeta] = useState<AssistantMeta | null>(null);
  const [diagnosticReply, setDiagnosticReply] = useState("");
  const [diagnosticError, setDiagnosticError] = useState("");
  const [diagnosticLoading, setDiagnosticLoading] = useState(false);
  const [selectedReportId, setSelectedReportId] = useState("");
  const [selectedCreatorId, setSelectedCreatorId] = useState("");
  const [selectedEventId, setSelectedEventId] = useState("");
  const [weeklyDays, setWeeklyDays] = useState(7);

  const { data, isLoading } = useQuery<AssistantSettings>({
    queryKey: ["admin-assistant-settings"],
    queryFn: () => apiClient.getAssistantSettings(),
  });
  const { data: overview } = useQuery<AssistantOverviewData>({
    queryKey: ["admin-assistant-overview"],
    queryFn: () => apiClient.getAssistantOverview(),
    refetchInterval: 30000,
  });
  const { data: reportOptions } = useQuery({
    queryKey: ["admin-assistant-tool-reports"],
    queryFn: () => apiClient.getAdminReports({ page: 1, page_size: 20, status: "pending" }),
  });
  const { data: creatorOptions } = useQuery({
    queryKey: ["admin-assistant-tool-creators"],
    queryFn: () => apiClient.getAdminUsers({ page: 1, page_size: 20, role: "creator", status: "active" }),
  });
  const { data: eventOptions } = useQuery({
    queryKey: ["admin-assistant-tool-events"],
    queryFn: () => apiClient.getAdminEvents({ page: 1, page_size: 20, status: "published" }),
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
  const reportSummaryMutation = useMutation({
    mutationFn: (reportId: string) => apiClient.generateAdminReportSummary(reportId),
    onError: (err: unknown) => {
      showAdminToast(err instanceof Error ? err.message : "生成失败", "error");
    },
  });
  const weeklyReportMutation = useMutation({
    mutationFn: (days: number) => apiClient.generateAdminWeeklyReport(days),
    onError: (err: unknown) => {
      showAdminToast(err instanceof Error ? err.message : "生成失败", "error");
    },
  });
  const creatorRecommendationMutation = useMutation({
    mutationFn: (userId: string) => apiClient.generateAdminCreatorRecommendation(userId),
    onError: (err: unknown) => {
      showAdminToast(err instanceof Error ? err.message : "生成失败", "error");
    },
  });
  const eventCopyMutation = useMutation({
    mutationFn: (eventId: string) => apiClient.generateAdminEventCopy(eventId),
    onError: (err: unknown) => {
      showAdminToast(err instanceof Error ? err.message : "生成失败", "error");
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

  const helpfulTotal =
    (overview?.overview.feedback_helpful || 0) +
    (overview?.overview.feedback_unhelpful || 0);
  const helpfulRate = helpfulTotal
    ? `${Math.round(
        ((overview?.overview.feedback_helpful || 0) / helpfulTotal) * 100,
      )}%`
    : "—";

  async function copyText(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      showAdminToast("已复制到剪贴板", "success");
    } catch {
      showAdminToast("复制失败，请手动复制", "error");
    }
  }

  async function runDiagnostic() {
    const prompt = diagnosticPrompt.trim();
    if (!prompt || diagnosticLoading) return;

    setDiagnosticLoading(true);
    setDiagnosticMeta(null);
    setDiagnosticReply("");
    setDiagnosticError("");

    try {
      const response = await fetch(`${API_BASE_URL}/assistant/chat/stream`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          messages: [{ role: "user", content: prompt }],
        }),
      });

      if (!response.ok || !response.body) {
        throw new Error("诊断请求失败，请检查助手是否已启用");
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      for (;;) {
        const { value, done } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        let marker = buffer.indexOf("\n\n");
        for (; marker !== -1; marker = buffer.indexOf("\n\n")) {
          const block = buffer.slice(0, marker);
          buffer = buffer.slice(marker + 2);

          const parsed = parseSSEBlock(block);
          if (!parsed) continue;

          const payload = JSON.parse(parsed.data);
          switch (parsed.event) {
            case "meta":
              setDiagnosticMeta(payload as AssistantMeta);
              break;
            case "token":
              setDiagnosticReply((prev) => `${prev}${payload.content ?? ""}`);
              break;
            case "error":
              setDiagnosticError(payload.message || "诊断请求失败");
              break;
            default:
              break;
          }
        }
      }
    } catch (err) {
      setDiagnosticError(err instanceof Error ? err.message : "诊断请求失败");
    } finally {
      setDiagnosticLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="AI Operations"
        title="AI 助手与内部工具"
        description="在这里管理站内 AI 助手的人设、检索来源，并为产品和运营团队生成可直接复用的 AI 内容。"
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

      <div className="grid gap-4 md:grid-cols-4">
        <AdminMetricCard
          label="索引文档"
          value={String(overview?.overview.indexed_documents ?? 0)}
          hint="当前检索知识库中的总块数"
          icon={Sparkles}
          tone="brand"
        />
        <AdminMetricCard
          label="最近同步"
          value={
            overview?.metrics.last_index_synced_at
              ? formatDateTime(overview.metrics.last_index_synced_at)
              : "未同步"
          }
          hint="后台索引最近一次刷新时间"
          icon={Save}
          tone="default"
        />
        <AdminMetricCard
          label="有帮助率"
          value={helpfulRate}
          hint="用户对回复的有帮助反馈占比"
          icon={Bot}
          tone="success"
        />
        <AdminMetricCard
          label="Fallback 次数"
          value={String(overview?.metrics.fallback_total ?? 0)}
          hint="模型未用上，直接退回站内检索的次数"
          icon={Loader2}
          tone={(overview?.metrics.fallback_total ?? 0) > 0 ? "warning" : "default"}
        />
        <AdminMetricCard
          label="媒体缓存"
          value={String(overview?.overview.media_cache_entries ?? 0)}
          hint="已缓存的图片理解结果数"
          icon={FileText}
          tone="default"
        />
        <AdminMetricCard
          label="图片分析次数"
          value={String(overview?.metrics.multimodal_requests_total ?? 0)}
          hint="多模态图片理解总调用次数"
          icon={Sparkles}
          tone="brand"
        />
        <AdminMetricCard
          label="视觉熔断状态"
          value={overview?.metrics.vision_circuit_state || "closed"}
          hint="视觉模型链路当前熔断状态"
          icon={Wand2}
          tone={
            (overview?.metrics.vision_circuit_state || "closed") === "open"
              ? "warning"
              : "success"
          }
        />
      </div>

      <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
        <CardHeader>
          <CardTitle className="text-lg">检索与索引概览</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-6 md:grid-cols-2">
          <div className="space-y-3 text-sm text-slate-600">
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>Embedding</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.overview.embedding_configured
                  ? overview?.overview.embedding_model || "已配置"
                  : "本地 fallback"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>Vision</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.overview.vision_configured
                  ? overview?.overview.vision_model || "已配置"
                  : "本地 fallback"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>检索块来源</span>
              <span className="text-right font-medium text-slate-900">
                {formatSourceCounts(overview?.overview.documents_by_source)}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>检索上限</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.overview.retrieval_limit ?? "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>向量扫描上限</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.overview.vector_scan_limit ?? "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4">
              <span>同步间隔</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.overview.sync_interval_sec
                  ? `${overview.overview.sync_interval_sec}s`
                  : "—"}
              </span>
            </div>
          </div>

          <div className="space-y-3 text-sm text-slate-600">
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>最近检索耗时</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.metrics.last_retrieval_duration_ms
                  ? `${overview.metrics.last_retrieval_duration_ms.toFixed(1)} ms`
                  : "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>最近首 token</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.metrics.last_first_token_latency_ms
                  ? `${overview.metrics.last_first_token_latency_ms.toFixed(1)} ms`
                  : "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>最近召回块数</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.metrics.last_retrieved_documents ?? "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>最近同步块数</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.metrics.last_indexed_documents ?? "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>最近图片分析耗时</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.metrics.last_multimodal_latency_ms
                  ? `${overview.metrics.last_multimodal_latency_ms.toFixed(1)} ms`
                  : "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>图片缓存命中</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.metrics.multimodal_cache_hits ?? "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>图片降级次数</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.metrics.multimodal_fallback_total ?? "—"}
              </span>
            </div>
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-3">
              <span>图片重试次数</span>
              <span className="text-right font-medium text-slate-900">
                {overview?.metrics.multimodal_retry_total ?? "—"}
              </span>
            </div>
            <div className="space-y-1">
              <p className="text-xs font-medium uppercase tracking-[0.18em] text-slate-500">
                最近同步错误
              </p>
              <p className="rounded-2xl bg-slate-50 px-3 py-2 text-sm text-slate-600">
                {overview?.metrics.last_index_error || "无"}
              </p>
            </div>
            <div className="space-y-1">
              <p className="text-xs font-medium uppercase tracking-[0.18em] text-slate-500">
                最近图片分析错误
              </p>
              <p className="rounded-2xl bg-slate-50 px-3 py-2 text-sm text-slate-600">
                {overview?.metrics.last_multimodal_error || "无"}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
        <CardHeader>
          <CardTitle className="text-lg">内部 AI 工具</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-6 xl:grid-cols-2">
          <div className="space-y-4 rounded-3xl border border-slate-200 bg-white p-5">
            <div className="flex items-center gap-3">
              <ShieldAlert className="h-5 w-5 text-rose-500" />
              <div>
                <p className="text-sm font-semibold text-slate-900">举报摘要与处理建议</p>
                <p className="text-xs text-slate-500">面向审核和治理场景，输出风险判断、建议动作和回复草稿。</p>
              </div>
            </div>
            <select
              value={selectedReportId}
              onChange={(e) => setSelectedReportId(e.target.value)}
              className="w-full rounded-2xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900"
            >
              <option value="">选择待处理举报</option>
              {(reportOptions?.reports || []).map((item: AdminReport) => (
                <option key={item.id} value={item.id}>
                  [{item.target_type}] {item.reason} · {item.reporter_username || item.reporter_id}
                </option>
              ))}
            </select>
            <Button
              type="button"
              disabled={!selectedReportId || reportSummaryMutation.isPending}
              onClick={() => reportSummaryMutation.mutate(selectedReportId)}
              className="bg-slate-950 text-white hover:bg-slate-800"
            >
              {reportSummaryMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Wand2 className="mr-2 h-4 w-4" />
              )}
              生成举报摘要
            </Button>
            <ToolResultCard
              result={reportSummaryMutation.data}
              loading={reportSummaryMutation.isPending}
              empty="选择一条待处理举报后，这里会生成结构化摘要和说明文案。"
              onCopy={copyText}
            />
          </div>

          <div className="space-y-4 rounded-3xl border border-slate-200 bg-white p-5">
            <div className="flex items-center gap-3">
              <FileText className="h-5 w-5 text-cyan-500" />
              <div>
                <p className="text-sm font-semibold text-slate-900">社区运营周报生成</p>
                <p className="text-xs text-slate-500">汇总指标、热门内容、圈子活动观察与跟进建议。</p>
              </div>
            </div>
            <div className="flex gap-3">
              <Input
                type="number"
                min={3}
                max={30}
                value={weeklyDays}
                onChange={(e) => setWeeklyDays(Number(e.target.value) || 7)}
              />
              <Button
                type="button"
                disabled={weeklyReportMutation.isPending}
                onClick={() => weeklyReportMutation.mutate(weeklyDays)}
                className="bg-slate-950 text-white hover:bg-slate-800"
              >
                {weeklyReportMutation.isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Sparkles className="mr-2 h-4 w-4" />
                )}
                生成周报
              </Button>
            </div>
            <ToolResultCard
              result={weeklyReportMutation.data}
              loading={weeklyReportMutation.isPending}
              empty="输入统计周期后，这里会生成可直接发给产品或运营同事的周报正文。"
              onCopy={copyText}
            />
          </div>

          <div className="space-y-4 rounded-3xl border border-slate-200 bg-white p-5">
            <div className="flex items-center gap-3">
              <Bot className="h-5 w-5 text-amber-500" />
              <div>
                <p className="text-sm font-semibold text-slate-900">创作者推荐理由生成</p>
                <p className="text-xs text-slate-500">为首页推荐位、专题或活动联动生成创作者推荐理由。</p>
              </div>
            </div>
            <select
              value={selectedCreatorId}
              onChange={(e) => setSelectedCreatorId(e.target.value)}
              className="w-full rounded-2xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900"
            >
              <option value="">选择创作者</option>
              {(creatorOptions?.users || []).map((item: AdminUser) => (
                <option key={item.id} value={item.id}>
                  @{item.username} · {item.furry_name || item.nickname || item.email}
                </option>
              ))}
            </select>
            <Button
              type="button"
              disabled={!selectedCreatorId || creatorRecommendationMutation.isPending}
              onClick={() => creatorRecommendationMutation.mutate(selectedCreatorId)}
              className="bg-slate-950 text-white hover:bg-slate-800"
            >
              {creatorRecommendationMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Wand2 className="mr-2 h-4 w-4" />
              )}
              生成推荐理由
            </Button>
            <ToolResultCard
              result={creatorRecommendationMutation.data}
              loading={creatorRecommendationMutation.isPending}
              empty="选择一个创作者后，这里会生成推荐理由和两版推荐文案。"
              onCopy={copyText}
            />
          </div>

          <div className="space-y-4 rounded-3xl border border-slate-200 bg-white p-5">
            <div className="flex items-center gap-3">
              <Megaphone className="h-5 w-5 text-emerald-500" />
              <div>
                <p className="text-sm font-semibold text-slate-900">活动文案生成</p>
                <p className="text-xs text-slate-500">为活动卡片、详情页或推送生成短文案和长文案。</p>
              </div>
            </div>
            <select
              value={selectedEventId}
              onChange={(e) => setSelectedEventId(e.target.value)}
              className="w-full rounded-2xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900"
            >
              <option value="">选择活动</option>
              {(eventOptions?.events || []).map((item: AdminEvent) => (
                <option key={item.id} value={item.id}>
                  {item.title} · {new Date(item.start_time).toLocaleDateString("zh-CN")}
                </option>
              ))}
            </select>
            <Button
              type="button"
              disabled={!selectedEventId || eventCopyMutation.isPending}
              onClick={() => eventCopyMutation.mutate(selectedEventId)}
              className="bg-slate-950 text-white hover:bg-slate-800"
            >
              {eventCopyMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Wand2 className="mr-2 h-4 w-4" />
              )}
              生成活动文案
            </Button>
            <ToolResultCard
              result={eventCopyMutation.data}
              loading={eventCopyMutation.isPending}
              empty="选择一个活动后，这里会生成宣传卖点、检查清单和两版活动文案。"
              onCopy={copyText}
            />
          </div>
        </CardContent>
      </Card>

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

          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardHeader>
              <CardTitle className="text-lg">诊断测试</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">测试问题</label>
                <Textarea
                  value={diagnosticPrompt}
                  onChange={(e) => setDiagnosticPrompt(e.target.value)}
                  rows={4}
                  placeholder="例如：推荐几个适合新人的圈子"
                />
                <p className="text-xs text-slate-500">
                  这里会以匿名请求调用助手，不写入后台管理员自己的会话历史。
                </p>
              </div>

              <div className="flex flex-wrap items-center gap-3">
                <Button
                  type="button"
                  onClick={() => void runDiagnostic()}
                  disabled={diagnosticLoading || !diagnosticPrompt.trim()}
                  className="bg-slate-950 text-white hover:bg-slate-800"
                >
                  {diagnosticLoading ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Sparkles className="mr-2 h-4 w-4" />
                  )}
                  运行诊断
                </Button>
                {diagnosticMeta && (
                  <>
                    <span className="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-700">
                      意图：{diagnosticMeta.intent_label || diagnosticMeta.intent || "综合导览"}
                    </span>
                    <span className="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-700">
                      模式：{diagnosticMeta.fallback ? "站内检索 fallback" : `${diagnosticMeta.provider || "AI"} + 检索`}
                    </span>
                    {formatSourceCounts(diagnosticMeta.source_counts) && (
                      <span className="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-700">
                        召回：{formatSourceCounts(diagnosticMeta.source_counts)}
                      </span>
                    )}
                  </>
                )}
              </div>

              {diagnosticError && (
                <div className="rounded-2xl bg-rose-50 px-4 py-3 text-sm text-rose-700">
                  {diagnosticError}
                </div>
              )}

              <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                <p className="text-xs font-medium uppercase tracking-[0.18em] text-slate-500">
                  回复预览
                </p>
                <div className="mt-3 whitespace-pre-wrap text-sm leading-6 text-slate-700">
                  {diagnosticReply || "运行诊断后，这里会显示当前问法下的实际回复。"}
                </div>
              </div>
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
