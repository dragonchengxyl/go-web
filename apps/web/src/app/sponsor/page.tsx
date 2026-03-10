'use client';

import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api-client';

interface SponsorInfo {
  monthly_goal: number;
  current_raised: number;
  alipay_qr_url: string;
  wechat_qr_url: string;
  message: string;
}

const THANK_YOU_LIST = [
  '匿名 #A7F2',
  '匿名 #B3D9',
  '匿名 #C1E8',
  '匿名 #D4AF',
  '匿名 #E6B2',
];

export default function SponsorPage() {
  const [info, setInfo] = useState<SponsorInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [mounted, setMounted] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    apiClient.getSponsorInfo()
      .then(setInfo)
      .catch(() => setInfo(null))
      .finally(() => {
        setLoading(false);
        // Delay mount animation slightly for progress bar transition
        setTimeout(() => setMounted(true), 100);
      });
  }, []);

  function handleCopy() {
    const text = '感谢支持！请转账至以上收款码，留言备注"赞助"。';
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }).catch(() => {});
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">加载中...</p>
      </div>
    );
  }

  if (!info) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">暂无赞助信息</p>
      </div>
    );
  }

  const progress = info.monthly_goal > 0
    ? Math.min(100, (info.current_raised / info.monthly_goal) * 100)
    : 0;
  const remaining = Math.max(0, info.monthly_goal - info.current_raised);

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 pt-24">
      <h1 className="text-3xl font-bold text-center mb-2">支持站长</h1>
      <p className="text-center text-gray-500 mb-8">感谢每一位愿意投喂的同好 🐾</p>

      {/* Message */}
      <div className="bg-orange-50 border border-orange-200 rounded-lg p-4 mb-8 text-center text-gray-700 dark:bg-orange-950/30 dark:border-orange-800 dark:text-orange-200">
        {info.message}
      </div>

      {/* Progress bar */}
      <div className="mb-8">
        <div className="flex justify-between text-sm text-gray-600 dark:text-gray-400 mb-1">
          <span>本月服务器开销进度</span>
          <span>¥{info.current_raised.toFixed(2)} / ¥{info.monthly_goal.toFixed(2)}</span>
        </div>
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-4 overflow-hidden">
          <div
            className="bg-orange-400 h-4 rounded-full transition-all duration-1000 ease-out"
            style={{ width: mounted ? `${progress}%` : '0%' }}
          />
        </div>
        <div className="flex justify-between mt-1">
          <p className="text-center text-sm text-gray-500 flex-1">
            已达成 {progress.toFixed(1)}%
          </p>
          <p className="text-sm text-gray-500">
            本月剩余 <span className="font-medium text-orange-600">¥{remaining.toFixed(2)}</span>
          </p>
        </div>
      </div>

      {/* QR codes */}
      <div className="grid grid-cols-2 gap-6 mb-8">
        {info.alipay_qr_url && (
          <div className="flex flex-col items-center gap-2">
            <img
              src={info.alipay_qr_url}
              alt="支付宝收款码"
              className="w-40 h-40 object-contain border rounded-lg"
            />
            <span className="text-sm text-blue-600 font-medium">支付宝</span>
          </div>
        )}
        {info.wechat_qr_url && (
          <div className="flex flex-col items-center gap-2">
            <img
              src={info.wechat_qr_url}
              alt="微信赞赏码"
              className="w-40 h-40 object-contain border rounded-lg"
            />
            <span className="text-sm text-green-600 font-medium">微信</span>
          </div>
        )}
      </div>

      {/* Copy button */}
      <div className="text-center mb-8">
        <button
          onClick={handleCopy}
          className="px-4 py-2 text-sm border rounded-lg hover:bg-muted transition-colors"
        >
          {copied ? '✓ 已复制转账提示' : '复制转账提示文字'}
        </button>
      </div>

      {/* Thank you list */}
      <div className="border rounded-xl p-5 bg-card">
        <h2 className="font-semibold text-center mb-4 text-gray-600 dark:text-gray-400">本月赞助鸣谢</h2>
        <div className="flex flex-wrap justify-center gap-3">
          {THANK_YOU_LIST.map(name => (
            <span key={name} className="px-3 py-1 bg-orange-50 dark:bg-orange-950/30 text-orange-700 dark:text-orange-300 text-sm rounded-full border border-orange-200 dark:border-orange-800">
              {name}
            </span>
          ))}
        </div>
      </div>

      <p className="text-center text-xs text-gray-400 mt-10">
        所有赞助均用于服务器维护，不做其他用途。感谢您的支持！
      </p>
    </div>
  );
}
