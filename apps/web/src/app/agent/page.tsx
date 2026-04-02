"use client";

import Link from "next/link";
import { Suspense, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bot, Loader2, PlayCircle, Sparkles } from "lucide-react";
import { apiClient, CreateAgentRunInput, AgentRun, AgentRunStatus } from "@/lib/api-client";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

const STATUS_LABELS: Record<AgentRunStatus, string> = {
  queued: "排队中",
  running: "运行中",
  waiting_approval: "待审批",
  completed: "已完成",
  failed: "失败",
  cancelled: "已取消",
};

function buildInitialForm(searchParams: URLSearchParams): CreateAgentRunInput {
  const scenario = searchParams.get("scenario") || "post_agent";
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

  return {
    title: draftTitle ? `润色草稿：${draftTitle}` : "发帖 Agent 任务",
    goal: "请帮我整理这条动态草稿，给出标题、正文润色、标签和可见性建议。",
    scenario,
    context_snapshot: Object.keys(contextSnapshot).length > 0 ? contextSnapshot : undefined,
  };
}

function StatusBadge({ status }: { status: AgentRunStatus }) {
  const className =
    status === "completed"
      ? "bg-emerald-100 text-emerald-700"
      : status === "failed"
        ? "bg-rose-100 text-rose-700"
        : status === "cancelled"
          ? "bg-slate-200 text-slate-700"
          : status === "waiting_approval"
            ? "bg-amber-100 text-amber-700"
            : "bg-cyan-100 text-cyan-700";
  return (
    <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${className}`}>
      {STATUS_LABELS[status]}
    </span>
  );
}

function shouldPollRuns(runs?: AgentRun[]) {
  return (runs || []).some((run) => run.status === "queued" || run.status === "running");
}

function AgentWorkspaceContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const { isLoggedIn } = useAuth();
  const initialForm = useMemo(() => buildInitialForm(searchParams), [searchParams]);
  const [form, setForm] = useState<CreateAgentRunInput>(initialForm);
  const [message, setMessage] = useState("");
  const hasExternalContext = Boolean(form.context_snapshot && Object.keys(form.context_snapshot).length > 0);

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
      setMessage("任务已创建，Agent 会在后台继续执行。");
      queryClient.invalidateQueries({ queryKey: ["agent-runs"] });
      router.push(`/agent/runs/${detail.run.id}`);
    },
  });

  if (!isLoggedIn) {
    return (
      <div className="mx-auto flex min-h-[60vh] max-w-3xl flex-col items-center justify-center px-4 pt-24 pb-12 text-center">
        <div className="rounded-full bg-orange-100 p-4 text-orange-700">
          <Bot className="h-8 w-8" />
        </div>
        <h1 className="mt-6 text-3xl font-semibold text-slate-950">Agent Workspace</h1>
        <p className="mt-3 max-w-xl text-sm leading-7 text-slate-500">
          这个页面用于发起和查看 Agent 任务运行。当前版本先开放登录用户使用。
        </p>
        <div className="mt-8 flex gap-3">
          <Button asChild>
            <Link href="/login">去登录</Link>
          </Button>
          <Button asChild variant="outline">
            <Link href="/feed">返回动态</Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-4 pt-24 pb-12">
      <div className="mb-8 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-orange-500">
            Site Agent
          </p>
          <h1 className="mt-2 text-4xl font-semibold tracking-tight text-slate-950">AI Agent</h1>
          <p className="mt-3 max-w-2xl text-sm leading-7 text-slate-500">
            这是一个独立的大页面，用来发起和管理 Agent 任务。当前开放的是 `发帖 Agent`，后续会继续扩到圈子与活动场景。
          </p>
        </div>
      </div>

      <div className="mb-6 grid gap-4 lg:grid-cols-3">
        <Card className="border-orange-200 bg-orange-50/70">
          <CardHeader>
            <CardTitle className="text-base text-slate-950">发帖 Agent</CardTitle>
          </CardHeader>
          <CardContent className="text-sm leading-6 text-slate-600">
            当前可用。适合整理标题、正文、标签和可见性建议，并在批准后回填到草稿。
          </CardContent>
        </Card>
        <Card className="border-slate-200">
          <CardHeader>
            <CardTitle className="text-base text-slate-950">圈子 Agent</CardTitle>
          </CardHeader>
          <CardContent className="text-sm leading-6 text-slate-500">
            即将开放。会围绕圈子规则、氛围和发帖方向给出连续执行建议。
          </CardContent>
        </Card>
        <Card className="border-slate-200">
          <CardHeader>
            <CardTitle className="text-base text-slate-950">活动 Agent</CardTitle>
          </CardHeader>
          <CardContent className="text-sm leading-6 text-slate-500">
            即将开放。会处理活动摘要、是否参加判断和行前准备清单。
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
        <Card className="border-slate-200">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-slate-950">
              <Sparkles className="h-5 w-5 text-orange-500" />
              发起任务
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {message ? (
              <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
                {message}
              </div>
            ) : null}
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-900">标题</label>
              <Input
                value={form.title || ""}
                onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
                placeholder="比如：润色我这条草稿"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-900">目标</label>
              <Textarea
                value={form.goal}
                onChange={(e) => setForm((prev) => ({ ...prev, goal: e.target.value }))}
                rows={5}
                placeholder="告诉 Agent 你希望它帮你完成什么"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-900">当前场景</label>
              <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
                发帖 Agent
              </div>
            </div>
            {hasExternalContext ? (
              <div className="rounded-2xl border border-orange-200 bg-orange-50/70 px-4 py-3">
                <p className="text-sm font-medium text-slate-900">已接收外部上下文</p>
                <p className="mt-1 text-sm leading-6 text-slate-600">
                  当前任务已经关联了一份页面草稿信息，Agent 会在后台自动读取并使用。
                </p>
              </div>
            ) : null}
            <Button
              onClick={() => createMutation.mutate(form)}
              disabled={createMutation.isPending || !form.goal?.trim()}
              className="bg-slate-950 text-white hover:bg-slate-800"
            >
              {createMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <PlayCircle className="mr-2 h-4 w-4" />
              )}
              创建运行
            </Button>
            {createMutation.error ? (
              <p className="text-sm text-rose-600">
                {createMutation.error instanceof Error ? createMutation.error.message : "创建失败"}
              </p>
            ) : null}
          </CardContent>
        </Card>

        <Card className="border-slate-200">
          <CardHeader>
            <CardTitle className="text-slate-950">最近运行</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {isLoading ? (
              <div className="flex items-center gap-2 text-sm text-slate-500">
                <Loader2 className="h-4 w-4 animate-spin" />
                正在读取运行记录
              </div>
            ) : null}
            {!isLoading && !(data?.runs?.length) ? (
              <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
                还没有运行记录。先创建一个发帖 Agent 任务。
              </div>
            ) : null}
            {(data?.runs || []).map((run: AgentRun) => (
              <Link
                key={run.id}
                href={`/agent/runs/${run.id}`}
                className="block rounded-2xl border border-slate-200 bg-white px-4 py-4 transition-colors hover:border-orange-300 hover:bg-orange-50/40"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-semibold text-slate-900">{run.title}</p>
                    <p className="mt-1 line-clamp-2 text-sm leading-6 text-slate-500">{run.goal}</p>
                  </div>
                  <StatusBadge status={run.status} />
                </div>
                <div className="mt-3 flex flex-wrap gap-3 text-xs text-slate-400">
                  <span>{new Date(run.updated_at).toLocaleString("zh-CN")}</span>
                </div>
                {run.latest_summary ? (
                  <p className="mt-2 text-sm leading-6 text-slate-500">{run.latest_summary}</p>
                ) : null}
              </Link>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default function AgentWorkspacePage() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-6xl px-4 pt-24 pb-12">
          <div className="flex items-center gap-2 text-sm text-slate-500">
            <Loader2 className="h-4 w-4 animate-spin" />
            正在加载 Agent 工作台
          </div>
        </div>
      }
    >
      <AgentWorkspaceContent />
    </Suspense>
  );
}
