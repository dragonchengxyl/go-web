"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  ArrowRight,
  Bot,
  Check,
  Clock3,
  Loader2,
  ShieldCheck,
  Sparkles,
  Square,
  Workflow,
  X,
} from "lucide-react";
import {
  AgentApproval,
  AgentArtifact,
  AgentRun,
  AgentRunEvent,
  AgentStep,
  AgentToolCall,
  apiClient,
} from "@/lib/api-client";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  AgentContextChips,
  AgentEmptyState,
  AgentMetricCard,
  AgentProgressBar,
  AgentScenarioBadge,
  AgentSectionHeader,
  AgentStatusBadge,
  AgentStepStatusBadge,
  AgentSurface,
  calculateRunProgress,
  formatAgentDateTime,
  formatAgentDuration,
  formatAgentElapsed,
  getRunStatusNarrative,
} from "@/components/agent/agent-ui";
import { writeStoredPostDraft } from "@/lib/post-draft";
import { cn } from "@/lib/utils";

const APPROVAL_STATUS_LABELS: Record<AgentApproval["status"], string> = {
  pending: "待确认",
  approved: "已通过",
  rejected: "已跳过",
};

function buildReplayHref(run: AgentRun) {
  const params = new URLSearchParams();
  params.set("scenario", run.scenario);
  params.set("title", run.title);
  params.set("goal", run.goal);

  Object.entries(run.context_snapshot || {}).forEach(([key, value]) => {
    if (value == null) return;
    params.set(key, String(value));
  });

  return `/agent?${params.toString()}`;
}

function StreamBadge({ mode }: { mode: "connecting" | "live" | "fallback" }) {
  const className =
    mode === "live"
      ? "border-emerald-300/30 bg-emerald-300/10 text-emerald-100"
      : mode === "fallback"
        ? "border-amber-300/30 bg-amber-300/10 text-amber-100"
        : "border-cyan-300/30 bg-cyan-300/10 text-cyan-100";

  const label =
    mode === "live"
      ? "已同步"
      : mode === "fallback"
        ? "更新中"
        : "同步中";

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.14em]",
        className,
      )}
    >
      <span className={cn("h-1.5 w-1.5 rounded-full bg-current", mode !== "fallback" ? "animate-pulse" : "")} />
      {label}
    </span>
  );
}

function TimelineStepCard({ step }: { step: AgentStep }) {
  return (
    <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-3">
            <span className="rounded-full border border-white/10 bg-white/8 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-white/62">
              步骤 {step.step_index}
            </span>
            <p className="text-sm font-semibold text-white">{step.title}</p>
          </div>
        </div>
        <AgentStepStatusBadge status={step.status} />
      </div>

      {step.summary ? (
        <p className="mt-4 text-sm leading-6 text-white/64">{step.summary}</p>
      ) : null}
      {step.error_text ? (
        <div className="mt-4 rounded-[18px] border border-rose-300/20 bg-rose-300/8 px-3 py-3 text-sm leading-6 text-rose-100">
          {step.error_text}
        </div>
      ) : null}

      <div className="mt-4 flex flex-wrap gap-3 text-xs text-white/40">
        <span>开始 {formatAgentDateTime(step.started_at || step.created_at)}</span>
        <span>结束 {formatAgentDateTime(step.completed_at)}</span>
        <span>耗时 {formatAgentDuration(step.started_at || step.created_at, step.completed_at)}</span>
      </div>
    </div>
  );
}

