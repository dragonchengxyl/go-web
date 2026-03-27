'use client'

import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import {
  AlertTriangle,
  ClipboardList,
  Disc3,
  Flag,
  MessageSquare,
  ShieldCheck,
} from 'lucide-react'
import { AdminAudioWork, AdminReport, AuditLog, apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import {
  AdminSectionCard,
  AdminWorkspaceCard,
  formatDateTime,
} from '@/components/admin/admin-dashboard-kit'

interface PostRow {
  id: string
  title: string
  content: string
  author_username?: string
  moderation_status: string
  created_at: string
}

interface CommentRow {
  id: string
  author_username?: string
  content: string
  commentable_type: string
  commentable_id: string
  created_at: string
  is_deleted: boolean
}

function QueueItem({
  eyebrow,
  title,
  description,
  meta,
}: {
  eyebrow: string
  title: string
  description?: string
  meta: string
}) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-4">
      <p className="text-[11px] uppercase tracking-[0.24em] text-slate-400">
        {eyebrow}
      </p>
      <p className="mt-2 font-medium text-slate-950">{title}</p>
      {description ? (
        <p className="mt-2 line-clamp-2 text-sm leading-6 text-slate-500">
          {description}
        </p>
      ) : null}
      <p className="mt-3 text-xs text-slate-400">{meta}</p>
    </div>
  )
}

