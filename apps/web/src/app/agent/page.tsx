"use client";

import Link from "next/link";
import { Suspense, useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowRight,
  Bot,
  CheckCircle2,
  Loader2,
  PlayCircle,
  Radar,
  ShieldCheck,
  Sparkles,
  Workflow,
  XCircle,
} from "lucide-react";
import { AgentRun, AgentRunStatus, CreateAgentRunInput, apiClient } from "@/lib/api-client";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  AgentContextChips,
  AgentMetricCard,
  AgentProgressBar,
  AgentScenarioBadge,
  AgentSectionHeader,
  AgentStatusBadge,
  AgentSurface,
  calculateRunProgress,
  formatAgentDateTime,
  getAgentScenarioPreset,
  getRunStatusIcon,
  getRunStatusNarrative,
  getScenarioShowcaseItems,
} from "@/components/agent/agent-ui";
import { cn } from "@/lib/utils";

function buildScenarioDefaults(
  scenario: string,
  contextSnapshot?: Record<string, unknown>,
) {
  const preset = getAgentScenarioPreset(scenario);
  const draftTitle = String(contextSnapshot?.draft_title || "").trim();
  const groupName = String(contextSnapshot?.group_name || "").trim();

  if (scenario === "group_agent") {
    return {
      title: groupName ? `分析圈子：${groupName}` : preset.defaultTitle,
      goal: preset.defaultGoal,
    };
  }

  if (scenario === "event_agent") {
    return {
      title: preset.defaultTitle,
      goal: preset.defaultGoal,
    };
  }

  return {
    title: draftTitle ? `润色草稿：${draftTitle}` : preset.defaultTitle,
    goal: preset.defaultGoal,
  };
}

function buildInitialForm(searchParams: URLSearchParams): CreateAgentRunInput {
  const scenario = searchParams.get("scenario") || "post_agent";
  const titleOverride = searchParams.get("title") || "";
  const goalOverride = searchParams.get("goal") || "";
  const draftTitle = searchParams.get("draft_title") || "";
  const draftContent = searchParams.get("draft_content") || "";
  const draftTags = searchParams.get("draft_tags") || "";
  const groupName = searchParams.get("group_name") || "";
  const visibility = searchParams.get("visibility") || "";
  const sourcePath = searchParams.get("source_path") || "";

  const contextSnapshot: Record<string, unknown> = {};
  if (draftTitle) contextSnapshot.draft_title = draftTitle;
  if (draftContent) contextSnapshot.draft_content = draftContent;
  if (draftTags) contextSnapshot.draft_tags = draftTags;
  if (groupName) contextSnapshot.group_name = groupName;
  if (visibility) contextSnapshot.visibility = visibility;
  if (sourcePath) contextSnapshot.source_path = sourcePath;

  const defaults = buildScenarioDefaults(scenario, contextSnapshot);

  return {
    title: titleOverride || defaults.title,
    goal: goalOverride || defaults.goal,
    scenario,
    context_snapshot: Object.keys(contextSnapshot).length > 0 ? contextSnapshot : undefined,
  };
}

function shouldPollRuns(runs?: AgentRun[]) {
  return (runs || []).some((run) => run.status === "queued" || run.status === "running");
}

function countRunsByStatus(runs: AgentRun[], status: AgentRunStatus) {
  return runs.filter((run) => run.status === status).length;
}

function getRunActionLabel(status: AgentRunStatus) {
  if (status === "waiting_approval") return "审阅结果";
  if (status === "queued" || status === "running") return "观看执行";
  if (status === "failed") return "查看问题";
  if (status === "completed") return "查看交付";
  return "查看记录";
}

function statusPriority(status: AgentRunStatus) {
  if (status === "waiting_approval") return 0;
  if (status === "running") return 1;
  if (status === "queued") return 2;
  if (status === "failed") return 3;
  if (status === "completed") return 4;
  return 5;
}

