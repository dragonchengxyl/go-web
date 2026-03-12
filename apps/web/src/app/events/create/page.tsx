'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { useAuth } from '@/contexts/auth-context';
import { Button } from '@/components/ui/button';

export default function CreateEventPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [form, setForm] = useState({
    title: '',
    description: '',
    location: '',
    is_online: false,
    start_time: '',
    end_time: '',
    max_capacity: 0,
    tags: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!form.title || !form.start_time || !form.end_time) {
      setError('请填写标题、开始时间和结束时间');
      return;
    }
    setLoading(true);
    try {
      const event = await apiClient.createEvent({
        title: form.title,
        description: form.description,
        location: form.location,
        is_online: form.is_online,
        start_time: new Date(form.start_time).toISOString(),
        end_time: new Date(form.end_time).toISOString(),
        max_capacity: Number(form.max_capacity) || 0,
        tags: form.tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
      });
      router.push(`/events/${event.id}`);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '发起失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const inputCls =
    'w-full rounded-xl bg-white/5 border border-white/10 px-4 py-2.5 text-white placeholder-white/30 focus:outline-none focus:ring-2 focus:ring-purple-500/50';

  if (!authLoading && user && !user.email_verified_at) {
    return (
      <main className="min-h-screen bg-gradient-to-br from-slate-950 via-purple-950/30 to-slate-950 px-4 py-10">
        <div className="mx-auto max-w-xl">
          <div className="rounded-2xl border border-white/10 bg-white/5 p-8 text-center">
            <h1 className="text-2xl font-bold text-white mb-2">先验证邮箱</h1>
            <p className="text-white/60 mb-6">发起活动前需要先完成邮箱验证。</p>
            <div className="flex justify-center gap-3">
              <Button onClick={() => router.push('/settings')} className="bg-purple-600 hover:bg-purple-500 text-white">
                去设置验证
              </Button>
              <Button variant="outline" onClick={() => router.push('/events')} className="border-white/20 text-white/80">
                返回活动
              </Button>
            </div>
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-950 via-purple-950/30 to-slate-950 px-4 py-10">
      <div className="mx-auto max-w-xl">
        <h1 className="text-2xl font-bold text-white mb-6">发起活动</h1>
        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm text-white/60 mb-1">活动标题 *</label>
            <input
              className={inputCls}
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              placeholder="给活动起个名字"
              required
            />
          </div>

          <div>
            <label className="block text-sm text-white/60 mb-1">活动描述</label>
            <textarea
              className={inputCls + ' resize-none min-h-[100px]'}
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              placeholder="介绍一下活动内容、注意事项等"
            />
          </div>

          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="is_online"
              checked={form.is_online}
              onChange={(e) => setForm({ ...form, is_online: e.target.checked })}
              className="w-4 h-4 rounded accent-purple-500"
            />
            <label htmlFor="is_online" className="text-sm text-white/70">
              线上活动（无需填写地点）
            </label>
          </div>

          {!form.is_online && (
            <div>
              <label className="block text-sm text-white/60 mb-1">活动地点</label>
              <input
                className={inputCls}
                value={form.location}
                onChange={(e) => setForm({ ...form, location: e.target.value })}
                placeholder="城市 / 具体地址"
              />
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-white/60 mb-1">开始时间 *</label>
              <input
                type="datetime-local"
                className={inputCls}
                value={form.start_time}
                onChange={(e) => setForm({ ...form, start_time: e.target.value })}
                required
              />
            </div>
            <div>
              <label className="block text-sm text-white/60 mb-1">结束时间 *</label>
              <input
                type="datetime-local"
                className={inputCls}
                value={form.end_time}
                onChange={(e) => setForm({ ...form, end_time: e.target.value })}
                required
              />
            </div>
          </div>

          <div>
            <label className="block text-sm text-white/60 mb-1">人数上限（0 = 不限）</label>
            <input
              type="number"
              min={0}
              className={inputCls}
              value={form.max_capacity}
              onChange={(e) => setForm({ ...form, max_capacity: Number(e.target.value) })}
            />
          </div>

          <div>
            <label className="block text-sm text-white/60 mb-1">标签（逗号分隔）</label>
            <input
              className={inputCls}
              value={form.tags}
              onChange={(e) => setForm({ ...form, tags: e.target.value })}
              placeholder="如：换装, 聚餐, 游戏"
            />
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <Button
            type="submit"
            disabled={loading}
            className="w-full bg-purple-600 hover:bg-purple-500 text-white py-2.5"
          >
            {loading ? '发布中…' : '发布活动'}
          </Button>
        </form>
      </div>
    </main>
  );
}