export default function AdminGovernancePage() {
  const { data: postsData } = useQuery<{ posts: PostRow[]; total: number }>({
    queryKey: ['admin-dashboard-governance-posts'],
    queryFn: () => apiClient.get('/admin/posts?status=pending&page_size=5'),
  })

  const { data: reportsData } = useQuery<{ reports: AdminReport[]; total: number }>({
    queryKey: ['admin-dashboard-governance-reports'],
    queryFn: () => apiClient.getAdminReports({ status: 'pending', page_size: 5 }),
  })

  const { data: audioData } = useQuery<{ works: AdminAudioWork[]; total: number }>({
    queryKey: ['admin-dashboard-governance-audio'],
    queryFn: () => apiClient.getAdminAudioWorks({ status: 'pending', page_size: 5 }),
  })

  const { data: commentsData } = useQuery<{ comments: CommentRow[]; total: number }>({
    queryKey: ['admin-dashboard-governance-comments'],
    queryFn: () => apiClient.get('/admin/comments?page=1&page_size=5'),
  })

  const { data: auditData } = useQuery<{ logs: AuditLog[]; total: number }>({
    queryKey: ['admin-dashboard-governance-audit'],
    queryFn: () => apiClient.getAdminAuditLogs({ page: 1, page_size: 6 }),
  })

  const queuePressure =
    (postsData?.total ?? 0) + (reportsData?.total ?? 0) + (audioData?.total ?? 0)

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Governance Dashboard"
        title="内容治理盘"
        description="把帖子审核、举报处理、音频治理、评论巡检和审计流聚成同一个值班面，先看队列压力，再进入具体工作台。"
        actions={
          <>
            <Button asChild variant="outline">
              <Link href="/admin/audit-logs">查看审计日志</Link>
            </Button>
            <Button asChild className="border border-[#c5dfd3] bg-[#edf8f2] text-[#21584e] hover:bg-[#e3f3eb]">
              <Link href="/admin/moderation">进入审核台</Link>
            </Button>
          </>
        }
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminMetricCard
          label="待审帖子"
          value={(postsData?.total ?? 0).toLocaleString()}
          hint="需要优先处理的文本内容队列"
          icon={ShieldCheck}
          tone={(postsData?.total ?? 0) > 0 ? 'warning' : 'success'}
        />
        <AdminMetricCard
          label="待处理举报"
          value={(reportsData?.total ?? 0).toLocaleString()}
          hint="涉及封禁、删除和闭环动作"
          icon={Flag}
          tone={(reportsData?.total ?? 0) > 0 ? 'danger' : 'success'}
        />
        <AdminMetricCard
          label="待审音频"
          value={(audioData?.total ?? 0).toLocaleString()}
          hint="创作者发布前的公开分发关口"
          icon={Disc3}
          tone={(audioData?.total ?? 0) > 0 ? 'warning' : 'default'}
        />
        <AdminMetricCard
          label="评论巡检"
          value={(commentsData?.total ?? 0).toLocaleString()}
          hint="当前评论管理接口返回的总记录"
          icon={MessageSquare}
          tone="default"
        />
      </div>

      <div className="rounded-3xl border border-slate-200 bg-white p-5 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-[11px] uppercase tracking-[0.24em] text-slate-400">
              Queue Pressure
            </p>
            <h2 className="mt-2 text-xl font-semibold text-slate-950">
              当前治理压力 {queuePressure.toLocaleString()}
            </h2>
            <p className="mt-2 text-sm leading-6 text-slate-500">
              这个数字越高，越需要先处理审核、举报和音频分发关口，避免内容积压和风险扩散。
            </p>
          </div>
          <Badge
            variant="outline"
            className="w-fit rounded-full border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-600"
          >
            最近审计记录 {(auditData?.logs?.length ?? 0).toLocaleString()} 条
          </Badge>
        </div>
      </div>

      <div className="grid gap-4 xl:grid-cols-5">
        <AdminWorkspaceCard
          href="/admin/moderation"
          title="帖子审核"
          description="处理待审帖子、快速通过或封禁，并回到具体内容上下文。"
          icon={ShieldCheck}
          tone="warning"
          metrics={[
            { label: '待审', value: (postsData?.total ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/reports"
          title="举报处理"
          description="集中处理用户举报与动作闭环，适合高风险值班时段。"
          icon={Flag}
          tone="danger"
          metrics={[
            { label: '待处理', value: (reportsData?.total ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/audio-works"
          title="音频治理"
          description="处理音频作品审核、下架与发布前的治理环节。"
          icon={Disc3}
          tone="warning"
          metrics={[
            { label: '待审', value: (audioData?.total ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/comments"
          title="评论管理"
          description="巡检评论区质量，必要时快速下线高风险评论。"
          icon={MessageSquare}
          tone="default"
          metrics={[
            { label: '记录', value: (commentsData?.total ?? 0).toLocaleString() },
          ]}
        />
        <AdminWorkspaceCard
          href="/admin/audit-logs"
          title="审计留痕"
          description="追踪最近后台关键动作，帮助回溯风险处置链路。"
          icon={ClipboardList}
          tone="brand"
          metrics={[
            { label: '总量', value: (auditData?.total ?? 0).toLocaleString() },
          ]}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <AdminSectionCard
          title="优先内容队列"
          description="先看最新待审帖子和待处理举报，减少风险内容在站内停留时间。"
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/reports">进入举报台</Link>
            </Button>
          }
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-900">待审帖子</p>
              {(postsData?.posts ?? []).slice(0, 3).map((post) => (
                <QueueItem
                  key={post.id}
                  eyebrow="Post"
                  title={post.title || '未命名帖子'}
                  description={post.content}
                  meta={`@${post.author_username || '未知作者'} · ${formatDateTime(post.created_at)}`}
                />
              ))}
              {(postsData?.posts?.length ?? 0) === 0 ? (
                <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                  当前没有待审帖子。
                </p>
              ) : null}
            </div>

            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-900">待处理举报</p>
              {(reportsData?.reports ?? []).slice(0, 3).map((report) => (
                <QueueItem
                  key={report.id}
                  eyebrow="Report"
                  title={report.reason}
                  description={report.description}
                  meta={`${report.target_type} · ${report.reporter_username || '未知用户'} · ${formatDateTime(report.created_at)}`}
                />
              ))}
              {(reportsData?.reports?.length ?? 0) === 0 ? (
                <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                  当前没有待处理举报。
                </p>
              ) : null}
            </div>
          </div>
        </AdminSectionCard>

        <AdminSectionCard
          title="音频与评论巡检"
          description="补齐非文本内容和评论区的值班视角，避免只盯帖子审核。"
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/audio-works">进入音频治理</Link>
            </Button>
          }
        >
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-900">待审音频</p>
              {(audioData?.works ?? []).slice(0, 3).map((work) => (
                <QueueItem
                  key={work.id}
                  eyebrow="Audio"
                  title={work.title}
                  description={work.description}
                  meta={`@${work.author_username || work.author_id} · ${formatDateTime(work.published_at)}`}
                />
              ))}
              {(audioData?.works?.length ?? 0) === 0 ? (
                <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                  当前没有待审音频作品。
                </p>
              ) : null}
            </div>

            <div className="space-y-3">
              <p className="text-sm font-medium text-slate-900">最新评论巡检</p>
              {(commentsData?.comments ?? []).slice(0, 3).map((comment) => (
                <QueueItem
                  key={comment.id}
                  eyebrow="Comment"
                  title={comment.author_username ? `@${comment.author_username}` : '匿名评论'}
                  description={comment.content}
                  meta={`${comment.commentable_type} · ${formatDateTime(comment.created_at)}`}
                />
              ))}
              {(commentsData?.comments?.length ?? 0) === 0 ? (
                <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                  当前没有评论巡检数据。
                </p>
              ) : null}
            </div>
          </div>
        </AdminSectionCard>
      </div>

      <AdminSectionCard
        title="最近治理动作"
        description="值班时看这一列，可以快速判断今天后台做了哪些关键动作，是否存在异常集中的治理操作。"
      >
        <div className="space-y-3">
          {(auditData?.logs ?? []).map((log) => (
            <div
              key={log.id}
              className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-slate-50/70 px-4 py-4 md:flex-row md:items-center md:justify-between"
            >
              <div className="min-w-0">
                <p className="font-medium text-slate-950">
                  {log.action} · {log.resource}
                </p>
                <p className="mt-1 text-sm text-slate-500">
                  {log.username || '未知操作者'} · {log.resource_id || '无资源 ID'}
                </p>
              </div>
              <div className="flex items-center gap-2 text-xs text-slate-400">
                {log.error_message ? (
                  <Badge className="rounded-full bg-rose-100 text-rose-700">
                    <AlertTriangle className="mr-1 h-3.5 w-3.5" />
                    有错误
                  </Badge>
                ) : null}
                <span>{formatDateTime(log.created_at)}</span>
              </div>
            </div>
          ))}
          {(auditData?.logs?.length ?? 0) === 0 ? (
            <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
              当前没有可展示的审计记录。
            </p>
          ) : null}
        </div>
      </AdminSectionCard>
    </div>
  )
}