function ScenarioLaunchCard({
  scenarioKey,
  active,
  disabled,
  onSelect,
}: {
  scenarioKey: string;
  active: boolean;
  disabled?: boolean;
  onSelect: (scenarioKey: string) => void;
}) {
  const preset = getAgentScenarioPreset(scenarioKey);
  const Icon = preset.icon;

  return (
    <button
      type="button"
      disabled={disabled}
      onClick={() => onSelect(scenarioKey)}
      className={cn(
        "group relative overflow-hidden rounded-[24px] border px-4 py-4 text-left transition-all duration-200",
        "bg-[linear-gradient(180deg,rgba(255,255,255,0.08),rgba(255,255,255,0.04))]",
        active
          ? "border-cyan-300/45 shadow-[0_0_0_1px_rgba(103,232,249,0.2),0_24px_60px_rgba(8,145,178,0.15)]"
          : "border-white/10 hover:border-white/18 hover:bg-white/8",
        disabled ? "cursor-not-allowed opacity-45" : "",
      )}
    >
      <div
        className={cn(
          "pointer-events-none absolute inset-0 bg-gradient-to-br opacity-80",
          preset.accentClassName,
          active ? "opacity-100" : "opacity-55",
        )}
      />
      <div className="relative">
        <div className="flex items-start justify-between gap-3">
          <div className="rounded-[18px] border border-white/12 bg-slate-950/35 p-3 text-white shadow-[inset_0_1px_0_rgba(255,255,255,0.08)]">
            <Icon className="h-5 w-5" />
          </div>
          <span className="rounded-full border border-white/12 bg-white/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-white/78">
            {disabled ? "规划中" : active ? "当前选中" : "可启动"}
          </span>
        </div>
        <p className="mt-5 text-base font-semibold text-white">{preset.label}</p>
        <p className="mt-2 text-sm leading-6 text-white/68">{preset.summary}</p>
        <div className="mt-4 flex flex-wrap gap-2">
          {preset.outputs.slice(0, 3).map((output) => (
            <span
              key={output}
              className="rounded-full border border-white/10 bg-white/8 px-2.5 py-1 text-[11px] text-white/74"
            >
              {output}
            </span>
          ))}
        </div>
      </div>
    </button>
  );
}

function AgentRunBoardCard({ run }: { run: AgentRun }) {
  const Icon = getRunStatusIcon(run.status);
  const progress = calculateRunProgress(run.status);
  const isHot = run.status === "waiting_approval";
  const isLive = run.status === "queued" || run.status === "running";
  const isFailed = run.status === "failed";

  return (
    <Link
      href={`/agent/runs/${run.id}`}
      className={cn(
        "group block rounded-[26px] border p-5 transition-all duration-200",
        "bg-[linear-gradient(180deg,rgba(255,255,255,0.08),rgba(255,255,255,0.04))]",
        isHot
          ? "border-amber-300/30 hover:border-amber-200/45"
          : isLive
            ? "border-cyan-300/25 hover:border-cyan-200/40"
            : isFailed
              ? "border-rose-300/25 hover:border-rose-200/45"
              : "border-white/10 hover:border-white/18",
      )}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <AgentStatusBadge status={run.status} pulse />
            <AgentScenarioBadge scenario={run.scenario} />
          </div>
          <p className="mt-4 truncate text-lg font-semibold tracking-tight text-white">
            {run.title}
          </p>
          <p className="mt-2 line-clamp-3 text-sm leading-6 text-white/62">
            {run.latest_summary || run.goal}
          </p>
        </div>
        <div className="rounded-[18px] border border-white/10 bg-white/6 p-3 text-white/70">
          <Icon className="h-5 w-5" />
        </div>
      </div>

      <div className="mt-5">
        <div className="flex items-center justify-between text-[11px] uppercase tracking-[0.16em] text-white/48">
          <span>推进进度</span>
          <span>{progress}%</span>
        </div>
        <AgentProgressBar value={progress} className="mt-2" />
        <p className="mt-3 text-sm leading-6 text-white/56">{getRunStatusNarrative(run.status)}</p>
      </div>

      <div className="mt-5 flex flex-wrap items-center gap-3 text-xs text-white/42">
        <span>更新时间 {formatAgentDateTime(run.updated_at)}</span>
        <span>尝试 {run.attempt_count}/{run.max_attempts}</span>
        {run.next_retry_at ? <span>下次重试 {formatAgentDateTime(run.next_retry_at)}</span> : null}
      </div>

      <div className="mt-5 flex items-center justify-between">
        <span className="text-sm text-white/48">{run.id.slice(0, 8)}</span>
        <span className="inline-flex items-center gap-1.5 text-sm font-medium text-white">
          {getRunActionLabel(run.status)}
          <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
        </span>
      </div>
    </Link>
  );
}

function AgentWorkspaceContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const { isLoggedIn } = useAuth();
  const initialForm = useMemo(() => buildInitialForm(searchParams), [searchParams]);
  const [form, setForm] = useState<CreateAgentRunInput>(initialForm);

  useEffect(() => {
    setForm(initialForm);
  }, [initialForm]);

  const { data, isLoading } = useQuery({
    queryKey: ["agent-runs"],
    queryFn: () => apiClient.getAgentRuns(1, 20),
    enabled: isLoggedIn,
    refetchInterval: (query) => {
      const runs = (query.state.data as { runs?: AgentRun[] } | undefined)?.runs;
      return shouldPollRuns(runs) ? 1500 : false;
    },
  });

  const createMutation = useMutation({
    mutationFn: (payload: CreateAgentRunInput) => apiClient.createAgentRun(payload),
    onSuccess: (detail) => {
      queryClient.invalidateQueries({ queryKey: ["agent-runs"] });
      router.push(`/agent/runs/${detail.run.id}`);
    },
  });

  const runs = useMemo(() => {
    const next = [...(data?.runs || [])];
    return next.sort((left, right) => {
      const priority = statusPriority(left.status) - statusPriority(right.status);
      if (priority !== 0) return priority;
      return new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime();
    });
  }, [data?.runs]);

  const activeRuns = runs.filter((run) => run.status === "queued" || run.status === "running");
  const attentionRuns = runs.filter(
    (run) => run.status === "waiting_approval" || run.status === "failed",
  );
  const archiveRuns = runs.filter(
    (run) => run.status === "completed" || run.status === "cancelled",
  );
  const featuredRun = attentionRuns[0] || activeRuns[0] || runs[0];

  const queuedCount = countRunsByStatus(runs, "queued");
  const runningCount = countRunsByStatus(runs, "running");
  const waitingApprovalCount = countRunsByStatus(runs, "waiting_approval");
  const completedCount = countRunsByStatus(runs, "completed");
  const failedCount = countRunsByStatus(runs, "failed");
  const terminalRuns = completedCount + failedCount + countRunsByStatus(runs, "cancelled");
  const successRate = terminalRuns > 0 ? `${Math.round((completedCount / terminalRuns) * 100)}%` : "—";
  const selectedScenario = getAgentScenarioPreset(form.scenario || "post_agent");
  const showcaseItems = getScenarioShowcaseItems();

  function handleScenarioChange(nextScenario: string) {
    setForm((prev) => {
      const prevScenario = String(prev.scenario || "post_agent");
      const prevDefaults = buildScenarioDefaults(prevScenario, prev.context_snapshot);
      const nextDefaults = buildScenarioDefaults(nextScenario, prev.context_snapshot);
      const shouldReplaceTitle = !String(prev.title || "").trim() || prev.title === prevDefaults.title;
      const shouldReplaceGoal = !String(prev.goal || "").trim() || prev.goal === prevDefaults.goal;

      return {
        ...prev,
        scenario: nextScenario,
        title: shouldReplaceTitle ? nextDefaults.title : prev.title,
        goal: shouldReplaceGoal ? nextDefaults.goal : prev.goal,
      };
    });
  }

  function handleQuickGoal(nextGoal: string) {
    setForm((prev) => ({
      ...prev,
      title: String(prev.title || "").trim() ? prev.title : selectedScenario.defaultTitle,
      goal: nextGoal,
    }));
  }

  if (!isLoggedIn) {
    return (
      <div className="agent-page-shell relative min-h-screen overflow-hidden">
        <div className="agent-page-grid pointer-events-none absolute inset-0" />
        <div className="pointer-events-none absolute left-[-8rem] top-32 h-72 w-72 rounded-full bg-cyan-400/20 blur-3xl" />
        <div className="pointer-events-none absolute right-[-6rem] top-24 h-80 w-80 rounded-full bg-orange-400/18 blur-3xl" />
        <div className="mx-auto max-w-5xl px-4 pb-16 pt-24">
          <AgentSurface className="px-6 py-8 sm:px-8 sm:py-10">
            <div className="grid gap-8 lg:grid-cols-[1.05fr_0.95fr] lg:items-center">
              <div>
                <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-cyan-200/70">
                  Agent
                </p>
                <h1 className="mt-4 text-4xl font-semibold tracking-tight text-white sm:text-5xl">
                  用 Agent 推进你的当前任务
                </h1>
                <p className="mt-5 max-w-2xl text-base leading-8 text-white/66">
                  登录后可以创建任务、查看进度、审阅结果，并在需要时继续完成下一步。
                </p>
                <div className="mt-8 flex flex-wrap gap-3">
                  <Button asChild className="h-11 rounded-full bg-white text-slate-950 hover:bg-white/90">
                    <Link href="/login">登录后开始</Link>
                  </Button>
                  <Button
                    asChild
                    variant="outline"
                    className="h-11 rounded-full border-white/16 bg-white/6 text-white hover:bg-white/10 hover:text-white"
                  >
                    <Link href="/feed">返回社区首页</Link>
                  </Button>
                </div>
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <AgentMetricCard
                  icon={Radar}
                  label="状态同步"
                  value="自动更新"
                  meta="你可以随时看到当前进度。"
                  tone="cyan"
                />
                <AgentMetricCard
                  icon={Sparkles}
                  label="结果确认"
                  value="由你决定"
                  meta="需要继续应用结果时再确认。"
                  tone="amber"
                />
                <AgentMetricCard
                  icon={ShieldCheck}
                  label="快速开始"
                  value="场景化"
                  meta="按目标选择合适的 Agent。"
                  tone="emerald"
                />
                <AgentMetricCard
                  icon={Workflow}
                  label="任务记录"
                  value="集中查看"
                  meta="最近运行和结果都在这里。"
                  tone="slate"
                />
              </div>
            </div>
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
        <section className="grid gap-6 xl:grid-cols-[1.08fr_0.92fr]">
          <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
            <div className="grid gap-8 xl:grid-cols-[1.02fr_0.98fr]">
              <div>
                <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-cyan-200/70">
                  当前任务
                </p>
                <h1 className="mt-4 max-w-3xl text-4xl font-semibold tracking-tight text-white sm:text-5xl">
                  Agent 工作台
                </h1>
                <p className="mt-5 max-w-2xl text-base leading-8 text-white/66">
                  选择场景，补充目标，让 Agent 帮你推进当前任务。
                </p>
                <div className="mt-6 flex flex-wrap gap-2.5">
                  <span className="rounded-full border border-white/10 bg-white/8 px-3 py-1.5 text-xs text-white/74">
                    创建任务
                  </span>
                  <span className="rounded-full border border-white/10 bg-white/8 px-3 py-1.5 text-xs text-white/74">
                    查看进度
                  </span>
                  <span className="rounded-full border border-white/10 bg-white/8 px-3 py-1.5 text-xs text-white/74">
                    审阅结果
                  </span>
                </div>
                {form.context_snapshot ? (
                  <div className="mt-6 rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                    <p className="text-sm font-medium text-white">已挂载外部上下文</p>
                    <p className="mt-2 text-sm leading-6 text-white/58">
                      当前任务会读取你从其他页面带过来的草稿信息或来源信息，避免从零重新描述。
                    </p>
                    <AgentContextChips snapshot={form.context_snapshot} className="mt-4" />
                  </div>
                ) : null}
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <AgentMetricCard
                  icon={Radar}
                  label="执行中"
                  value={queuedCount + runningCount}
                  meta="正在排队或执行的任务数。"
                  tone="cyan"
                />
                <AgentMetricCard
                  icon={Sparkles}
                  label="待你决策"
                  value={waitingApprovalCount}
                  meta="已有结果，等待你审阅和决定是否回填。"
                  tone="amber"
                />
                <AgentMetricCard
                  icon={CheckCircle2}
                  label="已完成"
                  value={completedCount}
                  meta={`基于最近 ${runs.length || 0} 条运行统计。`}
                  tone="emerald"
                />
                <AgentMetricCard
                  icon={XCircle}
                  label="成功率"
                  value={successRate}
                  meta={terminalRuns > 0 ? "已结束任务的完成占比。" : "当前还没有足够样本。"}
                  tone="rose"
                />
              </div>
            </div>
          </AgentSurface>

          <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
            <AgentSectionHeader
              eyebrow="可用场景"
              title="当前可调度的 Agent 能力"
              description="根据当前目标选择合适的 Agent。"
            />
            <div className="mt-6 grid gap-3">
              {showcaseItems.map((item) => {
                const Icon = item.icon;
                return (
                  <div
                    key={item.scenario.key}
                    className="flex items-start gap-4 rounded-[24px] border border-white/10 bg-white/6 px-4 py-4"
                  >
                    <div className="rounded-[18px] border border-white/10 bg-slate-950/35 p-3 text-white/82">
                      <Icon className="h-5 w-5" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <p className="text-sm font-semibold text-white">{item.scenario.label}</p>
                        <span className="rounded-full border border-white/10 bg-white/8 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.14em] text-white/66">
                          {item.availability}
                        </span>
                      </div>
                      <p className="mt-2 text-sm leading-6 text-white/56">{item.scenario.summary}</p>
                    </div>
                  </div>
                );
              })}

              {featuredRun ? (
                <Link
                  href={`/agent/runs/${featuredRun.id}`}
                  className="group mt-2 rounded-[24px] border border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.10),rgba(255,255,255,0.04))] px-4 py-4 transition-colors hover:border-white/16"
                >
                  <div className="flex flex-wrap items-center gap-2">
                    <AgentStatusBadge status={featuredRun.status} pulse />
                    <AgentScenarioBadge scenario={featuredRun.scenario} />
                  </div>
                  <p className="mt-4 text-lg font-semibold tracking-tight text-white">
                    {featuredRun.title}
                  </p>
                  <p className="mt-2 text-sm leading-6 text-white/58">
                    {featuredRun.latest_summary || featuredRun.goal}
                  </p>
                  <div className="mt-4 flex items-center justify-between text-sm text-white/60">
                    <span>{formatAgentDateTime(featuredRun.updated_at)}</span>
                    <span className="inline-flex items-center gap-1 text-white">
                      {getRunActionLabel(featuredRun.status)}
                      <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
                    </span>
                  </div>
                </Link>
              ) : null}
            </div>
          </AgentSurface>
        </section>

        <section className="mt-8 grid gap-6 xl:grid-cols-[1.08fr_0.92fr]">
          <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
            <AgentSectionHeader
              eyebrow={selectedScenario.eyebrow}
              title="启动一轮新的 Agent 任务"
              description="先选场景，再描述你希望 Agent 帮你完成什么。"
            />

            <div className="mt-6 grid gap-3 lg:grid-cols-3">
              <ScenarioLaunchCard
                scenarioKey="post_agent"
                active={form.scenario === "post_agent"}
                onSelect={handleScenarioChange}
              />
              <ScenarioLaunchCard
                scenarioKey="group_agent"
                active={form.scenario === "group_agent"}
                onSelect={handleScenarioChange}
              />
              <ScenarioLaunchCard
                scenarioKey="event_agent"
                active={false}
                disabled
                onSelect={handleScenarioChange}
              />
            </div>

            <div className="mt-7 grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
              <div className="space-y-5">
                <div>
                  <label className="text-sm font-medium text-white">任务标题</label>
                  <Input
                    value={form.title || ""}
                    onChange={(event) =>
                      setForm((prev) => ({ ...prev, title: event.target.value }))
                    }
                    placeholder={selectedScenario.defaultTitle}
                    className="mt-2 h-12 rounded-2xl border-white/10 bg-white/8 px-4 text-white placeholder:text-white/34 focus-visible:ring-cyan-300/50"
                  />
                </div>

                <div>
                  <label className="text-sm font-medium text-white">任务目标</label>
                  <Textarea
                    value={form.goal}
                    onChange={(event) =>
                      setForm((prev) => ({ ...prev, goal: event.target.value }))
                    }
                    rows={7}
                    placeholder={selectedScenario.promptPlaceholder}
                    className="mt-2 min-h-[220px] rounded-[24px] border-white/10 bg-white/8 px-4 py-3 text-white placeholder:text-white/34 focus-visible:ring-cyan-300/50"
                  />
                  <div className="mt-3 flex flex-wrap gap-2">
                    {selectedScenario.quickGoals.map((goal) => (
                      <button
                        key={goal}
                        type="button"
                        onClick={() => handleQuickGoal(goal)}
                        className="rounded-full border border-white/10 bg-white/6 px-3 py-1.5 text-xs text-white/70 transition-colors hover:bg-white/10"
                      >
                        {goal}
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              <div className="space-y-4">
                <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                  <p className="text-sm font-semibold text-white">这轮 Agent 会交付什么</p>
                  <p className="mt-2 text-sm leading-6 text-white/58">
                    {selectedScenario.contextHint}
                  </p>
                  <div className="mt-4 flex flex-wrap gap-2">
                    {selectedScenario.outputs.map((output) => (
                      <span
                        key={output}
                        className="rounded-full border border-white/10 bg-white/8 px-2.5 py-1 text-xs text-white/76"
                      >
                        {output}
                      </span>
                    ))}
                  </div>
                </div>

                <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                  <p className="text-sm font-semibold text-white">进行方式</p>
                  <div className="mt-4 space-y-4">
                    <div className="flex gap-3">
                      <div className="mt-1 rounded-full border border-white/10 bg-white/8 p-2 text-cyan-200">
                        <Bot className="h-4 w-4" />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">1. 明确任务目标</p>
                        <p className="mt-1 text-sm leading-6 text-white/56">
                          先确认标题、目标和当前上下文。
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-3">
                      <div className="mt-1 rounded-full border border-white/10 bg-white/8 p-2 text-cyan-200">
                        <Workflow className="h-4 w-4" />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">2. 生成结果</p>
                        <p className="mt-1 text-sm leading-6 text-white/56">
                          运行过程中会逐步更新状态、步骤和结果。
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-3">
                      <div className="mt-1 rounded-full border border-white/10 bg-white/8 p-2 text-cyan-200">
                        <ShieldCheck className="h-4 w-4" />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">3. 审阅并确认</p>
                        <p className="mt-1 text-sm leading-6 text-white/56">
                          如果需要继续应用结果，你可以在最后一步确认。
                        </p>
                      </div>
                    </div>
                  </div>
                </div>

                {form.context_snapshot ? (
                  <div className="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
                    <p className="text-sm font-semibold text-white">当前上下文挂载</p>
                    <AgentContextChips snapshot={form.context_snapshot} className="mt-4" />
                  </div>
                ) : null}
              </div>
            </div>

            <div className="mt-7 flex flex-col gap-4 border-t border-white/10 pt-6 sm:flex-row sm:items-center sm:justify-between">
              <div className="text-sm leading-6 text-white/58">
                {createMutation.error ? (
                  <span className="text-rose-200">
                    {createMutation.error instanceof Error
                      ? createMutation.error.message
                      : "创建任务失败，请稍后重试。"}
                  </span>
                ) : (
                  <span>
                    创建后可在详情页查看状态、结果和下一步动作。
                  </span>
                )}
              </div>
              <Button
                onClick={() => createMutation.mutate(form)}
                disabled={createMutation.isPending || !String(form.goal || "").trim()}
                className="h-12 rounded-full bg-white px-6 text-slate-950 hover:bg-white/90"
              >
                {createMutation.isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <PlayCircle className="mr-2 h-4 w-4" />
                )}
                启动 Agent 任务
              </Button>
            </div>
          </AgentSurface>

          <div className="space-y-6">
            <AgentSurface className="px-6 py-7 sm:px-8 sm:py-8">
              <AgentSectionHeader
                eyebrow="最近任务"
                title="运行板"
                description="需要处理的任务会优先出现在前面。"
              />

              <Tabs defaultValue="attention" className="mt-6">
                <TabsList className="h-auto w-full justify-start gap-2 rounded-full bg-white/6 p-1">
                  <TabsTrigger
                    value="attention"
                    className="rounded-full px-4 py-2 text-xs text-white/68 data-[state=active]:bg-white data-[state=active]:text-slate-950"
                  >
                    需关注 {attentionRuns.length}
                  </TabsTrigger>
                  <TabsTrigger
                    value="live"
                    className="rounded-full px-4 py-2 text-xs text-white/68 data-[state=active]:bg-white data-[state=active]:text-slate-950"
                  >
                    执行中 {activeRuns.length}
                  </TabsTrigger>
                  <TabsTrigger
                    value="archive"
                    className="rounded-full px-4 py-2 text-xs text-white/68 data-[state=active]:bg-white data-[state=active]:text-slate-950"
                  >
                    历史 {archiveRuns.length}
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="attention" className="mt-5 space-y-4">
                  {!attentionRuns.length ? (
                    <div className="rounded-[24px] border border-dashed border-white/12 bg-white/5 px-4 py-5 text-sm text-white/56">
                      当前没有需要你处理的运行。新的审批或失败任务会优先出现在这里。
                    </div>
                  ) : null}
                  {attentionRuns.map((run) => (
                    <AgentRunBoardCard key={run.id} run={run} />
                  ))}
                </TabsContent>

                <TabsContent value="live" className="mt-5 space-y-4">
                  {!activeRuns.length ? (
                    <div className="rounded-[24px] border border-dashed border-white/12 bg-white/5 px-4 py-5 text-sm text-white/56">
                      当前没有正在执行的任务。启动一轮新的 Agent，就能看到实时进度出现在这里。
                    </div>
                  ) : null}
                  {activeRuns.map((run) => (
                    <AgentRunBoardCard key={run.id} run={run} />
                  ))}
                </TabsContent>

                <TabsContent value="archive" className="mt-5 space-y-4">
                  {isLoading ? (
                    <div className="flex items-center gap-2 text-sm text-white/56">
                      <Loader2 className="h-4 w-4 animate-spin" />
                      正在读取最近运行
                    </div>
                  ) : null}
                  {!isLoading && !archiveRuns.length ? (
                    <div className="rounded-[24px] border border-dashed border-white/12 bg-white/5 px-4 py-5 text-sm text-white/56">
                      还没有历史记录。完成更多任务后可以在这里统一查看。
                    </div>
                  ) : null}
                  {archiveRuns.map((run) => (
                    <AgentRunBoardCard key={run.id} run={run} />
                  ))}
                </TabsContent>
              </Tabs>
            </AgentSurface>
          </div>
        </section>
      </div>
    </div>
  );
}

export default function AgentWorkspacePage() {
  return (
    <Suspense
      fallback={
        <div className="agent-page-shell relative min-h-screen overflow-hidden">
          <div className="agent-page-grid pointer-events-none absolute inset-0" />
          <div className="mx-auto max-w-7xl px-4 pb-16 pt-24">
            <AgentSurface className="px-6 py-8 sm:px-8">
              <div className="flex items-center gap-2 text-sm text-white/60">
                <Loader2 className="h-4 w-4 animate-spin" />
                正在加载
              </div>
            </AgentSurface>
          </div>
        </div>
      }
    >
      <AgentWorkspaceContent />
    </Suspense>
  );
}
