'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Mail, ArrowLeft, CheckCircle } from 'lucide-react';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'}/auth/forgot-password`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email }),
        }
      );
      const data = await res.json();
      if (data.code !== 0) throw new Error(data.message || '请求失败');
      setSent(true);
    } catch (err: any) {
      setError(err.message || '发送失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  }

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
          <p className="text-white/80 text-lg mb-8">毛毛们的温暖家园</p>
          <div className="space-y-3 text-sm text-white/70 max-w-xs">
            <p>找回密码后继续探索</p>
            <p>分享你的兽设与创作</p>
            <p>结识来自世界各地的同好</p>
          </div>
        </div>
      </div>

      {/* Form panel */}
      <div className="flex items-center justify-center px-8 py-16 bg-background">
        <div className="w-full max-w-sm">
          <Link
            href="/login"
            className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground mb-8 transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            返回登录
          </Link>

          {sent ? (
            <div className="text-center">
              <div className="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mx-auto mb-4">
                <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-400" />
              </div>
              <h2 className="text-2xl font-bold mb-2">邮件已发送</h2>
              <p className="text-muted-foreground text-sm mb-6">
                如果 <span className="font-medium text-foreground">{email}</span> 已在我们的系统中注册，
                您将收到一封包含重置链接的邮件。请查收邮箱（含垃圾邮件文件夹）。
              </p>
              <Link href="/login">
                <Button className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
                  返回登录
                </Button>
              </Link>
            </div>
          ) : (
            <>
              <div className="mb-8">
                <h2 className="text-2xl font-bold mb-1">忘记密码</h2>
                <p className="text-muted-foreground text-sm">
                  输入您的注册邮箱，我们将发送密码重置链接
                </p>
              </div>

              <form onSubmit={handleSubmit} className="space-y-5">
                {error && (
                  <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg">
                    {error}
                  </div>
                )}

                <div className="space-y-2">
                  <label htmlFor="email" className="text-sm font-medium">邮箱</label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="email"
                      type="email"
                      placeholder="请输入注册邮箱"
                      className="pl-10"
                      value={email}
                      onChange={e => setEmail(e.target.value)}
                      required
                    />
                  </div>
                </div>

                <Button
                  type="submit"
                  className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 animate-glow-pulse"
                  disabled={loading}
                >
                  {loading ? '发送中...' : '发送重置链接'}
                </Button>
              </form>

              <p className="text-sm text-center text-muted-foreground mt-6">
                想起密码了？{' '}
                <Link href="/login" className="text-primary hover:underline font-medium">
                  立即登录
                </Link>
              </p>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
