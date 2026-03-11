'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Users, FileText, Lock, Globe, Crown, Shield, UserPlus, UserMinus } from 'lucide-react';
import { apiClient, Group, GroupMember } from '@/lib/api-client';
import { Button } from '@/components/ui/button';

const ROLE_CONFIG: Record<string, { label: string; icon: typeof Crown; color: string }> = {
  owner: { label: '圈主', icon: Crown, color: 'text-yellow-500' },
  moderator: { label: '管理员', icon: Shield, color: 'text-blue-500' },
  member: { label: '成员', icon: Users, color: 'text-muted-foreground' },
};

const GRADIENTS = [
  'from-purple-500 to-teal-400',
  'from-teal-400 to-blue-500',
  'from-orange-400 to-pink-500',
  'from-blue-500 to-indigo-600',
  'from-green-400 to-teal-500',
  'from-pink-500 to-purple-600',
];
function hashGradient(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) hash = (hash * 31 + str.charCodeAt(i)) | 0;
  return GRADIENTS[Math.abs(hash) % GRADIENTS.length];
}

export default function GroupDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [group, setGroup] = useState<Group | null>(null);
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [joining, setJoining] = useState(false);
  const [isMember, setIsMember] = useState(false);
  const [myId, setMyId] = useState<string | null>(null);
  const [error, setError] = useState('');
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      apiClient.setToken(token);
      setIsLoggedIn(true);
      apiClient.getMe().then(me => {
        setMyId(me?.id ?? null);
      }).catch(() => {});
    }
    if (!id) return;
    Promise.all([apiClient.getGroup(id), apiClient.listGroupMembers(id)])
      .then(([g, mRes]) => {
        setGroup(g);
        const list = mRes.members ?? [];
        setMembers(list);
        if (token) {
          apiClient.getMe().then(me => {
            setIsMember(list.some(m => m.user_id === me?.id));
          }).catch(() => {});
        }
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
      const [updated, mRes] = await Promise.all([
        apiClient.getGroup(id),
        apiClient.listGroupMembers(id),
      ]);
      setGroup(updated);
      setMembers(mRes.members ?? []);
      setIsMember(true);
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
      const [updated, mRes] = await Promise.all([
        apiClient.getGroup(id),
        apiClient.listGroupMembers(id),
      ]);
      setGroup(updated);
      setMembers(mRes.members ?? []);
      setIsMember(false);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '退出失败');
    } finally {
      setJoining(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4">
        <div className="h-8 w-32 bg-muted animate-pulse rounded mb-6" />
        <div className="h-48 bg-muted animate-pulse rounded-2xl mb-4" />
        <div className="h-32 bg-muted animate-pulse rounded-2xl" />
      </div>
    );
  }

  if (!group) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4 text-center py-16 text-muted-foreground">
        <p className="text-xl font-medium mb-2">圈子不存在</p>
        <p className="text-sm mb-6">该圈子可能已被删除</p>
        <Link href="/groups"><Button variant="outline">返回圈子列表</Button></Link>
      </div>
    );
  }

  const isOwner = myId && members.find(m => m.user_id === myId)?.role === 'owner';
  const gradient = hashGradient(group.id);

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-12">
      {/* Back */}
      <Link href="/groups">
        <Button variant="ghost" size="sm" className="mb-6 -ml-2">
          <ArrowLeft className="h-4 w-4 mr-1" />返回圈子
        </Button>
      </Link>

      {/* Header card */}
      <div className="bg-card border rounded-2xl overflow-hidden mb-5">
        {/* Banner */}
        <div className={`h-24 bg-gradient-to-br ${gradient} opacity-60`} />

        <div className="px-6 pb-6">
          {/* Avatar row */}
          <div className="flex items-end justify-between -mt-8 mb-4">
            <div className="w-16 h-16 rounded-2xl bg-card border-4 border-background shadow-md flex items-center justify-center text-2xl bg-gradient-to-br from-brand-purple/20 to-brand-teal/20">
              🐾
            </div>
            <div className="flex items-center gap-1.5">
              {group.privacy === 'private' && (
                <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-muted text-muted-foreground border border-border">
                  <Lock className="h-3 w-3" />私密
                </span>
              )}
              {group.privacy === 'public' && (
                <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-green-500/10 text-green-600 dark:text-green-400 border border-green-500/20">
                  <Globe className="h-3 w-3" />公开
                </span>
              )}
            </div>
          </div>

          <h1 className="text-xl font-bold mb-1">{group.name}</h1>

          <div className="flex gap-4 text-sm text-muted-foreground mb-3">
            <span className="flex items-center gap-1.5">
              <Users className="h-3.5 w-3.5" />{group.member_count} 成员
            </span>
            <span className="flex items-center gap-1.5">
              <FileText className="h-3.5 w-3.5" />{group.post_count} 帖子
            </span>
          </div>

          {group.tags?.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mb-3">
              {group.tags.map(t => (
                <Link key={t} href={`/tags/${encodeURIComponent(t)}`}>
                  <span className="text-xs px-2 py-0.5 rounded-full bg-primary/10 text-primary hover:bg-primary/20 transition-colors cursor-pointer">
                    #{t}
                  </span>
                </Link>
              ))}
            </div>
          )}

          {group.description && (
            <p className="text-sm text-muted-foreground leading-relaxed">{group.description}</p>
          )}
        </div>
      </div>

      {/* Join/Leave */}
      {!isOwner && isLoggedIn && (
        <div className="mb-5">
          {error && <p className="text-destructive text-sm mb-2">{error}</p>}
          {isMember ? (
            <Button
              onClick={handleLeave}
              disabled={joining}
              variant="outline"
              className="w-full"
              size="lg"
            >
              <UserMinus className="h-4 w-4 mr-2" />
              {joining ? '处理中…' : '退出圈子'}
            </Button>
          ) : (
            <Button
              onClick={handleJoin}
              disabled={joining}
              size="lg"
              className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110"
            >
              <UserPlus className="h-4 w-4 mr-2" />
              {joining ? '处理中…' : '加入圈子'}
            </Button>
          )}
        </div>
      )}

      {!isLoggedIn && (
        <Link href="/login" className="block mb-5">
          <Button size="lg" className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
            <UserPlus className="h-4 w-4 mr-2" />
            登录后加入
          </Button>
        </Link>
      )}

      {/* Members */}
      {members.length > 0 && (
        <div className="bg-card border rounded-2xl p-6">
          <h2 className="font-semibold mb-4">成员（{members.length}）</h2>
          <div className="space-y-2">
            {members.map(m => {
              const roleConf = ROLE_CONFIG[m.role] ?? ROLE_CONFIG.member;
              const RoleIcon = roleConf.icon;
              const mGradient = hashGradient(m.user_id);
              return (
                <Link key={m.user_id} href={`/users/${m.user_id}`}
                  className="flex items-center justify-between p-2.5 rounded-xl hover:bg-muted/60 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div className={`w-9 h-9 rounded-full bg-gradient-to-br ${mGradient} flex items-center justify-center text-xs text-white font-bold flex-shrink-0`}>
                      {m.user_id.slice(0, 2).toUpperCase()}
                    </div>
                    <span className="text-sm font-medium">{m.user_id.slice(0, 8)}…</span>
                  </div>
                  <span className={`flex items-center gap-1 text-xs ${roleConf.color}`}>
                    <RoleIcon className="h-3 w-3" />
                    {roleConf.label}
                  </span>
                </Link>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
