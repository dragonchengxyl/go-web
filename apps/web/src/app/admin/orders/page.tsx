'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'

interface Order {
  id: number
  order_no: string
  user_id: number
  username: string
  total_amount: number
  final_amount: number
  status: string
  payment_status: string
  payment_method: string
  created_at: string
}

export default function AdminOrdersPage() {
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')

  const { data: orders, isLoading } = useQuery<Order[]>({
    queryKey: ['admin-orders', search, statusFilter],
    queryFn: async () => {
      const response = await apiClient.get('/admin/orders', {
        params: { search, status: statusFilter !== 'all' ? statusFilter : undefined },
      })
      return response.data.data
    },
  })

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-500'
      case 'pending':
        return 'bg-yellow-500'
      case 'cancelled':
        return 'bg-red-500'
      default:
        return 'bg-gray-500'
    }
  }

  const getPaymentStatusColor = (status: string) => {
    switch (status) {
      case 'paid':
        return 'bg-green-500'
      case 'unpaid':
        return 'bg-yellow-500'
      case 'refunded':
        return 'bg-blue-500'
      default:
        return 'bg-gray-500'
    }
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <Link href="/admin">
          <Button variant="outline" className="mb-4">← 返回后台首页</Button>
        </Link>
        <h1 className="text-3xl font-bold mb-4">订单管理</h1>

        <div className="flex gap-4 mb-4">
          <div className="flex-1 max-w-md">
            <Input
              placeholder="搜索订单号或用户名..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border rounded-md"
          >
            <option value="all">全部状态</option>
            <option value="pending">待处理</option>
            <option value="completed">已完成</option>
            <option value="cancelled">已取消</option>
          </select>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>订单列表</CardTitle>
        </CardHeader>
        <CardContent>
          {!orders || orders.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              {search ? '未找到匹配的订单' : '暂无订单'}
            </div>
          ) : (
            <div className="space-y-4">
              {orders.map((order) => (
                <div key={order.id} className="p-4 border rounded-lg">
                  <div className="flex justify-between items-start mb-3">
                    <div>
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="font-bold">订单号: {order.order_no}</h3>
                        <Badge className={getStatusColor(order.status)}>
                          {order.status}
                        </Badge>
                        <Badge className={getPaymentStatusColor(order.payment_status)}>
                          {order.payment_status}
                        </Badge>
                      </div>
                      <div className="flex gap-6 text-sm text-gray-600">
                        <span>用户: {order.username}</span>
                        <span>支付方式: {order.payment_method}</span>
                        <span>创建: {new Date(order.created_at).toLocaleString('zh-CN')}</span>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-lg font-bold text-red-600">
                        ¥{order.final_amount.toFixed(2)}
                      </p>
                      {order.total_amount !== order.final_amount && (
                        <p className="text-sm text-gray-500 line-through">
                          ¥{order.total_amount.toFixed(2)}
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="flex justify-end">
                    <Link href={`/admin/orders/${order.id}`}>
                      <Button variant="outline" size="sm">查看详情</Button>
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
