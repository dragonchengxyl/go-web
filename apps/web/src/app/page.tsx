'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiClient, Post } from '@/lib/api-client';
import { PostGalleryCard } from '@/components/post/post-gallery-card';
import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';
import { Palette, Users, MessageSquare, Calendar, Sparkles, TrendingUp } from 'lucide-react';

const FEATURES = [
  {
    icon: Palette,
    title: '创作展示',
    desc: '绘画、文字、音乐，皆可分享',
    gradient: 'from-brand-purple to-brand-teal',
    href: '/posts/create',
  },
  {
    icon: Users,
    title: '兽设社交',
    desc: '关注同好，构建你的毛圈',
    gradient: 'from-brand-teal to-brand-coral',
    href: '/explore',
  },
  {
    icon: Calendar,
    title: '同好活动',
    desc: '线上线下，精彩活动不错过',
    gradient: 'from-brand-coral to-brand-purple',
    href: '/events',
  },
  {
    icon: MessageSquare,
    title: '实时私聊',
    desc: '与全球毛毛即时互动交流',
    gradient: 'from-indigo-500 to-brand-purple',
    href: '/messages',
  },
];

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.1 } },
};
const itemVariants = {
  hidden: { opacity: 0, y: 24 },
  show: { opacity: 1, y: 0, transition: { duration: 0.4, ease: 'easeOut' } },
};

function GallerySkeleton() {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
      {[1, 2, 3, 4, 5, 6].map(i => (
        <div key={i} className="rounded-xl overflow-hidden">
          <div className="aspect-[4/3] bg-muted animate-pulse" />
          <div className="p-3 space-y-2">
            <div className="h-3 bg-muted animate-pulse rounded w-3/4" />
            <div className="h-3 bg-muted animate-pulse rounded w-1/2" />
          </div>
        </div>
      ))}
    </div>
  );
}

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
      <section className="relative overflow-hidden py-28 md:py-36 px-4 text-center">
        <div className="absolute inset-0 bg-gradient-to-br from-brand-purple/20 via-brand-teal/10 to-background bg-[length:200%_200%] animate-gradient-shift -z-10" />
        <div className="absolute top-16 left-1/4 w-72 h-72 bg-brand-purple/20 rounded-full blur-3xl animate-float -z-10" />
        <div className="absolute bottom-8 right-1/4 w-56 h-56 bg-brand-teal/20 rounded-full blur-3xl animate-float -z-10" style={{ animationDelay: '1.5s' }} />
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-brand-coral/10 rounded-full blur-3xl -z-10" />

        <motion.div
          initial={{ opacity: 0, y: 36 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, ease: 'easeOut' }}
        >
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-medium mb-6">
            <Sparkles className="h-3 w-3" />
            Furry 创作者社区
          </div>
          <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold mb-6 leading-tight tracking-tight">
            展示你的{' '}
            <span className="bg-gradient-to-r from-brand-purple via-brand-teal to-brand-coral bg-clip-text text-transparent">
              兽设与创作
            </span>
          </h1>
          <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-10 leading-relaxed">
            与来自全球的毛毛们共享创作、活动与生活。这里是你的温暖家园。
          </p>
          <div className="flex items-center justify-center gap-4 flex-wrap">
            <Link href="/register">
              <Button
                size="lg"
                className="px-8 bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 animate-glow-pulse transition-all shadow-lg shadow-brand-purple/20"
              >
                立即免费加入
              </Button>
            </Link>
            <Link href="/explore">
              <Button
                size="lg"
                variant="outline"
                className="px-8 backdrop-blur-sm bg-background/60 hover:bg-background/80"
              >
                浏览创作
              </Button>
            </Link>
          </div>
        </motion.div>
      </section>

      {/* Features */}
      <section className="container mx-auto px-4 py-16">
        <motion.div
          variants={containerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.15 }}
          className="grid grid-cols-2 md:grid-cols-4 gap-4 max-w-4xl mx-auto mb-20"
        >
          {FEATURES.map(({ icon: Icon, title, desc, gradient, href }) => (
            <motion.div key={title} variants={itemVariants}>
              <Link href={href} className="group block h-full">
                <div className="text-center p-5 rounded-2xl bg-card border hover:shadow-md hover:border-primary/30 transition-all h-full">
                  <div
                    className={`w-12 h-12 rounded-2xl bg-gradient-to-br ${gradient} flex items-center justify-center mx-auto mb-3 shadow-sm group-hover:scale-110 transition-transform duration-200`}
                  >
                    <Icon className="h-6 w-6 text-white" />
                  </div>
                  <div className="text-sm font-semibold mb-1">{title}</div>
                  <p className="text-xs text-muted-foreground leading-relaxed">{desc}</p>
                </div>
              </Link>
            </motion.div>
          ))}
        </motion.div>

        {/* Gallery section */}
        <div className="max-w-5xl mx-auto">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-primary" />
              <h2 className="text-2xl font-bold">热门创作</h2>
            </div>
            <Link href="/explore">
              <Button variant="ghost" size="sm" className="text-muted-foreground">
                查看全部 →
              </Button>
            </Link>
          </div>

          {loading ? (
            <GallerySkeleton />
          ) : posts.length > 0 ? (
            <motion.div
              variants={containerVariants}
              initial="hidden"
              whileInView="show"
              viewport={{ once: true, amount: 0.1 }}
              className="grid grid-cols-2 md:grid-cols-3 gap-4"
            >
              {posts.map(post => (
                <motion.div key={post.id} variants={itemVariants}>
                  <PostGalleryCard post={post} />
                </motion.div>
              ))}
            </motion.div>
          ) : (
            <div className="text-center py-16 text-muted-foreground">
              <p>暂无内容，成为第一个发帖的人吧！</p>
              <Link href="/posts/create" className="mt-4 inline-block">
                <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
                  发布第一条动态
                </Button>
              </Link>
            </div>
          )}
        </div>
      </section>

      {/* Bottom CTA */}
      <section className="bg-gradient-to-br from-brand-purple/10 via-brand-teal/5 to-background border-t py-20 px-4 mt-8">
        <div className="container mx-auto text-center max-w-2xl">
          <h3 className="text-3xl font-bold mb-4">准备好加入了吗？</h3>
          <p className="text-muted-foreground mb-8 text-lg">注册完全免费，立刻开始你的 Furry 创作之旅</p>
          <div className="flex items-center justify-center gap-4 flex-wrap">
            <Link href="/register">
              <Button
                size="lg"
                className="px-10 bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 transition-all"
              >
                免费注册
              </Button>
            </Link>
            <Link href="/login">
              <Button size="lg" variant="outline" className="px-10">
                已有账号？登录
              </Button>
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
