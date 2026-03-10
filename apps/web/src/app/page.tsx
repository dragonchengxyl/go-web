'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';
import { Button } from '@/components/ui/button';

export default function HomePage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiClient.getExplore(1, 6)
      .then(data => setPosts(data.posts ?? []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen">
      {/* Hero */}
      <section className="bg-gradient-to-br from-primary/10 via-background to-background py-24 px-4 text-center">
        <h1 className="text-4xl md:text-6xl font-bold mb-6">
          欢迎来到 <span className="text-primary">Furry 同好社区</span>
        </h1>
        <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-10">
          这里是毛毛们的温暖家园。分享你的兽设、创作与日常，结识来自世界各地的同好伙伴。
        </p>
        <div className="flex items-center justify-center gap-4 flex-wrap">
          <Link href="/register">
            <Button size="lg" className="px-8">立即加入</Button>
          </Link>
          <Link href="/explore">
            <Button size="lg" variant="outline" className="px-8">浏览内容</Button>
          </Link>
        </div>
      </section>

      {/* Hot posts preview */}
      <section className="container mx-auto px-4 py-16">
        <h2 className="text-2xl font-bold mb-8 text-center">热门动态</h2>
        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-4xl mx-auto">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="h-40 bg-muted animate-pulse rounded-xl" />
            ))}
          </div>
        ) : posts.length > 0 ? (
          <div className="max-w-2xl mx-auto space-y-4">
            {posts.map(post => <PostCard key={post.id} post={post} />)}
          </div>
        ) : (
          <p className="text-center text-muted-foreground">暂无内容，成为第一个发帖的人吧！</p>
        )}
        <div className="text-center mt-8">
          <Link href="/explore">
            <Button variant="outline">查看更多内容</Button>
          </Link>
        </div>
      </section>

      {/* Stats / CTA */}
      <section className="bg-muted/40 border-t py-16 px-4">
        <div className="container mx-auto grid grid-cols-1 md:grid-cols-3 gap-8 text-center mb-12">
          <div>
            <div className="text-4xl font-bold text-primary mb-2">🐾</div>
            <div className="text-lg font-semibold">兽设分享</div>
            <p className="text-sm text-muted-foreground mt-1">展示你的独特角色设定</p>
          </div>
          <div>
            <div className="text-4xl font-bold text-primary mb-2">🎨</div>
            <div className="text-lg font-semibold">创作展示</div>
            <p className="text-sm text-muted-foreground mt-1">绘画、文字、音乐，皆可分享</p>
          </div>
          <div>
            <div className="text-4xl font-bold text-primary mb-2">💬</div>
            <div className="text-lg font-semibold">同好交流</div>
            <p className="text-sm text-muted-foreground mt-1">与全球毛毛实时私信互动</p>
          </div>
        </div>

        <div className="text-center">
          <h3 className="text-2xl font-bold mb-4">准备好加入了吗？</h3>
          <p className="text-muted-foreground mb-6">注册完全免费，立刻开始你的 Furry 之旅</p>
          <Link href="/register">
            <Button size="lg" className="px-10">免费注册</Button>
          </Link>
        </div>
      </section>
    </div>
  );
}
