'use client';

import { motion } from 'framer-motion';
import { GameCard } from '@/components/game/game-card';

const FEATURED_GAMES = [
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
];

export function FeaturedGames() {
  return (
    <section className="py-20 bg-muted/30">
      <div className="container mx-auto px-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-4xl font-bold mb-4">精选游戏</h2>
          <p className="text-muted-foreground text-lg">
            探索我们最新和最受欢迎的游戏作品
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {FEATURED_GAMES.map((game, index) => (
            <motion.div
              key={game.id}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
            >
              <GameCard game={game} />
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
