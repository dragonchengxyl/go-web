'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Tag } from 'lucide-react';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';
import { Button } from '@/components/ui/button';

export default function TagPage() {
  const params = useParams();
  const tag = decodeURIComponent(params.tag as string);

  const [posts, setPosts] = useState<Post[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadPosts(1);
  }, [tag]);

  async function loadPosts(p: number) {
    setLoading(true);
    try {
      const data = await apiClient.getExplore(p, 20, tag);
      if (p === 1) {
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
    }
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <div className="flex items-center gap-3 mb-6">
        <Link href="/explore">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="h-4 w-4 mr-1" />返回
          </Button>
        </Link>
        <div className="flex items-center gap-2">
          <Tag className="h-5 w-5 text-primary" />
          <h1 className="text-xl font-bold">{tag}</h1>
          {total > 0 && (
            <span className="text-sm text-muted-foreground">({total} 篇帖子)</span>
          )}
        </div>
      </div>

      {loading && posts.length === 0 ? (
        <div className="space-y-3">
          {[1, 2, 3, 4].map(i => <div key={i} className="h-32 bg-muted animate-pulse rounded-xl" />)}
        </div>
      ) : posts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <Tag className="h-12 w-12 mx-auto mb-4 opacity-30" />
          <p>该标签下暂无内容</p>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {posts.map(post => <PostCard key={post.id} post={post} />)}
          </div>
          {posts.length < total && (
            <div className="text-center mt-6">
              <Button variant="outline" onClick={() => loadPosts(page + 1)} disabled={loading}>
                {loading ? '加载中...' : '加载更多'}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
