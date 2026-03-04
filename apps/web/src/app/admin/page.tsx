'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface DashboardStats {
  online_users: number
  today_new_users: number
  today_downloads: number
  today_revenue: number
  total_users: number
  total_games: number
  total_orders: number
}

interface PopularGame {
  id: number
  title: string
  downloads: number
  revenue: number
}

export default function AdminDashboardPage() {
  const { data: stats, isLoading } = useQuery<DashboardStats>({
    queryKey: ['admin-stats'],
    queryFn: async () => {
      const response = await apiClient.get('/admin/stats/dashboard')
      return response
    },
  })

  const { data: popularGames } = useQuery<PopularGame[]>({
    queryKey: ['admin-popular-games'],
    queryFn: async () => {
      const response = await apiClient.get('/admin/stats/popular-games')
      return response
    },
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
        <Link href="/">
          <button className="text-sm text-gray-600 hover:text-gray-900">
            返回前台
          </button>
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-500">
              在线用户
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.online_users || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-500">
              今日新增
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.today_new_users || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-500">
              今日下载
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.today_downloads || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-500">
              今日收入
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">
              ¥{stats?.today_revenue?.toFixed(2) || '0.00'}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">总用户数</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_users || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">游戏总数</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_games || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">订单总数</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_orders || 0}</div>
          </CardContent>
        </Card>
      </div>

      <Card className="mb-8">
        <CardHeader>
          <CardTitle>热门游戏</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {popularGames?.map((game, index) => (
              <div key={game.id} className="flex items-center justify-between pb-4 border-b last:border-b-0">
                <div className="flex items-center gap-4">
                  <span className="text-2xl font-bold text-gray-400">#{index + 1}</span>
                  <div>
                    <p className="font-medium">{game.title}</p>
                    <p className="text-sm text-gray-500">{game.downloads} 次下载</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="font-bold text-green-600">¥{game.revenue.toFixed(2)}</p>
                  <p className="text-sm text-gray-500">收入</p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Tabs defaultValue="games" className="space-y-6">
        <TabsList>
          <TabsTrigger value="games">游戏管理</TabsTrigger>
          <TabsTrigger value="music">音乐管理</TabsTrigger>
          <TabsTrigger value="users">用户管理</TabsTrigger>
          <TabsTrigger value="orders">订单管理</TabsTrigger>
        </TabsList>

        <TabsContent value="games">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>游戏列表</CardTitle>
                <Link href="/admin/games">
                  <button className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
                    进入游戏管理
                  </button>
                </Link>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-gray-500">点击右上角进入完整的游戏管理界面</p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="music">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>音乐专辑</CardTitle>
                <Link href="/admin/music">
                  <button className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
                    进入音乐管理
                  </button>
                </Link>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-gray-500">点击右上角进入完整的音乐管理界面</p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="users">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>用户列表</CardTitle>
                <Link href="/admin/users">
                  <button className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
                    进入用户管理
                  </button>
                </Link>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-gray-500">点击右上角进入完整的用户管理界面</p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="orders">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>订单列表</CardTitle>
                <Link href="/admin/orders">
                  <button className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
                    进入订单管理
                  </button>
                </Link>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-gray-500">点击右上角进入完整的订单管理界面</p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
