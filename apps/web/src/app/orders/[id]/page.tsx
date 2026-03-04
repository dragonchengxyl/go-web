'use client'

import { useQuery } from '@tanstack/react-query'
import { useParams, useRouter } from 'next/navigation'
import { apiClient } from '@/lib/api-client'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

interface OrderDetail {
  id: number
  order_no: string
  user_id: number
  total_amount: number
  discount_amount: number
  final_amount: number
  status: string
  payment_method: string
  payment_status: string
  payment_time: string
  created_at: string
  updated_at: string
  items: Array<{
    id: number
    product_id: number
    product_name: string
    quantity: number
    price: number
    subtotal: number
  }>
}

export default function OrderDetailPage() {
  const params = useParams()
  const router = useRouter()
  const orderId = params.id as string

  const { data: order, isLoading } = useQuery({
    queryKey: ['order', orderId],
    queryFn: () => apiClient.getOrder(orderId),
  })

  const handlePayment = async () => {
    try {
      await apiClient.payOrder(orderId, order?.payment_method || 'alipay')
      alert('支付成功！')
      router.push('/orders')
    } catch (error) {
      alert('支付失败，请重试')
    }
  }

  const handleCancel = async () => {
    if (!confirm('确定要取消订单吗？')) return

    try {
      await apiClient.post(`/orders/${orderId}/cancel`)
      alert('订单已取消')
      router.push('/orders')
    } catch (error) {
      alert('取消失败，请重试')
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

  if (!order) {
    return (
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-8">
            <div className="text-center">订单不存在</div>
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
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          <Button variant="outline" onClick={() => router.back()} className="mb-4">
            返回
          </Button>

      <Card className="mb-6">
        <CardHeader>
          <div className="flex justify-between items-start">
            <div>
              <CardTitle>订单详情</CardTitle>
              <p className="text-sm text-gray-500 mt-2">订单号: {order.order_no}</p>
            </div>
            <div className="flex gap-2">
              <Badge className={
                order.status === 'completed' ? 'bg-green-500' :
                order.status === 'pending' ? 'bg-yellow-500' :
                order.status === 'cancelled' ? 'bg-red-500' : 'bg-gray-500'
              }>
                {order.status}
              </Badge>
              <Badge className={
                order.payment_status === 'paid' ? 'bg-green-500' :
                order.payment_status === 'unpaid' ? 'bg-yellow-500' :
                order.payment_status === 'refunded' ? 'bg-blue-500' : 'bg-gray-500'
              }>
                {order.payment_status}
              </Badge>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-gray-500">创建时间</p>
              <p className="font-medium">{new Date(order.created_at).toLocaleString('zh-CN')}</p>
            </div>
            {order.payment_time && (
              <div>
                <p className="text-gray-500">支付时间</p>
                <p className="font-medium">{new Date(order.payment_time).toLocaleString('zh-CN')}</p>
              </div>
            )}
            <div>
              <p className="text-gray-500">支付方式</p>
              <p className="font-medium">{order.payment_method}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="mb-6">
        <CardHeader>
          <CardTitle>商品清单</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {order.items.map((item: any) => (
              <div key={item.id} className="flex justify-between items-center pb-4 border-b last:border-b-0">
                <div>
                  <p className="font-medium">{item.product_name}</p>
                  <p className="text-sm text-gray-500">单价: ¥{(item.price / 100).toFixed(2)}</p>
                  <p className="text-sm text-gray-500">数量: {item.quantity}</p>
                </div>
                <p className="font-bold">¥{(item.subtotal / 100).toFixed(2)}</p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>费用明细</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <div className="flex justify-between">
              <span>商品总额</span>
              <span>¥{(order.total_amount / 100).toFixed(2)}</span>
            </div>
            {order.discount_amount > 0 && (
              <div className="flex justify-between text-green-600">
                <span>优惠金额</span>
                <span>-¥{(order.discount_amount / 100).toFixed(2)}</span>
              </div>
            )}
            <div className="border-t pt-2 flex justify-between font-bold text-lg">
              <span>实付金额</span>
              <span className="text-red-600">¥{(order.final_amount / 100).toFixed(2)}</span>
            </div>
          </div>

          <div className="mt-6 flex justify-end gap-2">
            {order.payment_status === 'unpaid' && order.status !== 'cancelled' && (
              <>
                <Button variant="outline" onClick={handleCancel}>
                  取消订单
                </Button>
                <Button onClick={handlePayment}>
                  继续支付
                </Button>
              </>
            )}
          </div>
        </CardContent>
      </Card>
        </div>
      </main>
      <Footer />
    </div>
  )
}
