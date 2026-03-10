'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import {
  AreaChart, Area, BarChart, Bar, XAxis, YAxis, Tooltip,
  ResponsiveContainer, CartesianGrid,
} from 'recharts'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Users, TrendingUp, FileText, Flag } from 'lucide-react'

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

interface Post {
  id: string
  title: string
  content: string
  author_username: string
  created_at: string
  moderation_status: string
}

interface Report {
  id: string
  target_type: string
  reason: string
  reporter_username: string
  created_at: string
}

export default function AdminDashboardPage() {
  const queryClient = useQueryClient()

  const { data: stats } = useQuery<DashboardStats>({
    queryKey: ['admin-stats'],
    queryFn: () => apiClient.get('/admin/stats/dashboard'),
  })

  const { data: growthData } = useQuery<ChartPoint[]>({
    queryKey: ['admin-user-growth', 30],
    queryFn: () => apiClient.get('/admin/stats/user-growth?days=30'),
    initialData: [],
  })

  const { data: postsData } = useQuery<{ posts: Post[]; total: number }>({
    queryKey: ['admin-posts-pending'],
    queryFn: () => apiClient.get('/admin/posts?status=pending&page_size=5'),
  })

  const { data: reportsData } = useQuery<{ reports: Report[]; total: number }>({
    queryKey: ['admin-reports-pending'],
    queryFn: () => apiClient.get('/admin/reports?status=pending&page_size=5'),
  })

  const approveMutation = useMutation({
    mutationFn: (id: string) =>
      apiClient.put(`/admin/posts/${id}/moderation`, { status: 'approved' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-posts-pending'] }),
  })

  const blockMutation = useMutation({
    mutationFn: (id: string) =>
      apiClient.put(`/admin/posts/${id}/moderation`, { status: 'blocked' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-posts-pending'] }),
  })

  const metrics = [
    { label: '总用户', value: stats?.total_users ?? 0, icon: Users, color: 'text-blue-600' },
    { label: '今日新增', value: stats?.new_users_today ?? 0, icon: TrendingUp, color: 'text-green-600' },
    { label: '帖子总数', value: stats?.total_posts ?? 0, icon: FileText, color: 'text-purple-600' },
    { label: '待处理举报', value: stats?.total_reports ?? 0, icon: Flag, color: 'text-red-600', alert: true },
  ]

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-white">总览</h1>

      {/* Metric cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {metrics.map(({ label, value, icon: Icon, color, alert }) => (
          <Card key={label}>
            <CardContent className="pt-5">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-gray-500">{label}</span>
                <Icon size={18} className={color} />
              </div>
              <div className={`text-3xl font-bold ${alert && value > 0 ? 'text-red-500' : 'text-gray-900 dark:text-white'}`}>
                {value.toLocaleString()}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">用户增长（30天）</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={growthData}>
                <defs>
                  <linearGradient id="purpleGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="date" tick={{ fontSize: 11 }} tickFormatter={(v) => v.slice(5)} />
                <YAxis tick={{ fontSize: 11 }} />
                <Tooltip formatter={(v) => [v, '新用户']} />
                <Area type="monotone" dataKey="value" stroke="#8b5cf6" fill="url(#purpleGrad)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">待审核帖子（最新 5 条）</CardTitle>
          </CardHeader>
          <CardContent>
            {(postsData?.posts ?? []).length === 0 ? (
              <div className="h-[200px] flex items-center justify-center text-gray-400 text-sm">暂无待审核帖子</div>
            ) : (
              <div className="space-y-2">
                {(postsData?.posts ?? []).map((p) => (
                  <div key={p.id} className="flex items-center justify-between py-1.5 border-b last:border-0">
                    <div className="flex-1 min-w-0 mr-3">
                      <p className="text-sm font-medium truncate">{p.title || p.content.slice(0, 40)}</p>
                      <p className="text-xs text-gray-400">@{p.author_username}</p>
                    </div>
                    <div className="flex gap-1 shrink-0">
                      <Button
                        size="sm"
                        variant="outline"
                        className="h-6 px-2 text-xs text-green-600 hover:text-green-700 hover:bg-green-50"
                        onClick={() => approveMutation.mutate(p.id)}
                        disabled={approveMutation.isPending}
                      >通过</Button>
                      <Button
                        size="sm"
                        variant="outline"
                        className="h-6 px-2 text-xs text-red-600 hover:text-red-700 hover:bg-red-50"
                        onClick={() => blockMutation.mutate(p.id)}
                        disabled={blockMutation.isPending}
                      >封禁</Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Latest reports */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">最新举报（待处理）</CardTitle>
        </CardHeader>
        <CardContent>
          {(reportsData?.reports ?? []).length === 0 ? (
            <div className="py-6 text-center text-gray-400 text-sm">暂无待处理举报</div>
          ) : (
            <div className="divide-y">
              {(reportsData?.reports ?? []).map((r) => (
                <div key={r.id} className="py-3 flex items-center gap-4">
                  <Badge variant="outline" className="shrink-0 text-xs">{r.target_type}</Badge>
                  <span className="flex-1 text-sm truncate">{r.reason}</span>
                  <span className="text-xs text-gray-400 shrink-0">
                    {r.reporter_username} · {new Date(r.created_at).toLocaleDateString('zh-CN')}
                  </span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
