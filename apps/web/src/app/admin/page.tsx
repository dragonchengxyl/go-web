'use client'

import Link from 'next/link'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  ArrowRight,
  BarChart3,
  CheckCircle2,
  Flag,
  FileText,
  ShieldCheck,
  TrendingUp,
  Users,
  XCircle,
} from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'

interface DashboardStats {
  total_users: number
  new_users_today: number
  total_posts: number
  total_reports: number
}

interface ChartPoint {
  date: string
  value: number
}

interface PostRow {
  id: string
  title: string
  content: string
  author_username: string
  created_at: string
  moderation_status: string
}

interface ReportRow {
  id: string
  target_type: string
  reason: string
  reporter_username: string
  created_at: string
  status: string
}

export default function AdminDashboardPage() {
  const queryClient = useQueryClient()

  const { data: stats } = useQuery<DashboardStats>({
    queryKey: ['admin-stats'],
    queryFn: () => apiClient.get('/admin/stats/dashboard'),
  })

  const { data: growthData = [] } = useQuery<ChartPoint[]>({
    queryKey: ['admin-user-growth', 30],
    queryFn: () => apiClient.get('/admin/stats/user-growth?days=30'),
  })

  const { data: postsData, isLoading: pendingPostsLoading } = useQuery<{ posts: PostRow[]; total: number }>({
    queryKey: ['admin-posts-pending'],
    queryFn: () => apiClient.get('/admin/posts?status=pending&page_size=5'),
  })

  const { data: reportsData, isLoading: reportsLoading } = useQuery<{ reports: ReportRow[]; total: number }>({
    queryKey: ['admin-reports-pending'],
    queryFn: () => apiClient.get('/admin/reports?status=pending&page_size=5'),
  })

  const moderateMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      apiClient.put(`/admin/posts/${id}/moderation`, { status }),
    onSuccess: (_, { status }) => {
      showAdminToast(status === 'approved' ? '帖子已通过审核' : '帖子已封禁', 'success')
      queryClient.invalidateQueries({ queryKey: ['admin-posts-pending'] })
      queryClient.invalidateQueries({ queryKey: ['admin-stats'] })
    },
    onError: () => {
      showAdminToast('操作失败，请重试', 'error')
    },
  })

  const postColumns: AdminColumn<PostRow>[] = [
    {
      key: 'content',
      header: '待审内容',
      className: 'min-w-[280px]',
      render: (post) => (
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <AdminStatusBadge value={post.moderation_status} label="待审核" />
            <span className="text-xs text-slate-400">@{post.author_username || '未知作者'}</span>
          </div>
          <div>
            <p className="font-medium text-slate-900">
              {post.title || '未命名帖子'}
            </p>
            <p className="mt-1 line-clamp-2 text-sm leading-6 text-slate-500">
              {post.content || '无正文'}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'created_at',
      header: '提交时间',
      className: 'whitespace-nowrap text-slate-500',
      render: (post) => new Date(post.created_at).toLocaleString('zh-CN'),
    },
    {
      key: 'actions',
      header: '处理',
      className: 'w-[210px]',
      render: (post) => (
        <div className="flex flex-wrap gap-2">
          <Button
            size="sm"
            variant="outline"
            className="border-emerald-200 text-emerald-700 hover:bg-emerald-50"
            disabled={moderateMutation.isPending}
            onClick={() => moderateMutation.mutate({ id: post.id, status: 'approved' })}
          >
            <CheckCircle2 className="mr-1 h-4 w-4" />
            通过
          </Button>
          <Button
            size="sm"
            variant="outline"
            className="border-rose-200 text-rose-700 hover:bg-rose-50"
            disabled={moderateMutation.isPending}
            onClick={() => moderateMutation.mutate({ id: post.id, status: 'blocked' })}
          >
            <XCircle className="mr-1 h-4 w-4" />
            封禁
          </Button>
        </div>
      ),
    },
  ]

  const reportColumns: AdminColumn<ReportRow>[] = [
    {
      key: 'target',
      header: '举报对象',
      className: 'min-w-[250px]',
      render: (report) => (
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <AdminStatusBadge value={report.status} label="待处理" />
            <AdminStatusBadge value={report.target_type} label={report.target_type} />
          </div>
          <p className="font-medium text-slate-900">{report.reason}</p>
          <p className="text-sm text-slate-500">举报人：{report.reporter_username || '未知用户'}</p>
        </div>
      ),
    },
    {
      key: 'created_at',
      header: '时间',
      className: 'whitespace-nowrap text-slate-500',
      render: (report) => new Date(report.created_at).toLocaleString('zh-CN'),
    },
    {
      key: 'link',
      header: '跳转',
      className: 'w-[120px]',
      render: () => (
        <Button asChild size="sm" variant="outline">
          <Link href="/admin/reports">
            去处理
            <ArrowRight className="ml-1 h-4 w-4" />
          </Link>
        </Button>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Operations Overview"
        title="运营总览"
        description="把核心指标、待办队列和治理入口集中在一个工作面里，适合日常巡检、审核和异常处理。"
        actions={
          <>
            <Button asChild variant="outline">
              <Link href="/admin/reports">查看举报队列</Link>
            </Button>
            <Button asChild className="bg-slate-950 text-white hover:bg-slate-800">
              <Link href="/admin/moderation">进入审核台</Link>
            </Button>
          </>
        }
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminMetricCard
          label="平台总用户"
          value={(stats?.total_users ?? 0).toLocaleString()}
          hint="累计注册用户规模"
          icon={Users}
          tone="brand"
        />
        <AdminMetricCard
          label="今日新增用户"
          value={(stats?.new_users_today ?? 0).toLocaleString()}
          hint="近 24 小时新增注册"
          icon={TrendingUp}
          tone="success"
        />
        <AdminMetricCard
          label="帖子存量"
          value={(stats?.total_posts ?? 0).toLocaleString()}
          hint="已入库内容总量"
          icon={FileText}
          tone="default"
        />
        <AdminMetricCard
          label="待处理举报"
          value={(stats?.total_reports ?? 0).toLocaleString()}
          hint="需要运营尽快处置"
          icon={Flag}
          tone={(stats?.total_reports ?? 0) > 0 ? 'danger' : 'warning'}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(320px,0.9fr)]">
        <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">30 天用户增长</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={280}>
              <AreaChart data={growthData}>
                <defs>
                  <linearGradient id="adminGrowth" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#0f172a" stopOpacity={0.28} />
                    <stop offset="95%" stopColor="#0f172a" stopOpacity={0.02} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                <XAxis
                  dataKey="date"
                  tick={{ fontSize: 12, fill: '#64748b' }}
                  tickFormatter={(value) => value.slice(5)}
                />
                <YAxis tick={{ fontSize: 12, fill: '#64748b' }} />
                <Tooltip formatter={(value) => [value, '新增用户']} />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke="#0f172a"
                  strokeWidth={2}
                  fill="url(#adminGrowth)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <div className="space-y-4">
          <AdminMetricCard
            label="待审帖子"
            value={(postsData?.total ?? 0).toLocaleString()}
            hint="建议优先处理，避免内容积压"
            icon={ShieldCheck}
            tone={(postsData?.total ?? 0) > 0 ? 'warning' : 'success'}
          />
          <AdminMetricCard
            label="待处理举报"
            value={(reportsData?.total ?? 0).toLocaleString()}
            hint="涉及违规反馈与封禁动作"
            icon={Flag}
            tone={(reportsData?.total ?? 0) > 0 ? 'danger' : 'success'}
          />

          <div className="rounded-3xl bg-slate-950 p-5 text-white shadow-[0_20px_60px_rgba(15,23,42,0.18)]">
            <p className="text-[11px] uppercase tracking-[0.24em] text-sky-300/80">
              Quick Actions
            </p>
            <h3 className="mt-2 text-xl font-semibold">今日运营动作</h3>
            <p className="mt-2 text-sm leading-6 text-slate-300">
              从这里快速进入高频工作面，减少在多个页面之间来回切换。
            </p>
            <div className="mt-5 grid gap-2">
              <Link
                href="/admin/moderation"
                className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm font-medium text-white transition-colors hover:bg-white/10"
              >
                去处理内容审核
              </Link>
              <Link
                href="/admin/reports"
                className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm font-medium text-white transition-colors hover:bg-white/10"
              >
                去处理举报闭环
              </Link>
              <Link
                href="/admin/users"
                className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm font-medium text-white transition-colors hover:bg-white/10"
              >
                去管理用户角色
              </Link>
            </div>
          </div>
        </div>
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <section className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-slate-950">待审核帖子</h2>
              <p className="text-sm text-slate-500">最新 5 条待处理内容</p>
            </div>
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/moderation">查看全部</Link>
            </Button>
          </div>
          <AdminDataTable
            data={postsData?.posts ?? []}
            columns={postColumns}
            keyExtractor={(post) => post.id}
            loading={pendingPostsLoading}
            empty={
              <AdminEmptyState
                title="没有待审核帖子"
                description="当前审核队列为空，可以把精力转到举报或用户治理。"
              />
            }
          />
        </section>

        <section className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-slate-950">最新待处理举报</h2>
              <p className="text-sm text-slate-500">优先处理高风险举报项</p>
            </div>
            <Button asChild variant="outline" size="sm">
              <Link href="/admin/reports">进入举报台</Link>
            </Button>
          </div>
          <AdminDataTable
            data={reportsData?.reports ?? []}
            columns={reportColumns}
            keyExtractor={(report) => report.id}
            loading={reportsLoading}
            empty={
              <AdminEmptyState
                title="没有待处理举报"
                description="当前举报队列为空，系统处于相对稳定状态。"
              />
            }
          />
        </section>
      </div>
    </div>
  )
}
