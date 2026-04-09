"use client";

import { ReactNode } from "react";
import {
  Bot,
  CalendarRange,
  CheckCircle2,
  Clock3,
  Layers3,
  LucideIcon,
  MessageSquareText,
  Orbit,
  Radar,
  Sparkles,
  Users,
  XCircle,
} from "lucide-react";
import { AgentRunStatus, AgentStepStatus } from "@/lib/api-client";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export type AgentScenarioPreset = {
  key: string;
  label: string;
  shortLabel: string;
  eyebrow: string;
  summary: string;
  promptPlaceholder: string;
  defaultTitle: string;
  defaultGoal: string;
  outputs: string[];
  quickGoals: string[];
  contextHint: string;
  icon: LucideIcon;
  accentClassName: string;
};

export const AGENT_STATUS_LABELS: Record<AgentRunStatus, string> = {
  queued: "排队中",
  running: "执行中",
  waiting_approval: "待你决策",
  completed: "已完成",
  failed: "执行失败",
  cancelled: "已取消",
};

export const AGENT_STEP_STATUS_LABELS: Record<AgentStepStatus, string> = {
  pending: "待处理",
  running: "执行中",
  completed: "已完成",
  failed: "失败",
  skipped: "已跳过",
};

const DEFAULT_SCENARIO_PRESET: AgentScenarioPreset = {
  key: "generic_agent",
  label: "通用 Agent",
  shortLabel: "通用",
  eyebrow: "General Operation",
  summary: "围绕目标、上下文和用户决策节点执行一轮任务。",
  promptPlaceholder: "例如：先读上下文，再给我一版更可执行的结果。",
  defaultTitle: "新的 Agent 任务",
  defaultGoal: "请基于当前上下文整理一份更清晰、更可执行的结果。",
  outputs: ["执行摘要", "结构化产物", "可审批动作"],
  quickGoals: [
    "先帮我梳理问题，再给出一个可执行版本。",
    "告诉我最值得优先处理的 3 件事。",
    "如果需要审批，请把理由讲清楚。",
  ],
  contextHint: "可以从页面上下文或历史运行中接入任务背景。",
  icon: Bot,
  accentClassName: "from-cyan-400/30 via-sky-500/20 to-orange-400/25",
};

export const AGENT_SCENARIO_PRESETS: Record<string, AgentScenarioPreset> = {
  post_agent: {
    key: "post_agent",
    label: "发帖 Agent",
    shortLabel: "发帖",
    eyebrow: "Content Launch",
    summary: "面向发布动作的 Agent，重点处理标题、正文结构、标签与可见性建议。",
    promptPlaceholder: "例如：保留我的语气，但让内容更适合发出去。",
    defaultTitle: "润色这条动态草稿",
    defaultGoal: "请帮我整理这条动态草稿，给出标题、正文润色、标签和可见性建议。",
    outputs: ["标题方案", "正文优化", "标签建议", "可见性建议", "可回填草稿"],
    quickGoals: [
      "保留原本语气，但让这条动态更容易被读完。",
      "给我 3 个不同风格的标题方向。",
      "帮我判断这条动态更适合公开还是仅关注者可见。",
    ],
    contextHint: "最适合从发帖草稿、圈子信息和图片分析结果中接入上下文。",
    icon: MessageSquareText,
    accentClassName: "from-orange-400/30 via-amber-400/20 to-cyan-400/25",
  },
  group_agent: {
    key: "group_agent",
    label: "圈子 Agent",
    shortLabel: "圈子",
    eyebrow: "Community Fit",
    summary: "围绕圈子规则、氛围与内容方向做判断，帮助用户决定是否加入和怎么发第一条内容。",
    promptPlaceholder: "例如：告诉我这个圈子值不值得加入，以及发什么最稳妥。",
    defaultTitle: "分析这个圈子值不值得加入",
    defaultGoal: "这个圈子适合我加入吗？如果加入，适合先发什么内容？",
    outputs: ["圈子氛围概览", "规则提炼", "加入建议", "发帖方向"],
    quickGoals: [
      "告诉我这个圈子的内容风格和潜在雷点。",
      "如果我要加入，第一条内容应该发什么更自然。",
      "把加入建议拆成值得加入、不值得加入、需要观察三种情况。",
    ],
    contextHint: "最适合从圈子简介、规则摘要和用户当前页面来源中接入上下文。",
    icon: Users,
    accentClassName: "from-cyan-400/30 via-teal-400/18 to-emerald-400/24",
  },
  event_agent: {
    key: "event_agent",
    label: "活动 Agent",
    shortLabel: "活动",
    eyebrow: "Event Readiness",
    summary: "面向活动摘要、参与判断和准备建议的任务型 Agent。",
    promptPlaceholder: "例如：帮我判断这场活动值不值得去，并列出准备清单。",
    defaultTitle: "分析这个活动是否值得参加",
    defaultGoal: "请帮我总结活动信息，并给出是否值得参加与行前准备建议。",
    outputs: ["活动摘要", "适配判断", "报名提示", "准备清单"],
    quickGoals: [
      "快速告诉我这场活动适不适合我。",
      "如果要去，帮我列一个准备清单。",
      "把活动重点压缩成一分钟能看完的版本。",
    ],
    contextHint: "更适合接活动页上下文、时间地点和报名规则。",
    icon: CalendarRange,
    accentClassName: "from-fuchsia-400/24 via-violet-400/18 to-cyan-400/24",
  },
};

