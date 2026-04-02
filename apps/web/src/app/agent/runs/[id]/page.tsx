"use client";

import Link from "next/link";
import { useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Check, Loader2, Square, X } from "lucide-react";
import { AgentRunStatus, apiClient } from "@/lib/api-client";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { writeStoredPostDraft } from "@/lib/post-draft";

const STATUS_LABELS: Record<AgentRunStatus, string> = {
  queued: "排队中",
  running: "运行中",
  waiting_approval: "待审批",
  completed: "已完成",
  failed: "失败",
  cancelled: "已取消",
};

export default function AgentRunDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { isLoggedIn } = useAuth();

  const { data, isLoading, error } = useQuery({
    queryKey: ["agent-run", params.id],
    queryFn: () => apiClient.getAgentRun(params.id),
    enabled: isLoggedIn && !!params.id,
    refetchInterval: (query) => {
      const status = (query.state.data as { run?: { status?: AgentRunStatus } } | undefined)?.run?.status;
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

    void apiClient
      .streamAgentRun(params.id, {
        signal: controller.signal,
        onSnapshot: (detail) => {
          queryClient.setQueryData(["agent-run", params.id], detail);
          queryClient.invalidateQueries({ queryKey: ["agent-runs"] });
        },
        onDone: () => {
          queryClient.invalidateQueries({ queryKey: ["agent-runs"] });
        },
      })
      .catch(() => {
        // keep query-based refresh as fallback
      });

    return () => controller.abort();
  }, [isLoggedIn, params.id, queryClient]);

  if (!isLoggedIn) {
    return (
      <div className="mx-auto max-w-3xl px-4 pt-24 pb-12">
        <p className="text-sm text-slate-500">请先登录后再查看 Agent 运行详情。</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl px-4 pt-24 pb-12">
      <div className="mb-6 flex items-center justify-between gap-4">
        <Button asChild variant="outline">
          <Link href="/agent">
            <ArrowLeft className="mr-2 h-4 w-4" />
            返回工作台
          </Link>
        </Button>
        {data?.run && data.run.status !== "completed" && data.run.status !== "failed" && data.run.status !== "cancelled" ? (
          <Button
            variant="outline"
            onClick={() => cancelMutation.mutate()}
            disabled={cancelMutation.isPending}
          >
            {cancelMutation.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Square className="mr-2 h-4 w-4" />
            )}
            取消任务
          </Button>
        ) : null}
      </div>

      {isLoading ? (
        <div className="flex items-center gap-2 text-sm text-slate-500">
          <Loader2 className="h-4 w-4 animate-spin" />
          正在读取运行详情
        </div>
      ) : null}

      {error ? (
        <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-4 text-sm text-rose-700">
          {error instanceof Error ? error.message : "读取运行详情失败"}
        </div>
      ) : null}

      {data?.run ? (
        <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <Card className="border-slate-200">
            <CardHeader>
              <CardTitle className="text-slate-950">{data.run.title}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-wrap gap-3 text-sm text-slate-500">
                <span>状态：{STATUS_LABELS[data.run.status]}</span>
                <span>场景：{data.run.scenario}</span>
                <span>更新时间：{new Date(data.run.updated_at).toLocaleString("zh-CN")}</span>
              </div>
              {(data.run.status === "queued" || data.run.status === "running") ? (
                <div className="rounded-2xl border border-cyan-200 bg-cyan-50 px-4 py-3 text-sm text-cyan-700">
                  Agent 正在后台执行，这个页面会自动刷新运行状态。
                </div>
              ) : null}
              <div>
                <p className="text-sm font-medium text-slate-900">任务目标</p>
                <p className="mt-2 whitespace-pre-wrap text-sm leading-7 text-slate-600">
                  {data.run.goal}
                </p>
              </div>
              {data.run.latest_summary ? (
                <div>
                  <p className="text-sm font-medium text-slate-900">当前摘要</p>
                  <p className="mt-2 text-sm leading-7 text-slate-600">{data.run.latest_summary}</p>
                </div>
              ) : null}
              {data.run.context_snapshot ? (
                <div>
                  <p className="text-sm font-medium text-slate-900">上下文快照</p>
                  <pre className="mt-2 whitespace-pre-wrap break-all rounded-2xl bg-slate-50 px-4 py-3 text-xs leading-6 text-slate-600">
                    {JSON.stringify(data.run.context_snapshot, null, 2)}
                  </pre>
                </div>
              ) : null}
            </CardContent>
          </Card>

          <div className="space-y-6">
            <Card className="border-slate-200">
              <CardHeader>
                <CardTitle className="text-slate-950">执行时间线</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {!(data.steps?.length) ? (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-4 text-sm text-slate-500">
                    还没有步骤记录。
                  </div>
                ) : null}
                {(data.steps || []).map((step) => (
                  <div key={step.id} className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                    <div className="flex items-center justify-between gap-3">
                      <div>
                        <p className="text-sm font-semibold text-slate-900">
                          {step.step_index}. {step.title}
                        </p>
                        <p className="mt-1 text-xs uppercase tracking-[0.16em] text-slate-400">
                          {step.kind} · {step.status}
                        </p>
                      </div>
                    </div>
                    {step.summary ? (
                      <p className="mt-3 text-sm leading-6 text-slate-600">{step.summary}</p>
                    ) : null}
                    {step.error_text ? (
                      <p className="mt-3 text-sm leading-6 text-rose-600">{step.error_text}</p>
                    ) : null}
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card className="border-slate-200">
              <CardHeader>
                <CardTitle className="text-slate-950">审批动作</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {!(data.approvals?.length) ? (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-4 text-sm text-slate-500">
                    当前没有待处理的审批动作。
                  </div>
                ) : null}
                {(data.approvals || []).map((approval) => (
                  <div key={approval.id} className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-sm font-semibold text-slate-900">{approval.title}</p>
                        <p className="mt-1 text-xs uppercase tracking-[0.16em] text-slate-400">
                          {approval.action_type} · {approval.status}
                        </p>
                      </div>
                    </div>
                    {approval.status === "pending" ? (
                      <div className="mt-4 flex flex-wrap gap-2">
                        <Button
                          size="sm"
                          onClick={() =>
                            approvalMutation.mutate({ approvalId: approval.id, decision: "approved" })
                          }
                          disabled={approvalMutation.isPending}
                          className="bg-slate-950 text-white hover:bg-slate-800"
                        >
                          {approvalMutation.isPending ? (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          ) : (
                            <Check className="mr-2 h-4 w-4" />
                          )}
                          批准并回填草稿
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() =>
                            approvalMutation.mutate({ approvalId: approval.id, decision: "rejected" })
                          }
                          disabled={approvalMutation.isPending}
                        >
                          <X className="mr-2 h-4 w-4" />
                          暂不回填
                        </Button>
                      </div>
                    ) : null}
                    {approval.payload ? (
                      <pre className="mt-3 whitespace-pre-wrap break-all rounded-2xl bg-slate-50 px-4 py-3 text-xs leading-6 text-slate-600">
                        {JSON.stringify(approval.payload, null, 2)}
                      </pre>
                    ) : null}
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card className="border-slate-200">
              <CardHeader>
                <CardTitle className="text-slate-950">只读工具调用</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {!(data.tool_calls?.length) ? (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-4 text-sm text-slate-500">
                    还没有工具调用记录。
                  </div>
                ) : null}
                {(data.tool_calls || []).map((call) => (
                  <div key={call.id} className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                    <div className="flex items-center justify-between gap-3">
                      <div>
                        <p className="text-sm font-semibold text-slate-900">{call.tool_name}</p>
                        <p className="mt-1 text-xs uppercase tracking-[0.16em] text-slate-400">
                          {call.access_level} · {call.status}
                        </p>
                      </div>
                    </div>
                    {call.output_data ? (
                      <pre className="mt-3 whitespace-pre-wrap break-all rounded-2xl bg-slate-50 px-4 py-3 text-xs leading-6 text-slate-600">
                        {JSON.stringify(call.output_data, null, 2)}
                      </pre>
                    ) : null}
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card className="border-slate-200">
              <CardHeader>
                <CardTitle className="text-slate-950">结构化产物</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {!(data.artifacts?.length) ? (
                  <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-4 text-sm text-slate-500">
                    当前还没有生成产物。
                  </div>
                ) : null}
                {(data.artifacts || []).map((artifact) => (
                  <div key={artifact.id} className="rounded-2xl border border-slate-200 bg-white px-4 py-4">
                    <p className="text-sm font-semibold text-slate-900">{artifact.title}</p>
                    <p className="mt-1 text-xs uppercase tracking-[0.16em] text-slate-400">
                      {artifact.kind}
                    </p>
                    {artifact.content ? (
                      <div className="mt-3 whitespace-pre-wrap text-sm leading-6 text-slate-600">
                        {artifact.content}
                      </div>
                    ) : null}
                    {artifact.structured_data ? (
                      <pre className="mt-3 whitespace-pre-wrap break-all rounded-2xl bg-slate-50 px-4 py-3 text-xs leading-6 text-slate-600">
                        {JSON.stringify(artifact.structured_data, null, 2)}
                      </pre>
                    ) : null}
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </div>
      ) : null}
    </div>
  );
}
