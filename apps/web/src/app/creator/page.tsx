'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { Heart, MessageCircle, Users, FileText, Gift, TrendingUp } from 'lucide-react';

interface CreatorStats {
  post_count: number;
  total_likes: number;
  total_comments: number;
  follower_count: number;
  following_count: number;
  tip_total_cents: number;
  tip_count: number;
}

function StatCard({ icon: Icon, label, value, sub }: { icon: any; label: string; value: string | number; sub?: string }) {
  return (
    <div className="bg-card border rounded-xl p-5">
      <div className="flex items-center gap-3 mb-3">
        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
          <Icon className="h-5 w-5 text-primary" />
        </div>
        <span className="text-sm text-muted-foreground">{label}</span>
      </div>
      <p className="text-3xl font-bold">{value}</p>
      {sub && <p className="text-xs text-muted-foreground mt-1">{sub}</p>}
    </div>
  );
}

export default function CreatorDashboard() {
  const router = useRouter();
  const [stats, setStats] = useState<CreatorStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) { router.push('/login'); return; }
    apiClient.setToken(token);
    apiClient.getCreatorStats()
      .then(setStats)
      .catch(e => setError(e.message || '加载失败'))
      .finally(() => setLoading(false));
  }, [router]);

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto pt-20 px-4">
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {[...Array(6)].map((_, i) => <div key={i} className="h-32 bg-muted animate-pulse rounded-xl" />)}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-3xl mx-auto pt-20 px-4 text-center py-16 text-destructive">{error}</div>
    );
  }

  if (!stats) return null;

  const tipYuan = (stats.tip_total_cents / 100).toFixed(2);

  return (
    <div className="max-w-3xl mx-auto pt-20 px-4 pb-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">创作者仪表盘</h1>
          <p className="text-sm text-muted-foreground mt-1">查看你的内容数据</p>
        </div>
        <Link href="/posts/create">
          <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors">
            发布内容
          </button>
        </Link>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        <StatCard icon={FileText} label="发布动态" value={stats.post_count} sub="累计发布" />
        <StatCard icon={Users} label="粉丝数" value={stats.follower_count} sub={`关注了 ${stats.following_count} 人`} />
        <StatCard icon={Heart} label="获赞总数" value={stats.total_likes} sub="所有动态累计" />
        <StatCard icon={MessageCircle} label="评论总数" value={stats.total_comments} sub="所有动态累计" />
        <StatCard icon={Gift} label="打赏收入" value={`¥${tipYuan}`} sub={`共 ${stats.tip_count} 笔打赏`} />
        <StatCard
          icon={TrendingUp}
          label="互动率"
          value={stats.post_count > 0 ? `${((stats.total_likes + stats.total_comments) / stats.post_count).toFixed(1)}` : '0'}
          sub="平均每篇互动"
        />
      </div>

      {/* Tips history link */}
      {stats.tip_count > 0 && (
        <div className="mt-6 p-4 bg-yellow-50 dark:bg-yellow-950/20 border border-yellow-200 dark:border-yellow-800 rounded-xl">
          <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
            你已收到 {stats.tip_count} 笔打赏，累计 ¥{tipYuan}
          </p>
        </div>
      )}
    </div>
  );
}
