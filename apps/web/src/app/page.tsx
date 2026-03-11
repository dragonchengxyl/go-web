'use client';

import { useEffect, useState, useRef, MouseEvent as ReactMouseEvent } from 'react';
import Link from 'next/link';
import { motion, useMotionValue, useSpring, AnimatePresence } from 'framer-motion';
import { apiClient, Post, Group, Event, LeaderboardEntry } from '@/lib/api-client';
import { PostGalleryCard } from '@/components/post/post-gallery-card';
import { Button } from '@/components/ui/button';
import {
  Palette, Users, MessageSquare, Calendar, Trophy, ArrowRight, Zap,
  Flame, Star, Globe2, Hash, ChevronRight,
} from 'lucide-react';

// ── Helpers ───────────────────────────────────────────────────────────────
const GRADIENTS = [
  'from-purple-500 to-teal-400',
  'from-teal-400 to-blue-500',
  'from-orange-400 to-pink-500',
  'from-blue-500 to-indigo-600',
  'from-green-400 to-teal-500',
  'from-pink-500 to-rose-400',
];
function hashGradient(str: string) {
  let h = 0;
  for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) | 0;
  return GRADIENTS[Math.abs(h) % GRADIENTS.length];
}

function formatDate(s: string) {
  const d = new Date(s);
  return `${d.getMonth() + 1}月${d.getDate()}日`;
}

// ── Animated blob ─────────────────────────────────────────────────────────
function Blob({ color, size, x, y, dur }: { color: string; size: number; x: [number, number]; y: [number, number]; dur: number }) {
  return (
    <motion.div
      className="absolute rounded-full blur-[120px] pointer-events-none"
      style={{ width: size, height: size, background: color, left: '50%', top: '50%' }}
      animate={{ x: [x[0], x[1], x[0]], y: [y[0], y[1], y[0]], scale: [1, 1.25, 1] }}
      transition={{ duration: dur, repeat: Infinity, ease: 'easeInOut' }}
    />
  );
}

// ── Spotlight card ────────────────────────────────────────────────────────
function SpotlightCard({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  const divRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ x: 0, y: 0, opacity: 0 });

  function handleMove(e: ReactMouseEvent<HTMLDivElement>) {
    if (!divRef.current) return;
    const rect = divRef.current.getBoundingClientRect();
    setPos({ x: e.clientX - rect.left, y: e.clientY - rect.top, opacity: 1 });
  }

  return (
    <div
      ref={divRef}
      onMouseMove={handleMove}
      onMouseLeave={() => setPos(p => ({ ...p, opacity: 0 }))}
      className={`relative overflow-hidden rounded-2xl border border-white/10 bg-white/5 backdrop-blur-sm ${className}`}
      style={{ '--x': `${pos.x}px`, '--y': `${pos.y}px` } as React.CSSProperties}
    >
      <div
        className="pointer-events-none absolute inset-0 transition-opacity duration-500 rounded-2xl"
        style={{
          opacity: pos.opacity,
          background: `radial-gradient(400px circle at ${pos.x}px ${pos.y}px, rgba(139,92,246,0.15), transparent 60%)`,
        }}
      />
      {children}
    </div>
  );
}

