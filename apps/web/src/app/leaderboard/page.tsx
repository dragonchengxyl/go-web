'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiClient, LeaderboardEntry, Post } from '@/lib/api-client';
import { PostGalleryCard } from '@/components/post/post-gallery-card';
import { Trophy, TrendingUp, Flame, Star, Crown } from 'lucide-react';
import { motion } from 'framer-motion';

type Tab = 'weekly' | 'alltime' | 'hotposts';

const GRADIENTS = [
  'from-purple-500 to-teal-400',
  'from-teal-400 to-blue-500',
  'from-orange-400 to-pink-500',
  'from-blue-500 to-indigo-600',
  'from-green-400 to-teal-500',
];
function hashGradient(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) hash = (hash * 31 + str.charCodeAt(i)) | 0;
  return GRADIENTS[Math.abs(hash) % GRADIENTS.length];
}

const RANK_BADGES: Record<number, { icon: typeof Crown; color: string; bg: string }> = {
  1: { icon: Crown, color: 'text-yellow-500', bg: 'bg-yellow-500/10' },
  2: { icon: Trophy, color: 'text-slate-400', bg: 'bg-slate-400/10' },
  3: { icon: Trophy, color: 'text-amber-600', bg: 'bg-amber-600/10' },
};

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};
const itemVariants = {
  hidden: { opacity: 0, x: -16 },
  show: { opacity: 1, x: 0, transition: { duration: 0.3, ease: 'easeOut' } },
};

