'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

interface Order {
  id: number
  order_no: string
  user_id: number
  total_amount: number
  discount_amount: number
  final_amount: number
  status: string
  payment_method: string
  payment_status: string
  created_at: string
  items: Array<{
    id: number
    product_id: number
    product_name: string
    quantity: number
    price: number
    subtotal: number
  }>
}

export default function OrdersPage() {
  const { data: ordersData, isLoading } = useQuery({
    queryKey: ['orders'],
    queryFn: () => apiClient.getOrders(),
  })

  const orders = ordersData?.orders || [];

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
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-8">
            <div className="text-center">加载中...</div>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        <div className="container mx-auto px-4 py-8">
          <h1 className="text-3xl font-bold mb-8">我的订单</h1>

          {!orders || orders.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <p className="text-gray-500 mb-4">暂无订单</p>
                <Link href="/games">
                  <Button>去逛逛</Button>
                </Link>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {orders.map((order) => (
                <Card key={order.id}>
                  <CardHeader>
                    <div className="flex justify-between items-start">
                      <div>
                        <CardTitle className="text-lg">订单号: {order.order_no}</CardTitle>
                        <p className="text-sm text-gray-500 mt-1">
                          {new Date(order.created_at).toLocaleString('zh-CN')}
                        </p>
                      </div>
                      <div className="flex gap-2">
                        <Badge className={getStatusColor(order.status)}>
                          {order.status}
                        </Badge>
                        <Badge className={getPaymentStatusColor(order.payment_status)}>
                          {order.payment_status}
                        </Badge>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2 mb-4">
                      {order.items.map((item) => (
                        <div key={item.id} className="flex justify-between items-center">
                          <div>
                            <p className="font-medium">{item.product_name}</p>
                            <p className="text-sm text-gray-500">数量: {item.quantity}</p>
                          </div>
                          <p className="font-medium">¥{(item.subtotal / 100).toFixed(2)}</p>
                        </div>
                      ))}
                    </div>
                    <div className="border-t pt-4 space-y-1">
                      <div className="flex justify-between text-sm">
                        <span>商品总额</span>
                        <span>¥{(order.total_amount / 100).toFixed(2)}</span>
                      </div>
                      {order.discount_amount > 0 && (
                        <div className="flex justify-between text-sm text-green-600">
                          <span>优惠金额</span>
                          <span>-¥{(order.discount_amount / 100).toFixed(2)}</span>
                        </div>
                      )}
                      <div className="flex justify-between font-bold text-lg">
                        <span>实付金额</span>
                        <span className="text-red-600">¥{(order.final_amount / 100).toFixed(2)}</span>
                      </div>
                    </div>
                    <div className="mt-4 flex justify-end gap-2">
                      <Link href={`/orders/${order.id}`}>
                        <Button variant="outline">查看详情</Button>
                      </Link>
                      {order.payment_status === 'unpaid' && (
                        <Button>继续支付</Button>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </main>
      <Footer />
    </div>
  )
}
