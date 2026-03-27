'use client'

import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ComposedChart,
  Legend,
  Line,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  BarChart3,
  Disc3,
  Flag,
  Gamepad2,
  Receipt,
  TrendingUp,
  Users,
} from 'lucide-react'
import { AdminGameOverview, apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { AdminMetricCard } from '@/components/admin/admin-metric-card'
import { AdminPageHeader } from '@/components/admin/admin-page-header'

interface ChartPoint {
  date: string
  value: number
}

interface DashboardStats {
  total_users: number
  new_users_today: number
  total_posts: number
  total_reports: number
}

const ranges = [
  { label: '7 天', days: 7 },
  { label: '30 天', days: 30 },
  { label: '90 天', days: 90 },
]

const pieColors = ['#2f7a67', '#45a486', '#f0b95d', '#de8364', '#7ab7a2']

export default function AdminAnalyticsPage() {
  const [days, setDays] = useState(30)

  const { data: growthData = [] } = useQuery<ChartPoint[]>({
    queryKey: ['admin-user-growth', days],
    queryFn: () => apiClient.get(`/admin/stats/user-growth?days=${days}`),
  })

  const { data: stats } = useQuery<DashboardStats>({
    queryKey: ['admin-stats'],
    queryFn: () => apiClient.get('/admin/stats/dashboard'),
  })

  const { data: pendingReportsData } = useQuery<{ total: number }>({
    queryKey: ['admin-analytics-pending-reports'],
    queryFn: () => apiClient.getAdminReports({ status: 'pending', page_size: 1 }),
  })

  const { data: pendingAudioData } = useQuery<{ total: number }>({
    queryKey: ['admin-analytics-pending-audio'],
    queryFn: () => apiClient.getAdminAudioWorks({ status: 'pending', page_size: 1 }),
  })

  const { data: ordersData } = useQuery<{ total: number }>({
    queryKey: ['admin-analytics-orders'],
    queryFn: () => apiClient.getAdminOrders({ page_size: 1, type: 'tip' }),
  })

  const { data: pendingOrdersData } = useQuery<{ total: number }>({
    queryKey: ['admin-analytics-pending-orders'],
    queryFn: () =>
      apiClient.getAdminOrders({ page_size: 1, type: 'tip', status: 'pending_payment' }),
  })

  const { data: privateGroupsData } = useQuery<{ total: number }>({
    queryKey: ['admin-analytics-private-groups'],
    queryFn: () => apiClient.getAdminGroups({ page_size: 1, privacy: 'private' }),
  })

  const { data: publishedEventsData } = useQuery<{ total: number }>({
    queryKey: ['admin-analytics-published-events'],
    queryFn: () => apiClient.getAdminEvents({ page_size: 1, status: 'published' }),
  })

  const { data: gamesOverview } = useQuery<AdminGameOverview>({
    queryKey: ['admin-analytics-games-overview'],
    queryFn: () => apiClient.getAdminGamesOverview(),
  })

  const cumulativeData = useMemo(() => {
    const totalUsers = stats?.total_users ?? 0
    const totalGrowth = growthData.reduce((sum, item) => sum + item.value, 0)
    let running = Math.max(0, totalUsers - totalGrowth)

    return growthData.map((item) => {
      running += item.value
      return {
        date: item.date,
        new_users: item.value,
        total_users: running,
      }
    })
  }, [growthData, stats?.total_users])

  const queueDistribution = useMemo(
    () => [
      { name: '待处理举报', value: pendingReportsData?.total ?? 0 },
      { name: '待审音频', value: pendingAudioData?.total ?? 0 },
      { name: '待支付订单', value: pendingOrdersData?.total ?? 0 },
      {
        name: '活跃房间',
        value:
          (gamesOverview?.hex_blitz.metrics.active_rooms ?? 0) +
          (gamesOverview?.doudizhu.metrics.active_rooms ?? 0),
      },
    ].filter((item) => item.value > 0),
    [gamesOverview, pendingAudioData?.total, pendingOrdersData?.total, pendingReportsData?.total],
  )

  const businessSnapshot = useMemo(
    () => [
      { label: '帖子', value: stats?.total_posts ?? 0, fill: '#2f7a67' },
      { label: '用户', value: stats?.total_users ?? 0, fill: '#45a486' },
      { label: '订单', value: ordersData?.total ?? 0, fill: '#f0b95d' },
      { label: '私密圈子', value: privateGroupsData?.total ?? 0, fill: '#de8364' },
      { label: '已发布活动', value: publishedEventsData?.total ?? 0, fill: '#7ab7a2' },
    ],
    [
      ordersData?.total,
      privateGroupsData?.total,
      publishedEventsData?.total,
      stats?.total_posts,
      stats?.total_users,
    ],
  )

  const activeGamePlayers =
    (gamesOverview?.hex_blitz.metrics.active_players ?? 0) +
    (gamesOverview?.doudizhu.metrics.active_players ?? 0)

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Business Metrics"
        title="数据分析"
        description="把增长、队列压力和业务快照放进更直观的图形视图里，适合值班巡检、周报截图和趋势复盘。"
      />

      <div className="flex flex-wrap gap-2">
        {ranges.map((range) => (
          <button
            key={range.days}
            onClick={() => setDays(range.days)}
            className={`rounded-full border px-4 py-2 text-sm font-medium transition-colors ${
              days === range.days
                ? 'border-[#21584e] bg-[#21584e] text-[#f4fffb]'
                : 'border-slate-200 bg-white text-slate-500 hover:border-slate-300 hover:text-slate-900'
            }`}
          >
            {range.label}
          </button>
        ))}
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <AdminMetricCard
          label="累计用户"
          value={(stats?.total_users ?? 0).toLocaleString()}
          hint="当前平台用户总量"
          icon={Users}
          tone="brand"
        />
        <AdminMetricCard
          label="今日新增"
          value={(stats?.new_users_today ?? 0).toLocaleString()}
          hint="近 24 小时注册增量"
          icon={TrendingUp}
          tone="success"
        />
        <AdminMetricCard
          label="待处理举报"
          value={(pendingReportsData?.total ?? 0).toLocaleString()}
          hint="治理盘里的第一优先队列"
          icon={Flag}
          tone={(pendingReportsData?.total ?? 0) > 0 ? 'warning' : 'success'}
        />
        <AdminMetricCard
          label="待审音频"
          value={(pendingAudioData?.total ?? 0).toLocaleString()}
          hint="创作者公开分发前的积压"
          icon={Disc3}
          tone={(pendingAudioData?.total ?? 0) > 0 ? 'warning' : 'default'}
        />
        <AdminMetricCard
          label="活跃玩家"
          value={activeGamePlayers.toLocaleString()}
          hint="来自 Hex Blitz 与斗地主的实时玩家数"
          icon={Gamepad2}
          tone={activeGamePlayers > 0 ? 'brand' : 'default'}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.45fr)_minmax(320px,0.95fr)]">
        <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">{days} 天新增与累计用户</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={340}>
              <ComposedChart data={cumulativeData}>
                <defs>
                  <linearGradient id="analyticsBarGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#73d4be" stopOpacity={0.9} />
                    <stop offset="95%" stopColor="#73d4be" stopOpacity={0.2} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#d8e9df" />
                <XAxis
                  dataKey="date"
                  tick={{ fontSize: 12, fill: '#64897e' }}
                  tickFormatter={(value) => value.slice(5)}
                />
                <YAxis
                  yAxisId="left"
                  tick={{ fontSize: 12, fill: '#64897e' }}
                />
                <YAxis
                  yAxisId="right"
                  orientation="right"
                  tick={{ fontSize: 12, fill: '#64897e' }}
                />
                <Tooltip />
                <Legend />
                <Bar
                  yAxisId="left"
                  dataKey="new_users"
                  name="新增用户"
                  radius={[10, 10, 0, 0]}
                  fill="url(#analyticsBarGradient)"
                />
                <Line
                  yAxisId="right"
                  type="monotone"
                  dataKey="total_users"
                  name="累计用户"
                  stroke="#21584e"
                  strokeWidth={3}
                  dot={{ r: 2, fill: '#21584e' }}
                  activeDot={{ r: 5 }}
                />
              </ComposedChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">当前队列压力分布</CardTitle>
          </CardHeader>
          <CardContent>
            {queueDistribution.length > 0 ? (
              <ResponsiveContainer width="100%" height={340}>
                <PieChart>
                  <Pie
                    data={queueDistribution}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    innerRadius={66}
                    outerRadius={118}
                    paddingAngle={4}
                  >
                    {queueDistribution.map((item, index) => (
                      <Cell
                        key={item.name}
                        fill={pieColors[index % pieColors.length]}
                      />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-[340px] items-center justify-center rounded-2xl border border-dashed border-slate-200 bg-slate-50/70 text-sm text-slate-500">
                当前没有明显的待处理压力。
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
        <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">业务体量快照</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={320}>
              <BarChart data={businessSnapshot}>
                <CartesianGrid strokeDasharray="3 3" stroke="#d8e9df" />
                <XAxis dataKey="label" tick={{ fontSize: 12, fill: '#64897e' }} />
                <YAxis tick={{ fontSize: 12, fill: '#64897e' }} />
                <Tooltip />
                <Bar dataKey="value" radius={[12, 12, 0, 0]}>
                  {businessSnapshot.map((item) => (
                    <Cell key={item.label} fill={item.fill} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <div className="space-y-4">
          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardHeader className="pb-2">
              <CardTitle className="flex items-center gap-2 text-lg">
                <BarChart3 className="h-5 w-5" />
                图表解读
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm leading-6 text-slate-600">
              <p>
                上方组合图适合同时观察阶段性拉新波动和平台累计用户规模，尤其适合做周报截图和活动复盘。
              </p>
              <div className="rounded-2xl bg-slate-50 p-4">
                <p className="font-medium text-slate-900">建议重点关注</p>
                <ul className="mt-2 space-y-2 text-sm text-slate-600">
                  <li>新增用户上涨但帖子存量增幅不明显，说明内容转化链路偏弱。</li>
                  <li>待处理举报、待审音频和待支付订单同时抬升时，说明运营与商业侧同时承压。</li>
                  <li>活跃玩家数上升但订单和内容未同步增长时，可以回看游戏带动的转化链路。</li>
                </ul>
              </div>
            </CardContent>
          </Card>

          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardHeader className="pb-2">
              <CardTitle className="text-lg">快照</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="rounded-2xl border border-slate-200 p-4">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">
                  Orders
                </p>
                <p className="mt-2 text-2xl font-semibold text-slate-950">
                  {(ordersData?.total ?? 0).toLocaleString()}
                </p>
                <p className="mt-2 text-sm text-slate-500">
                  待支付 {(pendingOrdersData?.total ?? 0).toLocaleString()} 笔
                </p>
              </div>
              <div className="rounded-2xl border border-slate-200 p-4">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">
                  Community
                </p>
                <p className="mt-2 text-2xl font-semibold text-slate-950">
                  {(privateGroupsData?.total ?? 0).toLocaleString()}
                </p>
                <p className="mt-2 text-sm text-slate-500">
                  私密圈子 · 已发布活动 {(publishedEventsData?.total ?? 0).toLocaleString()} 个
                </p>
              </div>
              <div className="rounded-2xl border border-slate-200 p-4">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">
                  Runtime
                </p>
                <p className="mt-2 text-2xl font-semibold text-slate-950">
                  {activeGamePlayers.toLocaleString()}
                </p>
                <p className="mt-2 text-sm text-slate-500">
                  当前游戏活跃玩家
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
