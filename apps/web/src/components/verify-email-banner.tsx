'use client';

import Link from 'next/link';
import { useState } from 'react';
import { Mail, X } from 'lucide-react';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';

export function VerifyEmailBanner() {
  const { user, isLoggedIn } = useAuth();
  const [dismissed, setDismissed] = useState(false);
  const [sending, setSending] = useState(false);
  const [message, setMessage] = useState('');

  if (!isLoggedIn || !user || user.email_verified_at || dismissed) {
    return null;
  }

  async function handleResend() {
    setSending(true);
    setMessage('');
    try {
      const data = await apiClient.resendVerification();
      setMessage(data.message || '验证邮件已发送');
    } catch (err: any) {
      setMessage(err.message || '发送失败，请稍后重试');
    } finally {
      setSending(false);
    }
  }

  return (
    <div className="border-b bg-amber-50 text-amber-950 dark:bg-amber-950/30 dark:text-amber-100">
      <div className="container mx-auto px-4 py-3 flex items-start justify-between gap-4">
        <div className="flex items-start gap-3 min-w-0">
          <Mail className="h-4 w-4 mt-0.5 shrink-0" />
          <div className="min-w-0">
            <p className="text-sm font-medium">邮箱尚未验证</p>
            <p className="text-sm opacity-80">
              建议先验证邮箱，便于找回密码并解锁发布内容。
            </p>
            {message && <p className="text-xs mt-1 opacity-80">{message}</p>}
          </div>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <Button size="sm" variant="outline" onClick={handleResend} disabled={sending}>
            {sending ? '发送中...' : '重新发送'}
          </Button>
          <Link href="/settings">
            <Button size="sm">去设置</Button>
          </Link>
          <button
            type="button"
            onClick={() => setDismissed(true)}
            className="rounded p-1 opacity-70 hover:opacity-100 transition-opacity"
            aria-label="关闭提示"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
