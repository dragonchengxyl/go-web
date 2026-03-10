'use client';

import { useEffect, useState, useCallback } from 'react';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';
import { Button } from '@/components/ui/button';

type AIFilter = 'all' | 'human' | 'ai';

export default function ExplorePage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [activeTag, setActiveTag] = useState('');
  const [hotTags, setHotTags] = useState<string[]>([]);
  const [aiFilter, setAIFilter] = useState<AIFilter>('all');

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) { apiClient.setToken(token); }
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

  function loadMore() {
    const next = page + 1;
    setPage(next);
    load(next, activeTag, false);
  }

  const filteredPosts = posts.filter(post => {
    if (aiFilter === 'ai') return post.content_labels?.is_ai_generated === true;
    if (aiFilter === 'human') return !post.content_labels?.is_ai_generated;
    return true;
  });

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <h1 className="text-2xl font-bold mb-4">发现</h1>

      {/* Hot tags */}
      {hotTags.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-4">
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

      {/* AI filter toggle */}
      <div className="flex gap-1 mb-6">
        {(['all', 'human', 'ai'] as const).map(f => (
          <button
            key={f}
            onClick={() => setAIFilter(f)}
            className={`px-3 py-1 rounded-full text-xs border transition-colors ${
              aiFilter === f
                ? 'bg-purple-600 text-white border-purple-600'
                : 'text-muted-foreground hover:border-purple-400'
            }`}
          >
            {f === 'all' ? '全部' : f === 'human' ? '人工创作' : 'AI 生成'}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-40 bg-muted animate-pulse rounded-lg" />
          ))}
        </div>
      ) : filteredPosts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <p>{activeTag ? `#${activeTag} 暂无内容` : '暂无内容'}</p>
        </div>
      ) : (
        <div className="space-y-4">
          {filteredPosts.map(post => (
            <PostCard key={post.id} post={post} />
          ))}
          {hasMore && aiFilter === 'all' && (
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
