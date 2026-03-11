'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Hash, TrendingUp } from 'lucide-react';
import { apiClient, Post } from '@/lib/api-client';
import { PostGalleryCard } from '@/components/post/post-gallery-card';
import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.07 } },
};
const itemVariants = {
  hidden: { opacity: 0, y: 16 },
  show: { opacity: 1, y: 0, transition: { duration: 0.3 } },
};

export default function TagPage() {
  const params = useParams();
  const tag = decodeURIComponent(params.tag as string);

  const [posts, setPosts] = useState<Post[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);

  useEffect(() => {
    loadPosts(1, true);
  }, [tag]);

  async function loadPosts(p: number, replace = false) {
    if (replace) setLoading(true);
    else setLoadingMore(true);
    try {
      const data = await apiClient.getExplore(p, 18, tag);
      if (replace || p === 1) {
        setPosts(data.posts ?? []);
      } else {
        setPosts(prev => [...prev, ...(data.posts ?? [])]);
      }
      setTotal(data.total ?? 0);
      setPage(p);
    } catch {
      // ignore
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }

  return (
    <div className="max-w-5xl mx-auto pt-20 px-4 pb-12">
      {/* Back */}
      <Link href="/explore">
        <Button variant="ghost" size="sm" className="mb-4 -ml-2">
          <ArrowLeft className="h-4 w-4 mr-1" />返回发现
        </Button>
      </Link>

      {/* Tag header */}
      <div className="flex items-center gap-3 mb-8">
        <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center">
          <Hash className="h-6 w-6 text-white" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">{tag}</h1>
          {!loading && (
            <p className="text-sm text-muted-foreground flex items-center gap-1">
              <TrendingUp className="h-3.5 w-3.5" />
              {total} 篇帖子
            </p>
          )}
        </div>
      </div>

      {loading ? (
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {[1,2,3,4,5,6,7,8,9].map(i => (
            <div key={i} className="rounded-xl overflow-hidden">
              <div className="aspect-[4/3] bg-muted animate-pulse" />
              <div className="p-3 space-y-2">
                <div className="h-3 bg-muted animate-pulse rounded w-3/4" />
                <div className="h-3 bg-muted animate-pulse rounded w-1/2" />
              </div>
            </div>
          ))}
        </div>
      ) : posts.length === 0 ? (
        <div className="text-center py-20 text-muted-foreground">
          <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Hash className="h-8 w-8 opacity-40" />
          </div>
          <p className="font-medium mb-2">该标签下暂无内容</p>
          <p className="text-sm mb-6">成为第一个用 #{tag} 发帖的人</p>
          <Link href="/posts/create">
            <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
              发布帖子
            </Button>
          </Link>
        </div>
      ) : (
        <>
          <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="show"
            className="grid grid-cols-2 md:grid-cols-3 gap-4"
          >
            {posts.map(post => (
              <motion.div key={post.id} variants={itemVariants}>
                <PostGalleryCard post={post} />
              </motion.div>
            ))}
          </motion.div>

          {posts.length < total && (
            <div className="text-center mt-8">
              <Button
                variant="outline"
                onClick={() => loadPosts(page + 1)}
                disabled={loadingMore}
                className="px-8"
              >
                {loadingMore ? '加载中...' : '加载更多'}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
