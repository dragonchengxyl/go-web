'use client';

import { useEffect, useState, useRef } from 'react';
import { apiClient, AudioWork, Post } from '@/lib/api-client';
import { AudioWorkCard } from '@/components/audio/audio-work-card';
import { PostCard } from '@/components/post/post-card';
import { PostCardSkeleton, Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { AudioLines, PenSquare, Compass, UserPlus } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

interface RecommendedUser {
  id: string;
  username: string;
  furry_name?: string;
  species?: string;
}

function RecommendedUsers() {
  const [users, setUsers] = useState<RecommendedUser[]>([]);
  const [following, setFollowing] = useState<Set<string>>(new Set());

  useEffect(() => {
    apiClient.getExplore(1, 12).then(res => {
      const posts: Post[] = res.posts || [];
      const seen = new Set<string>();
      const list: RecommendedUser[] = [];
      for (const p of posts) {
        if (!seen.has(p.author_id) && list.length < 6) {
          seen.add(p.author_id);
          list.push({
            id: p.author_id,
            username: p.author_username || p.author_id,
            furry_name: undefined,
            species: undefined,
          });
        }
      }
      setUsers(list);
    }).catch(() => {});
  }, []);

  async function handleFollow(userId: string) {
    try {
      if (following.has(userId)) {
        await apiClient.unfollowUser(userId);
        setFollowing(prev => { const s = new Set(prev); s.delete(userId); return s; });
      } else {
        await apiClient.followUser(userId);
        setFollowing(prev => new Set(prev).add(userId));
      }
    } catch {}
  }

  if (users.length === 0) return null;

  return (
    <div className="mt-8">
      <h2 className="text-base font-semibold mb-3 flex items-center gap-2">
        <UserPlus className="h-4 w-4 text-brand-purple" />
        推荐关注
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {users.map(u => (
          <div key={u.id} className="flex items-center justify-between p-3 bg-card border rounded-xl hover:shadow-sm transition-shadow">
            <Link href={`/users/${u.id}`} className="flex items-center gap-3 min-w-0">
              <div className="w-9 h-9 rounded-full bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center text-white font-bold text-sm flex-shrink-0">
                {(u.username)[0]?.toUpperCase()}
              </div>
              <div className="min-w-0">
                <p className="text-sm font-medium truncate">@{u.username}</p>
              </div>
            </Link>
            <Button
              size="sm"
              variant={following.has(u.id) ? 'outline' : 'default'}
              className={following.has(u.id) ? '' : 'bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110'}
              onClick={() => handleFollow(u.id)}
            >
              {following.has(u.id) ? '已关注' : '关注'}
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function FeedPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [audioWorks, setAudioWorks] = useState<AudioWork[]>([]);
  const [loading, setLoading] = useState(true);
  const [audioLoading, setAudioLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [firstLoad, setFirstLoad] = useState(true);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const pageRef = useRef(1);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) { apiClient.setToken(token); }
    loadPage(1);
    apiClient
      .listAudioWorks({ page: 1, page_size: 4 })
      .then((res) => setAudioWorks(res.items ?? []))
      .catch(() => setAudioWorks([]))
      .finally(() => setAudioLoading(false));
  }, []);

  async function loadPage(p: number) {
    if (p === 1) setLoading(true); else setLoadingMore(true);
    try {
      const res = await apiClient.getFeed(p, 20);
      const items = res.posts || [];
      if (p === 1) {
        setPosts(items);
        setFirstLoad(true);
      } else {
        setPosts(prev => [...prev, ...items]);
        setFirstLoad(false);
      }
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
          {[...Array(5)].map((_, i) => <PostCardSkeleton key={i} />)}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      {/* Create post CTA */}
      <div className="mb-6 flex items-center gap-3 p-4 bg-card border rounded-xl hover:shadow-sm transition-shadow">
        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-brand-purple to-brand-teal flex-shrink-0 opacity-60" />
        <Link href="/posts/create" className="flex-1">
          <div className="w-full px-4 py-2 text-sm text-muted-foreground bg-muted rounded-full cursor-pointer hover:bg-muted/70 hover:text-foreground transition-colors">
            分享你的furry日常...
          </div>
        </Link>
        <Link href="/posts/create">
          <Button
            size="sm"
            className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110"
          >
            <PenSquare className="h-4 w-4 mr-1" />
            发布
          </Button>
        </Link>
      </div>

      <section className="mb-6 rounded-2xl border bg-card p-4">
        <div className="mb-4 flex items-center justify-between">
          <div>
            <div className="mb-1 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
              <AudioLines className="h-4 w-4 text-emerald-500" />
              音频作品
            </div>
            <h2 className="text-base font-semibold">最新公开音频</h2>
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
            还没有公开音频作品，去音频创作页发布第一首吧。
          </div>
        )}
      </section>

      {posts.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <Compass className="h-16 w-16 mx-auto mb-4 opacity-30 animate-float" />
          <p className="text-lg font-medium mb-2">还没有关注流内容</p>
          <p className="text-sm mb-6">关注感兴趣的创作者，他们的动态会出现在这里</p>
          <div className="flex gap-3 justify-center">
            <Link href="/explore">
              <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
                <Compass className="h-4 w-4 mr-1" />
                发现创作者
              </Button>
            </Link>
            <Link href="/posts/create">
              <Button variant="outline">发布第一条动态</Button>
            </Link>
            <Link href="/audio/works">
              <Button variant="outline">浏览音频作品</Button>
            </Link>
          </div>
          <RecommendedUsers />
        </div>
      ) : (
        <motion.div
          className="space-y-4"
          initial={firstLoad ? 'hidden' : false}
          animate="show"
          variants={{ hidden: {}, show: { transition: { staggerChildren: 0.08 } } }}
        >
          <AnimatePresence>
            {posts.map(post => (
              <motion.div
                key={post.id}
                variants={{
                  hidden: { opacity: 0, y: 16 },
                  show: { opacity: 1, y: 0, transition: { duration: 0.3, ease: 'easeOut' } },
                }}
              >
                <PostCard post={post} />
              </motion.div>
            ))}
          </AnimatePresence>

          {/* Infinite scroll sentinel */}
          {hasMore && (
            <div ref={sentinelRef} className="py-4 flex justify-center">
              {loadingMore && (
                <div className="flex gap-1.5">
                  {[0, 1, 2].map(i => (
                    <motion.div
                      key={i}
                      className="w-2 h-2 rounded-full bg-brand-purple/60"
                      animate={{ y: [0, -6, 0] }}
                      transition={{ duration: 0.6, repeat: Infinity, delay: i * 0.15 }}
                    />
                  ))}
                </div>
              )}
            </div>
          )}
          {!hasMore && posts.length > 0 && (
            <p className="text-center text-sm text-muted-foreground py-4">已经看完啦 ✨</p>
          )}
        </motion.div>
      )}
    </div>
  );
}
