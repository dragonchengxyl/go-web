'use client';

export const dynamic = 'force-dynamic';

import { useState } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { useCartStore } from '@/lib/store/cart';
import { formatPrice } from '@/lib/utils';
import { Trash2, Plus, Minus, ShoppingBag, Loader2 } from 'lucide-react';
import { apiClient } from '@/lib/api-client';

export default function CartPage() {
  const { items, removeItem, updateQuantity, getTotalPrice, clearCart } = useCartStore();
  const [couponCode, setCouponCode] = useState('');
  const [appliedCoupon, setAppliedCoupon] = useState<{
    code: string;
    discount: number;
    discount_type: string;
  } | null>(null);
  const [couponError, setCouponError] = useState('');
  const [isValidating, setIsValidating] = useState(false);

  const totalPrice = getTotalPrice();

  const discount = appliedCoupon
    ? appliedCoupon.discount_type === 'percent'
      ? Math.round(totalPrice * appliedCoupon.discount) / 100
      : appliedCoupon.discount
    : 0;

  const finalPrice = totalPrice - discount;

  const handleValidateCoupon = async () => {
    if (!couponCode.trim()) return;
    setIsValidating(true);
    setCouponError('');
    setAppliedCoupon(null);
    try {
      const result = await apiClient.validateCoupon(couponCode.trim());
      if (result.valid) {
        setAppliedCoupon({
          code: couponCode.trim(),
          discount: result.discount,
          discount_type: result.discount_type,
        });
      } else {
        setCouponError('优惠券无效或已过期');
      }
    } catch (err: any) {
      setCouponError(err.message || '验证失败，请重试');
    } finally {
      setIsValidating(false);
    }
  };

  const handleRemoveCoupon = () => {
    setAppliedCoupon(null);
    setCouponCode('');
    setCouponError('');
  };

  if (items.length === 0) {
    return (
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-20">
            <div className="max-w-md mx-auto text-center">
              <ShoppingBag className="h-24 w-24 mx-auto text-muted-foreground mb-6" />
              <h1 className="text-3xl font-bold mb-4">购物车是空的</h1>
              <p className="text-muted-foreground mb-8">
                还没有添加任何商品，去看看有什么好玩的游戏吧！
              </p>
              <Button asChild size="lg">
                <Link href="/games">浏览游戏</Link>
              </Button>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        <div className="container mx-auto px-4 py-12">
          <h1 className="text-4xl font-bold mb-8">购物车</h1>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Cart Items */}
            <div className="lg:col-span-2 space-y-4">
              {items.map((item) => (
                <Card key={item.id}>
                  <CardContent className="p-6">
                    <div className="flex gap-4">
                      {/* Image */}
                      <div className="relative w-24 h-24 bg-muted rounded-lg overflow-hidden flex-shrink-0">
                        <div className="w-full h-full bg-gradient-to-br from-primary/20 to-secondary/20" />
                      </div>

                      {/* Info */}
                      <div className="flex-1 min-w-0">
                        <h3 className="font-bold text-lg mb-2">{item.name}</h3>
                        <p className="text-2xl font-bold text-primary">
                          {formatPrice(item.price)}
                        </p>
                      </div>

                      {/* Quantity Controls */}
                      <div className="flex flex-col items-end gap-4">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => removeItem(item.id)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>

                        <div className="flex items-center gap-2">
                          <Button
                            variant="outline"
                            size="icon"
                            onClick={() =>
                              updateQuantity(item.id, item.quantity - 1)
                            }
                          >
                            <Minus className="h-4 w-4" />
                          </Button>
                          <span className="w-12 text-center font-medium">
                            {item.quantity}
                          </span>
                          <Button
                            variant="outline"
                            size="icon"
                            onClick={() =>
                              updateQuantity(item.id, item.quantity + 1)
                            }
                          >
                            <Plus className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}

              <Button
                variant="outline"
                onClick={clearCart}
                className="w-full"
              >
                清空购物车
              </Button>
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
                    {appliedCoupon ? (
                      <div className="flex items-center justify-between p-2 bg-primary/10 border border-primary/30 rounded-lg">
                        <span className="text-sm font-medium text-primary">
                          {appliedCoupon.code} 已应用
                        </span>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-auto py-0 px-1 text-muted-foreground hover:text-destructive"
                          onClick={handleRemoveCoupon}
                        >
                          移除
                        </Button>
                      </div>
                    ) : (
                      <>
                        <div className="flex gap-2">
                          <Input
                            placeholder="输入优惠券代码"
                            value={couponCode}
                            onChange={(e) => {
                              setCouponCode(e.target.value);
                              setCouponError('');
                            }}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') handleValidateCoupon();
                            }}
                          />
                          <Button
                            variant="outline"
                            onClick={handleValidateCoupon}
                            disabled={isValidating || !couponCode.trim()}
                          >
                            {isValidating ? (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                              '验证'
                            )}
                          </Button>
                        </div>
                        {couponError && (
                          <p className="text-sm text-destructive">{couponError}</p>
                        )}
                      </>
                    )}
                  </div>

                  {/* Price Breakdown */}
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
                </CardContent>
                <CardFooter className="flex flex-col gap-2">
                  <Button asChild className="w-full" size="lg">
                    <Link
                      href={`/checkout${appliedCoupon ? `?coupon=${encodeURIComponent(appliedCoupon.code)}` : ''}`}
                    >
                      去结算
                    </Link>
                  </Button>
                  <Button asChild variant="outline" className="w-full">
                    <Link href="/games">继续购物</Link>
                  </Button>
                </CardFooter>
              </Card>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
