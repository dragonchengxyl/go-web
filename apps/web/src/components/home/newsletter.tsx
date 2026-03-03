'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Mail } from 'lucide-react';

export function Newsletter() {
  const [email, setEmail] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // TODO: Implement newsletter subscription
    console.log('Subscribe:', email);
  };

  return (
    <section className="py-20 bg-primary text-primary-foreground">
      <div className="container mx-auto px-4">
        <div className="max-w-2xl mx-auto text-center">
          <Mail className="h-12 w-12 mx-auto mb-4" />
          <h2 className="text-3xl font-bold mb-4">订阅我们的新闻</h2>
          <p className="text-lg mb-8 opacity-90">
            第一时间获取最新游戏发布、更新和特别优惠信息
          </p>
          <form onSubmit={handleSubmit} className="flex gap-4 max-w-md mx-auto">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="输入您的邮箱"
              className="flex-1 px-4 py-2 rounded-md text-foreground"
              required
            />
            <Button
              type="submit"
              variant="secondary"
              className="whitespace-nowrap"
            >
              订阅
            </Button>
          </form>
        </div>
      </div>
    </section>
  );
}
