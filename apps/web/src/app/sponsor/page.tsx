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

export default function SponsorPage() {
  const [info, setInfo] = useState<SponsorInfo | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiClient.getSponsorInfo()
      .then(setInfo)
      .catch(() => setInfo(null))
      .finally(() => setLoading(false));
  }, []);

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

  return (
    <div className="max-w-2xl mx-auto px-4 py-10">
      <h1 className="text-3xl font-bold text-center mb-2">支持站长</h1>
      <p className="text-center text-gray-500 mb-8">感谢每一位愿意投喂的同好 🐾</p>

      {/* Message */}
      <div className="bg-orange-50 border border-orange-200 rounded-lg p-4 mb-8 text-center text-gray-700">
        {info.message}
      </div>

      {/* Progress bar */}
      <div className="mb-8">
        <div className="flex justify-between text-sm text-gray-600 mb-1">
          <span>本月服务器开销进度</span>
          <span>¥{info.current_raised.toFixed(2)} / ¥{info.monthly_goal.toFixed(2)}</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-4">
          <div
            className="bg-orange-400 h-4 rounded-full transition-all duration-500"
            style={{ width: `${progress}%` }}
          />
        </div>
        <p className="text-center text-sm text-gray-500 mt-1">
          已达成 {progress.toFixed(1)}%
        </p>
      </div>

      {/* QR codes */}
      <div className="grid grid-cols-2 gap-6">
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

      <p className="text-center text-xs text-gray-400 mt-10">
        所有赞助均用于服务器维护，不做其他用途。感谢您的支持！
      </p>
    </div>
  );
}
