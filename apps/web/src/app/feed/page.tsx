'use client';

import { useEffect, useState, useRef } from 'react';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { PenSquare, Compass } from 'lucide-react';

export default function FeedPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const pageRef = useRef(1);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) { apiClient.setToken(token); }
    loadPage(1);
  }, []);

  async function loadPage(p: number) {
    if (p === 1) setLoading(true); else setLoadingMore(true);
    try {
      const res = await apiClient.getFeed(p, 20);
      const items = res.posts || [];
      if (p === 1) setPosts(items); else setPosts(prev => [...prev, ...items]);
      const more = items.length === 20;
      setHasMore(more);
      pageRef.current = p;
      return more;
    } catch {
      if (p === 1) setPosts([]);
      setHasMore(false);
      return false;
    } finally {
      if (p === 1) setLoading(false); else setLoadingMore(false);
    }
  }

  // Intersection Observer for infinite scroll
  useEffect(() => {
    if (!sentinelRef.current) return;
    const obs = new IntersectionObserver(entries => {
      if (entries[0].isIntersecting && !loadingMore) {
        loadPage(pageRef.current + 1);
      }
    }, { rootMargin: '200px' });
    obs.observe(sentinelRef.current);
    return () => obs.disconnect();
  }, [loadingMore]);

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4">
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-40 bg-muted animate-pulse rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      {/* Create post CTA */}
      <div className="mb-6 flex items-center gap-3 p-4 bg-card border rounded-lg">
        <div className="w-10 h-10 rounded-full bg-muted flex-shrink-0" />
        <Link href="/posts/create" className="flex-1">
          <div className="w-full px-4 py-2 text-sm text-muted-foreground bg-muted rounded-full cursor-pointer hover:bg-muted/80 transition-colors">
            分享你的furry日常...
          </div>
        </Link>
        <Link href="/posts/create">
          <Button size="sm">
            <PenSquare className="h-4 w-4 mr-1" />
            发布
          </Button>
        </Link>
      </div>

      {posts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <Compass className="h-16 w-16 mx-auto mb-4 opacity-30" />
          <p className="text-lg font-medium mb-2">还没有关注流内容</p>
          <p className="text-sm mb-6">关注感兴趣的创作者，他们的动态会出现在这里</p>
          <div className="flex gap-3 justify-center">
            <Link href="/explore">
              <Button>
                <Compass className="h-4 w-4 mr-1" />
                发现创作者
              </Button>
            </Link>
            <Link href="/posts/create">
              <Button variant="outline">发布第一条动态</Button>
            </Link>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {posts.map(post => (
            <PostCard key={post.id} post={post} />
          ))}
          {/* Infinite scroll sentinel */}
          {hasMore && (
            <div ref={sentinelRef} className="py-4 flex justify-center">
              {loadingMore && (
                <div className="flex gap-1.5">
                  {[0, 1, 2].map(i => (
                    <div key={i} className="w-2 h-2 rounded-full bg-muted-foreground/40 animate-bounce" style={{ animationDelay: `${i * 0.15}s` }} />
                  ))}
                </div>
              )}
            </div>
          )}
          {!hasMore && posts.length > 0 && (
            <p className="text-center text-sm text-muted-foreground py-4">已经看完啦 ✨</p>
          )}
        </div>
      )}
    </div>
  );
}