const STATUS_TONE_CLASSNAME: Record<AgentRunStatus, string> = {
  queued: "border-cyan-400/30 bg-cyan-400/10 text-cyan-100",
  running: "border-sky-400/30 bg-sky-400/10 text-sky-100",
  waiting_approval: "border-amber-300/35 bg-amber-300/10 text-amber-100",
  completed: "border-emerald-300/35 bg-emerald-300/10 text-emerald-100",
  failed: "border-rose-300/35 bg-rose-300/10 text-rose-100",
  cancelled: "border-slate-200/20 bg-slate-300/10 text-slate-100",
};

const STEP_TONE_CLASSNAME: Record<AgentStepStatus, string> = {
  pending: "border-amber-200 bg-amber-50 text-amber-700",
  running: "border-cyan-200 bg-cyan-50 text-cyan-700",
  completed: "border-emerald-200 bg-emerald-50 text-emerald-700",
  failed: "border-rose-200 bg-rose-50 text-rose-700",
  skipped: "border-slate-200 bg-slate-100 text-slate-600",
};

export function getAgentScenarioPreset(scenario?: string) {
  if (!scenario) return DEFAULT_SCENARIO_PRESET;
  return AGENT_SCENARIO_PRESETS[scenario] || {
    ...DEFAULT_SCENARIO_PRESET,
    key: scenario,
    label: scenario,
    shortLabel: scenario,
  };
}

export function AgentSurface({
  className,
  children,
}: {
  className?: string;
  children: ReactNode;
}) {
  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-[28px] border border-white/10 bg-[linear-gradient(180deg,rgba(15,23,42,0.82),rgba(15,23,42,0.68))] shadow-[0_24px_80px_rgba(2,8,23,0.28)] backdrop-blur-xl",
        className,
      )}
    >
      {children}
    </div>
  );
}

export function AgentStatusBadge({
  status,
  pulse = false,
  className,
}: {
  status: AgentRunStatus;
  pulse?: boolean;
  className?: string;
}) {
  return (
    <Badge
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.14em]",
        STATUS_TONE_CLASSNAME[status],
        className,
      )}
    >
      <span
        className={cn(
          "h-1.5 w-1.5 rounded-full bg-current",
          pulse && (status === "queued" || status === "running") ? "animate-pulse" : "",
        )}
      />
      {AGENT_STATUS_LABELS[status]}
    </Badge>
  );
}

export function AgentStepStatusBadge({
  status,
  className,
}: {
  status: AgentStepStatus;
  className?: string;
}) {
  return (
    <Badge
      className={cn(
        "rounded-full border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em]",
        STEP_TONE_CLASSNAME[status],
        className,
      )}
    >
      {AGENT_STEP_STATUS_LABELS[status]}
    </Badge>
  );
}

export function AgentScenarioBadge({
  scenario,
  className,
}: {
  scenario?: string;
  className?: string;
}) {
  const preset = getAgentScenarioPreset(scenario);
  return (
    <Badge
      className={cn(
        "rounded-full border border-white/12 bg-white/8 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-white/82",
        className,
      )}
    >
      {preset.shortLabel}
    </Badge>
  );
}

