'use client';

import { useEffect, useState } from 'react';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';
import { Button } from '@/components/ui/button';

export default function ExplorePage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  const load = async (p: number) => {
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
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(1);
  }, []);

  const loadMore = () => {
    const next = page + 1;
    setPage(next);
    load(next);
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
      <h1 className="text-2xl font-bold mb-6">发现</h1>
      {posts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <p>暂无内容</p>
        </div>
      ) : (
        <div className="space-y-4">
          {posts.map((post) => (
            <PostCard
              key={post.id}
              post={post}
              onLike={async () => {
                try {
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
                } catch {
                  // ignore auth errors
                }
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