// ── Marquee row ───────────────────────────────────────────────────────────
function MarqueeRow({ posts, reverse }: { posts: Post[]; reverse?: boolean }) {
  const doubled = [...posts, ...posts, ...posts, ...posts]; // ensure enough items
  return (
    <div className="relative overflow-hidden py-2">
      <div
        className={`flex gap-3 w-max ${reverse ? 'animate-marquee-reverse' : 'animate-marquee'}`}
        style={{ willChange: 'transform' }}
      >
        {doubled.map((p, i) => {
          const img = p.media_urls?.[0];
          const g = hashGradient(p.id + i);
          return (
            <Link key={`${p.id}-${i}`} href={`/posts/${p.id}`}>
              <div className="w-36 h-24 rounded-xl overflow-hidden flex-shrink-0 ring-1 ring-white/10 hover:ring-brand-purple/60 transition-all hover:scale-105">
                {img ? (
                  <img src={img} alt="" className="w-full h-full object-cover" />
                ) : (
                  <div className={`w-full h-full bg-gradient-to-br ${g} flex items-center justify-center`}>
                    <span className="text-white/60 text-xs font-medium px-2 text-center line-clamp-2">{p.title}</span>
                  </div>
                )}
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}

// ── Word stagger ──────────────────────────────────────────────────────────
const sentenceVariants = { hidden: {}, show: { transition: { staggerChildren: 0.08 } } };
const wordVariants = {
  hidden: { opacity: 0, y: 20, filter: 'blur(8px)' },
  show: { opacity: 1, y: 0, filter: 'blur(0px)', transition: { duration: 0.5, ease: 'easeOut' } },
};

function AnimatedTitle({ parts }: { parts: Array<{ text: string; gradient?: boolean }> }) {
  return (
    <motion.h1
      variants={sentenceVariants}
      initial="hidden"
      animate="show"
      className="text-5xl md:text-7xl lg:text-8xl font-black mb-8 leading-[1.05] tracking-tight"
    >
      {parts.map((part, pi) => (
        <span key={pi}>
          {part.text.split('').map((char, ci) => (
            <motion.span
              key={ci}
              variants={wordVariants}
              className={
                part.gradient
                  ? 'bg-gradient-to-r from-[#8B5CF6] via-[#06B6D4] to-[#F97316] bg-clip-text text-transparent inline-block'
                  : 'text-white inline-block'
              }
              style={{ display: 'inline-block' }}
            >
              {char === ' ' ? '\u00A0' : char}
            </motion.span>
          ))}
        </span>
      ))}
    </motion.h1>
  );
}

// ── FEATURES ──────────────────────────────────────────────────────────────
const FEATURES = [
  { icon: Palette, title: '原创创作', desc: '绘画、文字、音乐皆可分享，展示你独一无二的兽设', href: '/posts/create', accent: '#8B5CF6' },
  { icon: Users, title: '兽迷社群', desc: '加入圈子，关注同好，构建你的毛毛人脉', href: '/groups', accent: '#06B6D4' },
  { icon: Calendar, title: '同好活动', desc: '线上线下精彩活动，与更多毛毛面对面', href: '/events', accent: '#F97316' },
  { icon: MessageSquare, title: '实时私聊', desc: '与全球 Furry 即时沟通，零距离互动', href: '/messages', accent: '#EC4899' },
];

// ── MAIN PAGE ─────────────────────────────────────────────────────────────
export default function HomePage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [events, setEvents] = useState<Event[]>([]);
  const [topUsers, setTopUsers] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.allSettled([
      apiClient.getExplore(1, 18),
      apiClient.listGroups({ page_size: 6 }),
      apiClient.listEvents(1, 4),
      apiClient.getLeaderboard(5),
    ]).then(([postsRes, groupsRes, eventsRes, leaderRes]) => {
      if (postsRes.status === 'fulfilled') setPosts(postsRes.value.posts ?? []);
      if (groupsRes.status === 'fulfilled') setGroups(groupsRes.value.groups ?? []);
      if (eventsRes.status === 'fulfilled') setEvents(eventsRes.value.events ?? []);
      if (leaderRes.status === 'fulfilled') setLeaderboard(leaderRes.value ?? []);
      setLoading(false);
    });
  }, []);

  function setLeaderboard(data: LeaderboardEntry[]) {
    setTopUsers(data.slice(0, 5));
  }

  const marqueeRow1 = posts.slice(0, 9);
  const marqueeRow2 = posts.slice(9, 18);

  return (
    <div className="min-h-screen" style={{ background: '#0a0a0f', color: '#fff' }}>
      {/* ── HERO ─────────────────────────────────────────────────── */}
      <section className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden px-4 pt-16">
        {/* Blobs */}
        <Blob color="rgba(139,92,246,0.35)" size={700} x={[-200, 120]} y={[-180, 80]} dur={12} />
        <Blob color="rgba(6,182,212,0.25)" size={500} x={[100, -100]} y={[100, -60]} dur={9} />
        <Blob color="rgba(249,115,22,0.18)" size={400} x={[50, -150]} y={[-100, 120]} dur={14} />

        {/* Dot grid */}
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            backgroundImage: 'radial-gradient(rgba(255,255,255,0.06) 1px, transparent 1px)',
            backgroundSize: '32px 32px',
          }}
        />

        {/* Radial fade-out for grid */}
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            background: 'radial-gradient(ellipse 80% 60% at 50% 50%, transparent, #0a0a0f 80%)',
          }}
        />

        {/* Badge */}
        <motion.div
          initial={{ opacity: 0, y: -12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 0.5 }}
          className="relative z-10 flex items-center gap-2 px-4 py-1.5 rounded-full border border-white/10 bg-white/5 backdrop-blur-sm text-sm text-white/70 mb-8"
        >
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          Furry 创作者社区 · 现已开放
        </motion.div>

        {/* Title */}
        <div className="relative z-10 text-center max-w-5xl">
          <AnimatedTitle
            parts={[
              { text: '你的兽设，' },
              { text: '你的世界', gradient: true },
            ]}
          />

          <motion.p
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.8, duration: 0.6 }}
            className="text-lg md:text-xl text-white/50 max-w-xl mx-auto mb-10 leading-relaxed"
          >
            与来自全球的毛毛们共享创作与故事——这里是属于兽迷的温暖星球。
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 1, duration: 0.5 }}
            className="flex items-center justify-center gap-4 flex-wrap"
          >
            <Link href="/explore">
              <Button
                size="lg"
                className="px-8 h-12 text-base font-semibold rounded-xl bg-gradient-to-r from-[#8B5CF6] to-[#06B6D4] text-white border-0 hover:brightness-110 shadow-lg shadow-purple-500/25 transition-all"
              >
                <Flame className="h-4 w-4 mr-2" />
                探索创作
              </Button>
            </Link>
            <Link href="/posts/create">
              <Button
                size="lg"
                variant="outline"
                className="px-8 h-12 text-base font-semibold rounded-xl border-white/15 bg-white/5 text-white hover:bg-white/10 backdrop-blur-sm"
              >
                <Zap className="h-4 w-4 mr-2" />
                立即发帖
              </Button>
            </Link>
          </motion.div>
        </div>

        {/* Scroll hint */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1.5 }}
          className="absolute bottom-10 left-1/2 -translate-x-1/2 flex flex-col items-center gap-1 text-white/30"
        >
          <span className="text-xs">向下滚动</span>
          <motion.div
            animate={{ y: [0, 6, 0] }}
            transition={{ duration: 1.5, repeat: Infinity }}
            className="w-px h-8 bg-gradient-to-b from-white/30 to-transparent"
          />
        </motion.div>
      </section>

      {/* ── MARQUEE ──────────────────────────────────────────────── */}
      {!loading && posts.length > 3 && (
        <section className="py-12 overflow-hidden" style={{ background: '#0a0a0f' }}>
          <div className="mb-1">
            <MarqueeRow posts={marqueeRow1.length > 3 ? marqueeRow1 : posts} />
          </div>
          {marqueeRow2.length > 3 && (
            <MarqueeRow posts={marqueeRow2} reverse />
          )}
        </section>
      )}

      {/* ── FEATURES ─────────────────────────────────────────────── */}
      <section className="py-24 px-4 relative overflow-hidden" style={{ background: 'linear-gradient(to bottom, #0a0a0f, #0f0a1a)' }}>
        <div className="max-w-6xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="text-center mb-14"
          >
            <p className="text-xs uppercase tracking-[0.3em] text-white/30 mb-3">平台特色</p>
            <h2 className="text-3xl md:text-4xl font-bold text-white">为毛毛们精心打造</h2>
          </motion.div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {FEATURES.map(({ icon: Icon, title, desc, href, accent }, i) => (
              <motion.div
                key={title}
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.1, duration: 0.45 }}
              >
                <Link href={href}>
                  <SpotlightCard className="p-6 h-full group cursor-pointer hover:border-white/20 transition-all duration-300">
                    <div
                      className="w-12 h-12 rounded-2xl flex items-center justify-center mb-5 group-hover:scale-110 transition-transform duration-300"
                      style={{ background: `${accent}22`, border: `1px solid ${accent}44` }}
                    >
                      <Icon className="h-6 w-6" style={{ color: accent }} />
                    </div>
                    <h3 className="text-base font-semibold text-white mb-2">{title}</h3>
                    <p className="text-sm text-white/50 leading-relaxed">{desc}</p>
                    <div className="mt-4 flex items-center gap-1 text-xs font-medium" style={{ color: accent }}>
                      了解更多 <ChevronRight className="h-3 w-3" />
                    </div>
                  </SpotlightCard>
                </Link>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ── HOT POSTS ────────────────────────────────────────────── */}
      <section className="py-24 px-4" style={{ background: '#0f0a1a' }}>
        <div className="max-w-6xl mx-auto">
          <div className="flex items-center justify-between mb-10">
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4 }}
            >
              <p className="text-xs uppercase tracking-[0.3em] text-white/30 mb-1">社区精选</p>
              <h2 className="text-3xl font-bold text-white flex items-center gap-2">
                <Flame className="h-7 w-7 text-orange-400" />
                热门创作
              </h2>
            </motion.div>
            <Link href="/explore">
              <Button variant="ghost" className="text-white/40 hover:text-white gap-1">
                全部 <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>

          {loading ? (
            <div className="columns-2 md:columns-3 gap-4 space-y-4">
              {[1,2,3,4,5,6].map(i => (
                <div key={i} className="break-inside-avoid rounded-xl overflow-hidden mb-4">
                  <div className="bg-white/5 animate-pulse" style={{ height: `${160 + (i % 3) * 60}px` }} />
                </div>
              ))}
            </div>
          ) : posts.length > 0 ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={{ once: true, amount: 0.1 }}
              variants={{ hidden: {}, show: { transition: { staggerChildren: 0.06 } } }}
              className="grid grid-cols-2 md:grid-cols-3 gap-4"
            >
              {posts.slice(0, 12).map((post, i) => (
                <motion.div
                  key={post.id}
                  variants={{ hidden: { opacity: 0, y: 24 }, show: { opacity: 1, y: 0, transition: { duration: 0.4 } } }}
                >
                  <PostGalleryCard post={post} />
                </motion.div>
              ))}
            </motion.div>
          ) : (
            <div className="text-center py-20 text-white/30">
              <Globe2 className="h-12 w-12 mx-auto mb-4 opacity-30" />
              <p>暂无内容，成为第一个发帖的人！</p>
              <Link href="/posts/create" className="inline-block mt-4">
                <Button className="bg-gradient-to-r from-[#8B5CF6] to-[#06B6D4] text-white border-0">发布第一条动态</Button>
              </Link>
            </div>
          )}
        </div>
      </section>

      {/* ── GROUPS + LEADERBOARD ─────────────────────────────────── */}
      <section className="py-24 px-4" style={{ background: 'linear-gradient(to bottom, #0f0a1a, #0a0a0f)' }}>
        <div className="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-10">
          {/* Groups */}
          <div className="lg:col-span-2">
            <div className="flex items-center justify-between mb-6">
              <div>
                <p className="text-xs uppercase tracking-[0.3em] text-white/30 mb-1">加入社群</p>
                <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                  <Users className="h-6 w-6 text-teal-400" />
                  热门圈子
                </h2>
              </div>
              <Link href="/groups">
                <Button variant="ghost" className="text-white/40 hover:text-white gap-1 text-sm">
                  全部 <ArrowRight className="h-3 w-3" />
                </Button>
              </Link>
            </div>

            {groups.length === 0 && !loading ? (
              <div className="rounded-2xl border border-white/10 bg-white/5 p-8 text-center text-white/30">
                <Users className="h-8 w-8 mx-auto mb-3 opacity-40" />
                <p className="text-sm">暂无圈子</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {(loading ? Array(4).fill(null) : groups.slice(0, 6)).map((g, i) => (
                  <motion.div
                    key={g?.id ?? i}
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    transition={{ delay: i * 0.08 }}
                  >
                    {g ? (
                      <Link href={`/groups/${g.id}`}>
                        <div className="flex items-center gap-3 p-4 rounded-xl border border-white/8 bg-white/4 hover:bg-white/8 hover:border-white/15 transition-all group">
                          <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${hashGradient(g.id)} flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition-transform`}>
                            <span className="text-white font-bold text-sm">{g.name[0]}</span>
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="font-semibold text-sm text-white truncate">{g.name}</p>
                            <p className="text-xs text-white/40">{g.member_count} 成员 · {g.post_count} 帖子</p>
                          </div>
                          <ChevronRight className="h-4 w-4 text-white/20 group-hover:text-white/50 transition-colors" />
                        </div>
                      </Link>
                    ) : (
                      <div className="h-16 rounded-xl bg-white/5 animate-pulse" />
                    )}
                  </motion.div>
                ))}
              </div>
            )}
          </div>

          {/* Leaderboard */}
          <div>
            <div className="flex items-center justify-between mb-6">
              <div>
                <p className="text-xs uppercase tracking-[0.3em] text-white/30 mb-1">社区之星</p>
                <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                  <Trophy className="h-6 w-6 text-yellow-400" />
                  排行榜
                </h2>
              </div>
              <Link href="/leaderboard">
                <Button variant="ghost" className="text-white/40 hover:text-white gap-1 text-sm">
                  全部 <ArrowRight className="h-3 w-3" />
                </Button>
              </Link>
            </div>

            <div className="space-y-2">
              {(loading ? Array(5).fill(null) : topUsers).map((u, i) => {
                const medals = ['🥇', '🥈', '🥉'];
                return (
                  <motion.div
                    key={u?.user_id ?? i}
                    initial={{ opacity: 0, x: 20 }}
                    whileInView={{ opacity: 1, x: 0 }}
                    viewport={{ once: true }}
                    transition={{ delay: i * 0.07 }}
                  >
                    {u ? (
                      <Link href={`/users/${u.user_id}`}>
                        <div className="flex items-center gap-3 p-3 rounded-xl border border-white/8 bg-white/4 hover:bg-white/8 transition-all group">
                          <span className="text-lg w-7 text-center">{medals[i] ?? `${i + 1}`}</span>
                          <div className={`w-8 h-8 rounded-full bg-gradient-to-br ${hashGradient(u.user_id)} flex items-center justify-center flex-shrink-0`}>
                            <span className="text-white text-xs font-bold">{u.username[0]?.toUpperCase()}</span>
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-semibold text-white truncate">{u.username}</p>
                            <p className="text-xs text-white/40">{u.score.toLocaleString()} 积分</p>
                          </div>
                          {i === 0 && <Star className="h-4 w-4 text-yellow-400" />}
                        </div>
                      </Link>
                    ) : (
                      <div className="h-14 rounded-xl bg-white/5 animate-pulse" />
                    )}
                  </motion.div>
                );
              })}
            </div>
          </div>
        </div>
      </section>

      {/* ── EVENTS ───────────────────────────────────────────────── */}
      {(events.length > 0 || loading) && (
        <section className="py-24 px-4" style={{ background: '#0a0a0f' }}>
          <div className="max-w-6xl mx-auto">
            <div className="flex items-center justify-between mb-8">
              <div>
                <p className="text-xs uppercase tracking-[0.3em] text-white/30 mb-1">线上线下</p>
                <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                  <Calendar className="h-6 w-6 text-orange-400" />
                  近期活动
                </h2>
              </div>
              <Link href="/events">
                <Button variant="ghost" className="text-white/40 hover:text-white gap-1 text-sm">
                  全部 <ArrowRight className="h-3 w-3" />
                </Button>
              </Link>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {(loading ? Array(4).fill(null) : events.slice(0, 4)).map((e, i) => (
                <motion.div
                  key={e?.id ?? i}
                  initial={{ opacity: 0, y: 24 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.1 }}
                >
                  {e ? (
                    <Link href={`/events/${e.id}`}>
                      <div className="rounded-2xl border border-white/10 bg-white/5 overflow-hidden hover:border-orange-500/40 hover:bg-white/8 transition-all group h-full">
                        <div className={`h-24 bg-gradient-to-br ${hashGradient(e.id)} flex items-end p-3`}>
                          <span className="text-white/80 text-xs font-medium px-2 py-0.5 rounded-full bg-black/30 backdrop-blur-sm">
                            {e.is_online ? '🌐 线上' : '📍 线下'}
                          </span>
                        </div>
                        <div className="p-4">
                          <p className="font-semibold text-sm text-white line-clamp-2 mb-2 group-hover:text-orange-300 transition-colors">{e.title}</p>
                          <p className="text-xs text-white/40">{formatDate(e.start_time)}</p>
                          <p className="text-xs text-white/40 mt-1">{e.attendee_count} 人参加</p>
                        </div>
                      </div>
                    </Link>
                  ) : (
                    <div className="h-48 rounded-2xl bg-white/5 animate-pulse" />
                  )}
                </motion.div>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* ── BOTTOM CTA ───────────────────────────────────────────── */}
      <section className="relative py-32 px-4 overflow-hidden text-center" style={{ background: '#0a0a0f' }}>
        <Blob color="rgba(139,92,246,0.25)" size={600} x={[0, 60]} y={[0, 40]} dur={10} />
        <Blob color="rgba(6,182,212,0.18)" size={400} x={[50, -50]} y={[20, -30]} dur={8} />

        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            backgroundImage: 'radial-gradient(rgba(255,255,255,0.04) 1px, transparent 1px)',
            backgroundSize: '32px 32px',
          }}
        />
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            background: 'radial-gradient(ellipse 70% 60% at 50% 50%, transparent, #0a0a0f 85%)',
          }}
        />

        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="relative z-10"
        >
          <h2 className="text-4xl md:text-5xl font-black text-white mb-4">
            准备好了吗？
          </h2>
          <p className="text-lg text-white/50 mb-10 max-w-md mx-auto">
            加入 Furry 同好社区，与全球毛毛一起创作、分享、连接。
          </p>
          <div className="flex items-center justify-center gap-4 flex-wrap">
            <Link href="/register">
              <Button
                size="lg"
                className="px-10 h-13 text-base font-semibold rounded-xl bg-gradient-to-r from-[#8B5CF6] to-[#06B6D4] text-white border-0 hover:brightness-110 shadow-xl shadow-purple-500/30 transition-all"
              >
                免费注册
              </Button>
            </Link>
            <Link href="/explore">
              <Button
                size="lg"
                variant="outline"
                className="px-10 h-13 text-base font-semibold rounded-xl border-white/15 bg-white/5 text-white hover:bg-white/10"
              >
                先逛逛
              </Button>
            </Link>
          </div>
        </motion.div>
      </section>
    </div>
  );
}
