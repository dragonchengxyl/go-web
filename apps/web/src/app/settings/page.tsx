'use client';

import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Lock } from 'lucide-react';

interface UserProfile {
  id: string
  username: string
  email?: string
  email_verified_at?: string
  bio?: string
  furry_name?: string
  species?: string
}

interface BlockedUser {
  id: string
  username: string
  furry_name?: string
  species?: string
}

function ChangePasswordForm() {
  const [form, setForm] = useState({ old_password: '', new_password: '', confirm: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setSuccess(false);
    if (form.new_password !== form.confirm) {
      setError('两次输入的新密码不一致');
      return;
    }
    if (form.new_password.length < 8) {
      setError('新密码至少 8 位');
      return;
    }
    setLoading(true);
    try {
      await apiClient.put('/auth/password', {
        old_password: form.old_password,
        new_password: form.new_password,
      });
      setSuccess(true);
      setForm({ old_password: '', new_password: '', confirm: '' });
    } catch (err: any) {
      setError(err.message || '修改失败，请重试');
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg">{error}</div>
      )}
      {success && (
        <div className="bg-green-500/10 text-green-600 dark:text-green-400 text-sm p-3 rounded-lg">
          密码修改成功
        </div>
      )}
      <div className="space-y-2">
        <label className="text-sm font-medium">当前密码</label>
        <div className="relative">
          <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
          <Input
            type="password"
            className="pl-10"
            placeholder="请输入当前密码"
            value={form.old_password}
            onChange={e => setForm({ ...form, old_password: e.target.value })}
            required
          />
        </div>
      </div>
      <div className="space-y-2">
        <label className="text-sm font-medium">新密码</label>
        <div className="relative">
          <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
          <Input
            type="password"
            className="pl-10"
            placeholder="至少 8 位"
            value={form.new_password}
            onChange={e => setForm({ ...form, new_password: e.target.value })}
            required
          />
        </div>
      </div>
      <div className="space-y-2">
        <label className="text-sm font-medium">确认新密码</label>
        <div className="relative">
          <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
          <Input
            type="password"
            className="pl-10"
            placeholder="再次输入新密码"
            value={form.confirm}
            onChange={e => setForm({ ...form, confirm: e.target.value })}
            required
          />
        </div>
      </div>
      <Button
        type="submit"
        disabled={loading}
        className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110"
      >
        {loading ? '修改中...' : '修改密码'}
      </Button>
    </form>
  );
}

export default function SettingsPage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [blocked, setBlocked] = useState<BlockedUser[]>([]);
  const [loadingBlocked, setLoadingBlocked] = useState(false);
  const [unblockingId, setUnblockingId] = useState<string | null>(null);
  const [resendingVerification, setResendingVerification] = useState(false);
  const [verificationMessage, setVerificationMessage] = useState('');

  useEffect(() => {
    apiClient.getMe().then(setProfile).catch(() => {});
  }, []);

  async function loadBlocked() {
    if (loadingBlocked) return;
    setLoadingBlocked(true);
    try {
      const data = await apiClient.getBlockedUsers();
      setBlocked(data.users ?? []);
    } catch {
      setBlocked([]);
    } finally {
      setLoadingBlocked(false);
    }
  }

  async function handleUnblock(userId: string) {
    setUnblockingId(userId);
    try {
      await apiClient.unblockUser(userId);
      setBlocked(prev => prev.filter(u => u.id !== userId));
    } catch {
      // ignore
    } finally {
      setUnblockingId(null);
    }
  }

  async function handleResendVerification() {
    setResendingVerification(true);
    setVerificationMessage('');
    try {
      const data = await apiClient.resendVerification();
      setVerificationMessage(data.message || '验证邮件已发送');
    } catch (err: any) {
      setVerificationMessage(err.message || '发送失败，请稍后重试');
    } finally {
      setResendingVerification(false);
    }
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <h1 className="text-2xl font-bold mb-6">设置</h1>

      <Tabs defaultValue="account" onValueChange={(v) => { if (v === 'privacy') loadBlocked(); }}>
        <TabsList className="w-full mb-6">
          <TabsTrigger value="account" className="flex-1">账号</TabsTrigger>
          <TabsTrigger value="security" className="flex-1">安全</TabsTrigger>
          <TabsTrigger value="privacy" className="flex-1">隐私</TabsTrigger>
        </TabsList>

        {/* Account Tab */}
        <TabsContent value="account">
          <Card>
            <CardHeader>
              <CardTitle>账号信息</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="text-sm font-medium text-muted-foreground">用户名</label>
                <p className="mt-1 text-base">{profile?.username ?? '—'}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">邮箱</label>
                <p className="mt-1 text-base">{profile?.email ?? '—'}</p>
                {profile?.email && (
                  <div className="mt-2 flex items-center gap-3 text-sm">
                    <span className={profile.email_verified_at ? 'text-green-600 dark:text-green-400' : 'text-amber-600 dark:text-amber-400'}>
                      {profile.email_verified_at ? '已验证' : '未验证'}
                    </span>
                    {!profile.email_verified_at && (
                      <Button variant="outline" size="sm" disabled={resendingVerification} onClick={handleResendVerification}>
                        {resendingVerification ? '发送中...' : '重新发送验证邮件'}
                      </Button>
                    )}
                  </div>
                )}
                {verificationMessage && (
                  <p className="mt-2 text-sm text-muted-foreground">{verificationMessage}</p>
                )}
              </div>
              {profile?.furry_name && (
                <div>
                  <label className="text-sm font-medium text-muted-foreground">兽名</label>
                  <p className="mt-1 text-base">{profile.furry_name}</p>
                </div>
              )}
              {profile?.species && (
                <div>
                  <label className="text-sm font-medium text-muted-foreground">物种</label>
                  <p className="mt-1 text-base">{profile.species}</p>
                </div>
              )}
              <div className="pt-2 border-t">
                <p className="text-sm text-muted-foreground">
                  如需修改资料，请前往{' '}
                  <a href="/profile" className="text-primary hover:underline">个人资料页</a>。
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Security Tab */}
        <TabsContent value="security">
          <Card>
            <CardHeader>
              <CardTitle>修改密码</CardTitle>
            </CardHeader>
            <CardContent>
              <ChangePasswordForm />
            </CardContent>
          </Card>
        </TabsContent>

        {/* Privacy Tab */}
        <TabsContent value="privacy">
          <Card>
            <CardHeader>
              <CardTitle>屏蔽列表</CardTitle>
            </CardHeader>
            <CardContent>
              {loadingBlocked ? (
                <div className="space-y-2">
                  {[1, 2, 3].map(i => <div key={i} className="h-12 bg-muted animate-pulse rounded-lg" />)}
                </div>
              ) : blocked.length === 0 ? (
                <p className="text-center py-8 text-muted-foreground">你还没有屏蔽任何用户</p>
              ) : (
                <div className="space-y-2">
                  {blocked.map(user => (
                    <div key={user.id} className="flex items-center justify-between p-3 border rounded-lg">
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                          <span className="text-sm font-bold text-primary">
                            {(user.furry_name || user.username)[0]?.toUpperCase()}
                          </span>
                        </div>
                        <div>
                          <p className="text-sm font-medium">{user.furry_name || user.username}</p>
                          {user.furry_name && (
                            <p className="text-xs text-muted-foreground">@{user.username}</p>
                          )}
                        </div>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={unblockingId === user.id}
                        onClick={() => handleUnblock(user.id)}
                      >
                        解除屏蔽
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