export function AgentMetricCard({
  icon: Icon,
  label,
  value,
  meta,
  tone = "slate",
  className,
}: {
  icon: LucideIcon;
  label: string;
  value: string | number;
  meta: string;
  tone?: "cyan" | "amber" | "emerald" | "rose" | "slate";
  className?: string;
}) {
  const toneClassName =
    tone === "cyan"
      ? "border-cyan-400/18 bg-cyan-400/10 text-cyan-100"
      : tone === "amber"
        ? "border-amber-400/18 bg-amber-400/10 text-amber-100"
        : tone === "emerald"
          ? "border-emerald-400/18 bg-emerald-400/10 text-emerald-100"
          : tone === "rose"
            ? "border-rose-400/18 bg-rose-400/10 text-rose-100"
            : "border-white/10 bg-white/6 text-white/90";

  return (
    <div
      className={cn(
        "rounded-[22px] border px-4 py-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.05)]",
        toneClassName,
        className,
      )}
    >
      <div className="flex items-center justify-between gap-3">
        <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-white/62">
          {label}
        </p>
        <Icon className="h-4 w-4 text-white/60" />
      </div>
      <p className="mt-4 text-3xl font-semibold tracking-tight text-white">{value}</p>
      <p className="mt-2 text-sm leading-6 text-white/62">{meta}</p>
    </div>
  );
}

export function AgentSectionHeader({
  eyebrow,
  title,
  description,
  action,
  className,
}: {
  eyebrow?: string;
  title: string;
  description?: string;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between", className)}>
      <div>
        {eyebrow ? (
          <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-white/48">
            {eyebrow}
          </p>
        ) : null}
        <h2 className="mt-2 text-xl font-semibold tracking-tight text-white">{title}</h2>
        {description ? (
          <p className="mt-2 max-w-2xl text-sm leading-7 text-white/62">{description}</p>
        ) : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}

export function AgentEmptyState({
  icon: Icon,
  title,
  description,
  className,
}: {
  icon: LucideIcon;
  title: string;
  description: string;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "rounded-[24px] border border-dashed border-white/12 bg-white/5 px-5 py-6 text-white/72",
        className,
      )}
    >
      <div className="flex items-start gap-3">
        <div className="rounded-2xl border border-white/10 bg-white/6 p-3 text-white/78">
          <Icon className="h-5 w-5" />
        </div>
        <div>
          <p className="text-sm font-semibold text-white">{title}</p>
          <p className="mt-2 text-sm leading-6 text-white/58">{description}</p>
        </div>
      </div>
    </div>
  );
}

export function AgentContextChips({
  snapshot,
  className,
}: {
  snapshot?: Record<string, unknown>;
  className?: string;
}) {
  const entries = getContextSnapshotEntries(snapshot);
  if (entries.length === 0) return null;

  return (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {entries.map((entry) => (
        <div
          key={`${entry.label}-${entry.value}`}
          className="rounded-full border border-white/10 bg-white/6 px-3 py-1.5 text-xs text-white/74"
        >
          <span className="text-white/46">{entry.label}</span>
          <span className="mx-1 text-white/24">/</span>
          <span>{entry.value}</span>
        </div>
      ))}
    </div>
  );
}

export function AgentProgressBar({
  value,
  className,
}: {
  value: number;
  className?: string;
}) {
  const safe = Math.max(0, Math.min(100, value));
  return (
    <div className={cn("h-2 overflow-hidden rounded-full bg-white/10", className)}>
      <div
        className="h-full rounded-full bg-[linear-gradient(90deg,rgba(34,211,238,0.92),rgba(251,146,60,0.92))] transition-[width] duration-500"
        style={{ width: `${safe}%` }}
      />
    </div>
  );
}

