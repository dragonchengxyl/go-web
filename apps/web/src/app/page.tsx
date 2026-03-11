'use client';

import { useEffect, useState, useRef } from 'react';
import Link from 'next/link';
import { motion } from 'framer-motion';
import { apiClient, Post, Group, Event, LeaderboardEntry } from '@/lib/api-client';
import { PostGalleryCard } from '@/components/post/post-gallery-card';
import { Button } from '@/components/ui/button';
import {
  Palette, Users, MessageSquare, Calendar, Trophy,
  ArrowRight, Zap, Flame, Star, Globe2, ChevronRight,
} from 'lucide-react';

// ── Constants ─────────────────────────────────────────────────────────────

const BG = '#050505';
const BORDER = '#1a1a1a';

const GRADIENTS = [
  'from-violet-600 to-indigo-600',
  'from-cyan-500 to-blue-600',
  'from-orange-500 to-rose-600',
  'from-emerald-500 to-teal-600',
  'from-pink-500 to-fuchsia-600',
  'from-amber-500 to-orange-600',
];

function hashGradient(s: string) {
  let h = 0;
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) | 0;
  return GRADIENTS[Math.abs(h) % GRADIENTS.length];
}

function formatDate(s: string) {
  const d = new Date(s);
  return `${d.getMonth() + 1}月${d.getDate()}日`;
}

// ── Retro perspective grid hero background ────────────────────────────────

function RetroGrid() {
  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none" aria-hidden>
      {/* Base dark */}
      <div className="absolute inset-0" style={{ background: BG }} />

      {/* Perspective grid floor */}
      <div
        className="absolute inset-x-0 bottom-0 h-[70%] retro-grid"
        style={{
          maskImage: 'linear-gradient(to top, rgba(0,0,0,0.7) 0%, transparent 100%)',
          WebkitMaskImage: 'linear-gradient(to top, rgba(0,0,0,0.7) 0%, transparent 100%)',
          transform: 'perspective(600px) rotateX(40deg)',
          transformOrigin: 'bottom center',
        }}
      />

      {/* Top fade */}
      <div
        className="absolute inset-0"
        style={{
          background: `radial-gradient(ellipse 90% 55% at 50% 0%, transparent 40%, ${BG} 80%)`,
        }}
      />

      {/* Single accent glow — not a blob, more like a distant light source */}
      <div
        className="absolute"
        style={{
          width: 640,
          height: 320,
          top: '10%',
          left: '50%',
          transform: 'translateX(-50%)',
          background: 'radial-gradient(ellipse at center, rgba(139,92,246,0.18) 0%, transparent 70%)',
          filter: 'blur(40px)',
        }}
      />
    </div>
  );
}

// ── Marquee with edge fade ────────────────────────────────────────────────

