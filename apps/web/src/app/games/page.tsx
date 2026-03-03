import { Metadata } from 'next';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import { GameCard } from '@/components/game/game-card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search } from 'lucide-react';

export const metadata: Metadata = {
  title: '游戏列表 - 独立游戏工作室',
  description: '探索我们的所有游戏作品',
};

// Mock data - would come from API
const games = [
  {
    id: '1',
    title: '神秘森林',
    slug: 'mystery-forest',
    description: '探索充满谜题的神秘森林，揭开隐藏的秘密',
    coverImage: '/images/games/game1.jpg',
    price: 4900,
    tags: ['冒险', '解谜'],
  },
  {
    id: '2',
    title: '像素王国',
    slug: 'pixel-kingdom',
    description: '建造你的像素王国，成为最强大的统治者',
    coverImage: '/images/games/game2.jpg',
    price: 3900,
    tags: ['策略', '建造'],
  },
  {
    id: '3',
    title: '星际旅行',
    slug: 'space-journey',
    description: '驾驶飞船穿越星系，探索未知的宇宙',
    coverImage: '/images/games/game3.jpg',
    price: 5900,
    tags: ['科幻', '探索'],
  },
  {
    id: '4',
    title: '地下城冒险',
    slug: 'dungeon-adventure',
    description: '深入危险的地下城，寻找传说中的宝藏',
    coverImage: '/images/games/game4.jpg',
    price: 4500,
    tags: ['RPG', '冒险'],
  },
  {
    id: '5',
    title: '时间旅行者',
    slug: 'time-traveler',
    description: '穿越时空，改变历史的进程',
    coverImage: '/images/games/game5.jpg',
    price: 5500,
    tags: ['冒险', '科幻'],
  },
  {
    id: '6',
    title: '魔法学院',
    slug: 'magic-academy',
    description: '在魔法学院学习魔法，成为最强大的魔法师',
    coverImage: '/images/games/game6.jpg',
    price: 4200,
    tags: ['RPG', '魔法'],
  },
];

export default function GamesPage() {
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
                  />
                </div>
                <Button>搜索</Button>
              </div>
            </div>
          </div>
        </section>

        {/* Filters */}
        <section className="border-b">
          <div className="container mx-auto px-4 py-6">
            <div className="flex flex-wrap gap-2">
              <Button variant="outline" size="sm">
                全部
              </Button>
              <Button variant="ghost" size="sm">
                冒险
              </Button>
              <Button variant="ghost" size="sm">
                解谜
              </Button>
              <Button variant="ghost" size="sm">
                RPG
              </Button>
              <Button variant="ghost" size="sm">
                策略
              </Button>
              <Button variant="ghost" size="sm">
                科幻
              </Button>
            </div>
          </div>
        </section>

        {/* Games Grid */}
        <section className="py-12">
          <div className="container mx-auto px-4">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
              {games.map((game) => (
                <GameCard key={game.id} game={game} />
              ))}
            </div>

            {/* Pagination */}
            <div className="flex justify-center gap-2 mt-12">
              <Button variant="outline" disabled>
                上一页
              </Button>
              <Button variant="outline">1</Button>
              <Button variant="ghost">2</Button>
              <Button variant="ghost">3</Button>
              <Button variant="outline">下一页</Button>
            </div>
          </div>
        </section>
      </main>
      <Footer />
    </div>
  );
}
