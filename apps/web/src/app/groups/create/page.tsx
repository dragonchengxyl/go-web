'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export default function CreateGroupPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [form, setForm] = useState({
    name: '',
    description: '',
    tags: '',
    privacy: 'public' as 'public' | 'private',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!form.name.trim()) {
      setError('请填写圈子名称');
      return;
    }
    setLoading(true);
    try {
      const group = await apiClient.createGroup({
        name: form.name.trim(),
        description: form.description.trim(),
        tags: form.tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
        privacy: form.privacy,
      });
      router.push(`/groups/${group.id}`);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '创建失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const inputCls =
    'w-full rounded-xl bg-white/5 border border-white/10 px-4 py-2.5 text-white placeholder-white/30 focus:outline-none focus:ring-2 focus:ring-purple-500/50';

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-950 via-purple-950/30 to-slate-950 px-4 py-10">
      <div className="mx-auto max-w-lg">
        <h1 className="text-2xl font-bold text-white mb-6">创建圈子</h1>
        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm text-white/60 mb-1">圈子名称 *</label>
            <input
              className={inputCls}
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="给你的圈子起个名字"
              required
            />
          </div>

          <div>
            <label className="block text-sm text-white/60 mb-1">圈子简介</label>
            <textarea
              className={inputCls + ' resize-none min-h-[90px]'}
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              placeholder="介绍圈子的主题和规则"
            />
          </div>

          <div>
            <label className="block text-sm text-white/60 mb-1">标签（逗号分隔）</label>
            <input
              className={inputCls}
              value={form.tags}
              onChange={(e) => setForm({ ...form, tags: e.target.value })}
              placeholder="如：换装, 画作, 同人"
            />
          </div>

          <div>
            <label className="block text-sm text-white/60 mb-2">隐私设置</label>
            <div className="flex gap-3">
              {(['public', 'private'] as const).map((p) => (
                <button
                  key={p}
                  type="button"
                  onClick={() => setForm({ ...form, privacy: p })}
                  className={cn(
                    'flex-1 py-2.5 rounded-xl border text-sm transition-colors',
                    form.privacy === p
                      ? 'border-purple-500 bg-purple-500/20 text-purple-300'
                      : 'border-white/10 bg-white/5 text-white/50 hover:text-white'
                  )}
                >
                  {p === 'public' ? '🌐 公开圈子' : '🔒 私密圈子'}
                </button>
              ))}
            </div>
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <Button
            type="submit"
            disabled={loading}
            className="w-full bg-purple-600 hover:bg-purple-500 text-white py-2.5"
          >
            {loading ? '创建中…' : '创建圈子'}
          </Button>
        </form>
      </div>
    </main>
  );
}
