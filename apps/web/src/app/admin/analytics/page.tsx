'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { BarChart3, FileText, Flag, TrendingUp, Users } from 'lucide-react'
import { apiClient } from '@/lib/api-client'
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

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow="Business Metrics"
        title="数据分析"
        description="围绕增长、内容存量和治理压力做基础监控，适合日常巡检和周报截图。"
      />

      <div className="flex flex-wrap gap-2">
        {ranges.map((range) => (
          <button
            key={range.days}
            onClick={() => setDays(range.days)}
            className={`rounded-full border px-4 py-2 text-sm font-medium transition-colors ${
              days === range.days
                ? 'border-slate-950 bg-slate-950 text-white'
                : 'border-slate-200 bg-white text-slate-500 hover:border-slate-300 hover:text-slate-900'
            }`}
          >
            {range.label}
          </button>
        ))}
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
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
          label="累计帖子"
          value={(stats?.total_posts ?? 0).toLocaleString()}
          hint="内容资产沉淀规模"
          icon={FileText}
          tone="default"
        />
        <AdminMetricCard
          label="待处理举报"
          value={(stats?.total_reports ?? 0).toLocaleString()}
          hint="当前治理压力指标"
          icon={Flag}
          tone={(stats?.total_reports ?? 0) > 0 ? 'warning' : 'success'}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.4fr)_minmax(320px,0.8fr)]">
        <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">{days} 天用户增长趋势</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={320}>
              <AreaChart data={growthData}>
                <defs>
                  <linearGradient id="analyticsGradient" x1="0" y1="0" x2="0" y2="1">
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
                  fill="url(#analyticsGradient)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <div className="space-y-4">
          <Card className="rounded-3xl border-slate-200 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
            <CardHeader className="pb-2">
              <CardTitle className="flex items-center gap-2 text-lg">
                <BarChart3 className="h-5 w-5" />
                运营解读
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm leading-6 text-slate-600">
              <p>
                当前图表展示的是最近 {days} 天的注册增长数据。适合配合总用户数、待审队列和举报量做日常值班判断。
              </p>
              <div className="rounded-2xl bg-slate-50 p-4">
                <p className="font-medium text-slate-900">建议重点关注</p>
                <ul className="mt-2 space-y-2 text-sm text-slate-600">
                  <li>新增用户上涨但帖子增长不明显，说明转化链路可能偏弱。</li>
                  <li>举报量持续抬升时，优先联动审核页和用户管理页。</li>
                  <li>每周固定截图此页，可直接沉淀为运营周报素材。</li>
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
                  Users
                </p>
                <p className="mt-2 text-2xl font-semibold text-slate-950">
                  {(stats?.total_users ?? 0).toLocaleString()}
                </p>
              </div>
              <div className="rounded-2xl border border-slate-200 p-4">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">
                  Content
                </p>
                <p className="mt-2 text-2xl font-semibold text-slate-950">
                  {(stats?.total_posts ?? 0).toLocaleString()}
                </p>
              </div>
              <div className="rounded-2xl border border-slate-200 p-4">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">
                  Reports
                </p>
                <p className="mt-2 text-2xl font-semibold text-slate-950">
                  {(stats?.total_reports ?? 0).toLocaleString()}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