function UserRankCard({ entry }: { entry: LeaderboardEntry }) {
  const badge = RANK_BADGES[entry.rank];
  const gradient = hashGradient(entry.user_id);
  const initial = (entry.username[0] || '?').toUpperCase();

  return (
    <motion.div variants={itemVariants}>
      <Link href={`/users/${entry.user_id}`}>
        <div className="flex items-center gap-3 p-3.5 rounded-xl border border-border/50 bg-card hover:border-primary/30 hover:bg-accent/30 transition-all group">
          {/* Rank */}
          <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${badge ? badge.bg : 'bg-muted'}`}>
            {badge ? (
              <badge.icon className={`h-4 w-4 ${badge.color}`} />
            ) : (
              <span className="text-xs font-bold text-muted-foreground">{entry.rank}</span>
            )}
          </div>

          {/* Avatar */}
          <div className={`w-10 h-10 rounded-full bg-gradient-to-br ${gradient} flex items-center justify-center text-white font-bold text-sm flex-shrink-0`}>
            {initial}
          </div>

          {/* Info */}
          <div className="flex-1 min-w-0">
            <p className="font-semibold text-sm truncate group-hover:text-primary transition-colors">
              {entry.username}
            </p>
            <p className="text-xs text-muted-foreground">积分 #{entry.rank}</p>
          </div>

          {/* Score */}
          <div className="text-right flex-shrink-0">
            <p className="font-bold text-sm text-primary">{Math.round(entry.score).toLocaleString()}</p>
            <p className="text-xs text-muted-foreground">pts</p>
          </div>
        </div>
      </Link>
    </motion.div>
  );
}

export default function LeaderboardPage() {
  const [tab, setTab] = useState<Tab>('weekly');
  const [weekly, setWeekly] = useState<LeaderboardEntry[]>([]);
  const [allTime, setAllTime] = useState<LeaderboardEntry[]>([]);
  const [hotPosts, setHotPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) apiClient.setToken(token);

    Promise.all([
      apiClient.getWeeklyLeaderboard(20).catch(() => []),
      apiClient.getLeaderboard(20).catch(() => []),
      apiClient.getExplore(1, 12).catch(() => ({ posts: [] })),
    ]).then(([w, a, p]) => {
      setWeekly(w);
      setAllTime(a);
      setHotPosts(p.posts ?? []);
    }).finally(() => setLoading(false));
  }, []);

  const TABS = [
    { id: 'weekly' as Tab, label: '本周新星', icon: Flame },
    { id: 'alltime' as Tab, label: '总榜', icon: Trophy },
    { id: 'hotposts' as Tab, label: '热门创作', icon: TrendingUp },
  ];

  const currentList = tab === 'weekly' ? weekly : allTime;

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-12">
      {/* Header */}
      <div className="flex items-center gap-3 mb-8">
        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center">
          <Trophy className="h-5 w-5 text-white" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">排行榜</h1>
          <p className="text-sm text-muted-foreground">社区活跃用户与热门创作</p>
        </div>
      </div>

      {/* Top 3 podium (weekly) */}
      {tab !== 'hotposts' && !loading && currentList.length >= 3 && (
        <div className="flex items-end justify-center gap-3 mb-8">
          {/* 2nd */}
          <div className="flex-1 text-center">
            <div className="h-20 bg-gradient-to-t from-slate-500/20 to-transparent rounded-t-xl flex items-end justify-center pb-3">
              <Link href={`/users/${currentList[1].user_id}`}>
                <div className={`w-12 h-12 rounded-full bg-gradient-to-br ${hashGradient(currentList[1].user_id)} flex items-center justify-center text-white font-bold mx-auto mb-1`}>
                  {currentList[1].username[0]?.toUpperCase()}
                </div>
              </Link>
            </div>
            <div className="bg-slate-400/10 border border-slate-400/20 rounded-b-xl pt-2 pb-3 px-2">
              <Trophy className="h-4 w-4 text-slate-400 mx-auto mb-1" />
              <p className="text-xs font-semibold truncate">{currentList[1].username}</p>
              <p className="text-xs text-muted-foreground">{Math.round(currentList[1].score).toLocaleString()} pts</p>
            </div>
          </div>
          {/* 1st */}
          <div className="flex-1 text-center">
            <div className="h-28 bg-gradient-to-t from-yellow-500/20 to-transparent rounded-t-xl flex items-end justify-center pb-3">
              <Link href={`/users/${currentList[0].user_id}`}>
                <div className={`w-14 h-14 rounded-full bg-gradient-to-br ${hashGradient(currentList[0].user_id)} flex items-center justify-center text-white font-bold text-lg mx-auto mb-1 ring-2 ring-yellow-500/50`}>
                  {currentList[0].username[0]?.toUpperCase()}
                </div>
              </Link>
            </div>
            <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-b-xl pt-2 pb-3 px-2">
              <Crown className="h-4 w-4 text-yellow-500 mx-auto mb-1" />
              <p className="text-xs font-semibold truncate">{currentList[0].username}</p>
              <p className="text-xs text-muted-foreground">{Math.round(currentList[0].score).toLocaleString()} pts</p>
            </div>
          </div>
          {/* 3rd */}
          <div className="flex-1 text-center">
            <div className="h-16 bg-gradient-to-t from-amber-600/20 to-transparent rounded-t-xl flex items-end justify-center pb-3">
              <Link href={`/users/${currentList[2].user_id}`}>
                <div className={`w-11 h-11 rounded-full bg-gradient-to-br ${hashGradient(currentList[2].user_id)} flex items-center justify-center text-white font-bold mx-auto mb-1`}>
                  {currentList[2].username[0]?.toUpperCase()}
                </div>
              </Link>
            </div>
            <div className="bg-amber-600/10 border border-amber-600/20 rounded-b-xl pt-2 pb-3 px-2">
              <Star className="h-4 w-4 text-amber-600 mx-auto mb-1" />
              <p className="text-xs font-semibold truncate">{currentList[2].username}</p>
              <p className="text-xs text-muted-foreground">{Math.round(currentList[2].score).toLocaleString()} pts</p>
            </div>
          </div>
        </div>
      )}

      {/* Tab switcher */}
      <div className="flex gap-1 border-b mb-6">
        {TABS.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            onClick={() => setTab(id)}
            className={`flex items-center gap-1.5 px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors ${
              tab === id
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Icon className="h-3.5 w-3.5" />
            {label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3, 4, 5].map(i => <div key={i} className="h-16 bg-muted animate-pulse rounded-xl" />)}
        </div>
      ) : tab === 'hotposts' ? (
        hotPosts.length === 0 ? (
          <div className="text-center py-16 text-muted-foreground">
            <TrendingUp className="h-12 w-12 mx-auto mb-4 opacity-30" />
            <p>暂无热门创作</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            {hotPosts.map(post => (
              <PostGalleryCard key={post.id} post={post} />
            ))}
          </div>
        )
      ) : currentList.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <Trophy className="h-12 w-12 mx-auto mb-4 opacity-30" />
          <p>暂无排行数据</p>
        </div>
      ) : (
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="show"
          className="space-y-2"
        >
          {currentList.map(entry => (
            <UserRankCard key={entry.user_id} entry={entry} />
          ))}
        </motion.div>
      )}
    </div>
  );
}
