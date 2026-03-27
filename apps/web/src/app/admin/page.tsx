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
  Disc3,
  Flag,
  FileText,
  Gamepad2,
  Receipt,
  ShieldCheck,
  SlidersHorizontal,
  TrendingUp,
  Users,
  XCircle,
} from 'lucide-react'
import { AdminGameOverview, AuditLog, AdminSystemConfig, apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { AdminDataTable, type AdminColumn } from '@/components/admin/admin-data-table'
import { AdminEmptyState } from '@/components/admin/admin-empty-state'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'
import { AdminStatusBadge } from '@/components/admin/admin-status-badge'
import { showAdminToast } from '@/components/admin/admin-toast'
import {
  AdminSectionCard,
  AdminWorkspaceCard,
  formatDateTime,
} from '@/components/admin/admin-dashboard-kit'

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

function SignalCard({
  title,
  description,
  tone = 'default',
}: {
  title: string
  description: string
  tone?: 'default' | 'warning' | 'danger' | 'success'
}) {
  const className =
    tone === 'danger'
      ? 'border-rose-200 bg-rose-50'
      : tone === 'warning'
        ? 'border-amber-200 bg-amber-50'
        : tone === 'success'
          ? 'border-emerald-200 bg-emerald-50'
          : 'border-slate-200 bg-slate-50/70'

  return (
    <div className={`rounded-2xl border p-4 ${className}`}>
      <p className="font-medium text-slate-950">{title}</p>
      <p className="mt-3 text-sm leading-6 text-slate-500">{description}</p>
    </div>
  )
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

  const { data: pendingAudioData } = useQuery<{ total: number }>({
    queryKey: ['admin-dashboard-audio-pending-total'],
    queryFn: () => apiClient.getAdminAudioWorks({ status: 'pending', page_size: 1 }),
  })

  const { data: privateGroupsData } = useQuery<{ total: number }>({
    queryKey: ['admin-dashboard-private-groups-total'],
    queryFn: () => apiClient.getAdminGroups({ privacy: 'private', page_size: 1 }),
  })

  const { data: publishedEventsData } = useQuery<{ total: number }>({
    queryKey: ['admin-dashboard-published-events-total'],
    queryFn: () => apiClient.getAdminEvents({ status: 'published', page_size: 1 }),
  })

  const { data: ordersData } = useQuery<{ total: number }>({
    queryKey: ['admin-dashboard-orders-total'],
    queryFn: () => apiClient.getAdminOrders({ page_size: 1, type: 'tip' }),
  })

  const { data: pendingOrdersData } = useQuery<{ total: number }>({
    queryKey: ['admin-dashboard-orders-pending-total'],
    queryFn: () =>
      apiClient.getAdminOrders({ page_size: 1, type: 'tip', status: 'pending_payment' }),
  })

  const { data: failedOrdersData } = useQuery<{ total: number }>({
    queryKey: ['admin-dashboard-orders-failed-total'],
    queryFn: () =>
      apiClient.getAdminOrders({ page_size: 1, type: 'tip', status: 'failed' }),
  })

  const { data: gamesOverview } = useQuery<AdminGameOverview>({
    queryKey: ['admin-dashboard-games-overview'],
    queryFn: () => apiClient.getAdminGamesOverview(),
  })

  const { data: systemData } = useQuery<AdminSystemConfig>({
    queryKey: ['admin-dashboard-system-config'],
    queryFn: () => apiClient.getAdminSystemConfig(),
  })

  const { data: auditData } = useQuery<{ logs: AuditLog[]; total: number }>({
    queryKey: ['admin-dashboard-audit'],
    queryFn: () => apiClient.getAdminAuditLogs({ page: 1, page_size: 5 }),
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

  const queuePressure =
    (postsData?.total ?? 0) + (reportsData?.total ?? 0) + (pendingAudioData?.total ?? 0)
  const activeGameRooms =
    (gamesOverview?.hex_blitz.metrics.active_rooms ?? 0) +
    (gamesOverview?.doudizhu.metrics.active_rooms ?? 0)
  const activeGamePlayers =
    (gamesOverview?.hex_blitz.metrics.active_players ?? 0) +
    (gamesOverview?.doudizhu.metrics.active_players ?? 0)

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Operations Overview"
        title="运营总览"
        description="把核心指标、待办队列、仪表盘目录和跨域异常流集中在同一页，适合作为每天进入后台后的第一屏。"
        actions={
          <>
            <Button asChild variant="outline">
              <Link href="/admin/governance">进入治理盘</Link>
            </Button>
            <Button asChild className="border border-[#c5dfd3] bg-[#edf8f2] text-[#21584e] hover:bg-[#e3f3eb]">
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

      <AdminSectionCard
        title="仪表盘目录"
        description="后台不缺工作台，缺的是先去哪一块。先选领域仪表盘，再进入具体明细页处理。"
      >
        <div className="grid gap-4 xl:grid-cols-3">
          <AdminWorkspaceCard
            href="/admin/governance"
            title="内容治理盘"
            description="聚合帖子审核、举报、音频治理、评论巡检和审计流。"
            icon={ShieldCheck}
            tone="warning"
            metrics={[
              { label: '帖子', value: (postsData?.total ?? 0).toLocaleString() },
              { label: '举报', value: (reportsData?.total ?? 0).toLocaleString() },
              { label: '音频', value: (pendingAudioData?.total ?? 0).toLocaleString() },
            ]}
          />
          <AdminWorkspaceCard
            href="/admin/community"
            title="社区运营盘"
            description="统一看用户、圈子和活动，判断社区秩序与供给是否健康。"
            icon={Users}
            tone="brand"
            metrics={[
              { label: '用户', value: (stats?.total_users ?? 0).toLocaleString() },
              { label: '私密圈子', value: (privateGroupsData?.total ?? 0).toLocaleString() },
              { label: '已发布活动', value: (publishedEventsData?.total ?? 0).toLocaleString() },
            ]}
          />
          <AdminWorkspaceCard
            href="/admin/commerce"
            title="商业创作者盘"
            description="把订单、赞助和音频分发压力收在一个商业视角里。"
            icon={Receipt}
            tone="success"
            metrics={[
              { label: '订单', value: (ordersData?.total ?? 0).toLocaleString() },
              { label: '待支付', value: (pendingOrdersData?.total ?? 0).toLocaleString() },
              { label: '待审音频', value: (pendingAudioData?.total ?? 0).toLocaleString() },
            ]}
          />
          <AdminWorkspaceCard
            href="/admin/games"
            title="游戏运行态"
            description="同步观察 Hex Blitz 和斗地主的房间、对局与活跃状态。"
            icon={Gamepad2}
            tone="brand"
            metrics={[
              { label: '活跃房间', value: activeGameRooms.toLocaleString() },
              { label: '在线玩家', value: activeGamePlayers.toLocaleString() },
            ]}
          />
          <AdminWorkspaceCard
            href="/admin/platform"
            title="平台配置盘"
            description="集中看 AI 助手、系统配置、权限矩阵和后台审计动作。"
            icon={SlidersHorizontal}
            tone="default"
            metrics={[
              { label: '助手', value: systemData?.assistant.configured ? '已配置' : '未配置' },
              { label: '审计', value: (auditData?.total ?? 0).toLocaleString() },
              { label: '来源数', value: (systemData?.server.allow_origins.length ?? 0).toLocaleString() },
            ]}
          />
          <AdminWorkspaceCard
            href="/admin/analytics"
            title="数据分析"
            description="观察增长、内容规模和治理压力，适合做周报截图与趋势跟踪。"
            icon={BarChart3}
            tone="success"
            metrics={[
              { label: '今日新增', value: (stats?.new_users_today ?? 0).toLocaleString() },
              { label: '帖子', value: (stats?.total_posts ?? 0).toLocaleString() },
              { label: '举报', value: (stats?.total_reports ?? 0).toLocaleString() },
            ]}
          />
        </div>
      </AdminSectionCard>

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
            label="当前治理压力"
            value={queuePressure.toLocaleString()}
            hint="待审帖子 + 待处理举报 + 待审音频"
            icon={ShieldCheck}
            tone={queuePressure > 0 ? 'warning' : 'success'}
          />
          <AdminMetricCard
            label="游戏运行态"
            value={activeGameRooms.toLocaleString()}
            hint={`活跃玩家 ${activeGamePlayers.toLocaleString()} 人`}
            icon={Gamepad2}
            tone={activeGameRooms > 0 ? 'brand' : 'default'}
          />

          <div className="rounded-3xl border border-[#c8e3d7] bg-[linear-gradient(135deg,#f8fffb_0%,#ecf8f2_56%,#dff4eb_100%)] p-5 text-[#17342d] shadow-[0_20px_60px_rgba(34,94,80,0.12)]">
            <p className="text-[11px] uppercase tracking-[0.24em] text-[#5f8d81]">
              Quick Actions
            </p>
            <h3 className="mt-2 text-xl font-semibold">今天先处理什么</h3>
            <p className="mt-2 text-sm leading-6 text-[#5d8378]">
              先看治理和支付，再看系统配置和游戏运行态，减少在多个页面之间来回切换。
            </p>
            <div className="mt-5 grid gap-2">
              <Link
                href="/admin/governance"
                className="rounded-2xl border border-[#c5dfd3] bg-white px-4 py-3 text-sm font-medium text-[#21584e] transition-colors hover:bg-[#f0faf5]"
              >
                去内容治理盘
              </Link>
              <Link
                href="/admin/commerce"
                className="rounded-2xl border border-[#c5dfd3] bg-white px-4 py-3 text-sm font-medium text-[#21584e] transition-colors hover:bg-[#f0faf5]"
              >
                去商业创作者盘
              </Link>
              <Link
                href="/admin/platform"
                className="rounded-2xl border border-[#c5dfd3] bg-white px-4 py-3 text-sm font-medium text-[#21584e] transition-colors hover:bg-[#f0faf5]"
              >
                去平台配置盘
              </Link>
            </div>
            <div className="mt-5 flex flex-wrap gap-2">
              <Badge className="rounded-full border border-[#c5dfd3] bg-white text-[#52776d] hover:bg-white">
                赞助目标 {systemData?.sponsor.monthly_goal ? '已设置' : '未设置'}
              </Badge>
              <Badge className="rounded-full border border-[#f1d4c7] bg-[#fff5ef] text-[#bc5c3e] hover:bg-[#fff5ef]">
                支付失败 {(failedOrdersData?.total ?? 0).toLocaleString()} 笔
              </Badge>
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

      <AdminSectionCard
        title="跨域异常流"
        description="把治理、商业、平台和游戏四个域的风险信号收在一起，适合作为日常巡检和交班摘要。"
      >
        <div className="grid gap-6 xl:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
          <div className="grid gap-4 md:grid-cols-2">
            <SignalCard
              title="治理信号"
              tone={queuePressure > 0 ? 'warning' : 'success'}
              description={`当前累计治理压力 ${queuePressure.toLocaleString()}。如果还在抬升，优先去内容治理盘。`}
            />
            <SignalCard
              title="支付信号"
              tone={(failedOrdersData?.total ?? 0) > 0 ? 'danger' : 'default'}
              description={`支付失败 ${(failedOrdersData?.total ?? 0).toLocaleString()} 笔，待支付 ${(pendingOrdersData?.total ?? 0).toLocaleString()} 笔。`}
            />
            <SignalCard
              title="平台信号"
              tone={systemData?.assistant.configured ? 'default' : 'warning'}
              description={`AI 助手${systemData?.assistant.configured ? '已' : '未'}配置，允许来源 ${(systemData?.server.allow_origins.length ?? 0).toLocaleString()} 个。`}
            />
            <SignalCard
              title="游戏信号"
              tone={activeGameRooms > 0 ? 'default' : 'success'}
              description={`当前活跃房间 ${activeGameRooms.toLocaleString()} 个，活跃玩家 ${activeGamePlayers.toLocaleString()} 人。`}
            />
          </div>

          <div className="space-y-3">
            {(auditData?.logs ?? []).map((log) => (
              <div
                key={log.id}
                className="rounded-2xl border border-slate-200 bg-slate-50/70 px-4 py-4"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <p className="font-medium text-slate-950">
                    {log.action} · {log.resource}
                  </p>
                  <span className="text-xs text-slate-400">
                    {formatDateTime(log.created_at)}
                  </span>
                </div>
                <p className="mt-2 text-sm text-slate-500">
                  {log.username || '未知操作者'} · {log.resource_id || '无资源 ID'}
                </p>
                {log.error_message ? (
                  <p className="mt-2 text-sm text-rose-600">
                    错误：{log.error_message}
                  </p>
                ) : null}
              </div>
            ))}
            {(auditData?.logs?.length ?? 0) === 0 ? (
              <p className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
                当前没有最近后台动作记录。
              </p>
            ) : null}
          </div>
        </div>
      </AdminSectionCard>
    </div>
  )
}
