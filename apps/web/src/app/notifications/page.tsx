'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Bell, Heart, MessageCircle, UserPlus, Gift, CheckCheck } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { apiClient, Notification } from '@/lib/api-client';

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}小时前`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}天前`;
  return new Date(dateStr).toLocaleDateString('zh-CN');
}

function NotificationIcon({ type }: { type: Notification['type'] }) {
  switch (type) {
    case 'like':
      return <Heart className="h-4 w-4 text-red-500" />;
    case 'comment':
      return <MessageCircle className="h-4 w-4 text-blue-500" />;
    case 'follow':
      return <UserPlus className="h-4 w-4 text-green-500" />;
    case 'tip':
      return <Gift className="h-4 w-4 text-yellow-500" />;
    default:
      return <Bell className="h-4 w-4 text-muted-foreground" />;
  }
}

function notificationText(n: Notification): string {
  const actor = n.actor_username || '有人';
  switch (n.type) {
    case 'like':
      return `${actor} 点赞了你的帖子`;
    case 'comment':
      return `${actor} 评论了你的帖子`;
    case 'follow':
      return `${actor} 关注了你`;
    case 'tip':
      return `${actor} 给你打赏了`;
    case 'system':
      if (n.target_type === 'post_approved') return '你的帖子已通过审核';
      if (n.target_type === 'post_blocked') return '你的帖子未通过审核';
      if (n.target_type === 'report_reviewed') return '你提交的举报已处理';
      if (n.target_type === 'report_dismissed') return '你提交的举报已被忽略';
      if (n.target_type === 'report_post_blocked') return '你举报的帖子已被处理';
      if (n.target_type === 'report_comment_deleted') return '你举报的评论已被删除';
      if (n.target_type === 'report_user_banned') return '你举报的用户已被封禁';
      return '你有一条系统通知';
    default:
      return '你有一条新通知';
  }
}

function notificationLink(n: Notification): string {
  switch (n.type) {
    case 'like':
    case 'comment':
      return n.target_id ? `/posts/${n.target_id}` : '/feed';
    case 'follow':
      return n.target_id ? `/users/${n.target_id}` : '/feed';
    case 'system':
      if (n.target_type === 'post_approved' && n.target_id) return `/posts/${n.target_id}`;
      if (n.target_type === 'report_post_blocked' && n.target_id) return `/posts/${n.target_id}`;
      return '/creator';
    default:
      return '/feed';
  }
}

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) return;
    apiClient.setToken(token);
    loadNotifications();
  }, []);

  async function loadNotifications() {
    try {
      const data = await apiClient.getNotifications(1, 50);
      setNotifications(data.notifications ?? []);
      setTotal(data.total);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }

  async function handleMarkAllRead() {
    try {
      await apiClient.markNotificationsRead([]);
      setNotifications(prev => prev.map(n => ({ ...n, is_read: true })));
    } catch {
      // ignore
    }
  }

  async function handleMarkRead(id: string) {
    try {
      await apiClient.markNotificationsRead([id]);
      setNotifications(prev =>
        prev.map(n => (n.id === id ? { ...n, is_read: true } : n))
      );
    } catch {
      // ignore
    }
  }

  const unreadCount = notifications.filter(n => !n.is_read).length;

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Bell className="h-6 w-6" />
          通知
          {unreadCount > 0 && (
            <span className="ml-1 text-sm font-normal bg-red-500 text-white rounded-full px-2 py-0.5">
              {unreadCount}
            </span>
          )}
        </h1>
        {unreadCount > 0 && (
          <Button variant="ghost" size="sm" onClick={handleMarkAllRead} className="text-muted-foreground">
            <CheckCheck className="h-4 w-4 mr-1" />
            全部已读
          </Button>
        )}
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map(i => (
            <div key={i} className="h-16 rounded-lg bg-muted animate-pulse" />
          ))}
        </div>
      ) : notifications.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <Bell className="h-12 w-12 mx-auto mb-4 opacity-30" />
          <p>暂无通知</p>
        </div>
      ) : (
        <div className="space-y-1">
          {notifications.map(n => (
            <Link
              key={n.id}
              href={notificationLink(n)}
              onClick={() => !n.is_read && handleMarkRead(n.id)}
              className={`flex items-start gap-3 p-3 rounded-lg hover:bg-accent transition-colors ${
                !n.is_read ? 'bg-accent/50' : ''
              }`}
            >
              <div className="mt-0.5 w-8 h-8 rounded-full bg-muted flex items-center justify-center flex-shrink-0">
                <NotificationIcon type={n.type} />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm leading-snug">{notificationText(n)}</p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {timeAgo(n.created_at)}
                </p>
              </div>
              {!n.is_read && (
                <div className="w-2 h-2 rounded-full bg-primary flex-shrink-0 mt-2" />
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
