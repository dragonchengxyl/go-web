'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';
import { PostCardSkeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';
import { Palette, Users, MessageSquare } from 'lucide-react';

const FEATURES = [
  {
    icon: Palette,
    title: '创作展示',
    desc: '绘画、文字、音乐，皆可分享',
    gradient: 'from-brand-purple to-brand-teal',
  },
  {
    icon: Users,
    title: '兽设分享',
    desc: '展示你的独特角色设定',
    gradient: 'from-brand-teal to-brand-coral',
  },
  {
    icon: MessageSquare,
    title: '同好交流',
    desc: '与全球毛毛实时私信互动',
    gradient: 'from-brand-coral to-brand-purple',
  },
];

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.12 } },
};
const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0, transition: { duration: 0.4, ease: 'easeOut' } },
};

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
      <section className="relative overflow-hidden py-28 px-4 text-center">
        {/* Animated gradient background */}
        <div className="absolute inset-0 bg-gradient-to-br from-brand-purple/20 via-brand-teal/10 to-background bg-[length:200%_200%] animate-gradient-shift -z-10" />
        {/* Decorative blobs */}
        <div className="absolute top-16 left-1/4 w-64 h-64 bg-brand-purple/20 rounded-full blur-3xl animate-float -z-10" />
        <div className="absolute bottom-8 right-1/4 w-48 h-48 bg-brand-teal/20 rounded-full blur-3xl animate-float -z-10" style={{ animationDelay: '1.5s' }} />

        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
        >
          <h1 className="text-4xl md:text-6xl font-bold mb-6 leading-tight">
            欢迎来到{' '}
            <span className="bg-gradient-to-r from-brand-purple to-brand-teal bg-clip-text text-transparent">
              Furry 同好社区
            </span>
          </h1>
          <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-10">
            这里是毛毛们的温暖家园。分享你的兽设、创作与日常，结识来自世界各地的同好伙伴。
          </p>
          <div className="flex items-center justify-center gap-4 flex-wrap">
            <Link href="/register">
              <Button
                size="lg"
                className="px-8 bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 animate-glow-pulse transition-all"
              >
                立即加入
              </Button>
            </Link>
            <Link href="/explore">
              <Button
                size="lg"
                variant="outline"
                className="px-8 backdrop-blur-sm bg-background/60 hover:bg-background/80"
              >
                浏览内容
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
          viewport={{ once: true, amount: 0.2 }}
          className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto mb-16"
        >
          {FEATURES.map(({ icon: Icon, title, desc, gradient }) => (
            <motion.div
              key={title}
              variants={itemVariants}
              className="text-center p-6 rounded-xl bg-card border hover:shadow-md transition-shadow"
            >
              <div className={`w-14 h-14 rounded-2xl bg-gradient-to-br ${gradient} flex items-center justify-center mx-auto mb-4 shadow-sm`}>
                <Icon className="h-7 w-7 text-white" />
              </div>
              <div className="text-lg font-semibold mb-1">{title}</div>
              <p className="text-sm text-muted-foreground">{desc}</p>
            </motion.div>
          ))}
        </motion.div>

        {/* Hot posts */}
        <h2 className="text-2xl font-bold mb-8 text-center">热门动态</h2>
        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-4xl mx-auto">
            {[1, 2, 3, 4].map(i => <PostCardSkeleton key={i} />)}
          </div>
        ) : posts.length > 0 ? (
          <motion.div
            variants={containerVariants}
            initial="hidden"
            whileInView="show"
            viewport={{ once: true, amount: 0.1 }}
            className="max-w-2xl mx-auto space-y-4"
          >
            {posts.map(post => (
              <motion.div key={post.id} variants={itemVariants}>
                <PostCard post={post} />
              </motion.div>
            ))}
          </motion.div>
        ) : (
          <p className="text-center text-muted-foreground">暂无内容，成为第一个发帖的人吧！</p>
        )}
        <div className="text-center mt-8">
          <Link href="/explore">
            <Button variant="outline">查看更多内容</Button>
          </Link>
        </div>
      </section>

      {/* Bottom CTA */}
      <section className="bg-gradient-to-br from-brand-purple/10 via-brand-teal/5 to-background border-t py-16 px-4">
        <div className="container mx-auto text-center">
          <h3 className="text-2xl font-bold mb-4">准备好加入了吗？</h3>
          <p className="text-muted-foreground mb-6">注册完全免费，立刻开始你的 Furry 之旅</p>
          <Link href="/register">
            <Button
              size="lg"
              className="px-10 bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 transition-all"
            >
              免费注册
            </Button>
          </Link>
        </div>
      </section>
    </div>
  );
}
