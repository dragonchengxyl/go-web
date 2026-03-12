'use client';

import { Suspense, useEffect, useState } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { CheckCircle, XCircle, Loader2 } from 'lucide-react';

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get('token') ?? '';

  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('正在验证邮箱...');

  useEffect(() => {
    if (!token) {
      setStatus('error');
      setMessage('验证链接无效或缺少参数。');
      return;
    }

    apiClient.verifyEmail(token)
      .then((data) => {
        setStatus('success');
        setMessage(data.message || '邮箱验证成功');
      })
      .catch((err: any) => {
        setStatus('error');
        setMessage(err.message || '验证失败，链接可能已过期');
      });
  }, [token]);

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-md rounded-2xl border bg-card p-8 text-center shadow-sm">
        {status === 'loading' && (
          <>
            <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
            <h1 className="text-2xl font-bold mb-2">验证中</h1>
            <p className="text-muted-foreground">{message}</p>
          </>
        )}

        {status === 'success' && (
          <>
            <div className="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mx-auto mb-4">
              <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-400" />
            </div>
            <h1 className="text-2xl font-bold mb-2">邮箱已验证</h1>
            <p className="text-muted-foreground mb-6">{message}</p>
            <Link href="/feed">
              <Button className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
                前往社区
              </Button>
            </Link>
          </>
        )}

        {status === 'error' && (
          <>
            <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center mx-auto mb-4">
              <XCircle className="h-8 w-8 text-destructive" />
            </div>
            <h1 className="text-2xl font-bold mb-2">验证失败</h1>
            <p className="text-muted-foreground mb-6">{message}</p>
            <Link href="/settings">
              <Button variant="outline" className="w-full">
                返回设置
              </Button>
            </Link>
          </>
        )}
      </div>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center px-4">
        <div className="w-full max-w-md rounded-2xl border bg-card p-8 text-center shadow-sm">
          <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
          <h1 className="text-2xl font-bold mb-2">验证中</h1>
          <p className="text-muted-foreground">正在加载验证结果...</p>
        </div>
      </div>
    }>
      <VerifyEmailContent />
    </Suspense>
  );
}