function ToolTraceCard({ toolCall }: { toolCall: AgentToolCall }) {
  return (
    <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-sm font-semibold text-white">{toolCall.tool_name}</p>
          <p className="mt-2 text-xs text-white/42">工具调用记录</p>
        </div>
        <AgentStepStatusBadge status={toolCall.status} />
      </div>

      <div className="mt-4 flex flex-wrap gap-3 text-xs text-white/40">
        <span>开始 {formatAgentDateTime(toolCall.started_at || toolCall.created_at)}</span>
        <span>结束 {formatAgentDateTime(toolCall.completed_at)}</span>
        <span>
          耗时 {formatAgentDuration(toolCall.started_at || toolCall.created_at, toolCall.completed_at)}
        </span>
      </div>

      {toolCall.error_text ? (
        <div className="mt-4 rounded-[18px] border border-rose-300/20 bg-rose-300/8 px-3 py-3 text-sm leading-6 text-rose-100">
          {toolCall.error_text}
        </div>
      ) : null}

      {toolCall.output_data ? (
        <details className="mt-4 overflow-hidden rounded-[18px] border border-white/10 bg-slate-950/40">
          <summary className="cursor-pointer list-none px-3 py-3 text-sm font-medium text-white/78">
            查看工具输出
          </summary>
          <pre className="overflow-x-auto border-t border-white/10 px-3 py-3 text-xs leading-6 text-white/62">
            {JSON.stringify(toolCall.output_data, null, 2)}
          </pre>
        </details>
      ) : null}
    </div>
  );
}

function ArtifactCard({
  artifact,
  compact = false,
}: {
  artifact: AgentArtifact;
  compact?: boolean;
}) {
  const content = artifact.content?.trim();
  const preview = compact && content && content.length > 280
    ? `${content.slice(0, 280).trim()}...`
    : content;

  return (
    <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-sm font-semibold text-white">{artifact.title}</p>
        </div>
        <span className="rounded-full border border-white/10 bg-white/8 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.14em] text-white/58">
          结果
        </span>
      </div>

      {preview ? (
        <div className="mt-4 whitespace-pre-wrap text-sm leading-6 text-white/68">
          {preview}
        </div>
      ) : null}

      {!preview && artifact.kind === "draft_patch" ? (
        <p className="mt-4 text-sm leading-6 text-white/58">
          这项结果会在确认后应用到草稿中。
        </p>
      ) : null}

      {!preview && artifact.structured_data ? (
        <details className="mt-4 overflow-hidden rounded-[18px] border border-white/10 bg-slate-950/40">
          <summary className="cursor-pointer list-none px-3 py-3 text-sm font-medium text-white/78">
            查看结构化数据
          </summary>
          <pre className="overflow-x-auto border-t border-white/10 px-3 py-3 text-xs leading-6 text-white/62">
            {JSON.stringify(artifact.structured_data, null, 2)}
          </pre>
        </details>
      ) : null}
    </div>
  );
}

function ApprovalCard({
  approval,
  onApprove,
  onReject,
  isPending,
}: {
  approval: AgentApproval;
  onApprove: (approvalId: string) => void;
  onReject: (approvalId: string) => void;
  isPending: boolean;
}) {
  const isActionable = approval.status === "pending";

  return (
    <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
      <div className="flex flex-wrap items-center gap-2">
        <span className="rounded-full border border-white/10 bg-white/8 px-2.5 py-1 text-[10px] font-semibold tracking-[0.14em] text-white/60">
          {APPROVAL_STATUS_LABELS[approval.status]}
        </span>
      </div>
      <p className="mt-4 text-sm font-semibold text-white">{approval.title}</p>
      <p className="mt-2 text-sm leading-6 text-white/58">
        {approval.status === "pending"
          ? "这一步需要你做最终决定。确认后会继续应用这轮结果。"
          : approval.status === "approved"
            ? "这一轮审批已经通过，结果已经处理完成。"
            : "你保留了这轮结果，但不会自动执行回填。"}
      </p>

      {isActionable ? (
        <div className="mt-4 flex flex-wrap gap-2">
          <Button
            size="sm"
            onClick={() => onApprove(approval.id)}
            disabled={isPending}
            className="rounded-full bg-white text-slate-950 hover:bg-white/90"
          >
            {isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Check className="mr-2 h-4 w-4" />
            )}
            批准并继续
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => onReject(approval.id)}
            disabled={isPending}
            className="rounded-full border-white/16 bg-white/6 text-white hover:bg-white/10 hover:text-white"
          >
            <X className="mr-2 h-4 w-4" />
            暂不执行
          </Button>
        </div>
      ) : null}
    </div>
  );
}

