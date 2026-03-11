'use client';

import { Suspense, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Lock, CheckCircle, XCircle } from 'lucide-react';

function ResetPasswordForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get('token') ?? '';

  const [form, setForm] = useState({ new_password: '', confirm: '' });
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState('');

  if (!token) {
    return (
      <div className="text-center">
        <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center mx-auto mb-4">
          <XCircle className="h-8 w-8 text-destructive" />
        </div>
        <h2 className="text-2xl font-bold mb-2">链接无效</h2>
        <p className="text-muted-foreground text-sm mb-6">
          此密码重置链接无效或已过期。请重新申请。
        </p>
        <Link href="/forgot-password">
          <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
            重新发送重置邮件
          </Button>
        </Link>
      </div>
    );
  }

  if (done) {
    return (
      <div className="text-center">
        <div className="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mx-auto mb-4">
          <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-400" />
        </div>
        <h2 className="text-2xl font-bold mb-2">密码重置成功</h2>
        <p className="text-muted-foreground text-sm mb-6">
          您的密码已更新，请使用新密码登录。
        </p>
        <Link href="/login">
          <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
            去登录
          </Button>
        </Link>
      </div>
    );
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    if (form.new_password !== form.confirm) {
      setError('两次输入的密码不一致');
      return;
    }
    setLoading(true);
    try {
      const res = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'}/auth/reset-password`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ token, new_password: form.new_password }),
        }
      );
      const data = await res.json();
      if (data.code !== 0) throw new Error(data.message || '重置失败');
      setDone(true);
    } catch (err: any) {
      setError(err.message || '重置失败，链接可能已过期');
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <div className="mb-8">
        <h2 className="text-2xl font-bold mb-1">设置新密码</h2>
        <p className="text-muted-foreground text-sm">请输入您的新密码</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        {error && (
          <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg">{error}</div>
        )}

        <div className="space-y-2">
          <label htmlFor="new_password" className="text-sm font-medium">新密码</label>
          <div className="relative">
            <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
            <Input
              id="new_password"
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
          <label htmlFor="confirm" className="text-sm font-medium">确认新密码</label>
          <div className="relative">
            <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
            <Input
              id="confirm"
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
          className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 animate-glow-pulse"
        >
          {loading ? '重置中...' : '确认重置密码'}
        </Button>
      </form>

      <p className="text-sm text-center text-muted-foreground mt-6">
        <Link href="/login" className="text-primary hover:underline font-medium">
          返回登录
        </Link>
      </p>
    </>
  );
}

export default function ResetPasswordPage() {
  return (
    <div className="min-h-screen grid md:grid-cols-2">
      {/* Brand panel */}
      <div className="hidden md:flex flex-col items-center justify-center bg-gradient-to-br from-brand-purple to-brand-teal p-12 relative overflow-hidden">
        <div className="absolute top-16 left-8 w-48 h-48 bg-white/10 rounded-full blur-3xl" />
        <div className="absolute bottom-16 right-8 w-32 h-32 bg-white/10 rounded-full blur-3xl" />
        <div className="relative z-10 text-center text-white">
          <div className="w-20 h-20 rounded-2xl bg-white/20 backdrop-blur-sm flex items-center justify-center text-4xl font-bold mx-auto mb-6 shadow-lg">
            F
          </div>
          <h1 className="text-3xl font-bold mb-3">Furry 同好社区</h1>
          <p className="text-white/80 text-lg">毛毛们的温暖家园</p>
        </div>
      </div>

      {/* Form panel */}
      <div className="flex items-center justify-center px-8 py-16 bg-background">
        <div className="w-full max-w-sm">
          <Suspense fallback={<div className="h-64 bg-muted animate-pulse rounded-xl" />}>
            <ResetPasswordForm />
          </Suspense>
        </div>
      </div>
    </div>
  );
}
