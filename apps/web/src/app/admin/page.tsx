'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface DashboardStats {
  total_users: number
  new_users_today: number
  total_downloads: number
  downloads_today: number
  total_revenue: number
  revenue_today: number
  total_games: number
  total_orders: number
}

interface PopularGame {
  id: string
  title: string
  slug: string
  downloads: number
}

export default function AdminDashboardPage() {
  const { data: stats, isLoading } = useQuery<DashboardStats>({
    queryKey: ['admin-stats'],
    queryFn: () => apiClient.get('/admin/stats/dashboard'),
  })

  const { data: popularGames } = useQuery<PopularGame[]>({
    queryKey: ['admin-popular-games'],
    queryFn: () => apiClient.get('/admin/stats/popular-games'),
  })

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">管理后台</h1>
        <Link href="/" className="text-sm text-gray-600 hover:text-gray-900">
          返回前台
        </Link>
      </div>

      {/* Today stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-500">今日新增用户</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.new_users_today ?? 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-500">今日下载</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.downloads_today ?? 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-500">今日收入</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">
              ¥{stats?.revenue_today?.toFixed(2) ?? '0.00'}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-500">总收入</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">
              ¥{stats?.total_revenue?.toFixed(2) ?? '0.00'}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Cumulative stats */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 mb-8">
        <Card>
          <CardHeader><CardTitle className="text-sm font-medium">总用户数</CardTitle></CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_users ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-sm font-medium">游戏总数</CardTitle></CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_games ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-sm font-medium">总下载量</CardTitle></CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_downloads ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-sm font-medium">订单总数</CardTitle></CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_orders ?? 0}</div>
          </CardContent>
        </Card>
      </div>

      {/* Popular games */}
      <Card className="mb-8">
        <CardHeader>
          <CardTitle>热门游戏 (下载量)</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {(popularGames ?? []).map((game, index) => (
              <div key={game.id} className="flex items-center justify-between pb-4 border-b last:border-b-0">
                <div className="flex items-center gap-4">
                  <span className="text-2xl font-bold text-gray-400">#{index + 1}</span>
                  <div>
                    <p className="font-medium">{game.title}</p>
                    <p className="text-sm text-gray-500">{game.downloads} 次下载</p>
                  </div>
                </div>
              </div>
            ))}
            {(!popularGames || popularGames.length === 0) && (
              <p className="text-gray-500 text-center py-4">暂无数据</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Quick links */}
      <Tabs defaultValue="games" className="space-y-6">
        <TabsList>
          <TabsTrigger value="games">游戏管理</TabsTrigger>
          <TabsTrigger value="music">音乐管理</TabsTrigger>
          <TabsTrigger value="users">用户管理</TabsTrigger>
          <TabsTrigger value="orders">订单管理</TabsTrigger>
          <TabsTrigger value="comments">评论审核</TabsTrigger>
        </TabsList>

        {[
          { value: 'games', label: '游戏列表', href: '/admin/games' },
          { value: 'music', label: '音乐专辑', href: '/admin/music' },
          { value: 'users', label: '用户列表', href: '/admin/users' },
          { value: 'orders', label: '订单列表', href: '/admin/orders' },
          { value: 'comments', label: '评论审核', href: '/admin/comments' },
        ].map(({ value, label, href }) => (
          <TabsContent key={value} value={value}>
            <Card>
              <CardHeader>
                <div className="flex justify-between items-center">
                  <CardTitle>{label}</CardTitle>
                  <Link
                    href={href}
                    className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                  >
                    进入管理
                  </Link>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-gray-500">点击右上角进入完整的管理界面</p>
              </CardContent>
            </Card>
          </TabsContent>
        ))}
      </Tabs>
    </div>
  )
}
