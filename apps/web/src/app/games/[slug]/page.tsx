import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import Image from 'next/image';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ShoppingCart, Download, Heart, Share2 } from 'lucide-react';
import { formatPrice } from '@/lib/utils';

// This would come from API in real implementation
async function getGame(slug: string) {
  // Mock data for now
  return {
    id: '1',
    title: '神秘森林',
    slug: 'mystery-forest',
    description: '探索充满谜题的神秘森林，揭开隐藏的秘密。在这个充满魔法和冒险的世界中，你将扮演一位勇敢的探险家，解开古老的谜题，发现失落的文明。',
    longDescription: `## 游戏简介

在《神秘森林》中，你将踏上一段充满惊奇的冒险之旅。这片古老的森林隐藏着无数秘密，等待着勇敢的探险家去发现。

### 核心玩法

- **探索**: 在广阔的开放世界中自由探索
- **解谜**: 解开各种精心设计的谜题
- **收集**: 寻找隐藏的宝藏和神秘物品
- **战斗**: 与森林中的生物战斗

### 游戏特色

- 精美的像素艺术风格
- 动态天气系统
- 丰富的剧情和角色
- 多个结局等你发现`,
    coverImage: '/images/games/game1.jpg',
    screenshots: [
      '/images/games/screenshot1.jpg',
      '/images/games/screenshot2.jpg',
      '/images/games/screenshot3.jpg',
    ],
    price: 4900,
    originalPrice: 5900,
    tags: ['冒险', '解谜', '像素', '单人'],
    releaseDate: '2025-12-01',
    developer: '独立游戏工作室',
    isPublished: true,
    rating: 4.8,
    reviewCount: 1234,
  };
}

export async function generateMetadata({
  params,
}: {
  params: { slug: string };
}): Promise<Metadata> {
  const game = await getGame(params.slug);

  return {
    title: `${game.title} - 独立游戏工作室`,
    description: game.description,
    openGraph: {
      title: game.title,
      description: game.description,
      images: [game.coverImage],
    },
  };
}

export default async function GameDetailPage({
  params,
}: {
  params: { slug: string };
}) {
  const game = await getGame(params.slug);

  if (!game) {
    notFound();
  }

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
                {game.tags.map((tag) => (
                  <Badge key={tag} variant="secondary">
                    {tag}
                  </Badge>
                ))}
              </div>
              <h1 className="text-5xl font-bold mb-4">{game.title}</h1>
              <p className="text-xl text-muted-foreground mb-6">
                {game.description}
              </p>
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-3xl font-bold">
                    {formatPrice(game.price)}
                  </span>
                  {game.originalPrice && (
                    <span className="text-lg text-muted-foreground line-through">
                      {formatPrice(game.originalPrice)}
                    </span>
                  )}
                </div>
                <Button size="lg" className="gap-2">
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
                  <TabsTrigger value="screenshots">截图</TabsTrigger>
                  <TabsTrigger value="reviews">评价</TabsTrigger>
                </TabsList>
                <TabsContent value="about" className="mt-6">
                  <div className="prose prose-lg dark:prose-invert max-w-none">
                    <div
                      dangerouslySetInnerHTML={{
                        __html: game.longDescription.replace(/\n/g, '<br />'),
                      }}
                    />
                  </div>
                </TabsContent>
                <TabsContent value="screenshots" className="mt-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {game.screenshots.map((screenshot, index) => (
                      <div
                        key={index}
                        className="relative aspect-video bg-muted rounded-lg overflow-hidden"
                      >
                        <div className="w-full h-full bg-gradient-to-br from-primary/20 to-secondary/20" />
                      </div>
                    ))}
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
                    <dt className="text-sm text-muted-foreground">开发商</dt>
                    <dd className="font-medium">{game.developer}</dd>
                  </div>
                  <div>
                    <dt className="text-sm text-muted-foreground">发行日期</dt>
                    <dd className="font-medium">{game.releaseDate}</dd>
                  </div>
                  <div>
                    <dt className="text-sm text-muted-foreground">评分</dt>
                    <dd className="font-medium">
                      ⭐ {game.rating} ({game.reviewCount} 评价)
                    </dd>
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
