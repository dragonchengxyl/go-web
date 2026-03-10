'use client';

import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface UserProfile {
  id: string
  username: string
  email?: string
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

export default function SettingsPage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [blocked, setBlocked] = useState<BlockedUser[]>([]);
  const [loadingBlocked, setLoadingBlocked] = useState(false);
  const [unblockingId, setUnblockingId] = useState<string | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) apiClient.setToken(token);
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

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <h1 className="text-2xl font-bold mb-6">设置</h1>

      <Tabs defaultValue="account" onValueChange={(v) => { if (v === 'privacy') loadBlocked(); }}>
        <TabsList className="w-full mb-6">
          <TabsTrigger value="account" className="flex-1">账号</TabsTrigger>
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
                <p className="text-sm text-muted-foreground mt-1">
                  密码修改功能即将上线，敬请期待。
                </p>
              </div>
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
