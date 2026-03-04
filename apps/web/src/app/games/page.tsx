'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { GameCard } from '@/components/game/game-card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search } from 'lucide-react';
import { apiClient } from '@/lib/api-client';

export default function GamesPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [page, setPage] = useState(1);

  // Fetch games
  const { data: gamesData, isLoading: gamesLoading } = useQuery({
    queryKey: ['games', page, searchQuery, selectedTag],
    queryFn: () => apiClient.getGames({
      page,
      page_size: 12,
      search: searchQuery || undefined,
      tag: selectedTag || undefined,
    }),
  });

  // Fetch products to get prices
  const { data: productsData } = useQuery({
    queryKey: ['products', 'game'],
    queryFn: () => apiClient.getProducts({ product_type: 'game', is_active: true }),
  });

  // Merge games with product prices
  const games = gamesData?.games?.map(game => {
    const product = productsData?.products?.find(p => p.entity_id === game.id);
    return {
      ...game,
      price: product?.price_cents || 0,
      productId: product?.id,
    };
  }) || [];

  const tags = ['冒险', '解谜', 'RPG', '策略', '科幻', '动作'];

  return (
    <div className="min-h-screen">
      <Header />
      <main className="pt-16">
        {/* Hero Section */}
        <section className="bg-gradient-to-br from-primary/10 to-secondary/10 py-20">
          <div className="container mx-auto px-4">
            <div className="max-w-3xl mx-auto text-center">
              <h1 className="text-5xl font-bold mb-6">探索我们的游戏</h1>
              <p className="text-xl text-muted-foreground mb-8">
                发现独特而富有创意的游戏体验
              </p>

              {/* Search Bar */}
              <div className="flex gap-2 max-w-xl mx-auto">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    type="search"
                    placeholder="搜索游戏..."
                    className="pl-10"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                  />
                </div>
                <Button onClick={() => setPage(1)}>搜索</Button>
              </div>
            </div>
          </div>
        </section>

        {/* Filters */}
        <section className="border-b">
          <div className="container mx-auto px-4 py-6">
            <div className="flex flex-wrap gap-2">
              <Button
                variant={selectedTag === null ? 'outline' : 'ghost'}
                size="sm"
                onClick={() => setSelectedTag(null)}
              >
                全部
              </Button>
              {tags.map(tag => (
                <Button
                  key={tag}
                  variant={selectedTag === tag ? 'outline' : 'ghost'}
                  size="sm"
                  onClick={() => setSelectedTag(tag)}
                >
                  {tag}
                </Button>
              ))}
            </div>
          </div>
        </section>

        {/* Games Grid */}
        <section className="py-12">
          <div className="container mx-auto px-4">
            {gamesLoading ? (
              <div className="text-center py-12">加载中...</div>
            ) : games.length === 0 ? (
              <div className="text-center py-12 text-muted-foreground">
                暂无游戏
              </div>
            ) : (
              <>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                  {games.map((game) => (
                    <GameCard key={game.id} game={game} />
                  ))}
                </div>

                {/* Pagination */}
                {gamesData && gamesData.total > 12 && (
                  <div className="flex justify-center gap-2 mt-12">
                    <Button
                      variant="outline"
                      disabled={page === 1}
                      onClick={() => setPage(p => p - 1)}
                    >
                      上一页
                    </Button>
                    <Button variant="outline">{page}</Button>
                    <Button
                      variant="outline"
                      disabled={page * 12 >= gamesData.total}
                      onClick={() => setPage(p => p + 1)}
                    >
                      下一页
                    </Button>
                  </div>
                )}
              </>
            )}
          </div>
        </section>
      </main>
      <Footer />
    </div>
  );
}
