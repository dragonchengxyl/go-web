'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useCartStore } from '@/lib/store/cart';
import { formatPrice } from '@/lib/utils';
import { CreditCard, Smartphone, Loader2 } from 'lucide-react';

export default function CheckoutPage() {
  const router = useRouter();
  const { items, getTotalPrice, clearCart } = useCartStore();
  const [isProcessing, setIsProcessing] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState<'alipay' | 'wechat' | 'stripe'>('alipay');
  const [couponCode, setCouponCode] = useState('');

  const totalPrice = getTotalPrice();
  const discount = 0; // TODO: Implement coupon logic
  const finalPrice = totalPrice - discount;

  const handleCheckout = async () => {
    setIsProcessing(true);

    try {
      // Create order
      const response = await fetch('/api/v1/orders', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
        body: JSON.stringify({
          items: items.map(item => ({
            product_id: item.productId,
            quantity: item.quantity,
          })),
          coupon_code: couponCode || undefined,
          idempotency_key: `order_${Date.now()}`,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        const orderId = data.data.id;

        // Process payment
        const paymentResponse = await fetch(`/api/v1/orders/${orderId}/pay`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
          },
          body: JSON.stringify({
            payment_method: paymentMethod,
          }),
        });

        if (paymentResponse.ok) {
          clearCart();
          router.push(`/orders/${orderId}?success=true`);
        } else {
          throw new Error('Payment failed');
        }
      } else {
        throw new Error('Order creation failed');
      }
    } catch (error) {
      console.error('Checkout error:', error);
      alert('结账失败，请重试');
    } finally {
      setIsProcessing(false);
    }
  };

  if (items.length === 0) {
    router.push('/cart');
    return null;
  }

  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        <div className="container mx-auto px-4 py-12">
          <h1 className="text-4xl font-bold mb-8">结算</h1>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Payment Method */}
            <div className="lg:col-span-2 space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>选择支付方式</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <button
                    onClick={() => setPaymentMethod('alipay')}
                    className={`w-full p-4 border-2 rounded-lg flex items-center gap-4 transition-colors ${
                      paymentMethod === 'alipay'
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                  >
                    <div className="w-12 h-12 bg-blue-500 rounded-lg flex items-center justify-center">
                      <CreditCard className="h-6 w-6 text-white" />
                    </div>
                    <div className="flex-1 text-left">
                      <div className="font-semibold">支付宝</div>
                      <div className="text-sm text-muted-foreground">
                        使用支付宝扫码支付
                      </div>
                    </div>
                  </button>

                  <button
                    onClick={() => setPaymentMethod('wechat')}
                    className={`w-full p-4 border-2 rounded-lg flex items-center gap-4 transition-colors ${
                      paymentMethod === 'wechat'
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                  >
                    <div className="w-12 h-12 bg-green-500 rounded-lg flex items-center justify-center">
                      <Smartphone className="h-6 w-6 text-white" />
                    </div>
                    <div className="flex-1 text-left">
                      <div className="font-semibold">微信支付</div>
                      <div className="text-sm text-muted-foreground">
                        使用微信扫码支付
                      </div>
                    </div>
                  </button>

                  <button
                    onClick={() => setPaymentMethod('stripe')}
                    className={`w-full p-4 border-2 rounded-lg flex items-center gap-4 transition-colors ${
                      paymentMethod === 'stripe'
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                  >
                    <div className="w-12 h-12 bg-purple-500 rounded-lg flex items-center justify-center">
                      <CreditCard className="h-6 w-6 text-white" />
                    </div>
                    <div className="flex-1 text-left">
                      <div className="font-semibold">信用卡 / Stripe</div>
                      <div className="text-sm text-muted-foreground">
                        支持国际信用卡支付
                      </div>
                    </div>
                  </button>
                </CardContent>
              </Card>

              {/* Order Items */}
              <Card>
                <CardHeader>
                  <CardTitle>订单商品 ({items.length})</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {items.map((item) => (
                    <div key={item.id} className="flex items-center gap-4">
                      <div className="w-16 h-16 bg-muted rounded-lg flex-shrink-0">
                        <div className="w-full h-full bg-gradient-to-br from-primary/20 to-secondary/20 rounded-lg" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="font-medium truncate">{item.name}</div>
                        <div className="text-sm text-muted-foreground">
                          数量: {item.quantity}
                        </div>
                      </div>
                      <div className="font-semibold">
                        {formatPrice(item.price * item.quantity)}
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            </div>

            {/* Order Summary */}
            <div>
              <Card>
                <CardHeader>
                  <CardTitle>订单摘要</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Coupon */}
                  <div className="space-y-2">
                    <label className="text-sm font-medium">优惠券代码</label>
                    <div className="flex gap-2">
                      <Input
                        placeholder="输入优惠券"
                        value={couponCode}
                        onChange={(e) => setCouponCode(e.target.value)}
                      />
                      <Button variant="outline" size="sm">
                        应用
                      </Button>
                    </div>
                  </div>

                  {/* Price */}
                  <div className="space-y-2 pt-4 border-t">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">小计</span>
                      <span>{formatPrice(totalPrice)}</span>
                    </div>
                    {discount > 0 && (
                      <div className="flex justify-between text-primary">
                        <span>优惠</span>
                        <span>-{formatPrice(discount)}</span>
                      </div>
                    )}
                    <div className="flex justify-between text-lg font-bold pt-2 border-t">
                      <span>总计</span>
                      <span>{formatPrice(finalPrice)}</span>
                    </div>
                  </div>

                  <Button
                    onClick={handleCheckout}
                    disabled={isProcessing}
                    className="w-full"
                    size="lg"
                  >
                    {isProcessing ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        处理中...
                      </>
                    ) : (
                      `支付 ${formatPrice(finalPrice)}`
                    )}
                  </Button>

                  <p className="text-xs text-center text-muted-foreground">
                    点击支付即表示您同意我们的服务条款
                  </p>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
