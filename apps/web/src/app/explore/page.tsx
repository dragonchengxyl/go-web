'use client';

import { useEffect, useState, useCallback } from 'react';
import { apiClient, AudioWork, Post } from '@/lib/api-client';
import { AudioWorkCard } from '@/components/audio/audio-work-card';
import { PostCard } from '@/components/post/post-card';
import { PostCardSkeleton, Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import Link from 'next/link';
import { AudioLines } from 'lucide-react';

type AIFilter = 'all' | 'human' | 'ai';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.08 } },
};
const itemVariants = {
  hidden: { opacity: 0, y: 16 },
  show: { opacity: 1, y: 0, transition: { duration: 0.3, ease: 'easeOut' } },
};

export default function ExplorePage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [audioWorks, setAudioWorks] = useState<AudioWork[]>([]);
  const [loading, setLoading] = useState(true);
  const [audioLoading, setAudioLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [activeTag, setActiveTag] = useState('');
  const [hotTags, setHotTags] = useState<string[]>([]);
  const [aiFilter, setAIFilter] = useState<AIFilter>('all');

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) { apiClient.setToken(token); }
    apiClient.getHotTags().then(setHotTags).catch(() => {});
    apiClient
      .listAudioWorks({ page: 1, page_size: 4 })
      .then((res) => setAudioWorks(res.items ?? []))
      .catch(() => setAudioWorks([]))
      .finally(() => setAudioLoading(false));
  }, []);

  const load = useCallback(async (p: number, tag: string, reset: boolean) => {
    if (reset) setLoading(true);
    try {
      const res = await apiClient.getExplore(p, 20, tag || undefined);
      const items = res.posts || [];
      if (reset) setPosts(items); else setPosts(prev => [...prev, ...items]);
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
      <h1 className="text-2xl font-bold mb-6 bg-gradient-to-r from-brand-purple to-brand-teal bg-clip-text text-transparent">
        发现创作
      </h1>

      <section className="mb-6 rounded-2xl border bg-card p-4">
        <div className="mb-4 flex items-center justify-between">
          <div>
            <div className="mb-1 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
              <AudioLines className="h-4 w-4 text-emerald-500" />
              音频分发
            </div>
            <h2 className="text-base font-semibold">最新音频作品</h2>
          </div>
          <Link href="/audio/works" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
            查看全部
          </Link>
        </div>

        {audioLoading ? (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {[0, 1].map((item) => (
              <div key={item} className="space-y-3 rounded-2xl border border-border/60 bg-background p-3">
                <Skeleton className="aspect-[16/10] w-full rounded-xl" />
                <Skeleton className="h-5 w-2/3" />
                <Skeleton className="h-4 w-1/2" />
              </div>
            ))}
          </div>
        ) : audioWorks.length > 0 ? (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {audioWorks.map((work) => (
              <AudioWorkCard key={work.id} work={work} compact />
            ))}
          </div>
        ) : (
          <div className="rounded-2xl border border-dashed border-border bg-muted/10 px-4 py-6 text-sm text-muted-foreground">
            暂无公开音频作品。
          </div>
        )}
      </section>

      {/* Hot tags */}
      {hotTags.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-4">
          <button
            onClick={() => setActiveTag('')}
            className={cn(
              'px-3 py-1 rounded-full text-sm border transition-all duration-200',
              activeTag === ''
                ? 'bg-gradient-to-r from-brand-purple to-brand-teal text-white border-transparent shadow-sm'
                : 'hover:border-brand-purple/50 text-muted-foreground'
            )}
          >
            全部
          </button>
          {hotTags.map(tag => (
            <button
              key={tag}
              onClick={() => handleTagClick(tag)}
              className={cn(
                'px-3 py-1 rounded-full text-sm border transition-all duration-200',
                activeTag === tag
                  ? 'bg-gradient-to-r from-brand-purple to-brand-teal text-white border-transparent shadow-sm'
                  : 'hover:border-brand-purple/50 text-muted-foreground'
              )}
            >
              #{tag}
            </button>
          ))}
        </div>
      )}

      {/* AI filter */}
      <div className="flex gap-1 mb-6">
        {(['all', 'human', 'ai'] as const).map(f => (
          <button
            key={f}
            onClick={() => setAIFilter(f)}
            className={cn(
              'px-3 py-1 rounded-full text-xs border transition-all duration-200',
              aiFilter === f
                ? 'bg-gradient-to-r from-brand-purple to-brand-teal text-white border-transparent'
                : 'text-muted-foreground hover:border-brand-purple/40'
            )}
          >
            {f === 'all' ? '全部' : f === 'human' ? '人工创作' : 'AI 生成'}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => <PostCardSkeleton key={i} />)}
        </div>
      ) : filteredPosts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <p>{activeTag ? `#${activeTag} 暂无内容` : '暂无内容'}</p>
        </div>
      ) : (
        <motion.div
          className="space-y-4"
          variants={containerVariants}
          initial="hidden"
          animate="show"
        >
          {filteredPosts.map(post => (
            <motion.div key={post.id} variants={itemVariants}>
              <PostCard post={post} />
            </motion.div>
          ))}
          {hasMore && aiFilter === 'all' && (
            <div className="text-center pt-4">
              <Button variant="outline" onClick={loadMore}>加载更多</Button>
            </div>
          )}
        </motion.div>
      )}
    </div>
  );
}