function MarqueeRow({ posts, reverse }: { posts: Post[]; reverse?: boolean }) {
  const items = [...posts, ...posts, ...posts, ...posts];
  return (
    <div
      className="overflow-hidden py-1.5"
      style={{
        maskImage: 'linear-gradient(to right, transparent, black 12%, black 88%, transparent)',
        WebkitMaskImage: 'linear-gradient(to right, transparent, black 12%, black 88%, transparent)',
      }}
    >
      <div
        className={`flex gap-3 w-max ${reverse ? 'animate-marquee-reverse' : 'animate-marquee'}`}
        style={{ willChange: 'transform' }}
      >
        {items.map((p, i) => {
          const img = p.media_urls?.[0];
          const g = hashGradient(p.id + i);
          return (
            <Link key={`${p.id}-${i}`} href={`/posts/${p.id}`}>
              <div
                className="w-40 h-26 rounded-lg overflow-hidden flex-shrink-0 transition-transform duration-300 hover:scale-[1.03]"
                style={{ height: 104, border: `1px solid ${BORDER}` }}
              >
                {img ? (
                  <img src={img} alt="" className="w-full h-full object-cover" />
                ) : (
                  <div className={`w-full h-full bg-gradient-to-br ${g} flex items-center justify-center`}>
                    <span className="text-white/50 text-[10px] font-medium px-2 text-center line-clamp-2 leading-relaxed">
                      {p.title}
                    </span>
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

// ── Bento feature card ────────────────────────────────────────────────────

interface FeatureDef {
  icon: React.ElementType;
  title: string;
  desc: string;
  href: string;
  accent: string;
  area: string;
  large?: boolean;
}

const FEATURES: FeatureDef[] = [
  {
    icon: Palette,
    title: '原创创作',
    desc: '绘画、文字、音乐、视频皆可分享。展示你独一无二的兽设，让世界看见你的创意。',
    href: '/posts/create',
    accent: '#8B5CF6',
    area: 'create',
    large: true,
  },
  {
    icon: Users,
    title: '兽迷社群',
    desc: '加入圈子，关注同好，构建你的毛圈人脉。',
    href: '/groups',
    accent: '#06B6D4',
    area: 'social',
  },
  {
    icon: Calendar,
    title: '同好活动',
    desc: '线上线下精彩活动，与更多毛毛面对面。',
    href: '/events',
    accent: '#F97316',
    area: 'events',
  },
  {
    icon: MessageSquare,
    title: '实时私聊',
    desc: '与全球 Furry 即时沟通，零距离互动。',
    href: '/messages',
    accent: '#EC4899',
    area: 'chat',
  },
];

function BentoCard({ feat }: { feat: FeatureDef }) {
  const divRef = useRef<HTMLDivElement>(null);
  const [spotlight, setSpotlight] = useState({ x: 0, y: 0, op: 0 });
  const Icon = feat.icon;

  return (
    <Link href={feat.href} className={`bento-${feat.area} block group`}>
      <div
        ref={divRef}
        className="relative h-full overflow-hidden rounded-2xl transition-all duration-500 hover:shadow-[0_0_40px_-12px_rgba(139,92,246,0.25)]"
        style={{ background: '#0d0d0d', border: `1px solid ${BORDER}` }}
        onMouseMove={e => {
          const r = divRef.current?.getBoundingClientRect();
          if (r) setSpotlight({ x: e.clientX - r.left, y: e.clientY - r.top, op: 1 });
        }}
        onMouseLeave={() => setSpotlight(p => ({ ...p, op: 0 }))}
      >
        {/* spotlight */}
        <div
          className="pointer-events-none absolute inset-0 rounded-2xl transition-opacity duration-500"
          style={{
            opacity: spotlight.op,
            background: `radial-gradient(320px circle at ${spotlight.x}px ${spotlight.y}px, ${feat.accent}18, transparent 60%)`,
          }}
        />

        {/* top accent line */}
        <div
          className="absolute top-0 left-0 right-0 h-px opacity-0 group-hover:opacity-100 transition-opacity duration-500"
          style={{ background: `linear-gradient(to right, transparent, ${feat.accent}60, transparent)` }}
        />

        <div className={`relative p-6 flex flex-col h-full ${feat.large ? 'min-h-[160px]' : 'min-h-[120px]'}`}>
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center mb-5 flex-shrink-0 transition-transform duration-300 group-hover:scale-105"
            style={{ background: `${feat.accent}14`, border: `1px solid ${feat.accent}30` }}
          >
            <Icon className="h-5 w-5" style={{ color: feat.accent }} />
          </div>

          <div className="flex-1">
            <h3 className="text-[15px] font-semibold text-white mb-1.5 heading-tight">{feat.title}</h3>
            <p className="text-[13px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)' }}>
              {feat.desc}
            </p>
          </div>

          <div
            className="mt-4 flex items-center gap-1 text-xs font-medium opacity-0 group-hover:opacity-100 transition-all duration-300 translate-x-0 group-hover:translate-x-1"
            style={{ color: feat.accent }}
          >
            进入 <ChevronRight className="h-3 w-3" />
          </div>
        </div>
      </div>
    </Link>
  );
}

// ── Stat counter ──────────────────────────────────────────────────────────

function StatPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center gap-2 text-sm" style={{ color: 'rgba(255,255,255,0.35)' }}>
      <span className="font-medium text-white/60">{value}</span>
      <span>{label}</span>
    </div>
  );
}

// ── Section header ────────────────────────────────────────────────────────

function SectionLabel({ eyebrow, title, icon: Icon, iconColor }: {
  eyebrow: string; title: string; icon: React.ElementType; iconColor: string;
}) {
  return (
    <div>
      <p className="text-[11px] uppercase tracking-[0.25em] mb-1" style={{ color: 'rgba(255,255,255,0.25)' }}>
        {eyebrow}
      </p>
      <h2 className="text-2xl font-bold text-white flex items-center gap-2 heading-tight">
        <Icon className="h-5 w-5" style={{ color: iconColor }} />
        {title}
      </h2>
    </div>
  );
}

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
    ]).then(([p, g, e, l]) => {
      if (p.status === 'fulfilled') setPosts(p.value.posts ?? []);
      if (g.status === 'fulfilled') setGroups(g.value.groups ?? []);
      if (e.status === 'fulfilled') setEvents(e.value.events ?? []);
      if (l.status === 'fulfilled') setTopUsers((l.value ?? []).slice(0, 5));
      setLoading(false);
    });
  }, []);

  const row1 = posts.slice(0, 9);
  const row2 = posts.slice(9, 18);

  return (
    <div className="noise-overlay" style={{ background: BG, color: '#fff', minHeight: '100vh' }}>

      {/* ── HERO ────────────────────────────────────────────────── */}
      <section className="relative min-h-screen flex flex-col items-center justify-center px-4 pt-16 overflow-hidden">
        <RetroGrid />

        {/* Content */}
        <div className="relative z-10 text-center max-w-4xl mx-auto">
          {/* Badge */}
          <motion.div
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: 'easeOut' }}
            className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full text-xs font-medium mb-10"
            style={{
              background: 'rgba(255,255,255,0.04)',
              border: `1px solid ${BORDER}`,
              color: 'rgba(255,255,255,0.5)',
            }}
          >
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
            Furry 创作者社区 · 现已开放
          </motion.div>

          {/* Title — tight tracked */}
          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.15, ease: [0.16, 1, 0.3, 1] }}
            className="font-black mb-6 heading-tight"
            style={{ fontSize: 'clamp(3rem, 8vw, 6rem)', lineHeight: 1.04 }}
          >
            你的兽设
            <br />
            <span
              style={{
                backgroundImage: 'linear-gradient(135deg, #a78bfa 0%, #38bdf8 50%, #fb923c 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
              }}
            >
              你的世界
            </span>
          </motion.h1>

          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.7, delay: 0.35 }}
            className="text-base md:text-lg max-w-md mx-auto mb-10 leading-relaxed font-warmth"
            style={{ color: 'rgba(255,255,255,0.42)' }}
          >
            与来自全球的毛毛们共享创作与故事——这里是属于兽迷的温暖星球。
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.5 }}
            className="flex items-center justify-center gap-3 flex-wrap"
          >
            <Link href="/explore">
              <button
                className="inline-flex items-center gap-2 px-7 py-3 rounded-xl text-sm font-semibold text-white transition-all duration-300 hover:brightness-110 hover:shadow-[0_0_30px_-6px_rgba(139,92,246,0.5)] active:scale-[0.98]"
                style={{ background: 'linear-gradient(135deg, #7c3aed, #0ea5e9)' }}
              >
                <Flame className="h-4 w-4" />
                探索创作
              </button>
            </Link>
            <Link href="/posts/create">
              <button
                className="inline-flex items-center gap-2 px-7 py-3 rounded-xl text-sm font-semibold transition-all duration-300 hover:bg-white/10 active:scale-[0.98]"
                style={{
                  background: 'rgba(255,255,255,0.05)',
                  border: `1px solid ${BORDER}`,
                  color: 'rgba(255,255,255,0.75)',
                }}
              >
                <Zap className="h-4 w-4" />
                立即发帖
              </button>
            </Link>
          </motion.div>

          {/* Stats line */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.8 }}
            className="flex items-center justify-center gap-6 mt-12 flex-wrap"
          >
            <StatPill label="创作帖子" value="∞" />
            <span style={{ color: BORDER }}>·</span>
            <StatPill label="兽迷社群" value="开放" />
            <span style={{ color: BORDER }}>·</span>
            <StatPill label="实时互动" value="24/7" />
          </motion.div>
        </div>

        {/* Scroll indicator */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1.2 }}
          className="absolute bottom-8 left-1/2 -translate-x-1/2 flex flex-col items-center gap-1.5"
          style={{ color: 'rgba(255,255,255,0.2)' }}
        >
          <motion.div
            animate={{ y: [0, 5, 0] }}
            transition={{ duration: 1.8, repeat: Infinity, ease: 'easeInOut' }}
            className="w-px h-10"
            style={{ background: 'linear-gradient(to bottom, rgba(255,255,255,0.3), transparent)' }}
          />
        </motion.div>
      </section>

      {/* ── MARQUEE ─────────────────────────────────────────────── */}
      {!loading && posts.length > 4 && (
        <section
          className="py-10 overflow-hidden"
          style={{ background: BG, borderTop: `1px solid ${BORDER}`, borderBottom: `1px solid ${BORDER}` }}
        >
          <MarqueeRow posts={row1.length > 3 ? row1 : posts} />
          {row2.length > 3 && <MarqueeRow posts={row2} reverse />}
        </section>
      )}

      {/* ── BENTO FEATURES ──────────────────────────────────────── */}
      <section className="py-24 px-4" style={{ background: BG }}>
        <div className="max-w-5xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="mb-10"
          >
            <p className="text-[11px] uppercase tracking-[0.25em] mb-2" style={{ color: 'rgba(255,255,255,0.25)' }}>
              平台特色
            </p>
            <h2 className="text-3xl font-bold text-white heading-tight">为毛毛们精心打造</h2>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="bento-grid"
          >
            {FEATURES.map(feat => <BentoCard key={feat.area} feat={feat} />)}
          </motion.div>
        </div>
      </section>

      {/* ── HOT POSTS ───────────────────────────────────────────── */}
      <section className="py-24 px-4" style={{ background: '#080808', borderTop: `1px solid ${BORDER}` }}>
        <div className="max-w-6xl mx-auto">
          <div className="flex items-center justify-between mb-10">
            <SectionLabel eyebrow="社区精选" title="热门创作" icon={Flame} iconColor="#f97316" />
            <Link href="/explore">
              <button
                className="flex items-center gap-1 text-sm transition-colors hover:text-white"
                style={{ color: 'rgba(255,255,255,0.3)' }}
              >
                全部 <ArrowRight className="h-3.5 w-3.5" />
              </button>
            </Link>
          </div>

          {loading ? (
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
              {[...Array(6)].map((_, i) => (
                <div
                  key={i}
                  className="rounded-xl animate-pulse"
                  style={{ height: 200 + (i % 3) * 40, background: '#111' }}
                />
              ))}
            </div>
          ) : posts.length > 0 ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={{ once: true, amount: 0.08 }}
              variants={{ hidden: {}, show: { transition: { staggerChildren: 0.05 } } }}
              className="grid grid-cols-2 md:grid-cols-3 gap-3"
            >
              {posts.slice(0, 12).map(post => (
                <motion.div
                  key={post.id}
                  variants={{ hidden: { opacity: 0, y: 20 }, show: { opacity: 1, y: 0, transition: { duration: 0.4 } } }}
                >
                  <PostGalleryCard post={post} />
                </motion.div>
              ))}
            </motion.div>
          ) : (
            <div
              className="text-center py-20 rounded-2xl"
              style={{ border: `1px solid ${BORDER}`, color: 'rgba(255,255,255,0.25)' }}
            >
              <Globe2 className="h-10 w-10 mx-auto mb-4 opacity-30" />
              <p className="text-sm mb-4">暂无内容，成为第一个发帖的人！</p>
              <Link href="/posts/create">
                <button
                  className="px-6 py-2 rounded-lg text-sm font-medium text-white"
                  style={{ background: 'linear-gradient(135deg, #7c3aed, #0ea5e9)' }}
                >
                  发布动态
                </button>
              </Link>
            </div>
          )}
        </div>
      </section>

      {/* ── GROUPS + LEADERBOARD ────────────────────────────────── */}
      <section className="py-24 px-4" style={{ background: BG, borderTop: `1px solid ${BORDER}` }}>
        <div className="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-5 gap-12">
          {/* Groups — 3 cols */}
          <div className="lg:col-span-3">
            <div className="flex items-center justify-between mb-8">
              <SectionLabel eyebrow="加入社群" title="热门圈子" icon={Users} iconColor="#06B6D4" />
              <Link href="/groups">
                <button className="flex items-center gap-1 text-sm transition-colors hover:text-white" style={{ color: 'rgba(255,255,255,0.3)' }}>
                  全部 <ArrowRight className="h-3.5 w-3.5" />
                </button>
              </Link>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {(loading ? Array(4).fill(null) : groups.slice(0, 6)).map((g, i) => (
                <motion.div
                  key={g?.id ?? i}
                  initial={{ opacity: 0, y: 12 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.06 }}
                >
                  {g ? (
                    <Link href={`/groups/${g.id}`}>
                      <div
                        className="flex items-center gap-3 p-3.5 rounded-xl transition-all duration-300 hover:shadow-[0_0_20px_-8px_rgba(6,182,212,0.2)] group"
                        style={{ background: '#0d0d0d', border: `1px solid ${BORDER}` }}
                      >
                        <div className={`w-9 h-9 rounded-lg bg-gradient-to-br ${hashGradient(g.id)} flex items-center justify-center flex-shrink-0 transition-transform duration-300 group-hover:scale-105`}>
                          <span className="text-white font-bold text-sm">{g.name[0]}</span>
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="font-semibold text-[13px] text-white truncate">{g.name}</p>
                          <p className="text-[11px] mt-0.5" style={{ color: 'rgba(255,255,255,0.3)' }}>
                            {g.member_count} 成员
                          </p>
                        </div>
                        <ChevronRight className="h-3.5 w-3.5 opacity-20 group-hover:opacity-50 transition-opacity" />
                      </div>
                    </Link>
                  ) : (
                    <div className="h-[60px] rounded-xl animate-pulse" style={{ background: '#0d0d0d' }} />
                  )}
                </motion.div>
              ))}
            </div>
          </div>

          {/* Leaderboard — 2 cols */}
          <div className="lg:col-span-2">
            <div className="flex items-center justify-between mb-8">
              <SectionLabel eyebrow="社区之星" title="排行榜" icon={Trophy} iconColor="#facc15" />
              <Link href="/leaderboard">
                <button className="flex items-center gap-1 text-sm transition-colors hover:text-white" style={{ color: 'rgba(255,255,255,0.3)' }}>
                  全部 <ArrowRight className="h-3.5 w-3.5" />
                </button>
              </Link>
            </div>

            <div className="space-y-1.5">
              {(loading ? Array(5).fill(null) : topUsers).map((u, i) => {
                const medals = ['🥇', '🥈', '🥉', '4', '5'];
                const isTop = i < 3;
                return (
                  <motion.div
                    key={u?.user_id ?? i}
                    initial={{ opacity: 0, x: 16 }}
                    whileInView={{ opacity: 1, x: 0 }}
                    viewport={{ once: true }}
                    transition={{ delay: i * 0.07 }}
                  >
                    {u ? (
                      <Link href={`/users/${u.user_id}`}>
                        <div
                          className="flex items-center gap-3 px-3.5 py-3 rounded-xl transition-all duration-300 group"
                          style={{
                            background: isTop ? 'rgba(250,204,21,0.04)' : '#0d0d0d',
                            border: `1px solid ${isTop ? 'rgba(250,204,21,0.12)' : BORDER}`,
                          }}
                        >
                          <span className="text-base w-6 text-center flex-shrink-0">
                            {isTop ? medals[i] : <span style={{ color: 'rgba(255,255,255,0.2)', fontSize: 12 }}>{i + 1}</span>}
                          </span>
                          <div className={`w-7 h-7 rounded-full bg-gradient-to-br ${hashGradient(u.user_id)} flex items-center justify-center flex-shrink-0`}>
                            <span className="text-white text-[10px] font-bold">{u.username[0]?.toUpperCase()}</span>
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-[13px] font-semibold text-white truncate">{u.username}</p>
                            <p className="text-[11px]" style={{ color: 'rgba(255,255,255,0.3)' }}>
                              {u.score.toLocaleString()} 积分
                            </p>
                          </div>
                          {i === 0 && <Star className="h-3.5 w-3.5 text-yellow-400 flex-shrink-0" />}
                        </div>
                      </Link>
                    ) : (
                      <div className="h-[52px] rounded-xl animate-pulse" style={{ background: '#0d0d0d' }} />
                    )}
                  </motion.div>
                );
              })}
            </div>
          </div>
        </div>
      </section>

      {/* ── EVENTS ──────────────────────────────────────────────── */}
      {(events.length > 0 || loading) && (
        <section className="py-24 px-4" style={{ background: '#080808', borderTop: `1px solid ${BORDER}` }}>
          <div className="max-w-6xl mx-auto">
            <div className="flex items-center justify-between mb-8">
              <SectionLabel eyebrow="线上线下" title="近期活动" icon={Calendar} iconColor="#f97316" />
              <Link href="/events">
                <button className="flex items-center gap-1 text-sm transition-colors hover:text-white" style={{ color: 'rgba(255,255,255,0.3)' }}>
                  全部 <ArrowRight className="h-3.5 w-3.5" />
                </button>
              </Link>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
              {(loading ? Array(4).fill(null) : events).map((ev, i) => (
                <motion.div
                  key={ev?.id ?? i}
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.09 }}
                >
                  {ev ? (
                    <Link href={`/events/${ev.id}`}>
                      <div
                        className="rounded-2xl overflow-hidden transition-all duration-400 hover:shadow-[0_0_30px_-10px_rgba(249,115,22,0.2)] hover:-translate-y-0.5 group"
                        style={{ border: `1px solid ${BORDER}`, background: '#0d0d0d' }}
                      >
                        <div className={`h-20 bg-gradient-to-br ${hashGradient(ev.id)} flex items-end p-3`}>
                          <span
                            className="text-white/80 text-[10px] font-medium px-2 py-0.5 rounded-full"
                            style={{ background: 'rgba(0,0,0,0.35)', backdropFilter: 'blur(4px)' }}
                          >
                            {ev.is_online ? '🌐 线上' : '📍 线下'}
                          </span>
                        </div>
                        <div className="p-4">
                          <p className="font-semibold text-[13px] text-white line-clamp-2 mb-1.5 group-hover:text-orange-300 transition-colors heading-tight">
                            {ev.title}
                          </p>
                          <p className="text-[11px]" style={{ color: 'rgba(255,255,255,0.3)' }}>
                            {formatDate(ev.start_time)} · {ev.attendee_count} 人参加
                          </p>
                        </div>
                      </div>
                    </Link>
                  ) : (
                    <div className="h-40 rounded-2xl animate-pulse" style={{ background: '#0d0d0d' }} />
                  )}
                </motion.div>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* ── BOTTOM CTA ──────────────────────────────────────────── */}
      <section
        className="relative py-36 px-4 text-center overflow-hidden"
        style={{ background: BG, borderTop: `1px solid ${BORDER}` }}
      >
        {/* Architectural glow — single centered, tight */}
        <div
          className="absolute inset-x-0 top-0 h-px"
          style={{ background: 'linear-gradient(to right, transparent, rgba(139,92,246,0.4), transparent)' }}
        />
        <div
          className="absolute top-0 left-1/2 -translate-x-1/2"
          style={{
            width: 500,
            height: 200,
            background: 'radial-gradient(ellipse at center top, rgba(139,92,246,0.12), transparent 70%)',
          }}
        />

        <motion.div
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="relative z-10"
        >
          <p className="text-[11px] uppercase tracking-[0.25em] mb-4" style={{ color: 'rgba(255,255,255,0.25)' }}>
            加入我们
          </p>
          <h2
            className="font-black mb-4 heading-tight"
            style={{ fontSize: 'clamp(2.5rem, 6vw, 4rem)', lineHeight: 1.1 }}
          >
            准备好了吗？
          </h2>
          <p
            className="text-base max-w-sm mx-auto mb-10 leading-relaxed font-warmth"
            style={{ color: 'rgba(255,255,255,0.38)' }}
          >
            加入 Furry 同好社区，与全球毛毛一起创作、分享、连接。
          </p>
          <div className="flex items-center justify-center gap-3 flex-wrap">
            <Link href="/register">
              <button
                className="px-8 py-3 rounded-xl text-sm font-semibold text-white transition-all duration-300 hover:brightness-110 hover:shadow-[0_0_30px_-8px_rgba(124,58,237,0.5)] active:scale-[0.98]"
                style={{ background: 'linear-gradient(135deg, #7c3aed, #0ea5e9)' }}
              >
                免费注册
              </button>
            </Link>
            <Link href="/explore">
              <button
                className="px-8 py-3 rounded-xl text-sm font-semibold transition-all duration-300 hover:bg-white/10 active:scale-[0.98]"
                style={{
                  background: 'rgba(255,255,255,0.04)',
                  border: `1px solid ${BORDER}`,
                  color: 'rgba(255,255,255,0.6)',
                }}
              >
                先逛逛
              </button>
            </Link>
          </div>
        </motion.div>
      </section>
    </div>
  );
}
