'use client';

import { useParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ShoppingCart, Heart, Share2 } from 'lucide-react';
import { formatPrice } from '@/lib/utils';
import { apiClient } from '@/lib/api-client';
import { useCartStore } from '@/lib/store/cart';

export default function GameDetailPage() {
  const params = useParams();
  const slug = params.slug as string;
  const addItem = useCartStore((state) => state.addItem);

  // Fetch game data
  const { data: game, isLoading } = useQuery({
    queryKey: ['game', slug],
    queryFn: () => apiClient.getGameBySlug(slug),
  });

  // Fetch product data for price
  const { data: productsData } = useQuery({
    queryKey: ['products', 'game'],
    queryFn: () => apiClient.getProducts({ product_type: 'game', is_active: true }),
  });

  if (isLoading) {
    return (
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-12 text-center">
            加载中...
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  if (!game) {
    return (
      <div className="min-h-screen">
        <Header />
        <main className="pt-16">
          <div className="container mx-auto px-4 py-12 text-center">
            游戏不存在
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const product = productsData?.products?.find(p => p.entity_id === game.id);
  const price = product?.price_cents || 0;

  const handleAddToCart = () => {
    if (product) {
      addItem({
        id: game.id,
        productId: product.id,
        name: game.title,
        price,
        coverImage: game.cover_key || '/images/games/default.jpg',
      });
    }
  };

  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        {/* Hero Section */}
        <section className="relative h-[60vh] overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-primary/20 to-secondary/20" />
          <div className="absolute inset-0 bg-gradient-to-t from-background via-background/50 to-transparent" />
          <div className="relative container mx-auto px-4 h-full flex items-end pb-12">
            <div className="max-w-3xl">
              <div className="flex gap-2 mb-4">
                {game.tags?.map((tag: string) => (
                  <Badge key={tag} variant="secondary">
                    {tag}
                  </Badge>
                ))}
              </div>
              <h1 className="text-5xl font-bold mb-4">{game.title}</h1>
              <p className="text-xl text-muted-foreground mb-6">
                {game.subtitle || game.description}
              </p>
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-3xl font-bold">
                    {formatPrice(price)}
                  </span>
                </div>
                <Button size="lg" className="gap-2" onClick={handleAddToCart} disabled={!product}>
                  <ShoppingCart className="h-5 w-5" />
                  加入购物车
                </Button>
                <Button size="lg" variant="outline" className="gap-2">
                  <Heart className="h-5 w-5" />
                  收藏
                </Button>
              </div>
            </div>
          </div>
        </section>

        {/* Content Section */}
        <section className="container mx-auto px-4 py-12">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Main Content */}
            <div className="lg:col-span-2">
              <Tabs defaultValue="about" className="w-full">
                <TabsList className="w-full justify-start">
                  <TabsTrigger value="about">关于游戏</TabsTrigger>
                  <TabsTrigger value="reviews">评价</TabsTrigger>
                </TabsList>
                <TabsContent value="about" className="mt-6">
                  <div className="prose prose-lg dark:prose-invert max-w-none">
                    <p>{game.description}</p>
                  </div>
                </TabsContent>
                <TabsContent value="reviews" className="mt-6">
                  <div className="text-center text-muted-foreground py-12">
                    评价系统开发中...
                  </div>
                </TabsContent>
              </Tabs>
            </div>

            {/* Sidebar */}
            <div className="space-y-6">
              {/* Game Info */}
              <div className="bg-card border rounded-lg p-6">
                <h3 className="font-bold text-lg mb-4">游戏信息</h3>
                <dl className="space-y-3">
                  <div>
                    <dt className="text-sm text-muted-foreground">引擎</dt>
                    <dd className="font-medium">{game.engine || '未知'}</dd>
                  </div>
                  {game.release_date && (
                    <div>
                      <dt className="text-sm text-muted-foreground">发行日期</dt>
                      <dd className="font-medium">{new Date(game.release_date).toLocaleDateString('zh-CN')}</dd>
                    </div>
                  )}
                  <div>
                    <dt className="text-sm text-muted-foreground">类型</dt>
                    <dd className="font-medium">{game.genre?.join(', ') || '未分类'}</dd>
                  </div>
                </dl>
              </div>

              {/* Actions */}
              <div className="bg-card border rounded-lg p-6">
                <h3 className="font-bold text-lg mb-4">分享游戏</h3>
                <Button variant="outline" className="w-full gap-2">
                  <Share2 className="h-4 w-4" />
                  分享
                </Button>
              </div>
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </div>
  );
}