export function getContextSnapshotEntries(snapshot?: Record<string, unknown>) {
  if (!snapshot) return [];

  const labels: Record<string, string> = {
    draft_title: "标题草稿",
    draft_content: "正文草稿",
    draft_tags: "草稿标签",
    group_name: "目标圈子",
    visibility: "可见性",
    source_path: "来源页面",
    ai_generated: "AI 标记",
    image_tags: "图片标签",
    image_alt_notes: "图片说明",
  };

  const preferredOrder = [
    "group_name",
    "source_path",
    "draft_title",
    "draft_content",
    "draft_tags",
    "visibility",
    "image_tags",
    "image_alt_notes",
    "ai_generated",
  ];

  return Object.entries(snapshot)
    .sort(([keyA], [keyB]) => {
      const indexA = preferredOrder.indexOf(keyA);
      const indexB = preferredOrder.indexOf(keyB);
      const safeA = indexA === -1 ? preferredOrder.length + 1 : indexA;
      const safeB = indexB === -1 ? preferredOrder.length + 1 : indexB;
      return safeA - safeB;
    })
    .flatMap(([key, rawValue]) => {
      if (rawValue == null) return [];
      const value = String(rawValue).trim();
      if (!value) return [];
      return [
        {
          key,
          label: labels[key] || key,
          value: value.length > 48 ? `${value.slice(0, 48).trim()}...` : value,
        },
      ];
    });
}

export function formatAgentDateTime(value?: string) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleString("zh-CN");
}

function formatDuration(durationMs: number) {
  if (durationMs < 1000) return `${durationMs}ms`;
  if (durationMs < 60_000) return `${(durationMs / 1000).toFixed(1)}s`;
  if (durationMs < 3_600_000) return `${(durationMs / 60_000).toFixed(1)}m`;
  return `${(durationMs / 3_600_000).toFixed(1)}h`;
}

export function formatAgentDuration(startedAt?: string, completedAt?: string) {
  if (!startedAt || !completedAt) return "—";
  const start = new Date(startedAt).getTime();
  const end = new Date(completedAt).getTime();
  const durationMs = end - start;
  if (!Number.isFinite(durationMs) || durationMs < 0) return "—";
  return formatDuration(durationMs);
}

export function formatAgentElapsed(startedAt?: string, completedAt?: string) {
  if (!startedAt) return "—";
  const start = new Date(startedAt).getTime();
  const end = completedAt ? new Date(completedAt).getTime() : Date.now();
  const durationMs = end - start;
  if (!Number.isFinite(durationMs) || durationMs < 0) return "—";
  return formatDuration(durationMs);
}

export function calculateRunProgress(
  status: AgentRunStatus,
  steps?: Array<{ status: AgentStepStatus }>,
) {
  if (status === "completed" || status === "failed" || status === "cancelled") {
    return 100;
  }

  if (steps && steps.length > 0) {
    const completed = steps.filter(
      (step) => step.status === "completed" || step.status === "skipped",
    ).length;
    const running = steps.some((step) => step.status === "running");
    const base = Math.round((completed / steps.length) * 100);
    if (status === "waiting_approval") return Math.max(base, 92);
    if (running) return Math.min(Math.max(base + 12, 28), 88);
    if (status === "queued") return Math.max(base, 12);
    return Math.max(base, 18);
  }

  if (status === "waiting_approval") return 92;
  if (status === "running") return 46;
  if (status === "queued") return 14;
  return 0;
}

export function getRunStatusNarrative(status: AgentRunStatus) {
  if (status === "queued") return "任务已收下，正在等待 worker 接手。";
  if (status === "running") return "Agent 正在执行步骤、调用工具并生成产物。";
  if (status === "waiting_approval") return "Agent 已经完成核心推理，当前等待你的最终决定。";
  if (status === "completed") return "本轮任务已经交付完成，结果可继续复用。";
  if (status === "failed") return "本轮任务在执行过程中遇到错误，需要查看原因。";
  return "任务已被停止，当前不会继续推进。";
}

export function getRunStatusIcon(status: AgentRunStatus) {
  if (status === "completed") return CheckCircle2;
  if (status === "failed") return XCircle;
  if (status === "waiting_approval") return Sparkles;
  if (status === "running") return Radar;
  if (status === "queued") return Clock3;
  return Layers3;
}

export function getScenarioShowcaseItems() {
  return [
    {
      scenario: AGENT_SCENARIO_PRESETS.post_agent,
      availability: "开放中",
      icon: MessageSquareText,
    },
    {
      scenario: AGENT_SCENARIO_PRESETS.group_agent,
      availability: "开放中",
      icon: Users,
    },
    {
      scenario: AGENT_SCENARIO_PRESETS.event_agent,
      availability: "规划中",
      icon: Orbit,
    },
  ];
}
