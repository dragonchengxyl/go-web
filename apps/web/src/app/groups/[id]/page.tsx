'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { apiClient, Group, GroupMember } from '@/lib/api-client';
import { Button } from '@/components/ui/button';

export default function GroupDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [group, setGroup] = useState<Group | null>(null);
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [joining, setJoining] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!id) return;
    Promise.all([apiClient.getGroup(id), apiClient.listGroupMembers(id)])
      .then(([g, mRes]) => {
        setGroup(g);
        setMembers(mRes.members ?? []);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [id]);

  const handleJoin = async () => {
    if (!id) return;
    setJoining(true);
    setError('');
    try {
      await apiClient.joinGroup(id);
      const updated = await apiClient.getGroup(id);
      setGroup(updated);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '加入失败');
    } finally {
      setJoining(false);
    }
  };

  const handleLeave = async () => {
    if (!id) return;
    setJoining(true);
    setError('');
    try {
      await apiClient.leaveGroup(id);
      const updated = await apiClient.getGroup(id);
      setGroup(updated);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '退出失败');
    } finally {
      setJoining(false);
    }
  };

  if (loading) {
    return (
      <main className="min-h-screen bg-slate-950 flex items-center justify-center">
        <div className="w-10 h-10 rounded-full border-2 border-purple-500 border-t-transparent animate-spin" />
      </main>
    );
  }

  if (!group) {
    return (
      <main className="min-h-screen bg-slate-950 flex items-center justify-center">
        <p className="text-white/50">圈子不存在或已被删除</p>
      </main>
    );
  }

  const ROLE_LABEL: Record<string, string> = {
    owner: '圈主',
    moderator: '管理员',
    member: '成员',
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-950 via-purple-950/30 to-slate-950 px-4 py-10">
      <div className="mx-auto max-w-2xl">
        {/* Header */}
        <div className="flex items-start gap-4 mb-6">
          <div className="w-16 h-16 rounded-2xl bg-purple-600/40 flex items-center justify-center text-3xl shrink-0">
            🐾
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <h1 className="text-2xl font-bold text-white">{group.name}</h1>
              {group.privacy === 'private' && (
                <span className="text-xs px-1.5 py-0.5 rounded bg-white/10 text-white/40">🔒 私密</span>
              )}
            </div>
            <div className="mt-1 flex gap-4 text-sm text-white/50">
              <span>👥 {group.member_count} 成员</span>
              <span>📝 {group.post_count} 帖子</span>
            </div>
          </div>
        </div>

        {/* Tags */}
        {group.tags?.length > 0 && (
          <div className="flex flex-wrap gap-1.5 mb-4">
            {group.tags.map((t) => (
              <span key={t} className="text-xs px-2 py-0.5 rounded-full bg-purple-500/20 text-purple-300">
                #{t}
              </span>
            ))}
          </div>
        )}

        {/* Description */}
        {group.description && (
          <div className="rounded-2xl border border-white/10 bg-white/5 p-5 mb-6">
            <p className="text-white/70 whitespace-pre-wrap leading-relaxed">{group.description}</p>
          </div>
        )}

        {/* Join / Leave */}
        <div className="mb-8">
          {error && <p className="text-red-400 text-sm mb-2">{error}</p>}
          <div className="flex gap-3">
            <Button
              onClick={handleJoin}
              disabled={joining}
              className="flex-1 bg-purple-600 hover:bg-purple-500 text-white py-2.5"
            >
              {joining ? '处理中…' : '加入圈子'}
            </Button>
            <Button
              onClick={handleLeave}
              disabled={joining}
              variant="outline"
              className="border-white/20 text-white/70 hover:text-white"
            >
              退出
            </Button>
          </div>
        </div>

        {/* Members */}
        {members.length > 0 && (
          <div>
            <h2 className="text-lg font-semibold text-white mb-3">成员（{members.length}）</h2>
            <div className="space-y-2">
              {members.map((m) => (
                <div
                  key={m.user_id}
                  className="flex items-center justify-between rounded-xl bg-white/5 px-4 py-2.5"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-purple-600/40 flex items-center justify-center text-xs text-purple-200">
                      {m.user_id.slice(0, 2).toUpperCase()}
                    </div>
                    <span className="text-sm text-white/70">{m.user_id.slice(0, 8)}…</span>
                  </div>
                  <span className="text-xs text-white/40">{ROLE_LABEL[m.role] ?? m.role}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