export default function AgentRunDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { isLoggedIn } = useAuth();
  const [streamMode, setStreamMode] = useState<"connecting" | "live" | "fallback">("connecting");
  const [liveEvent, setLiveEvent] = useState<AgentRunEvent | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ["agent-run", params.id],
    queryFn: () => apiClient.getAgentRun(params.id),
    enabled: isLoggedIn && !!params.id,
    refetchInterval: (query) => {
      const status = (query.state.data as { run?: AgentRun } | undefined)?.run?.status;
      return status === "queued" || status === "running" ? 1200 : false;
    },
  });

  const cancelMutation = useMutation({
    mutationFn: () => apiClient.cancelAgentRun(params.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agent-run", params.id] });
      queryClient.invalidateQueries({ queryKey: ["agent-runs"] });
    },
  });

  const approvalMutation = useMutation({
    mutationFn: ({
      approvalId,
      decision,
    }: {
      approvalId: string;
      decision: "approved" | "rejected";
    }) => apiClient.decideAgentApproval(params.id, { approval_id: approvalId, decision }),
    onSuccess: (result, variables) => {
      queryClient.setQueryData(["agent-run", params.id], result.detail);
      queryClient.invalidateQueries({ queryKey: ["agent-runs"] });

      if (variables.decision === "approved" && result.apply_payload) {
        writeStoredPostDraft({
          title: String(result.apply_payload.title || ""),
          content: String(result.apply_payload.content || ""),
          tags: String(result.apply_payload.tags || ""),
          visibility: String(result.apply_payload.visibility || "public"),
          source: "agent",
          agent_run_id: params.id,
        });
        router.push("/posts/create?agent_applied=1");
      }
    },
  });

  useEffect(() => {
    if (!isLoggedIn || !params.id) return;
    const controller = new AbortController();
    setStreamMode("connecting");

    void apiClient
      .streamAgentRun(params.id, {
        signal: controller.signal,
        onSnapshot: (detail) => {
          setStreamMode("live");
          queryClient.setQueryData(["agent-run", params.id], detail);
          queryClient.invalidateQueries({ queryKey: ["agent-runs"] });
        },
        onUpdate: (event) => {
          setStreamMode("live");
          setLiveEvent(event);
        },
        onDone: () => {
          queryClient.invalidateQueries({ queryKey: ["agent-runs"] });
        },
        onError: () => {
          setStreamMode("fallback");
        },
      })
      .catch(() => {
        setStreamMode("fallback");
      });

    return () => controller.abort();
  }, [isLoggedIn, params.id, queryClient]);

  const run = data?.run;
  const steps = data?.steps || [];
  const toolCalls = data?.tool_calls || [];
  const approvals = data?.approvals || [];
  const artifacts = data?.artifacts || [];
  const pendingApproval = approvals.find((approval) => approval.status === "pending");
  const progress = run ? calculateRunProgress(run.status, steps) : 0;
  const activeStep = steps.find((step) => step.status === "running");
  const replayHref = run ? buildReplayHref(run) : "/agent";

  const artifactPreview = artifacts.slice(0, 2);
  const canCancel = Boolean(
    run && run.status !== "completed" && run.status !== "failed" && run.status !== "cancelled",
  );

  if (!isLoggedIn) {
    return (
      <div className="agent-page-shell relative min-h-screen overflow-hidden">
        <div className="agent-page-grid pointer-events-none absolute inset-0" />
        <div className="mx-auto max-w-5xl px-4 pb-16 pt-24">
          <AgentSurface className="px-6 py-8 sm:px-8">
            <p className="text-sm text-white/62">请先登录后再查看 Agent 任务详情。</p>
          </AgentSurface>
        </div>
      </div>
    );
  }

  return (
    <div className="agent-page-shell relative min-h-screen overflow-hidden">
      <div className="agent-page-grid pointer-events-none absolute inset-0" />
      <div className="pointer-events-none absolute left-[-8rem] top-28 h-72 w-72 rounded-full bg-cyan-400/18 blur-3xl" />
      <div className="pointer-events-none absolute right-[-6rem] top-24 h-80 w-80 rounded-full bg-orange-400/14 blur-3xl" />
      <div className="pointer-events-none absolute bottom-0 left-1/2 h-96 w-[38rem] -translate-x-1/2 rounded-full bg-sky-500/10 blur-3xl" />

      <div className="mx-auto max-w-7xl px-4 pb-16 pt-24">
        <div className="grid gap-6 xl:grid-cols-[1.08fr_0.92fr]">
          <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Button
                asChild
                variant="outline"
                className="rounded-full border-white/16 bg-white/6 text-white hover:bg-white/10 hover:text-white"
              >
                  <Link href="/agent">
                    <ArrowLeft className="mr-2 h-4 w-4" />
                    返回工作台
                  </Link>
                </Button>

              {run ? (
                <Button
                  asChild
                  variant="outline"
                  className="rounded-full border-white/16 bg-white/6 text-white hover:bg-white/10 hover:text-white"
                >
                  <Link href={replayHref}>
                    基于本次再开一轮
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </Link>
                </Button>
              ) : null}
            </div>

            {isLoading ? (
              <div className="mt-8 flex items-center gap-2 text-sm text-white/60">
                <Loader2 className="h-4 w-4 animate-spin" />
                正在载入任务详情
              </div>
            ) : null}

            {error ? (
              <div className="mt-8 rounded-[24px] border border-rose-300/20 bg-rose-300/8 px-4 py-4 text-sm text-rose-100">
                {error instanceof Error ? error.message : "读取任务详情失败"}
              </div>
            ) : null}

            {run ? (
              <>
                <div className="mt-8 flex flex-wrap items-center gap-2">
                  <AgentStatusBadge status={run.status} pulse />
                  <AgentScenarioBadge scenario={run.scenario} />
                  <StreamBadge mode={streamMode} />
                </div>

                <h1 className="mt-5 text-4xl font-semibold tracking-tight text-white sm:text-[2.9rem]">
                  {run.title}
                </h1>
                <p className="mt-5 max-w-3xl whitespace-pre-wrap text-base leading-8 text-white/68">
                  {run.goal}
                </p>

                <div className="mt-6 flex flex-wrap gap-3 text-xs text-white/44">
                  <span>任务编号 {run.id}</span>
                  <span>创建于 {formatAgentDateTime(run.created_at)}</span>
                  <span>最近更新 {formatAgentDateTime(run.updated_at)}</span>
                  <span>尝试 {run.attempt_count}/{run.max_attempts}</span>
                </div>

                <div className="mt-6 rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <p className="text-sm font-semibold text-white">当前进度</p>
                    <span className="text-sm text-white/62">{progress}%</span>
                  </div>
                  <AgentProgressBar value={progress} className="mt-3" />
                  <p className="mt-3 text-sm leading-6 text-white/58">{getRunStatusNarrative(run.status)}</p>
                  {activeStep ? (
                    <p className="mt-2 text-sm text-cyan-100">
                      当前步骤：{activeStep.title}
                    </p>
                  ) : liveEvent?.summary ? (
                    <p className="mt-2 text-sm text-cyan-100">{liveEvent.summary}</p>
                  ) : null}
                </div>

                {run.latest_summary ? (
                  <div className="mt-6 rounded-[24px] border border-cyan-300/18 bg-cyan-300/8 px-4 py-4">
                    <p className="text-sm font-semibold text-white">最新摘要</p>
                    <p className="mt-3 text-sm leading-7 text-white/68">{run.latest_summary}</p>
                  </div>
                ) : null}

                {run.last_error ? (
                  <div className="mt-6 rounded-[24px] border border-rose-300/20 bg-rose-300/8 px-4 py-4">
                    <p className="text-sm font-semibold text-rose-100">最近一次错误</p>
                    <p className="mt-3 text-sm leading-7 text-rose-50/90">{run.last_error}</p>
                    <p className="mt-3 text-xs text-white/40">
                      错误时间 {formatAgentDateTime(run.last_error_at)}
                      {run.next_retry_at ? ` · 下次重试 ${formatAgentDateTime(run.next_retry_at)}` : ""}
                    </p>
                  </div>
                ) : null}
              </>
            ) : null}
          </AgentSurface>

          <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
            <AgentSectionHeader
              eyebrow="任务状态"
              title="运行概览"
              description="集中查看当前进度、时长和下一步操作。"
            />

            {run ? (
              <>
                <div className="mt-6 grid gap-4 sm:grid-cols-2">
                  <AgentMetricCard
                    icon={Clock3}
                    label="已运行"
                    value={formatAgentElapsed(run.started_at || run.created_at, run.completed_at)}
                    meta="从任务开始到现在或结束。"
                    tone="cyan"
                  />
                  <AgentMetricCard
                    icon={Workflow}
                    label="步骤数"
                    value={steps.length}
                    meta="包含已完成、运行中和失败步骤。"
                    tone="amber"
                  />
                  <AgentMetricCard
                    icon={Bot}
                    label="工具轨迹"
                    value={toolCalls.length}
                    meta="本轮任务使用过的工具次数。"
                    tone="slate"
                  />
                  <AgentMetricCard
                    icon={Sparkles}
                    label="产物数"
                    value={artifacts.length}
                    meta="当前可查看的结果数量。"
                    tone="emerald"
                  />
                </div>

                <div className="mt-6 rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                  <p className="text-sm font-semibold text-white">状态更新</p>
                  <p className="mt-3 text-sm leading-6 text-white/58">
                    {streamMode === "live"
                      ? "当前状态会持续更新。"
                      : streamMode === "fallback"
                        ? "当前状态会继续更新，可能会有轻微延迟。"
                        : "正在同步最新状态。"}
                  </p>
                  {liveEvent?.summary ? (
                    <div className="mt-4 rounded-[18px] border border-white/10 bg-slate-950/40 px-3 py-3 text-sm text-white/70">
                      {liveEvent.summary}
                    </div>
                  ) : null}
                  {cancelMutation.error ? (
                    <p className="mt-4 text-sm text-rose-100">
                      {cancelMutation.error instanceof Error
                        ? cancelMutation.error.message
                        : "取消任务失败"}
                    </p>
                  ) : null}
                </div>

                <div className="mt-6 flex flex-wrap gap-3">
                  {canCancel ? (
                    <Button
                      variant="outline"
                      onClick={() => cancelMutation.mutate()}
                      disabled={cancelMutation.isPending}
                      className="rounded-full border-white/16 bg-white/6 text-white hover:bg-white/10 hover:text-white"
                    >
                      {cancelMutation.isPending ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : (
                        <Square className="mr-2 h-4 w-4" />
                      )}
                      取消本轮任务
                    </Button>
                  ) : null}
                  <Button
                    asChild
                    variant="outline"
                    className="rounded-full border-white/16 bg-white/6 text-white hover:bg-white/10 hover:text-white"
                  >
                    <Link href={replayHref}>把当前配置带回启动页</Link>
                  </Button>
                </div>
              </>
            ) : null}
          </AgentSurface>
        </div>

        <div className="mt-8 grid gap-6 xl:grid-cols-[1.05fr_0.95fr]">
          <div>
            <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
              <Tabs defaultValue="overview">
                <TabsList className="h-auto w-full justify-start gap-2 rounded-full bg-white/6 p-1">
                  <TabsTrigger
                    value="overview"
                    className="rounded-full px-4 py-2 text-xs text-white/68 data-[state=active]:bg-white data-[state=active]:text-slate-950"
                  >
                    任务概览
                  </TabsTrigger>
                  <TabsTrigger
                    value="operations"
                    className="rounded-full px-4 py-2 text-xs text-white/68 data-[state=active]:bg-white data-[state=active]:text-slate-950"
                  >
                    执行过程
                  </TabsTrigger>
                  <TabsTrigger
                    value="results"
                    className="rounded-full px-4 py-2 text-xs text-white/68 data-[state=active]:bg-white data-[state=active]:text-slate-950"
                  >
                    交付结果
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="overview" className="mt-6 space-y-6">
                  <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                    <p className="text-sm font-semibold text-white">任务目标</p>
                    <p className="mt-3 text-sm leading-7 text-white/66">
                      {run?.goal || "—"}
                    </p>
                  </div>

                  {artifactPreview.length > 0 ? (
                    <div className="space-y-4">
                      <div className="flex items-center justify-between gap-3">
                        <p className="text-sm font-semibold text-white">结果预览</p>
                        <span className="text-xs text-white/42">前 {artifactPreview.length} 项</span>
                      </div>
                      {artifactPreview.map((artifact) => (
                        <ArtifactCard key={artifact.id} artifact={artifact} compact />
                      ))}
                    </div>
                  ) : (
                    <AgentEmptyState
                      icon={Sparkles}
                      title="结果还在路上"
                      description="结构化产物生成后会先出现在这里，完整版会归档到结果页签。"
                    />
                  )}
                </TabsContent>

                <TabsContent value="operations" className="mt-6 space-y-6">
                  <div>
                    <div className="mb-4 flex items-center justify-between gap-3">
                      <p className="text-sm font-semibold text-white">步骤时间线</p>
                      <span className="text-xs text-white/42">{steps.length} 条</span>
                    </div>
                    {steps.length ? (
                      <div className="space-y-4">
                        {steps.map((step) => (
                          <TimelineStepCard key={step.id} step={step} />
                        ))}
                      </div>
                    ) : (
                      <AgentEmptyState
                        icon={Workflow}
                        title="还没有步骤记录"
                        description="任务开始推进后，时间线会从这里展开。"
                      />
                    )}
                  </div>

                  <div>
                    <div className="mb-4 flex items-center justify-between gap-3">
                      <p className="text-sm font-semibold text-white">工具轨迹</p>
                      <span className="text-xs text-white/42">{toolCalls.length} 次</span>
                    </div>
                    {toolCalls.length ? (
                      <div className="space-y-4">
                        {toolCalls.map((toolCall) => (
                          <ToolTraceCard key={toolCall.id} toolCall={toolCall} />
                        ))}
                      </div>
                    ) : (
                      <AgentEmptyState
                        icon={Bot}
                        title="还没有工具调用"
                        description="如果这一轮任务需要读取外部信息或生成中间产物，工具轨迹会记录在这里。"
                      />
                    )}
                  </div>
                </TabsContent>

                <TabsContent value="results" className="mt-6 space-y-6">
                  <div>
                    <div className="mb-4 flex items-center justify-between gap-3">
                      <p className="text-sm font-semibold text-white">结构化产物</p>
                      <span className="text-xs text-white/42">{artifacts.length} 项</span>
                    </div>
                    {artifacts.length ? (
                      <div className="space-y-4">
                        {artifacts.map((artifact) => (
                          <ArtifactCard key={artifact.id} artifact={artifact} />
                        ))}
                      </div>
                    ) : (
                      <AgentEmptyState
                        icon={Sparkles}
                        title="还没有产物"
                        description="Agent 生成可交付内容后，结果会集中出现在这里。"
                      />
                    )}
                  </div>

                  <div>
                    <div className="mb-4 flex items-center justify-between gap-3">
                      <p className="text-sm font-semibold text-white">审批记录</p>
                      <span className="text-xs text-white/42">{approvals.length} 条</span>
                    </div>
                    {approvals.length ? (
                      <div className="space-y-4">
                        {approvals.map((approval) => (
                          <ApprovalCard
                            key={approval.id}
                            approval={approval}
                            isPending={approvalMutation.isPending}
                            onApprove={(approvalId) =>
                              approvalMutation.mutate({ approvalId, decision: "approved" })
                            }
                            onReject={(approvalId) =>
                              approvalMutation.mutate({ approvalId, decision: "rejected" })
                            }
                          />
                        ))}
                      </div>
                    ) : (
                      <AgentEmptyState
                        icon={ShieldCheck}
                        title="暂时没有审批节点"
                        description="如果这轮任务需要你确认结果，这里会优先出现相应动作。"
                      />
                    )}
                  </div>
                </TabsContent>
              </Tabs>
            </AgentSurface>
          </div>

          <div className="space-y-6">
            <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
              <AgentSectionHeader
                eyebrow="下一步"
                title="下一步动作"
                description="需要你确认的结果会优先出现在这里。"
              />

              <div className="mt-6">
                {pendingApproval ? (
                  <ApprovalCard
                    approval={pendingApproval}
                    isPending={approvalMutation.isPending}
                    onApprove={(approvalId) =>
                      approvalMutation.mutate({ approvalId, decision: "approved" })
                    }
                    onReject={(approvalId) =>
                      approvalMutation.mutate({ approvalId, decision: "rejected" })
                    }
                  />
                ) : approvals.length ? (
                  <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                    <p className="text-sm font-semibold text-white">当前没有待处理审批</p>
                    <p className="mt-2 text-sm leading-6 text-white/56">
                      最近一次审批已经处理完成。你可以继续查看执行记录，或基于本次配置重新发起一轮任务。
                    </p>
                  </div>
                ) : (
                  <AgentEmptyState
                    icon={ShieldCheck}
                    title="暂时没有需要你拍板的动作"
                    description="当 Agent 需要用户确认结果是否回填时，这里会优先出现决策卡。"
                  />
                )}

                {approvalMutation.error ? (
                  <p className="mt-4 text-sm text-rose-100">
                    {approvalMutation.error instanceof Error
                      ? approvalMutation.error.message
                      : "处理审批失败"}
                  </p>
                ) : null}
              </div>
            </AgentSurface>

            <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
              <AgentSectionHeader
                eyebrow="上下文"
                title="关联上下文"
                description="当前任务会参考这些信息。"
              />

              <div className="mt-6 rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                {run?.context_snapshot ? (
                  <>
                    <p className="text-sm text-white/60">
                      这轮任务会参考以下信息：
                    </p>
                    <AgentContextChips snapshot={run.context_snapshot} className="mt-4" />
                  </>
                ) : (
                  <p className="text-sm leading-6 text-white/56">
                    本轮任务没有额外挂载上下文，主要依据你手动输入的目标推进。
                  </p>
                )}
              </div>
            </AgentSurface>

            <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
              <AgentSectionHeader
                eyebrow="提示"
                title="当前任务值得注意的点"
                description="先看这里，可以更快判断当前情况。"
              />

              <div className="mt-6 grid gap-3">
                <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                  <p className="text-sm font-semibold text-white">状态解读</p>
                  <p className="mt-2 text-sm leading-6 text-white/56">
                    {run ? getRunStatusNarrative(run.status) : "—"}
                  </p>
                </div>
                <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                  <p className="text-sm font-semibold text-white">当前推进</p>
                  <p className="mt-2 text-sm leading-6 text-white/56">
                    {activeStep
                      ? `正在执行「${activeStep.title}」。`
                      : pendingApproval
                        ? "核心产物已准备好，当前等待你决定是否执行下一步动作。"
                        : run?.status === "completed"
                          ? "本轮任务已经结束，可以查看结果或基于它继续发起下一轮。"
                          : "当前没有正在运行的步骤。"}
                  </p>
                </div>
                <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                  <p className="text-sm font-semibold text-white">用户动作建议</p>
                  <p className="mt-2 text-sm leading-6 text-white/56">
                    {pendingApproval
                      ? "先审阅结果，再决定是否让系统回填。"
                      : run?.status === "failed"
                        ? "先查看失败原因，再基于本次配置重新发起一轮。"
                        : "当前可以继续观察执行，或返回工作台启动新的任务。"}
                  </p>
                </div>
              </div>
            </AgentSurface>
          </div>
        </div>
      </div>
    </div>
  );
}
