'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useMutation } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Mail, Lock, User, Check, X } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { cn } from '@/lib/utils';

function PasswordHints({ password }: { password: string }) {
  const rules = [
    { label: '至少 8 个字符', ok: password.length >= 8 },
    { label: '包含小写字母', ok: /[a-z]/.test(password) },
    { label: '包含数字', ok: /[0-9]/.test(password) },
  ];

  if (!password) return null;

  return (
    <ul className="mt-1.5 space-y-1">
      {rules.map((r) => (
        <li key={r.label} className={cn('flex items-center gap-1.5 text-xs', r.ok ? 'text-green-500' : 'text-muted-foreground')}>
          {r.ok ? <Check className="w-3 h-3" /> : <X className="w-3 h-3" />}
          {r.label}
        </li>
      ))}
    </ul>
  );
}

export default function RegisterPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  });
  const [error, setError] = useState('');

  const registerMutation = useMutation({
    mutationFn: () => apiClient.register(formData.username, formData.email, formData.password),
    onSuccess: () => {
      router.push('/login?registered=true');
    },
    onError: (err: any) => {
      setError(err.message || '注册失败，请重试');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (formData.password !== formData.confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }
    registerMutation.mutate();
  };

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
          <h1 className="text-3xl font-bold mb-3">加入 Furry 社区</h1>
          <p className="text-white/80 text-lg mb-8">毛毛们的温暖家园</p>
          <div className="space-y-3 text-sm text-white/70 max-w-xs">
            <p>注册完全免费</p>
            <p>分享你的兽设与创作</p>
            <p>结识来自世界各地的同好</p>
          </div>
        </div>
      </div>

      {/* Form panel */}
      <div className="flex items-center justify-center px-8 py-16 bg-background">
        <div className="w-full max-w-sm">
          <div className="mb-8">
            <h2 className="text-2xl font-bold mb-1">创建账号</h2>
            <p className="text-muted-foreground text-sm">加入我们的 Furry 社区</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-lg">
                {error}
              </div>
            )}

            <div className="space-y-2">
              <label htmlFor="username" className="text-sm font-medium">用户名</label>
              <div className="relative">
                <User className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="username"
                  type="text"
                  placeholder="3-20 位字母、数字、下划线"
                  className="pl-10"
                  value={formData.username}
                  onChange={e => setFormData({ ...formData, username: e.target.value })}
                  required
                />
              </div>
            </div>

            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium">邮箱</label>
              <div className="relative">
                <Mail className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  placeholder="请输入邮箱"
                  className="pl-10"
                  value={formData.email}
                  onChange={e => setFormData({ ...formData, email: e.target.value })}
                  required
                />
              </div>
            </div>

            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium">密码</label>
              <div className="relative">
                <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="password"
                  type="password"
                  placeholder="至少 8 位，含小写字母和数字"
                  className="pl-10"
                  value={formData.password}
                  onChange={e => setFormData({ ...formData, password: e.target.value })}
                  required
                />
              </div>
              <PasswordHints password={formData.password} />
            </div>

            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-sm font-medium">确认密码</label>
              <div className="relative">
                <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder="请再次输入密码"
                  className="pl-10"
                  value={formData.confirmPassword}
                  onChange={e => setFormData({ ...formData, confirmPassword: e.target.value })}
                  required
                />
              </div>
              {formData.confirmPassword && formData.password !== formData.confirmPassword && (
                <p className="text-xs text-destructive flex items-center gap-1">
                  <X className="w-3 h-3" /> 两次密码不一致
                </p>
              )}
            </div>

            <Button
              type="submit"
              className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 animate-glow-pulse"
              disabled={registerMutation.isPending}
            >
              {registerMutation.isPending ? '注册中...' : '注册'}
            </Button>
          </form>

          <p className="text-sm text-center text-muted-foreground mt-6">
            已有账号？{' '}
            <Link href="/login" className="text-primary hover:underline font-medium">
              立即登录
            </Link>
          </p>
          <p className="text-xs text-center text-muted-foreground mt-3">
            注册即表示您同意我们的{' '}
            <Link href="/terms" className="hover:underline">服务条款</Link>
            {' '}和{' '}
            <Link href="/privacy" className="hover:underline">隐私政策</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
