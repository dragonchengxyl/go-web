'use client';

import { useEffect, useState, useCallback } from 'react';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';
import { Button } from '@/components/ui/button';

export default function ExplorePage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [activeTag, setActiveTag] = useState('');
  const [hotTags, setHotTags] = useState<string[]>([]);
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) { apiClient.setToken(token); setIsLoggedIn(true); }
    apiClient.getHotTags().then(setHotTags).catch(() => {});
  }, []);

  const load = useCallback(async (p: number, tag: string, reset: boolean) => {
    if (reset) setLoading(true);
    try {
      const res = await apiClient.getExplore(p, 20, tag || undefined);
      const items = res.posts || [];
      if (reset) {
        setPosts(items);
      } else {
        setPosts(prev => [...prev, ...items]);
      }
      setHasMore(items.length === 20);
    } catch {
      if (reset) setPosts([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    setPage(1);
    load(1, activeTag, true);
  }, [activeTag, load]);

  function handleTagClick(tag: string) {
    setActiveTag(prev => prev === tag ? '' : tag);
  }

  function handleLike(post: Post) {
    if (!isLoggedIn) return;
    const fn = post.is_liked_by_me ? apiClient.unlikePost(post.id) : apiClient.likePost(post.id);
    fn.then(() => {
      setPosts(prev => prev.map(p =>
        p.id === post.id
          ? { ...p, is_liked_by_me: !p.is_liked_by_me, like_count: p.is_liked_by_me ? p.like_count - 1 : p.like_count + 1 }
          : p
      ));
    }).catch(() => {});
  }

  function loadMore() {
    const next = page + 1;
    setPage(next);
    load(next, activeTag, false);
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <h1 className="text-2xl font-bold mb-4">发现</h1>

      {/* Hot tags */}
      {hotTags.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-6">
          <button
            onClick={() => setActiveTag('')}
            className={`px-3 py-1 rounded-full text-sm border transition-colors ${activeTag === '' ? 'bg-primary text-primary-foreground border-primary' : 'hover:border-primary text-muted-foreground'}`}
          >
            全部
          </button>
          {hotTags.map(tag => (
            <button
              key={tag}
              onClick={() => handleTagClick(tag)}
              className={`px-3 py-1 rounded-full text-sm border transition-colors ${activeTag === tag ? 'bg-primary text-primary-foreground border-primary' : 'hover:border-primary text-muted-foreground'}`}
            >
              #{tag}
            </button>
          ))}
        </div>
      )}

      {loading ? (
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-40 bg-muted animate-pulse rounded-lg" />
          ))}
        </div>
      ) : posts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <p>{activeTag ? `#${activeTag} 暂无内容` : '暂无内容'}</p>
        </div>
      ) : (
        <div className="space-y-4">
          {posts.map(post => (
            <PostCard
              key={post.id}
              post={post}
              onLike={() => handleLike(post)}
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
