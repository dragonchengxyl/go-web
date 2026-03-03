import Link from 'next/link';
import Image from 'next/image';
import { formatPrice } from '@/lib/utils';
import { ShoppingCart } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface Game {
  id: string;
  title: string;
  slug: string;
  description: string;
  coverImage: string;
  price: number;
  tags: string[];
}

interface GameCardProps {
  game: Game;
}

export function GameCard({ game }: GameCardProps) {
  return (
    <div className="group relative bg-card rounded-lg overflow-hidden border hover:shadow-lg transition-all duration-300">
      {/* Cover Image */}
      <Link href={`/games/${game.slug}`}>
        <div className="relative aspect-video overflow-hidden bg-muted">
          <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent z-10" />
          <div className="w-full h-full bg-gradient-to-br from-primary/20 to-secondary/20" />
          <div className="absolute bottom-4 left-4 z-20">
            {game.tags.map((tag) => (
              <span
                key={tag}
                className="inline-block bg-primary/80 text-primary-foreground text-xs px-2 py-1 rounded mr-2"
              >
                {tag}
              </span>
            ))}
          </div>
        </div>
      </Link>

      {/* Content */}
      <div className="p-6">
        <Link href={`/games/${game.slug}`}>
          <h3 className="text-xl font-bold mb-2 group-hover:text-primary transition-colors">
            {game.title}
          </h3>
        </Link>
        <p className="text-muted-foreground text-sm mb-4 line-clamp-2">
          {game.description}
        </p>

        {/* Footer */}
        <div className="flex items-center justify-between">
          <span className="text-2xl font-bold">
            {formatPrice(game.price)}
          </span>
          <Button size="sm">
            <ShoppingCart className="mr-2 h-4 w-4" />
            购买
          </Button>
        </div>
      </div>
    </div>
  );
}
