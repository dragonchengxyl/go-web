"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Loader2, Square } from "lucide-react";
import { AgentRunStatus, apiClient } from "@/lib/api-client";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

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
  const queryClient = useQueryClient();
  const { isLoggedIn } = useAuth();

  const { data, isLoading, error } = useQuery({
    queryKey: ["agent-run", params.id],
    queryFn: () => apiClient.getAgentRun(params.id),
    enabled: isLoggedIn && !!params.id,
  });

  const cancelMutation = useMutation({
    mutationFn: () => apiClient.cancelAgentRun(params.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agent-run", params.id] });
      queryClient.invalidateQueries({ queryKey: ["agent-runs"] });
    },
  });

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
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      ) : null}
    </div>
  );
}
