'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Gift } from 'lucide-react';

interface TipModalProps {
  toUserId: string
  toUsername: string
  onClose: () => void
  onSuccess?: () => void
}

export function TipModal({ toUserId, toUsername, onClose, onSuccess }: TipModalProps) {
  const [amount, setAmount] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const presets = [1, 5, 10, 50];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const amountNum = parseFloat(amount);
    if (!amountNum || amountNum <= 0) {
      setError('请输入有效金额');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const order = await apiClient.createTip(toUserId, amountNum, message);
      // Redirect to payment
      const payRes = await apiClient.payTipAlipay(order.id);
      if (payRes.pay_url) {
        window.open(payRes.pay_url, '_blank');
      }
      onSuccess?.();
      onClose();
    } catch (err: any) {
      setError(err.message || '打赏失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-background border rounded-xl w-full max-w-md p-6">
        <div className="flex items-center gap-2 mb-4">
          <Gift className="h-5 w-5 text-primary" />
          <h2 className="font-bold text-lg">打赏 {toUsername}</h2>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label>快速选择金额</Label>
            <div className="grid grid-cols-4 gap-2 mt-2">
              {presets.map((p) => (
                <Button
                  key={p}
                  type="button"
                  variant={amount === String(p) ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setAmount(String(p))}
                >
                  ¥{p}
                </Button>
              ))}
            </div>
          </div>

          <div>
            <Label htmlFor="amount">自定义金额（元）</Label>
            <Input
              id="amount"
              type="number"
              min="0.01"
              step="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="输入金额"
              className="mt-1"
            />
          </div>

          <div>
            <Label htmlFor="message">留言（可选）</Label>
            <Input
              id="message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="给创作者的一句话..."
              className="mt-1"
            />
          </div>

          {error && <p className="text-destructive text-sm">{error}</p>}

          <div className="flex gap-3 pt-2">
            <Button type="submit" disabled={loading} className="flex-1">
              {loading ? '处理中...' : `打赏 ¥${amount || '?'}`}
            </Button>
            <Button type="button" variant="outline" onClick={onClose}>
              取消
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
