'use client';

import { useEffect, useState } from 'react';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { PenSquare } from 'lucide-react';

export default function FeedPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  const loadFeed = async (p: number) => {
    try {
      const res = await apiClient.getFeed(p, 20);
      if (p === 1) {
        setPosts(res.posts || []);
      } else {
        setPosts((prev) => [...prev, ...(res.posts || [])]);
      }
      setHasMore((res.posts || []).length === 20);
    } catch {
      // fallback to explore if not logged in
      try {
        const res = await apiClient.getExplore(p, 20);
        if (p === 1) {
          setPosts(res.posts || []);
        } else {
          setPosts((prev) => [...prev, ...(res.posts || [])]);
        }
        setHasMore((res.posts || []).length === 20);
      } catch {
        setPosts([]);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFeed(1);
  }, []);

  const loadMore = () => {
    const next = page + 1;
    setPage(next);
    loadFeed(next);
  };

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
          <p className="text-lg mb-4">还没有动态</p>
          <p className="text-sm mb-6">关注更多创作者来获取精彩内容</p>
          <Link href="/explore">
            <Button>发现创作者</Button>
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {posts.map((post) => (
            <PostCard
              key={post.id}
              post={post}
              onLike={async () => {
                if (post.is_liked_by_me) {
                  await apiClient.unlikePost(post.id);
                } else {
                  await apiClient.likePost(post.id);
                }
                setPosts((prev) =>
                  prev.map((p) =>
                    p.id === post.id
                      ? {
                          ...p,
                          is_liked_by_me: !p.is_liked_by_me,
                          like_count: p.is_liked_by_me ? p.like_count - 1 : p.like_count + 1,
                        }
                      : p
                  )
                );
              }}
            />
          ))}
          {hasMore && (
            <div className="text-center pt-4">
              <Button variant="outline" onClick={loadMore}>
                加载更多
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
